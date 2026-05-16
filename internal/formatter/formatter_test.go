package formatter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/logstream-tail/internal/formatter"
	"github.com/logstream-tail/internal/logevent"
)

var sampleEvent = logevent.Event{
	Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	Source:    logevent.SourceCloudWatch,
	Severity:  logevent.SeverityError,
	Message:   "disk full",
}

func TestFormatter_PlainContainsFields(t *testing.T) {
	f := formatter.New()
	f.Style = formatter.StylePlain
	out := f.Format(sampleEvent)

	for _, want := range []string{"2024-06-01", "cloudwatch", "ERROR", "disk full"} {
		if !strings.Contains(out, want) {
			t.Errorf("plain output missing %q: %s", want, out)
		}
	}
}

func TestFormatter_JSONIsValidish(t *testing.T) {
	f := formatter.New()
	f.Style = formatter.StyleJSON
	out := f.Format(sampleEvent)

	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(out, "}") {
		t.Errorf("expected JSON object, got: %s", out)
	}
	for _, key := range []string{`"ts"`, `"source"`, `"severity"`, `"msg"`} {
		if !strings.Contains(out, key) {
			t.Errorf("JSON output missing key %s: %s", key, out)
		}
	}
}

func TestFormatter_ColoredContainsANSI(t *testing.T) {
	f := formatter.New()
	f.Style = formatter.StyleColored
	out := f.Format(sampleEvent)

	if !strings.Contains(out, "\033[") {
		t.Errorf("colored output missing ANSI codes: %s", out)
	}
}

func TestFormatter_HideSource(t *testing.T) {
	f := formatter.New()
	f.Style = formatter.StylePlain
	f.ShowSource = false
	out := f.Format(sampleEvent)

	if strings.Contains(out, "cloudwatch") {
		t.Errorf("expected source to be hidden, got: %s", out)
	}
}

func TestStyleFromString(t *testing.T) {
	cases := []struct {
		input string
		want  formatter.Style
		wantErr bool
	}{
		{"plain", formatter.StylePlain, false},
		{"colored", formatter.StyleColored, false},
		{"colour", formatter.StyleColored, false},
		{"json", formatter.StyleJSON, false},
		{"nope", formatter.StylePlain, true},
	}
	for _, tc := range cases {
		got, err := formatter.StyleFromString(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("StyleFromString(%q) error=%v wantErr=%v", tc.input, err, tc.wantErr)
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("StyleFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
