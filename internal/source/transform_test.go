package source

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
)

func makeTransformEvent(msg string) logevent.Event {
	return logevent.Event{
		Message:  msg,
		Severity: logevent.SeverityInfo,
		Source:   logevent.SourceCloudWatch,
	}
}

func TestTransformer_PassesEventsUnchanged(t *testing.T) {
	src := make(chan logevent.Event, 2)
	src <- makeTransformEvent("hello")
	src <- makeTransformEvent("world")
	close(src)

	tr := NewTransformer(src)
	tr.Run(context.Background())

	var got []string
	for ev := range tr.Out() {
		got = append(got, ev.Message)
	}
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Fatalf("unexpected events: %v", got)
	}
}

func TestTransformer_AppliesTransformFunc(t *testing.T) {
	src := make(chan logevent.Event, 1)
	src <- makeTransformEvent("hello")
	close(src)

	tr := NewTransformer(src, UpperCaseMessage)
	tr.Run(context.Background())

	ev := <-tr.Out()
	if ev.Message != "HELLO" {
		t.Fatalf("expected HELLO, got %q", ev.Message)
	}
}

func TestTransformer_DropsEventWhenFuncReturnsFalse(t *testing.T) {
	drop := func(ev logevent.Event) (logevent.Event, bool) { return ev, false }

	src := make(chan logevent.Event, 2)
	src <- makeTransformEvent("keep")
	src <- makeTransformEvent("drop")
	close(src)

	tr := NewTransformer(src,
		func(ev logevent.Event) (logevent.Event, bool) {
			if ev.Message == "drop" {
				return drop(ev)
			}
			return ev, true
		},
	)
	tr.Run(context.Background())

	var got []string
	for ev := range tr.Out() {
		got = append(got, ev.Message)
	}
	if len(got) != 1 || got[0] != "keep" {
		t.Fatalf("unexpected events: %v", got)
	}
}

func TestTransformer_ContextCancellation(t *testing.T) {
	src := make(chan logevent.Event) // never sends
	ctx, cancel := context.WithCancel(context.Background())

	tr := NewTransformer(src)
	tr.Run(ctx)
	cancel()

	select {
	case <-tr.Out():
	case <-time.After(time.Second):
		t.Fatal("output channel not closed after context cancellation")
	}
}

func TestRedactTransform(t *testing.T) {
	src := make(chan logevent.Event, 1)
	src <- makeTransformEvent("token=secret123 user=alice")
	close(src)

	tr := NewTransformer(src, RedactTransform("secret123"))
	tr.Run(context.Background())

	ev := <-tr.Out()
	if strings.Contains(ev.Message, "secret123") {
		t.Fatalf("secret not redacted: %q", ev.Message)
	}
	if !strings.Contains(ev.Message, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in message: %q", ev.Message)
	}
}
