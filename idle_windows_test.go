//go:build windows

package idle

import (
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SendInput lets us fake input on a CI desktop and verify the idle timer resets.
func TestWindowsInputResetsIdle(t *testing.T) {
	type mouseInput struct {
		typ                    uint32
		_                      uint32 // padding on 64-bit
		dx, dy                 int32
		mouseData, flags, time uint32
		extra                  uintptr
	}
	sendInput := windows.NewLazySystemDLL("user32.dll").NewProc("SendInput")

	time.Sleep(1200 * time.Millisecond)
	before, err := Get()
	if err != nil {
		t.Skipf("no interactive desktop: %v", err)
	}
	in := mouseInput{typ: 0, dx: 1, dy: 1, flags: 0x0001 /* MOUSEEVENTF_MOVE */}
	if r, _, e := sendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in)); r == 0 {
		t.Skipf("SendInput failed: %v", e)
	}
	after, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	if after >= before {
		t.Fatalf("idle did not reset: before=%v after=%v", before, after)
	}
}
