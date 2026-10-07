//go:build !darwin && !windows && !linux

package idle

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func get(context.Context) (time.Duration, error) {
	return 0, fmt.Errorf("%w: %s is not supported", ErrUnsupported, runtime.GOOS)
}
