package source

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
)

func makeMulticastEvent(msg string) logevent.Event {
	return logevent.Event{Message: msg, Source: logevent.CloudWatch}
}

func TestMulticast_BroadcastsToAllSubscribers(t *testing.T) {
	input := make(chan logevent.Event, 4)
	mc := NewMulticast(input, 8)

	sub1 := mc.Subscribe()
	sub2 := mc.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); mc.Run(ctx) }()

	events := []logevent.Event{
		makeMulticastEvent("alpha"),
		makeMulticastEvent("beta"),
	}
	for _, ev := range events {
		input <- ev
	}
	close(input)
	wg.Wait()

	for _, sub := range []<-chan logevent.Event{sub1, sub2} {
		var got []string
		for ev := range sub {
			got = append(got, ev.Message)
		}
		if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
			t.Errorf("subscriber received unexpected events: %v", got)
		}
	}
}

func TestMulticast_ContextCancellation(t *testing.T) {
	input := make(chan logevent.Event)
	mc := NewMulticast(input, 8)
	sub := mc.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() { mc.Run(ctx); close(done) }()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}

	// sub must be closed after Run returns
	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("expected subscriber channel to be closed")
		}
	default:
		t.Fatal("subscriber channel not closed")
	}
}

func TestMulticast_DefaultBufSize(t *testing.T) {
	input := make(chan logevent.Event)
	mc := NewMulticast(input, 0) // 0 should fall back to 64
	if mc.bufSize != 64 {
		t.Errorf("expected bufSize 64, got %d", mc.bufSize)
	}
}
