package source

import (
	"context"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// WindowConfig controls the behaviour of the sliding-window aggregator.
type WindowConfig struct {
	// Size is the duration of each window bucket.
	Size time.Duration
	// MaxEvents is the maximum number of events emitted per window.
	// Zero means unlimited.
	MaxEvents int
}

// DefaultWindowConfig returns a WindowConfig with sensible defaults.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Size:      5 * time.Second,
		MaxEvents: 500,
	}
}

// Window aggregates events from in over a sliding time window, emitting at
// most cfg.MaxEvents per window bucket. Events that exceed the cap are
// silently dropped. The window resets on each tick of cfg.Size.
func Window(ctx context.Context, in <-chan logevent.Event, cfg WindowConfig) <-chan logevent.Event {
	out := make(chan logevent.Event, 64)

	go func() {
		defer close(out)

		ticker := time.NewTicker(cfg.Size)
		defer ticker.Stop()

		count := 0

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				count = 0

			case ev, ok := <-in:
				if !ok {
					return
				}
				if cfg.MaxEvents > 0 && count >= cfg.MaxEvents {
					continue
				}
				count++
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}
