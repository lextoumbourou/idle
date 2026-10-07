//go:build linux

package idle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	mutterDest  = "org.gnome.Mutter.IdleMonitor"
	mutterPath  = "/org/gnome/Mutter/IdleMonitor/Core"
	mutterIface = "org.gnome.Mutter.IdleMonitor"
)

var mutter struct {
	sync.Mutex
	conn *dbus.Conn
	// addr overrides session bus discovery (used by tests).
	addr string
}

// connectSessionBus connects to the user's session bus. When
// DBUS_SESSION_BUS_ADDRESS is missing (cron, systemd units, `sudo -u`) it
// falls back to the conventional systemd socket path.
func connectSessionBus() (*dbus.Conn, error) {
	addr := mutter.addr
	if addr == "" {
		addr = os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	}
	if addr == "" {
		sock := fmt.Sprintf("/run/user/%d/bus", os.Getuid())
		if _, err := os.Stat(sock); err != nil {
			return nil, errors.New("no session bus (DBUS_SESSION_BUS_ADDRESS unset)")
		}
		addr = "unix:path=" + sock
	}
	// Not dbus.SessionBus(): that shares a process-wide connection we could
	// never repair after a bus restart.
	return dbus.Connect(addr)
}

func getMutter(ctx context.Context) (time.Duration, error) {
	mutter.Lock()
	defer mutter.Unlock()

	if mutter.conn == nil {
		c, err := connectSessionBus()
		if err != nil {
			return 0, err
		}
		mutter.conn = c
	}
	var ms uint64
	call := mutter.conn.Object(mutterDest, mutterPath).CallWithContext(ctx, mutterIface+".GetIdletime", 0)
	if call.Err == nil {
		call.Err = call.Store(&ms)
	}
	if call.Err != nil {
		// Drop the connection on anything but cancellation so a restarted bus
		// or compositor is picked up next time.
		if ctx.Err() == nil {
			mutter.conn.Close()
			mutter.conn = nil
		}
		return 0, fmt.Errorf("GetIdletime: %w", call.Err)
	}
	return time.Duration(ms) * time.Millisecond, nil
}
