package idle

import (
	"testing"
	"time"
)

func TestParseIOReg(t *testing.T) {
	const sample = `
+-o IOHIDSystem  <class IOHIDSystem, id 0x1000002cf, registered, matched, active, busy 0 (0 ms), retain 21>
{
"IOClass" = "IOHIDSystem"
"HIDIdleTime" = 21858429
}
+-o IOHIDSystem1  <class IOHIDSystem, id 0x1000002cf>
{
"HIDIdleTime" = 1000000
}`
	tests := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"takes the smallest of several entries", sample, time.Millisecond, false},
		{"single entry", `"HIDIdleTime" = 5000000000`, 5 * time.Second, false},
		{"no spaces", `"HIDIdleTime"=7`, 7, false},
		{"missing", `"IOClass" = "IOHIDSystem"`, 0, true},
		{"empty", ``, 0, true},
		{"overflow clamps", `"HIDIdleTime" = 18446744073709551615`, 1<<63 - 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIOReg([]byte(tt.in))
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got %v, %v; want %v, err=%v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
