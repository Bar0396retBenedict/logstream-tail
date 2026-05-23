package source

import (
	"context"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// BatchConfig controls how the Batcher accumulates and flushes events.
type BatchConfig struct {
	// MaxSize is the maximum number of events in a single batch.
	MaxSize int
	// MaxWait is the longest time to wait before flushing an incomplete batch.
	MaxWait time.Duration
}

// DefaultBatchConfig returns a BatchConfig with sensible defaults.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxSize: 100,
		MaxWait: 2 * time.Second,
	}
}

// NewBatcher reads individual events from in, groups them into slices of up to
// cfg.MaxSize events (or whatever arrived within cfg.MaxWait), and emits each
// batch as a burst of individual events onto the returned channel. This is
// useful for downstream sinks that benefit from processing events in groups
// (e.g. bulk-write APIs) while still presenting a uniform per-event channel
// interface to the rest of the pipeline.
func NewBatcher(ctx context.Context, in <-chan logevent.Event, cfg BatchConfig) <-chan logevent.Event {
	out := make(chan logevent.Event, cfg.MaxSize)

	go func() {
		defer close(out)

		batch := make([]logevent.Event, 0, cfg.MaxSize)
		ticker := time.NewTicker(cfg.MaxWait)
		defer ticker.Stop()

		flush := func() {
			for _, ev := range batch {
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				}
			}
			batch = batch[:0]
		}

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-in:
				if !ok {
					flush()
					return
				}
				batch = append(batch, ev)
				if len(batch) >= cfg.MaxSize {
					flush()
					ticker.Reset(cfg.MaxWait)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush()
				}
			}
		}
	}()

	return out
}
