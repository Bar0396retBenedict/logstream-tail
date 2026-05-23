package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeThrottleEvent(src logevent.EventSource, msg string) logevent.Event {
	return logevent.Event{
		Timestamp: time.Now(),
		Severity:  logevent.SeverityInfo,
		Source:    src,
		Message:   msg,
	}
}

func TestThrottler_PassesFirstEvent(t *testing.T) {
	cfg := NewThrottleConfig(WithThrottleInterval(500 * time.Millisecond))
	in := make(chan logevent.Event, 4)
	in <- makeThrottleEvent("svc-a", "first")
	close(in)

	out := NewThrottler(cfg, in)
	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Message != "first" {
		t.Errorf("unexpected message: %s", got[0].Message)
	}
}

func TestThrottler_DropsRapidDuplicates(t *testing.T) {
	cfg := NewThrottleConfig(WithThrottleInterval(10 * time.Second))
	in := make(chan logevent.Event, 8)
	for i := 0; i < 5; i++ {
		in <- makeThrottleEvent("svc-b", "burst")
	}
	close(in)

	out := NewThrottler(cfg, in)
	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 forwarded event, got %d", len(got))
	}
}

func TestThrottler_IndependentSources(t *testing.T) {
	cfg := NewThrottleConfig(WithThrottleInterval(10 * time.Second))
	in := make(chan logevent.Event, 8)
	in <- makeThrottleEvent("src-x", "x1")
	in <- makeThrottleEvent("src-y", "y1")
	in <- makeThrottleEvent("src-x", "x2") // dropped
	in <- makeThrottleEvent("src-y", "y2") // dropped
	close(in)

	out := NewThrottler(cfg, in)
	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events (one per source), got %d", len(got))
	}
}

func TestThrottler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := DefaultThrottleConfig()
	in := make(chan logevent.Event)
	out := NewThrottlerContext(ctx, cfg, in)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestDefaultThrottleConfig(t *testing.T) {
	cfg := DefaultThrottleConfig()
	if cfg.Interval <= 0 {
		t.Errorf("expected positive interval, got %v", cfg.Interval)
	}
}

func TestWithThrottleInterval_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for zero interval")
		}
	}()
	WithThrottleInterval(0)
}
