package source

import (
	"context"

	"github.com/your-org/logstream-tail/internal/logevent"
)

// SplitConfig controls how the Splitter routes events to named output channels.
type SplitConfig struct {
	// BufSize is the buffer depth for each output channel.
	BufSize int
}

// DefaultSplitConfig returns a SplitConfig with sensible defaults.
func DefaultSplitConfig() SplitConfig {
	return SplitConfig{
		BufSize: 64,
	}
}

// SplitFunc maps an event to a key. Events with the same key are routed to the
// same output channel. Returning an empty string drops the event.
type SplitFunc func(logevent.Event) string

// NewSplitter fans a single input channel out to multiple named output channels
// according to fn. Each unique key returned by fn gets its own dedicated
// channel. Unknown keys are created on first use.
//
// The returned map is populated before NewSplitter returns; callers must pass
// the expected set of keys so channels can be pre-allocated. Events whose key
// is not in keys are silently dropped.
func NewSplitter(
	ctx context.Context,
	in <-chan logevent.Event,
	keys []string,
	fn SplitFunc,
	cfg SplitConfig,
) map[string]<-chan logevent.Event {
	outs := make(map[string]chan logevent.Event, len(keys))
	result := make(map[string]<-chan logevent.Event, len(keys))
	for _, k := range keys {
		ch := make(chan logevent.Event, cfg.BufSize)
		outs[k] = ch
		result[k] = ch
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-in:
				if !ok {
					return
				}
				key := fn(ev)
				if ch, found := outs[key]; found {
					select {
					case ch <- ev:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return result
}
