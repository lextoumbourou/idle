package idle

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// To do: implement this.
func parseIORegOutput(raw []byte) (time.Duration, error) {
	var idleInNs string

	rawStr := string(raw)
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

// ToDo: Use cgo coolness instead of spawning a proc.
func Get() (time.Duration, error) {
	var output time.Duration

	ioRegOutput, err := exec.Command("ioreg", "-c", "IOHIDSystem").Output()
	if err != nil {
		return output, err
	}

	return parseIORegOutput(ioRegOutput)
}
