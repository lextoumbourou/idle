//go:build linux

package idle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/screensaver"
	"github.com/jezek/xgb/xproto"
)

func init() {
	// xgb logs connection chatter (e.g. a missing ~/.Xauthority) straight to
	// stderr. Failures are already reported through returned errors.
	xgb.Logger = log.New(io.Discard, "", 0)
}

// x11Conn is cached across calls: opening an X connection per poll is slow
// and, for a long-running daemon, wasteful. It is dropped on any error so the
// next call reconnects (the X server may have restarted).
var x11 struct {
	sync.Mutex
	conn *xgb.Conn
	root xproto.Drawable
}

func x11Reset() {
	if x11.conn != nil {
		x11.conn.Close()
		x11.conn = nil
	}
}

func getX11(ctx context.Context) (time.Duration, error) {
	if os.Getenv("DISPLAY") == "" {
		return 0, errors.New("DISPLAY is not set")
	}
	x11.Lock()
	defer x11.Unlock()

	type result struct {
		d   time.Duration
		err error
	}
	done := make(chan result, 1)
	// xgb has no context support, so run the blocking calls in a goroutine and
	// abandon them on cancellation. Closing the connection unblocks them.
	go func() {
		d, err := queryX11()
		done <- result{d, err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			x11Reset()
		}
		return r.d, r.err
	case <-ctx.Done():
		x11Reset()
		return 0, ctx.Err()
	}
}

// queryX11 must be called with x11.Mutex held.
func queryX11() (time.Duration, error) {
	if x11.conn == nil {
		c, err := xgb.NewConn()
		if err != nil {
			return 0, fmt.Errorf("connecting to X server: %w", err)
		}
		if err := screensaver.Init(c); err != nil {
			c.Close()
			return 0, fmt.Errorf("MIT-SCREEN-SAVER extension: %w", err)
		}
		x11.conn = c
		x11.root = xproto.Drawable(xproto.Setup(c).DefaultScreen(c).Root)
	}
	r, err := screensaver.QueryInfo(x11.conn, x11.root).Reply()
	if err != nil {
		return 0, fmt.Errorf("QueryInfo: %w", err)
	}
	return time.Duration(r.MsSinceUserInput) * time.Millisecond, nil
}
