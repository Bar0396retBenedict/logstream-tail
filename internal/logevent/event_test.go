package logevent

import (
	"strings"
	"testing"
	"time"
)

func TestSeverityString(t *testing.T) {
	cases := []struct {
		sev  Severity
		want string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARN"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{Severity(99), "UNKNOWN"},
	}
	for _, c := range cases {
		if got := c.sev.String(); got != c.want {
			t.Errorf("Severity(%d).String() = %q, want %q", c.sev, got, c.want)
		}
	}
}

func TestEventString(t *testing.T) {
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	e := Event{
		Timestamp: ts,
		Source:    SourceCloudWatch,
		Stream:    "/aws/lambda/my-function",
		Severity:  SeverityError,
		Message:   "something went wrong",
		Labels:    map[string]string{"region": "us-east-1"},
	}

	got := e.String()

	if !strings.Contains(got, "2024-06-01T12:00:00Z") {
		t.Errorf("expected timestamp in output, got: %s", got)
	}
	if !strings.Contains(got, "ERROR") {
		t.Errorf("expected severity ERROR in output, got: %s", got)
	}
	if !strings.Contains(got, string(SourceCloudWatch)) {
		t.Errorf("expected source cloudwatch in output, got: %s", got)
	}
	if !strings.Contains(got, "/aws/lambda/my-function") {
		t.Errorf("expected stream name in output, got: %s", got)
	}
	if !strings.Contains(got, "something went wrong") {
		t.Errorf("expected message in output, got: %s", got)
	}
}

func TestEventSourceConstants(t *testing.T) {
	if SourceCloudWatch != "cloudwatch" {
		t.Errorf("unexpected value for SourceCloudWatch: %s", SourceCloudWatch)
	}
	if SourceGCPLogging != "gcp" {
		t.Errorf("unexpected value for SourceGCPLogging: %s", SourceGCPLogging)
	}
}
