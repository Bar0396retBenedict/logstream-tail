package source

import (
	"context"

	"github.com/your-org/logstream-tail/internal/logevent"
)

// Sampler forwards only every Nth event, dropping the rest.
// This is useful for high-volume streams where full fidelity is not required.
type Sampler struct {
	in     <-chan logevent.Event
	out    chan logevent.Event
	config SampleConfig
	counter uint64
}

// SampleConfig controls the sampling behaviour.
type SampleConfig struct {
	// Rate is the keep-one-in-N ratio. A Rate of 1 passes every event.
	// A Rate of 5 passes every 5th event.
	Rate uint64
}

// DefaultSampleConfig returns a SampleConfig that passes every event.
func DefaultSampleConfig() SampleConfig {
	return SampleConfig{Rate: 1}
}

// NewSampler creates a Sampler that reads from in and writes sampled events
// to the returned channel. The goroutine stops when ctx is cancelled.
func NewSampler(ctx context.Context, in <-chan logevent.Event, cfg SampleConfig) <-chan logevent.Event {
	if cfg.Rate == 0 {
		cfg.Rate = 1
	}
	s := &Sampler{
		in:     in,
		out:    make(chan logevent.Event, 64),
		config: cfg,
	}
	go s.run(ctx)
	return s.out
}

func (s *Sampler) run(ctx context.Context) {
	defer close(s.out)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-s.in:
			if !ok {
				return
			}
			s.counter++
			if s.counter%s.config.Rate == 0 {
				select {
				case s.out <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
