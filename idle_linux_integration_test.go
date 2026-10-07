//go:build linux

package idle

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// TestX11 runs against a real Xvfb server and drives it with xdotool.
func TestX11(t *testing.T) {
	for _, bin := range []string{"Xvfb", "xdotool"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not installed", bin)
		}
	}
	const display = ":97"
	xvfb := exec.Command("Xvfb", display, "-screen", "0", "64x64x24", "-nolisten", "tcp")
	if err := xvfb.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { xvfb.Process.Kill(); xvfb.Wait() })
	t.Setenv("DISPLAY", display)
	t.Setenv("XDG_SESSION_TYPE", "x11")

	var d time.Duration
	var err error
	for i := 0; i < 50; i++ { // wait for the server to come up
		if d, err = Get(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Get never succeeded: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)
	if d, err = Get(); err != nil || d < time.Second || d > 10*time.Second {
		t.Fatalf("after 1.5s of nothing: got %v, %v", d, err)
	}

	if out, err := exec.Command("xdotool", "mousemove", "10", "10").CombinedOutput(); err != nil {
		t.Fatalf("xdotool: %v: %s", err, out)
	}
	if d, err = Get(); err != nil || d > 500*time.Millisecond {
		t.Fatalf("right after input: got %v, %v", d, err)
	}

	// Server dies: expect an error, not a hang or a stale value.
	xvfb.Process.Kill()
	xvfb.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = GetContext(ctx); err == nil {
		t.Fatal("expected an error after the X server died")
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

// fakeIdleMonitor mimics org.gnome.Mutter.IdleMonitor.
type fakeIdleMonitor struct{ ms uint64 }

func (f fakeIdleMonitor) GetIdletime() (uint64, *dbus.Error) { return f.ms, nil }

func TestMutter(t *testing.T) {
	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skip("dbus-daemon not installed")
	}
	cmd := exec.Command("dbus-daemon", "--session", "--nofork", "--print-address")
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	addr, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	addr = strings.TrimSpace(addr)

	srv, err := dbus.Connect(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	if err := srv.Export(fakeIdleMonitor{ms: 4321}, mutterPath, mutterIface); err != nil {
		t.Fatal(err)
	}
	reply, err := srv.RequestName(mutterDest, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		t.Fatalf("RequestName: %v %v", reply, err)
	}

	mutter.Lock()
	mutter.addr = addr
	mutter.Unlock()
	t.Cleanup(func() {
		mutter.Lock()
		mutter.addr = ""
		if mutter.conn != nil {
			mutter.conn.Close()
			mutter.conn = nil
		}
		mutter.Unlock()
	})
	t.Setenv("XDG_SESSION_TYPE", "wayland")
	os.Unsetenv("DISPLAY")

	d, err := Get()
	if err != nil || d != 4321*time.Millisecond {
		t.Fatalf("got %v, %v", d, err)
	}

	// Compositor goes away: error, and recovery once it is back.
	srv.ReleaseName(mutterDest)
	if _, err := Get(); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("want ErrUnsupported with no service, got %v", err)
	}
	if _, err := srv.RequestName(mutterDest, dbus.NameFlagDoNotQueue); err != nil {
		t.Fatal(err)
	}
	if d, err = Get(); err != nil || d != 4321*time.Millisecond {
		t.Fatalf("after recovery: got %v, %v", d, err)
	}
}
