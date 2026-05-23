package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeWindowEvent(msg string) logevent.Event {
	return logevent.Event{
		Message:   msg,
		Severity:  logevent.SeverityInfo,
		Timestamp: time.Now(),
		Source:    logevent.SourceCloudWatch,
	}
}

func TestWindow_PassesEventsUnderCap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 10)
	cfg := WindowConfig{Size: 5 * time.Second, MaxEvents: 10}
	out := Window(ctx, in, cfg)

	for i := 0; i < 5; i++ {
		in <- makeWindowEvent("msg")
	}
	close(in)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}

	if len(got) != 5 {
		t.Fatalf("expected 5 events, got %d", len(got))
	}
}

func TestWindow_DropsEventsOverCap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 20)
	cfg := WindowConfig{Size: 10 * time.Second, MaxEvents: 3}
	out := Window(ctx, in, cfg)

	for i := 0; i < 10; i++ {
		in <- makeWindowEvent("msg")
	}
	close(in)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 events (cap), got %d", len(got))
	}
}

func TestWindow_ResetsCountOnTick(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 20)
	cfg := WindowConfig{Size: 200 * time.Millisecond, MaxEvents: 2}
	out := Window(ctx, in, cfg)

	// Send 2 events in first window
	in <- makeWindowEvent("a")
	in <- makeWindowEvent("b")

	// Wait for window to reset
	time.Sleep(300 * time.Millisecond)

	// Send 2 more events in second window
	in <- makeWindowEvent("c")
	in <- makeWindowEvent("d")
	close(in)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}

	if len(got) != 4 {
		t.Fatalf("expected 4 events across two windows, got %d", len(got))
	}
}

func TestWindow_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan logevent.Event)
	cfg := DefaultWindowConfig()
	out := Window(ctx, in, cfg)

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

func TestDefaultWindowConfig(t *testing.T) {
	cfg := DefaultWindowConfig()
	if cfg.Size != 5*time.Second {
		t.Errorf("expected Size=5s, got %v", cfg.Size)
	}
	if cfg.MaxEvents != 500 {
		t.Errorf("expected MaxEvents=500, got %d", cfg.MaxEvents)
	}
}
