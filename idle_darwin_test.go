//go:build darwin

package idle

import (
	"context"
	"errors"
	"testing"
)

// CoreGraphics (session event state) and ioreg (HID system) are different
// sources and legitimately drift apart when there is no real input, e.g. on a
// CI VM, so only sanity-check both and log the difference. Skipped when there
// is no window server.
func TestCoreGraphicsAndIORegAreSane(t *testing.T) {
	cg, err := idleFromCoreGraphics()
	if err != nil {
		t.Skipf("CoreGraphics unavailable: %v", err)
	}
	io, err := idleFromIOReg(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cg < 0 || io < 0 {
		t.Fatalf("negative idle: CoreGraphics=%v ioreg=%v", cg, io)
	}
	t.Logf("CoreGraphics=%v ioreg=%v", cg, io)
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
