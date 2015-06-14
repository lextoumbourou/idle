package idle

import "testing"

type mockFetcher struct{}

func (mockFetcher) Fetch() ([]byte, error) {
	return []byte(`
+-o IOHIDSystem  <class IOHIDSystem, id 0x1000002cf, registered, matched, active, busy 0 (0 ms), retain 21>
{
"IOClass" = "IOHIDSystem"
"HIDIdleTime" = 21858429
}
+-o IOHIDSystem1  <class IOHIDSystem, id 0x1000002cf, registered, matched, active, busy 0 (0 ms), retain 21>
{
"IOClass" = "IOHIDSystem1"
"HIDIdleTime" = 1000000
}`), nil
}

func TestParseIdleFromIOReg(t *testing.T) {
	fetcher := mockFetcher{}
	result, _ := parseIdleFromIOReg(fetcher)
	if result.Nanoseconds() != int64(21858429) {
		t.Error("Failed to parse time from IOReg.")
	}
}
