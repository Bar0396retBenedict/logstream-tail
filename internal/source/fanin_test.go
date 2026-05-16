package source_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
	"github.com/user/logstream-tail/internal/source"
)

// stubSource emits a fixed set of events then returns.
type stubSource struct {
	name   string
	events []logevent.Event
}

func (s *stubSource) Name() string { return s.name }
func (s *stubSource) Close() error { return nil }
func (s *stubSource) Start(_ context.Context, out chan<- logevent.Event) error {
	for _, e := range s.events {
		out <- e
	}
	return nil
}

func TestFanIn_CollectsAllEvents(t *testing.T) {
	evA := logevent.Event{Source: logevent.SourceCloudWatch, Message: "a"}
	evB := logevent.Event{Source: logevent.SourceGCP, Message: "b"}

	fa := source.NewFanIn(
		&stubSource{name: "cw", events: []logevent.Event{evA}},
		&stubSource{name: "gcp", events: []logevent.Event{evB}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := fa.Stream(ctx)

	got := make(map[string]bool)
	for evt := range ch {
		got[evt.Message] = true
	}

	if !got["a"] || !got["b"] {
		t.Errorf("expected both events; got %v", got)
	}
}

func TestFanIn_ContextCancellation(t *testing.T) {
	blocking := &blockingSource{}
	fa := source.NewFanIn(blocking)

	ctx, cancel := context.WithCancel(context.Background())
	ch := fa.Stream(ctx)

	cancel()

	select {
	case <-ch:
		// channel closed after cancel — pass
	case <-time.After(2 * time.Second):
		t.Fatal("channel was not closed after context cancellation")
	}
}

// blockingSource blocks until the context is cancelled.
type blockingSource struct{}

func (b *blockingSource) Name() string { return "blocking" }
func (b *blockingSource) Close() error { return nil }
func (b *blockingSource) Start(ctx context.Context, _ chan<- logevent.Event) error {
	<-ctx.Done()
	return ctx.Err()
}
