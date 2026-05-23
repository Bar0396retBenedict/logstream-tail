package source_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/logstream-tail/internal/logevent"
	"github.com/your-org/logstream-tail/internal/source"
)

func makeSampleEvent(msg string) logevent.Event {
	return logevent.Event{
		Message:  msg,
		Severity: logevent.SeverityInfo,
		Source:   logevent.SourceCloudWatch,
	}
}

func TestSampler_Rate1_PassesAll(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 10)
	cfg := source.NewSampleConfig(source.WithSampleRate(1))
	out := source.NewSampler(ctx, in, cfg)

	const total = 8
	for i := 0; i < total; i++ {
		in <- makeSampleEvent("msg")
	}
	close(in)

	var got int
	for range out {
		got++
	}
	if got != total {
		t.Fatalf("expected %d events, got %d", total, got)
	}
}

func TestSampler_Rate5_DropsOthers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	in := make(chan logevent.Event, 20)
	cfg := source.NewSampleConfig(source.WithSampleRate(5))
	out := source.NewSampler(ctx, in, cfg)

	const total = 20
	for i := 0; i < total; i++ {
		in <- makeSampleEvent("msg")
	}
	close(in)

	var got int
	for range out {
		got++
	}
	// expect total/rate == 4
	if got != total/5 {
		t.Fatalf("expected %d events, got %d", total/5, got)
	}
}

func TestSampler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan logevent.Event) // never closed
	cfg := source.DefaultSampleConfig()
	out := source.NewSampler(ctx, in, cfg)

	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected channel to be closed after context cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for sampler to stop")
	}
}

func TestDefaultSampleConfig(t *testing.T) {
	cfg := source.DefaultSampleConfig()
	if cfg.Rate != 1 {
		t.Fatalf("expected default rate 1, got %d", cfg.Rate)
	}
}

func TestNewSampleConfig_AppliesOptions(t *testing.T) {
	cfg := source.NewSampleConfig(source.WithSampleRate(7))
	if cfg.Rate != 7 {
		t.Fatalf("expected rate 7, got %d", cfg.Rate)
	}
}
