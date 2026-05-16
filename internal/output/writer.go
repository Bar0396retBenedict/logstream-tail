// Package output handles writing formatted log events to an io.Writer,
// supporting optional line-buffering and graceful shutdown.
package output

import (
	"context"
	"io"
	"sync"

	"github.com/user/logstream-tail/internal/formatter"
	"github.com/user/logstream-tail/internal/logevent"
)

// Writer consumes log events from a channel and writes them to an io.Writer
// using the provided formatter.
type Writer struct {
	out       io.Writer
	fmt       *formatter.Formatter
	mu        sync.Mutex
}

// New creates a new Writer that formats events with f and writes to out.
func New(out io.Writer, f *formatter.Formatter) *Writer {
	return &Writer{out: out, fmt: f}
}

// Run reads events from ch until it is closed or ctx is cancelled.
// It returns any write error encountered.
func (w *Writer) Run(ctx context.Context, ch <-chan logevent.Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			if err := w.write(ev); err != nil {
				return err
			}
		}
	}
}

func (w *Writer) write(ev logevent.Event) error {
	line := w.fmt.Format(ev)
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := io.WriteString(w.out, line+"\n")
	return err
}
