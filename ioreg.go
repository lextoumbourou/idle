package idle

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var hidIdleRe = regexp.MustCompile(`"HIDIdleTime"\s*=\s*(\d+)`)

// parseIOReg extracts the idle time from `ioreg -c IOHIDSystem` output. A
// machine can expose several IOHIDSystem entries (one per HID client); the
// smallest value is the most recent input.
//
// It lives in a build-tag-free file so it can be tested on every platform.
func parseIOReg(out []byte) (time.Duration, error) {
	var best uint64
	found := false
	for _, m := range hidIdleRe.FindAllSubmatch(out, -1) {
		ns, err := strconv.ParseUint(string(m[1]), 10, 64)
		if err != nil {
			continue
		}
		if !found || ns < best {
			best, found = ns, true
		}
	}
	if !found {
		return 0, fmt.Errorf("idle: no HIDIdleTime in ioreg output")
	}
	if best > uint64(1<<63-1) {
		best = 1<<63 - 1
	}
	return time.Duration(best), nil
}
