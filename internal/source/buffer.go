package source

import (
	"context"
	"sync"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// BufferConfig controls the behaviour of the event buffer.
type BufferConfig struct {
	// Size is the maximum number of events held in memory before flushing.
	Size int
	// FlushInterval is the maximum time between flushes regardless of size.
	FlushInterval time.Duration
}

// DefaultBufferConfig returns sensible defaults for BufferConfig.
func DefaultBufferConfig() BufferConfig {
	return BufferConfig{
		Size:          256,
		FlushInterval: 500 * time.Millisecond,
	}
}

// Buffer batches incoming events and forwards them to an output channel in
// periodic flushes, reducing downstream pressure from bursty sources.
type Buffer struct {
	cfg BufferConfig
	in  <-chan logevent.Event
	out chan logevent.Event
}

// NewBuffer creates a Buffer that reads from in and writes to the returned
// channel. Call Run to start processing.
func NewBuffer(in <-chan logevent.Event, cfg BufferConfig) (*Buffer, <-chan logevent.Event) {
	out := make(chan logevent.Event, cfg.Size)
	return &Buffer{cfg: cfg, in: in, out: out}, out
}

// Run starts the buffering loop and blocks until ctx is cancelled or in is
// closed. The output channel is closed when Run returns.
func (b *Buffer) Run(ctx context.Context) {
	defer close(b.out)

	batch := make([]logevent.Event, 0, b.cfg.Size)
	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()

	var mu sync.Mutex

	flush := func() {
		mu.Lock()
		defer mu.Unlock()
		for _, ev := range batch {
			select {
			case b.out <- ev:
			case <-ctx.Done():
				return
			}
		}
		batch = batch[:0]
	}

	for {
		select {
		case ev, ok := <-b.in:
			if !ok {
				flush()
				return
			}
			mu.Lock()
			batch = append(batch, ev)
			full := len(batch) >= b.cfg.Size
			mu.Unlock()
			if full {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}
