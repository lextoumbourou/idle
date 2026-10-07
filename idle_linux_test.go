//go:build linux

package idle

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackendsForWaylandExcludesX11(t *testing.T) {
	for _, b := range backendsFor("wayland", false) {
		if b.name == "x11" {
			t.Fatal("x11 must not be used on wayland sessions")
		}
	}
	for _, b := range backendsFor("", true) {
		if b.name == "x11" {
			t.Fatal("x11 must not be used when only WAYLAND_DISPLAY is set")
		}
	}
	if got := backendsFor("x11", false); got[0].name != "x11" {
		t.Fatalf("x11 session should try x11 first, got %v", got[0].name)
	}
}

func TestTryBackendsFallbackAndPreference(t *testing.T) {
	mu.Lock()
	preferred = ""
	mu.Unlock()

	var calls []string
	bad := backend{"bad", func(context.Context) (time.Duration, error) {
		calls = append(calls, "bad")
		return 0, errors.New("nope")
	}}
	good := backend{"good", func(context.Context) (time.Duration, error) {
		calls = append(calls, "good")
		return 3 * time.Second, nil
	}}

	d, err := tryBackends(context.Background(), []backend{bad, good})
	if err != nil || d != 3*time.Second {
		t.Fatalf("got %v, %v", d, err)
	}
	calls = nil
	if _, err := tryBackends(context.Background(), []backend{bad, good}); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "good" {
		t.Fatalf("expected remembered backend to be tried first alone, got %v", calls)
	}
}

func TestTryBackendsAllFail(t *testing.T) {
	mu.Lock()
	preferred = ""
	mu.Unlock()
	bad := backend{"bad", func(context.Context) (time.Duration, error) { return 0, errors.New("boom") }}
	_, err := tryBackends(context.Background(), []backend{bad})
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GetContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
