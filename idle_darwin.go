//go:build darwin

package idle

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/ebitengine/purego"
)

const (
	coreGraphicsPath = "/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics"

	kCGEventSourceStateCombinedSessionState = 0
	kCGAnyInputEventType                    = ^uint32(0) // (CGEventType)~0
)

var (
	cgOnce sync.Once
	cgFunc func(stateID, eventType uint32) float64
	cgErr  error
)

// loadCoreGraphics binds CGEventSourceSecondsSinceLastEventType without cgo.
func loadCoreGraphics() {
	lib, err := purego.Dlopen(coreGraphicsPath, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		cgErr = err
		return
	}
	// RegisterLibFunc panics if the symbol is missing; turn that into an error.
	defer func() {
		if r := recover(); r != nil {
			cgErr = fmt.Errorf("binding CGEventSourceSecondsSinceLastEventType: %v", r)
		}
	}()
	purego.RegisterLibFunc(&cgFunc, lib, "CGEventSourceSecondsSinceLastEventType")
}

func idleFromCoreGraphics() (time.Duration, error) {
	cgOnce.Do(loadCoreGraphics)
	if cgErr != nil {
		return 0, cgErr
	}
	secs := cgFunc(kCGEventSourceStateCombinedSessionState, kCGAnyInputEventType)
	// The API yields a negative or huge value when there is no window server
	// (e.g. an SSH session with no GUI login).
	if secs < 0 || secs != secs || secs > 1e9 {
		return 0, fmt.Errorf("%w: CoreGraphics returned %v (no GUI session?)", ErrUnsupported, secs)
	}
	return time.Duration(secs * float64(time.Second)), nil
}

func idleFromIOReg(ctx context.Context) (time.Duration, error) {
	out, err := exec.CommandContext(ctx, "ioreg", "-c", "IOHIDSystem", "-d", "4").Output()
	if err != nil {
		return 0, fmt.Errorf("ioreg: %w", err)
	}
	return parseIOReg(out)
}

func get(ctx context.Context) (time.Duration, error) {
	d, cgE := idleFromCoreGraphics()
	if cgE == nil {
		return d, nil
	}
	d, ioE := idleFromIOReg(ctx)
	if ioE == nil {
		return d, nil
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	// ioreg works without a GUI session too, so reaching here means both
	// failed. Preserve ErrUnsupported if CoreGraphics said so.
	return 0, fmt.Errorf("idle: coregraphics: %w; ioreg: %v", cgE, ioE)
}
