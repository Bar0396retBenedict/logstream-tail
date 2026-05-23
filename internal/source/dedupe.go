package source

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
)

type dedupEntry struct {
	key     string
	seenAt  time.Time
}

type deduplicator struct {
	cfg     DedupeConfig
	entries []dedupEntry
	mu      sync.Mutex
}

func fingerprint(ev logevent.Event) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d|%s", ev.Source, ev.Severity, ev.Message)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (d *deduplicator) isDuplicate(key string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Evict expired entries.
	active := d.entries[:0]
	for _, e := range d.entries {
		if now.Sub(e.seenAt) < d.cfg.Window {
			active = append(active, e)
		}
	}
	d.entries = active

	for _, e := range d.entries {
		if e.key == key {
			return true
		}
	}

	// Evict oldest when at capacity.
	if len(d.entries) >= d.cfg.MaxTracked {
		d.entries = d.entries[1:]
	}
	d.entries = append(d.entries, dedupEntry{key: key, seenAt: now})
	return false
}

// NewDeduplicator returns a channel that forwards events from upstream while
// suppressing duplicates within the configured window.
func NewDeduplicator(ctx context.Context, upstream <-chan logevent.Event, cfg DedupeConfig) <-chan logevent.Event {
	out := make(chan logevent.Event, cap(upstream))
	d := &deduplicator{
		cfg:     cfg,
		entries: make([]dedupEntry, 0, cfg.MaxTracked),
	}

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-upstream:
				if !ok {
					return
				}
				if !d.isDuplicate(fingerprint(ev), time.Now()) {
					select {
					case out <- ev:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out
}
