//go:build linux

package idle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// backend is one way of asking the desktop how long the user has been idle.
type backend struct {
	name string
	get  func(ctx context.Context) (time.Duration, error)
}

var (
	x11Backend    = backend{"x11", getX11}
	mutterBackend = backend{"mutter", getMutter}

	// preferred remembers the backend that last worked so that steady-state
	// polling does a single round trip instead of probing every time.
	mu        sync.Mutex
	preferred string
)

// backendsFor returns the backends worth trying, most trustworthy first.
//
// On Wayland the X11 screensaver extension is deliberately excluded: XWayland
// only sees input delivered to X clients, so it would report a growing idle
// time while the user is busy in native Wayland windows. A wrong answer is
// worse than an error.
func backendsFor(sessionType string, wayland bool) []backend {
	switch {
	case wayland || strings.EqualFold(sessionType, "wayland"):
		return []backend{mutterBackend}
	case strings.EqualFold(sessionType, "x11"), strings.EqualFold(sessionType, "mir"):
		return []backend{x11Backend, mutterBackend}
	default:
		// tty, unspecified (e.g. launched from cron/systemd): try everything.
		return []backend{x11Backend, mutterBackend}
	}
}

func get(ctx context.Context) (time.Duration, error) {
	list := backendsFor(os.Getenv("XDG_SESSION_TYPE"), os.Getenv("WAYLAND_DISPLAY") != "" && os.Getenv("DISPLAY") == "")
	return tryBackends(ctx, list)
}

func tryBackends(ctx context.Context, list []backend) (time.Duration, error) {
	mu.Lock()
	pref := preferred
	mu.Unlock()

	// Move the last-known-good backend to the front.
	ordered := make([]backend, 0, len(list))
	for _, b := range list {
		if b.name == pref {
			ordered = append(ordered, b)
		}
	}
	for _, b := range list {
		if b.name != pref {
			ordered = append(ordered, b)
		}
	}

	var errs []error
	for _, b := range ordered {
		d, err := b.get(ctx)
		if err == nil {
			mu.Lock()
			preferred = b.name
			mu.Unlock()
			return d, nil
		}
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		errs = append(errs, fmt.Errorf("%s: %w", b.name, err))
	}
	return 0, fmt.Errorf("%w: %w", ErrUnsupported, errors.Join(errs...))
}
