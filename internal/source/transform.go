package source

import (
	"strings"

	"github.com/your-org/logstream-tail/internal/logevent"
)

// TransformFunc is a function that transforms a log event, returning the
// modified event and whether it should be forwarded downstream. Returning
// false drops the event from the stream.
type TransformFunc func(logevent.Event) (logevent.Event, bool)

// Transformer applies a chain of TransformFuncs to each event received from
// an upstream channel, forwarding results to the output channel.
type Transformer struct {
	src   <-chan logevent.Event
	fns   []TransformFunc
	out   chan logevent.Event
}

// NewTransformer creates a Transformer that reads from src and applies each
// TransformFunc in order. If any function drops an event (returns false) the
// remaining functions are skipped and the event is not forwarded.
func NewTransformer(src <-chan logevent.Event, fns ...TransformFunc) *Transformer {
	return &Transformer{
		src: src,
		fns: fns,
		out: make(chan logevent.Event, cap(src)+1),
	}
}

// Out returns the read-only output channel.
func (t *Transformer) Out() <-chan logevent.Event {
	return t.out
}

// Run starts processing events until src is closed or ctx is done.
func (t *Transformer) Run(ctx interface{ Done() <-chan struct{} }) {
	go func() {
		defer close(t.out)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-t.src:
				if !ok {
					return
				}
				result, keep := ev, true
				for _, fn := range t.fns {
					result, keep = fn(result)
					if !keep {
						break
					}
				}
				if keep {
					select {
					case t.out <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
}

// UpperCaseMessage is a built-in TransformFunc that upper-cases the event
// message. Useful mainly for testing and demonstration purposes.
func UpperCaseMessage(ev logevent.Event) (logevent.Event, bool) {
	ev.Message = strings.ToUpper(ev.Message)
	return ev, true
}

// RedactTransform returns a TransformFunc that replaces occurrences of each
// sensitive string in the event message with the literal text "[REDACTED]".
func RedactTransform(sensitive ...string) TransformFunc {
	return func(ev logevent.Event) (logevent.Event, bool) {
		for _, s := range sensitive {
			ev.Message = strings.ReplaceAll(ev.Message, s, "[REDACTED]")
		}
		return ev, true
	}
}
