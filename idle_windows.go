//go:build windows

package idle

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// lastInputInfo mirrors the Win32 LASTINPUTINFO structure.
type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	procLastInput    = user32.NewProc("GetLastInputInfo")
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procGetTickCount = kernel32.NewProc("GetTickCount")
)

func get(context.Context) (time.Duration, error) {
	info := lastInputInfo{cbSize: uint32(unsafe.Sizeof(lastInputInfo{}))}
	if r, _, err := procLastInput.Call(uintptr(unsafe.Pointer(&info))); r == 0 {
		return 0, fmt.Errorf("%w: GetLastInputInfo: %v", ErrUnsupported, err)
	}
	now, _, _ := procGetTickCount.Call()
	// Both values are 32-bit millisecond tick counts that wrap every ~49.7
	// days. Unsigned subtraction yields the right delta across a wrap.
	return time.Duration(uint32(now)-info.dwTime) * time.Millisecond, nil
}
