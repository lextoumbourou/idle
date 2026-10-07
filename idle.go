// Package idle reports how long it has been since the user last touched the
// keyboard or mouse.
//
// It is pure Go (no cgo) and supports:
//
//   - macOS: CoreGraphics via purego, falling back to parsing ioreg.
//   - Windows: GetLastInputInfo.
//   - Linux: the X11 MIT-SCREEN-SAVER extension, GNOME's Mutter IdleMonitor
//     and systemd-logind's IdleSinceHint, whichever works for the current
//     session.
//
// Idle time is a property of an interactive desktop session. Processes that
// are not part of one (SSH sessions, system services, containers, headless
// servers) get an error wrapping [ErrUnsupported], never a made-up zero.
package idle

import (
	"context"
	"errors"
	"time"
)

// ErrUnsupported is returned (wrapped) when idle time cannot be determined in
// the current environment, for example on a headless machine or on a Wayland
// compositor that does not expose it. Use [errors.Is] to test for it.
var ErrUnsupported = errors.New("idle: idle time is not available in this environment")

// Get returns the time since the last user input.
func Get() (time.Duration, error) {
	return GetContext(context.Background())
}

// GetContext is like [Get] but honours cancellation and deadlines of ctx for
// any I/O it performs (D-Bus calls, subprocesses, X11 requests).
//
// It is safe for concurrent use.
func GetContext(ctx context.Context) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	d, err := get(ctx)
	if err != nil {
		return 0, err
	}
	// Clock adjustments and rounding can produce tiny negatives; never leak them.
	return max(d, 0), nil
}
