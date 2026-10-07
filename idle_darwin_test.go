//go:build darwin

package idle

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Both macOS sources should agree to within a couple of seconds. Skipped when
// there is no window server (headless CI), where CoreGraphics reports no data.
func TestCoreGraphicsMatchesIOReg(t *testing.T) {
	cg, err := idleFromCoreGraphics()
	if err != nil {
		t.Skipf("CoreGraphics unavailable: %v", err)
	}
	io, err := idleFromIOReg(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cg - io; diff > 2*time.Second || diff < -2*time.Second {
		t.Fatalf("CoreGraphics=%v ioreg=%v", cg, io)
	}
}

func TestGetDarwin(t *testing.T) {
	d, err := Get()
	if err != nil && !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if err == nil && d < 0 {
		t.Fatalf("negative idle %v", d)
	}
}
