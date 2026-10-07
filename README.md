# idle

Cross-platform **idle time** detection in Go: how long since the user last
touched the keyboard or mouse. Pure Go, **no cgo**, cross-compiles anywhere.

<img src="https://lh6.googleusercontent.com/-sm9TUtep2xs/T3R7ZCDrJVI/AAAAAAAAAKQ/jaSnMOyRJGw/w856-h1228-no/2_b%2Bdata%2Bnieznana.jpg" height="400"><br>
[Untitled, Zdzisław Beksiński](http://www.wikiart.org/en/zdislav-beksinski)

[![CI](https://github.com/lextoumbourou/idle/actions/workflows/ci.yml/badge.svg)](https://github.com/lextoumbourou/idle/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lextoumbourou/idle.svg)](https://pkg.go.dev/github.com/lextoumbourou/idle)

Requires Go 1.26+.

## Usage

```go
import "github.com/lextoumbourou/idle"

d, err := idle.Get() // or idle.GetContext(ctx)
if errors.Is(err, idle.ErrUnsupported) {
	// headless box, SSH session, Wayland compositor without an idle API...
}
fmt.Println("idle for", d)
```

A ready-made CLI lives in [`cmd/idle`](./cmd/idle):

```
$ go run ./cmd/idle -watch 1s
1.002s
2.003s
```

Exit status is `2` when idle time is unavailable in the environment, `1` for
other errors.

## Platform support

| Platform | Method | Notes |
|---|---|---|
| macOS | `CGEventSourceSecondsSinceLastEventType` via [purego](https://github.com/ebitengine/purego); falls back to `ioreg` | Works over SSH too (via the `ioreg` fallback). |
| Windows | `GetLastInputInfo` | Only meaningful in an interactive session; Windows services (session 0) have no input. |
| Linux, X11 | MIT-SCREEN-SAVER extension ([xgb](https://github.com/jezek/xgb)) | Needs `DISPLAY` (and `XAUTHORITY` if non-default). |
| Linux, GNOME (Wayland or X11) | `org.gnome.Mutter.IdleMonitor.GetIdletime` on the session bus | Falls back to `/run/user/$UID/bus` when `DBUS_SESSION_BUS_ADDRESS` is unset. |
| Linux, other Wayland (KDE, sway, Hyprland…) | **Unsupported** → `ErrUnsupported` | These compositors only offer push-style `ext-idle-notify-v1`, no way to ask "how long?". |

On Wayland the X11 backend is intentionally *not* used: XWayland only sees input
sent to X clients and would silently report wrong values. You get an error
instead of a lie.

### Resilience

* Never panics at import time (Windows DLLs are lazily loaded).
* X11 and D-Bus connections are cached, dropped on any error, and re-established
  on the next call, so X server / compositor / bus restarts are survived.
* The last working Linux backend is tried first; others are fallbacks.
* All calls accept a `context.Context`. Safe for concurrent use.
* Windows tick-count wraparound (49.7 days) is handled.

## Testing

`go test ./...` runs everything it can on the current OS. On Linux, the tests
start a real `Xvfb` and drive it with `xdotool`, and run a private
`dbus-daemon` with a fake Mutter service; they skip if those tools are absent.
CI covers Linux, macOS and Windows (where `SendInput` simulates input).

## License

MIT
