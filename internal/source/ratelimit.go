package source

import (
	"context"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// RateLimiter wraps a log event channel and emits at most maxPerSec events
// per second, dropping excess events to prevent terminal flooding.
type RateLimiter struct {
	input      <-chan logevent.Event
	output     chan logevent.Event
	maxPerSec  int
	droppedTotal int
}

// NewRateLimiter creates a RateLimiter that forwards up to maxPerSec events
// per second from input. If maxPerSec <= 0, all events are forwarded.
func NewRateLimiter(input <-chan logevent.Event, maxPerSec int) *RateLimiter {
	return &RateLimiter{
		input:     input,
		output:    make(chan logevent.Event, 64),
		maxPerSec: maxPerSec,
	}
}

// Out returns the channel of rate-limited events.
func (r *RateLimiter) Out() <-chan logevent.Event {
	return r.output
}

// Dropped returns the total number of events dropped since the limiter started.
func (r *RateLimiter) Dropped() int {
	return r.droppedTotal
}

// Run starts forwarding events until ctx is cancelled. It should be called
// in its own goroutine.
func (r *RateLimiter) Run(ctx context.Context) {
	defer close(r.output)

	if r.maxPerSec <= 0 {
		// No limit — pass everything through.
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-r.input:
				if !ok {
					return
				}
				r.output <- ev
			}
		}
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count = 0
		case ev, ok := <-r.input:
			if !ok {
				return
			}
			if count < r.maxPerSec {
				r.output <- ev
				count++
			} else {
				r.droppedTotal++
			}
		}
	}
}
