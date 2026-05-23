package source

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
)

func makeDedupEvent(src, msg string, sev logevent.Severity) logevent.Event {
	return logevent.Event{
		Source:    src,
		Severity:  sev,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func TestDeduplicator_PassesUniqueEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	upstream := make(chan logevent.Event, 4)
	upstream <- makeDedupEvent("svc-a", "hello", logevent.SeverityInfo)
	upstream <- makeDedupEvent("svc-a", "world", logevent.SeverityInfo)
	upstream <- makeDedupEvent("svc-b", "hello", logevent.SeverityInfo)
	close(upstream)

	cfg := DefaultDedupeConfig()
	out := NewDeduplicator(ctx, upstream, cfg)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
}

func TestDeduplicator_SuppressDuplicates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	upstream := make(chan logevent.Event, 5)
	ev := makeDedupEvent("svc-a", "repeated", logevent.SeverityWarn)
	for i := 0; i < 5; i++ {
		upstream <- ev
	}
	close(upstream)

	cfg := DefaultDedupeConfig()
	out := NewDeduplicator(ctx, upstream, cfg)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 unique event, got %d", len(got))
	}
}

func TestDeduplicator_WindowExpiry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	upstream := make(chan logevent.Event, 2)
	cfg := DefaultDedupeConfig()
	cfg.Window = 50 * time.Millisecond

	out := NewDeduplicator(ctx, upstream, cfg)

	ev := makeDedupEvent("svc-a", "msg", logevent.SeverityInfo)
	upstream <- ev
	time.Sleep(100 * time.Millisecond)
	upstream <- ev
	close(upstream)

	var got []logevent.Event
	for e := range out {
		got = append(got, e)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events after window expiry, got %d", len(got))
	}
}

func TestDeduplicator_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	upstream := make(chan logevent.Event)
	cfg := DefaultDedupeConfig()
	out := NewDeduplicator(ctx, upstream, cfg)

	cancel()
	time.Sleep(50 * time.Millisecond)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected output channel to be closed after context cancel")
		}
	default:
		t.Fatal("output channel not closed after context cancel")
	}
}
