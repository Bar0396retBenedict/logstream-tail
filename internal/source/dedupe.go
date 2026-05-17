package source

import (
	"context"
	"sync"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// Deduplicator filters out duplicate log events within a configurable time
// window. Events are considered duplicates if they share the same source,
// message, and severity within the deduplication window.
//
// This is useful when multiple cloud sources may emit overlapping log entries
// (e.g., during a polling overlap or a source restart).
type Deduplicator struct {
	in     <-chan logevent.Event
	out    chan logevent.Event
	window time.Duration

	mu   sync.Mutex
	seen map[string]time.Time
}

// NewDeduplicator wraps an input channel and returns a Deduplicator that
// suppresses duplicate events within the given time window.
// A window of zero disables deduplication (all events pass through).
func NewDeduplicator(in <-chan logevent.Event, window time.Duration) *Deduplicator {
	return &Deduplicator{
		in:     in,
		out:    make(chan logevent.Event, cap(in)+1),
		window: window,
		seen:   make(map[string]time.Time),
	}
}

// Out returns the channel of deduplicated events.
func (d *Deduplicator) Out() <-chan logevent.Event {
	return d.out
}

// Run reads from the input channel, suppresses duplicates, and forwards unique
// events to the output channel. It returns when ctx is cancelled or the input
// channel is closed.
func (d *Deduplicator) Run(ctx context.Context) error {
	defer close(d.out)

	// Periodically evict expired entries to prevent unbounded memory growth.
	ticker := time.NewTicker(d.evictInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case ev, ok := <-d.in:
			if !ok {
				return nil
			}
			if d.window == 0 || d.admit(ev) {
				select {
				case d.out <- ev:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

		case now := <-ticker.C:
			d.evict(now)
		}
	}
}

// admit returns true if the event has not been seen within the dedup window,
// and records it as seen.
func (d *Deduplicator) admit(ev logevent.Event) bool {
	key := dedupKey(ev)
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	if last, exists := d.seen[key]; exists && now.Sub(last) < d.window {
		return false
	}
	d.seen[key] = now
	return true
}

// evict removes entries older than the dedup window.
func (d *Deduplicator) evict(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, ts := range d.seen {
		if now.Sub(ts) >= d.window {
			delete(d.seen, key)
		}
	}
}

// evictInterval returns how often to run the eviction sweep.
// It is capped between 1 second and 1 minute.
func (d *Deduplicator) evictInterval() time.Duration {
	interval := d.window / 2
	if interval < time.Second {
		interval = time.Second
	}
	if interval > time.Minute {
		interval = time.Minute
	}
	return interval
}

// dedupKey builds a string key that uniquely identifies a log event for the
// purpose of deduplication. It combines source, severity, and message.
func dedupKey(ev logevent.Event) string {
	return ev.Source + "\x00" + ev.Severity.String() + "\x00" + ev.Message
}
