package source

import (
	"context"
	"sync"

	"github.com/your-org/logstream-tail/internal/logevent"
)

// MergeConfig controls the behaviour of the Merger pipeline stage.
type MergeConfig struct {
	// BufSize is the capacity of the internal channel used to collect events
	// from all input sources before forwarding them downstream.
	BufSize int
}

// DefaultMergeConfig returns a MergeConfig with sensible defaults.
func DefaultMergeConfig() MergeConfig {
	return MergeConfig{
		BufSize: 256,
	}
}

// NewMerger fans-in multiple upstream channels into a single output channel.
// Events from all inputs are forwarded in arrival order; no sorting is applied.
// The output channel is closed once every input goroutine has finished and the
// context has been cancelled (or all inputs are drained).
//
// Usage:
//
//	out := NewMerger(ctx, DefaultMergeConfig(), chA, chB, chC)
//	for ev := range out {
//	    // handle ev
//	}
func NewMerger(ctx context.Context, cfg MergeConfig, inputs ...<-chan logevent.Event) <-chan logevent.Event {
	out := make(chan logevent.Event, cfg.BufSize)

	if len(inputs) == 0 {
		close(out)
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, in := range inputs {
		in := in // capture loop variable
		go func() {
			defer wg.Done()
			for {
				select {
				case ev, ok := <-in:
					if !ok {
						return
					}
					select {
					case out <- ev:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
