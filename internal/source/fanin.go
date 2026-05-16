package source

import (
	"context"
	"sync"

	"github.com/user/logstream-tail/internal/logevent"
)

// FanIn merges multiple Sources into a single channel of log events.
type FanIn struct {
	sources []Source
}

// NewFanIn creates a FanIn aggregator from the given sources.
func NewFanIn(sources ...Source) *FanIn {
	return &FanIn{sources: sources}
}

// Stream starts all sources concurrently and fans their events into the
// returned read-only channel. The channel is closed when all sources finish
// or the context is cancelled.
func (f *FanIn) Stream(ctx context.Context) <-chan logevent.Event {
	out := make(chan logevent.Event, 256)

	var wg sync.WaitGroup
	for _, s := range f.sources {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			// Each source writes into a per-source channel that we forward.
			perSource := make(chan logevent.Event, 64)
			go func() {
				_ = src.Start(ctx, perSource)
				close(perSource)
			}()
			for evt := range perSource {
				select {
				case out <- evt:
				case <-ctx.Done():
					return
				}
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Close calls Close on every registered source.
func (f *FanIn) Close() error {
	var firstErr error
	for _, s := range f.sources {
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
