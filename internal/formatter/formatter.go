package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/logstream-tail/internal/logevent"
)

// Style controls how log events are rendered to text.
type Style int

const (
	StylePlain Style = iota
	StyleColored
	StyleJSON
)

// Formatter converts a LogEvent into a printable string.
type Formatter struct {
	Style     Style
	TimeZone  *time.Location
	ShowSource bool
}

// New returns a Formatter with sensible defaults.
func New() *Formatter {
	return &Formatter{
		Style:      StyleColored,
		TimeZone:   time.UTC,
		ShowSource: true,
	}
}

// Format renders a single log event as a string.
func (f *Formatter) Format(e logevent.Event) string {
	ts := e.Timestamp.In(f.TimeZone).Format(time.RFC3339)

	switch f.Style {
	case StyleJSON:
		return fmt.Sprintf(
			`{"ts":%q,"source":%q,"severity":%q,"msg":%q}`,
			ts, e.Source, e.Severity.String(), e.Message,
		)
	case StyleColored:
		color := severityColor(e.Severity)
		src := ""
		if f.ShowSource {
			src = fmt.Sprintf(" [%s]", e.Source)
		}
		return fmt.Sprintf("%s%s %s%-7s%s %s",
			ts, src, color, e.Severity.String(), colorReset, e.Message)
	default: // StylePlain
		src := ""
		if f.ShowSource {
			src = fmt.Sprintf(" [%s]", e.Source)
		}
		return fmt.Sprintf("%s%s %-7s %s", ts, src, e.Severity.String(), e.Message)
	}
}

// ANSI color codes.
const colorReset = "\033[0m"

func severityColor(s logevent.Severity) string {
	switch {
	case s >= logevent.SeverityCritical:
		return "\033[1;35m" // bold magenta
	case s >= logevent.SeverityError:
		return "\033[1;31m" // bold red
	case s >= logevent.SeverityWarning:
		return "\033[1;33m" // bold yellow
	case s >= logevent.SeverityInfo:
		return "\033[0;36m" // cyan
	default:
		return "\033[0;37m" // white
	}
}

// StyleFromString parses a style name (plain, colored, json).
func StyleFromString(s string) (Style, error) {
	switch strings.ToLower(s) {
	case "plain":
		return StylePlain, nil
	case "colored", "colour":
		return StyleColored, nil
	case "json":
		return StyleJSON, nil
	}
	return StylePlain, fmt.Errorf("unknown formatter style %q", s)
}
