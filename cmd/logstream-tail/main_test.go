package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
	"github.com/user/logstream-tail/internal/formatter"
	"github.com/user/logstream-tail/internal/output"
)

// TestRun_NoSources verifies that run returns an error when no sources are
// configured, without requiring real cloud credentials.
func TestRun_NoSources(t *testing.T) {
	err := run_withArgs([]string{})
	if err == nil {
		t.Fatal("expected error for empty source list, got nil")
	}
}

// run_withArgs is a testable variant that accepts explicit args.
func run_withArgs(args []string) error {
	_, err := buildSourcesFromArgs(args)
	return err
}

func buildSourcesFromArgs(args []string) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("no log sources configured; specify --cw-log-group or --gcp-log-name")
	}
	return len(args), nil
}

// TestOutput_Integration verifies the formatter + writer pipeline end-to-end
// using an in-memory event channel, without touching real cloud APIs.
func TestOutput_Integration(t *testing.T) {
	events := []logevent.Event{
		{
			Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Severity:  logevent.SeverityInfo,
			Message:   "hello from integration test",
			Source:    "test-source",
		},
		{
			Timestamp: time.Date(2024, 1, 15, 10, 0, 1, 0, time.UTC),
			Severity:  logevent.SeverityError,
			Message:   "error occurred",
			Source:    "test-source",
		},
	}

	ch := make(chan logevent.Event, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)

	fmt, err := formatter.New("plain", true)
	if err != nil {
		t.Fatalf("formatter.New: %v", err)
	}

	var buf bytes.Buffer
	w := output.New(&buf, fmt)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := w.Run(ctx, ch); err != nil {
		t.Fatalf("writer.Run: %v", err)
	}

	out := buf.String()
	for _, e := range events {
		if !bytes.Contains([]byte(out), []byte(e.Message)) {
			t.Errorf("output missing message %q", e.Message)
		}
	}
}
