package idle

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type fetcher interface {
	Fetch() ([]byte, error)
}

type ioRegFetcher struct{}

func (ioRegFetcher) Fetch() ([]byte, error) {
	return exec.Command("ioreg", "-c", "IOHIDSystem").Output()
}

func parseIdleFromIOReg(f fetcher) (time.Duration, error) {
	var output time.Duration
	var idleInNs string

	ioRegOutput, err := f.Fetch()
	if err != nil {
		return output, err
	}

	rawStr := string(ioRegOutput)
	lines := strings.Split(rawStr, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "HIDIdleTime") {
			continue
		}

		cols := strings.Split(line, " ")
		idleInNs = cols[len(cols)-1]
		break
	}

	return time.ParseDuration(fmt.Sprintf("%sns", idleInNs))
}

// Get idle time for Darwin (OSX)
func Get() (time.Duration, error) {
	fetcher := ioRegFetcher{}
	return parseIdleFromIOReg(fetcher)
}
