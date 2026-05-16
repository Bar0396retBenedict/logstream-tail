package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// fakeCloudWatchClient implements cloudWatchLogsClient for testing.
type fakeCloudWatchClient struct {
	events []rawCWEvent
	calls  int
	err    error
}

func (f *fakeCloudWatchClient) FilterLogEvents(_ context.Context, _, _ string, _ time.Time) ([]rawCWEvent, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.events, nil
}

func TestCloudWatchSource_EmitsEvents(t *testing.T) {
	now := time.Now()
	fake := &fakeCloudWatchClient{
		events: []rawCWEvent{
			{Timestamp: now.UnixMilli(), Message: "hello from cw", Stream: "app/web"},
			{Timestamp: now.Add(time.Millisecond).UnixMilli(), Message: "second event", Stream: "app/web"},
		},
	}

	cfg := CloudWatchConfig{
		Config:   Config{PollInterval: 20 * time.Millisecond},
		Region:   "us-east-1",
		LogGroup: "/app/prod",
	}
	src := NewCloudWatchSource(cfg, fake)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	out := make(chan logevent.Event, 16)
	go src.Stream(ctx, out) //nolint:errcheck

	<-ctx.Done()
	close(out)

	var collected []logevent.Event
	for ev := range out {
		collected = append(collected, ev)
	}

	if len(collected) == 0 {
		t.Fatal("expected at least one event, got none")
	}
	if collected[0].Source != logevent.SourceCloudWatch {
		t.Errorf("expected source CloudWatch, got %v", collected[0].Source)
	}
	if collected[0].Labels["region"] != "us-east-1" {
		t.Errorf("expected region label us-east-1, got %s", collected[0].Labels["region"])
	}
}

func TestCloudWatchSource_FilterApplied(t *testing.T) {
	now := time.Now()
	fake := &fakeCloudWatchClient{
		events: []rawCWEvent{
			{Timestamp: now.UnixMilli(), Message: "keep", Stream: "s"},
			{Timestamp: now.UnixMilli(), Message: "drop", Stream: "s"},
		},
	}
	cfg := CloudWatchConfig{
		Config: Config{
			PollInterval: 20 * time.Millisecond,
			Filter: func(e logevent.Event) bool { return e.Message == "keep" },
		},
		Region:   "eu-west-1",
		LogGroup: "/svc",
	}
	src := NewCloudWatchSource(cfg, fake)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	out := make(chan logevent.Event, 16)
	go src.Stream(ctx, out) //nolint:errcheck
	<-ctx.Done()
	close(out)

	for ev := range out {
		if ev.Message == "drop" {
			t.Errorf("filtered event should not have been emitted: %s", ev.Message)
		}
	}
}

func TestCloudWatchSource_DefaultPollInterval(t *testing.T) {
	src := NewCloudWatchSource(CloudWatchConfig{}, &fakeCloudWatchClient{})
	if src.cfg.PollInterval != DefaultConfig.PollInterval {
		t.Errorf("expected default poll interval %v, got %v", DefaultConfig.PollInterval, src.cfg.PollInterval)
	}
}
