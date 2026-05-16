package source

import (
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// mockGCPClient satisfies GCPLoggingClient for testing.
type mockGCPClient struct {
	entries []GCPLogEntry
	calls   int
}

func (m *mockGCPClient) ListLogEntries(_ context.Context, _, _ string, _ time.Time) ([]GCPLogEntry, error) {
	m.calls++
	if m.calls == 1 {
		return m.entries, nil
	}
	return nil, nil
}

func TestGCPSource_EmitsEvents(t *testing.T) {
	now := time.Now()
	client := &mockGCPClient{
		entries: []GCPLogEntry{
			{Timestamp: now, Severity: "INFO", Payload: "hello gcp", LogName: "app"},
			{Timestamp: now, Severity: "ERROR", Payload: "oh no", LogName: "app"},
		},
	}
	cfg := Config{
		PollInterval:   20 * time.Millisecond,
		MinSeverity:    logevent.SeverityInfo,
		LookbackWindow: time.Minute,
	}
	src := NewGCPSource(client, "my-project", "app", cfg)
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := src.Stream(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var events []logevent.Event
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()
collect:
	for {
		select {
		case e, ok := <-ch:
			if !ok {
				break collect
			}
			events = append(events, e)
			if len(events) >= 2 {
				cancel()
			}
		case <-timer.C:
			cancel()
			break collect
		}
	}
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
	if events[0].Message != "hello gcp" {
		t.Errorf("unexpected message: %s", events[0].Message)
	}
}

func TestGCPSource_FilterApplied(t *testing.T) {
	now := time.Now()
	client := &mockGCPClient{
		entries: []GCPLogEntry{
			{Timestamp: now, Severity: "DEBUG", Payload: "verbose", LogName: "app"},
			{Timestamp: now, Severity: "ERROR", Payload: "critical", LogName: "app"},
		},
	}
	cfg := Config{
		PollInterval:   20 * time.Millisecond,
		MinSeverity:    logevent.SeverityError,
		LookbackWindow: time.Minute,
	}
	src := NewGCPSource(client, "my-project", "app", cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	ch, _ := src.Stream(ctx)
	var events []logevent.Event
	for e := range ch {
		events = append(events, e)
	}
	for _, e := range events {
		if e.Severity < logevent.SeverityError {
			t.Errorf("expected only ERROR+, got %s", e.Severity)
		}
	}
}

func TestGCPSource_DefaultPollInterval(t *testing.T) {
	client := &mockGCPClient{}
	src := NewGCPSource(client, "p", "l", Config{})
	if src.cfg.PollInterval != DefaultConfig.PollInterval {
		t.Errorf("expected default poll interval %v, got %v", DefaultConfig.PollInterval, src.cfg.PollInterval)
	}
}
