package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeBufferEvent(msg string) logevent.Event {
	return logevent.Event{
		Message:   msg,
		Timestamp: time.Now(),
		Severity:  logevent.SeverityInfo,
		Source:    logevent.SourceCloudWatch,
	}
}

func TestBuffer_ForwardsAllEvents(t *testing.T) {
	in := make(chan logevent.Event, 10)
	cfg := BufferConfig{Size: 4, FlushInterval: 50 * time.Millisecond}
	buf, out := NewBuffer(in, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go buf.Run(ctx)

	const n = 8
	for i := 0; i < n; i++ {
		in <- makeBufferEvent("msg")
	}
	close(in)

	var received int
	for range out {
		received++
	}
	if received != n {
		t.Fatalf("expected %d events, got %d", n, received)
	}
}

func TestBuffer_FlushesOnInterval(t *testing.T) {
	in := make(chan logevent.Event, 10)
	cfg := BufferConfig{Size: 100, FlushInterval: 60 * time.Millisecond}
	buf, out := NewBuffer(in, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go buf.Run(ctx)

	in <- makeBufferEvent("tick-flush")

	select {
	case ev := <-out:
		if ev.Message != "tick-flush" {
			t.Fatalf("unexpected message: %s", ev.Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for interval flush")
	}
	close(in)
}

func TestBuffer_ContextCancellation(t *testing.T) {
	in := make(chan logevent.Event)
	cfg := DefaultBufferConfig()
	buf, out := NewBuffer(in, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		buf.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// out should be closed
		_, ok := <-out
		if ok {
			t.Fatal("expected output channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}

func TestDefaultBufferConfig(t *testing.T) {
	cfg := DefaultBufferConfig()
	if cfg.Size <= 0 {
		t.Errorf("expected positive Size, got %d", cfg.Size)
	}
	if cfg.FlushInterval <= 0 {
		t.Errorf("expected positive FlushInterval, got %v", cfg.FlushInterval)
	}
}
