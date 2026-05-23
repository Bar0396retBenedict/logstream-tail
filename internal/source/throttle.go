package source

import (
	"context"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// ThrottleConfig controls how the Throttler behaves.
type ThrottleConfig struct {
	// Interval is the minimum duration between forwarded events per unique source.
	Interval time.Duration
}

// DefaultThrottleConfig returns a sensible default ThrottleConfig.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		Interval: 100 * time.Millisecond,
	}
}

// NewThrottler returns a pipeline stage that ensures at most one event per
// source is forwarded within the configured Interval. Events that arrive
// faster than the interval are silently dropped.
func NewThrottler(cfg ThrottleConfig, in <-chan logevent.Event) <-chan logevent.Event {
	out := make(chan logevent.Event, 64)
	go func() {
		defer close(out)
		last := make(map[string]time.Time)
		for ev := range in {
			key := string(ev.Source)
			if t, ok := last[key]; ok && time.Since(t) < cfg.Interval {
				continue // too soon — drop
			}
			last[key] = time.Now()
			out <- ev
		}
	}()
	return out
}

// NewThrottlerContext is like NewThrottler but respects ctx cancellation.
func NewThrottlerContext(ctx context.Context, cfg ThrottleConfig, in <-chan logevent.Event) <-chan logevent.Event {
	out := make(chan logevent.Event, 64)
	go func() {
		defer close(out)
		last := make(map[string]time.Time)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-in:
				if !ok {
					return
				}
				key := string(ev.Source)
				if t, seen := last[key]; seen && time.Since(t) < cfg.Interval {
					continue
				}
				last[key] = time.Now()
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
