// Command idle prints how long the user has been idle.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/lextoumbourou/idle"
)

func main() {
	watch := flag.Duration("watch", 0, "keep printing at this interval (e.g. 1s); 0 prints once and exits")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *watch); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, idle.ErrUnsupported) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, every time.Duration) error {
	for {
		qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		d, err := idle.GetContext(qctx)
		cancel()
		if err != nil {
			return err
		}
		fmt.Println(d.Round(time.Millisecond))
		if every <= 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(every):
		}
	}
}
