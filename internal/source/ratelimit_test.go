package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeEvent(msg string) logevent.Event {
	return logevent.Event{
		Timestamp: time.Now(),
		Severity:  logevent.SeverityInfo,
		Message:   msg,
		Source:    "test",
	}
}

func TestRateLimiter_NoLimit_PassesAll(t *testing.T) {
	input := make(chan logevent.Event, 10)
	rl := NewRateLimiter(input, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rl.Run(ctx)

	for i := 0; i < 5; i++ {
		input <- makeEvent("msg")
	}
	close(input)

	var received int
	for range rl.Out() {
		received++
	}
	if received != 5 {
		t.Fatalf("expected 5 events, got %d", received)
	}
}

func TestRateLimiter_DropsExcessEvents(t *testing.T) {
	input := make(chan logevent.Event, 20)
	const limit = 3
	rl := NewRateLimiter(input, limit)

	// Send more events than the per-second limit in one burst.
	for i := 0; i < 10; i++ {
		input <- makeEvent("msg")
	}
	close(input)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go rl.Run(ctx)

	var received int
	for range rl.Out() {
		received++
	}

	if received > limit {
		t.Fatalf("expected at most %d events, got %d", limit, received)
	}
	if rl.Dropped() != 10-received {
		t.Fatalf("expected %d dropped, got %d", 10-received, rl.Dropped())
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	input := make(chan logevent.Event)
	rl := NewRateLimiter(input, 100)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		rl.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("RateLimiter did not stop after context cancellation")
	}
}
