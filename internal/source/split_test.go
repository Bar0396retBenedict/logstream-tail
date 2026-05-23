package source

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
)

func makeSplitEvent(src, msg string) logevent.Event {
	return logevent.Event{Source: src, Message: msg}
}

func collectSplit(ch <-chan logevent.Event) []logevent.Event {
	var evs []logevent.Event
	for e := range ch {
		evs = append(evs, e)
	}
	return evs
}

func TestSplitter_RoutesEventsByKey(t *testing.T) {
	in := make(chan logevent.Event, 10)
	in <- makeSplitEvent("aws", "a")
	in <- makeSplitEvent("gcp", "b")
	in <- makeSplitEvent("aws", "c")
	in <- makeSplitEvent("gcp", "d")
	close(in)

	ctx := context.Background()
	outs := NewSplitter(ctx, in, []string{"aws", "gcp"}, func(e logevent.Event) string {
		return e.Source
	}, DefaultSplitConfig())

	awsEvs := collectSplit(outs["aws"])
	gcpEvs := collectSplit(outs["gcp"])

	if len(awsEvs) != 2 {
		t.Fatalf("expected 2 aws events, got %d", len(awsEvs))
	}
	if len(gcpEvs) != 2 {
		t.Fatalf("expected 2 gcp events, got %d", len(gcpEvs))
	}
}

func TestSplitter_DropsUnknownKeys(t *testing.T) {
	in := make(chan logevent.Event, 10)
	in <- makeSplitEvent("aws", "keep")
	in <- makeSplitEvent("unknown", "drop")
	close(in)

	ctx := context.Background()
	outs := NewSplitter(ctx, in, []string{"aws"}, func(e logevent.Event) string {
		return e.Source
	}, DefaultSplitConfig())

	awsEvs := collectSplit(outs["aws"])
	if len(awsEvs) != 1 {
		t.Fatalf("expected 1 aws event, got %d", len(awsEvs))
	}
}

func TestSplitter_ContextCancellation(t *testing.T) {
	in := make(chan logevent.Event)
	ctx, cancel := context.WithCancel(context.Background())

	outs := NewSplitter(ctx, in, []string{"aws"}, func(e logevent.Event) string {
		return e.Source
	}, DefaultSplitConfig())

	cancel()

	select {
	case _, ok := <-outs["aws"]:
		if ok {
			t.Fatal("expected channel to be closed after context cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestDefaultSplitConfig(t *testing.T) {
	cfg := DefaultSplitConfig()
	if cfg.BufSize <= 0 {
		t.Fatalf("expected positive BufSize, got %d", cfg.BufSize)
	}
}
