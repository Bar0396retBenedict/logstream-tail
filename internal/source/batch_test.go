package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeBatchEvent(msg string) logevent.Event {
	return logevent.Event{
		Message:   msg,
		Severity:  logevent.SeverityInfo,
		Timestamp: time.Now(),
		Source:    logevent.SourceCloudWatch,
	}
}

func TestBatcher_CollectsAndEmitsAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 10)
	for i := 0; i < 5; i++ {
		in <- makeBatchEvent("msg")
	}
	close(in)

	cfg := DefaultBatchConfig()
	out := NewBatcher(ctx, in, cfg)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}

	if len(got) != 5 {
		t.Fatalf("expected 5 events, got %d", len(got))
	}
}

func TestBatcher_FlushesOnMaxSize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := BatchConfig{MaxSize: 3, MaxWait: 10 * time.Second}
	in := make(chan logevent.Event, 20)

	for i := 0; i < 6; i++ {
		in <- makeBatchEvent("msg")
	}
	close(in)

	out := NewBatcher(ctx, in, cfg)

	var got []logevent.Event
	for ev := range out {
		got = append(got, ev)
	}

	if len(got) != 6 {
		t.Fatalf("expected 6 events, got %d", len(got))
	}
}

func TestBatcher_FlushesOnTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := BatchConfig{MaxSize: 100, MaxWait: 50 * time.Millisecond}
	in := make(chan logevent.Event, 10)
	in <- makeBatchEvent("a")
	in <- makeBatchEvent("b")

	out := NewBatcher(ctx, in, cfg)

	var got []logevent.Event
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case ev, ok := <-out:
			if !ok {
				goto done
			}
			got = append(got, ev)
			if len(got) == 2 {
				cancel()
			}
		case <-timeout:
			goto done
		}
	}
done:
	if len(got) != 2 {
		t.Fatalf("expected 2 events flushed by ticker, got %d", len(got))
	}
}

func TestBatcher_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := make(chan logevent.Event)
	cfg := DefaultBatchConfig()
	out := NewBatcher(ctx, in, cfg)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancellation")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestDefaultBatchConfig(t *testing.T) {
	cfg := DefaultBatchConfig()
	if cfg.MaxSize != 100 {
		t.Errorf("expected MaxSize 100, got %d", cfg.MaxSize)
	}
	if cfg.MaxWait != 2*time.Second {
		t.Errorf("expected MaxWait 2s, got %v", cfg.MaxWait)
	}
}
