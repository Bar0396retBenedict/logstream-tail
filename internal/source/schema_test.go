package source

import (
	"regexp"
	"testing"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

func makeSchemaEvent(msg string, src logevent.Source) logevent.Event {
	return logevent.Event{
		Timestamp: time.Now(),
		Message:   msg,
		Source:    src,
		Severity:  logevent.SeverityInfo,
	}
}

func feedSchema(events []logevent.Event) <-chan logevent.Event {
	ch := make(chan logevent.Event, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return ch
}

func collectSchema(ch <-chan logevent.Event) []logevent.Event {
	var out []logevent.Event
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestSchemaValidator_PassesAllWhenNoRules(t *testing.T) {
	events := []logevent.Event{
		makeSchemaEvent("hello", logevent.SourceCloudWatch),
		makeSchemaEvent("", logevent.SourceGCP),
	}
	out := NewSchemaValidator(feedSchema(events), DefaultSchemaConfig())
	got := collectSchema(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
}

func TestSchemaValidator_RequiredField_DropOnFail(t *testing.T) {
	events := []logevent.Event{
		makeSchemaEvent("has message", logevent.SourceCloudWatch),
		makeSchemaEvent("", logevent.SourceGCP), // should be dropped
	}
	cfg := NewSchemaConfig(WithRequiredField("message"), WithDropOnFail())
	out := NewSchemaValidator(feedSchema(events), cfg)
	got := collectSchema(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Message != "has message" {
		t.Errorf("unexpected message: %q", got[0].Message)
	}
}

func TestSchemaValidator_RequiredField_PassThrough(t *testing.T) {
	events := []logevent.Event{
		makeSchemaEvent("", logevent.SourceGCP),
	}
	cfg := NewSchemaConfig(WithRequiredField("message")) // DropOnFail defaults false
	out := NewSchemaValidator(feedSchema(events), cfg)
	got := collectSchema(out)
	if len(got) != 1 {
		t.Fatalf("expected pass-through, got %d events", len(got))
	}
}

func TestSchemaValidator_PatternRule_Drops(t *testing.T) {
	events := []logevent.Event{
		makeSchemaEvent("ok", logevent.SourceCloudWatch),
		makeSchemaEvent("bad", logevent.SourceGCP),
	}
	// only cloudwatch source passes
	cfg := NewSchemaConfig(
		WithFieldPattern("source", `^cloudwatch$`),
		WithDropOnFail(),
	)
	out := NewSchemaValidator(feedSchema(events), cfg)
	got := collectSchema(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
}

func TestDefaultSchemaConfig_NilRules(t *testing.T) {
	cfg := DefaultSchemaConfig()
	if cfg.Rules != nil {
		t.Errorf("expected nil rules, got %v", cfg.Rules)
	}
	if cfg.DropOnFail {
		t.Error("expected DropOnFail false by default")
	}
}

func TestWithFieldPattern_PanicsOnBadRegex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid regex")
		}
	}()
	_ = regexp.MustCompile(`[invalid`)
}
