package logevent

import (
	"fmt"
	"time"
)

// Source identifies the origin of a log event.
type Source string

const (
	SourceCloudWatch Source = "cloudwatch"
	SourceGCPLogging  Source = "gcp"
	SourceUnknown     Source = "unknown"
)

// Severity maps cloud-provider log levels to a common set.
type Severity int

const (
	SeverityDebug   Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARN"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Event is a normalised log record from any supported source.
type Event struct {
	Timestamp time.Time
	Source    Source
	Stream    string   // log group / log name / stream identifier
	Severity  Severity
	Message   string
	Labels    map[string]string // arbitrary key-value metadata
}

// String returns a human-readable, single-line representation suitable for
// terminal output.
func (e Event) String() string {
	return fmt.Sprintf("%s [%-8s] (%s) %s: %s",
		e.Timestamp.UTC().Format(time.RFC3339),
		e.Severity.String(),
		e.Source,
		e.Stream,
		e.Message,
	)
}
