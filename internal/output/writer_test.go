package output_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/formatter"
	"github.com/user/logstream-tail/internal/logevent"
	"github.com/user/logstream-tail/internal/output"
)

func makeEvent(msg, src string) logevent.Event {
	return logevent.Event{
		Timestamp: time.Now(),
		Severity:  logevent.SeverityInfo,
		Message:   msg,
		Source:    src,
	}
}

func TestWriter_WritesFormattedEvents(t *testing.T) {
	var buf bytes.Buffer
	f := formatter.New(formatter.StylePlain, false)
	w := output.New(&buf, f)

	ch := make(chan logevent.Event, 2)
	ch <- makeEvent("hello world", "cloudwatch")
	ch <- makeEvent("second line", "gcp")
	close(ch)

	ctx := context.Background()
	if err := w.Run(ctx, ch); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world' in output, got: %s", out)
	}
	if !strings.Contains(out, "second line") {
		t.Errorf("expected 'second line' in output, got: %s", out)
	}
}

func TestWriter_ContextCancellation(t *testing.T) {
	var buf bytes.Buffer
	f := formatter.New(formatter.StylePlain, false)
	w := output.New(&buf, f)

	ch := make(chan logevent.Event) // never sends
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := w.Run(ctx, ch)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestWriter_EachEventOnOwnLine(t *testing.T) {
	var buf bytes.Buffer
	f := formatter.New(formatter.StylePlain, false)
	w := output.New(&buf, f)

	ch := make(chan logevent.Event, 3)
	for i := 0; i < 3; i++ {
		ch <- makeEvent("msg", "src")
	}
	close(ch)

	_ = w.Run(context.Background(), ch)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), buf.String())
	}
}
