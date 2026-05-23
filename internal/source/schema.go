package source

import (
	"fmt"
	"regexp"

	"github.com/user/logstream-tail/internal/logevent"
)

// SchemaRule defines a single field validation rule applied to log events.
type SchemaRule struct {
	Field   string
	Pattern *regexp.Regexp
	Required bool
}

// SchemaConfig holds configuration for the schema validator.
type SchemaConfig struct {
	Rules      []SchemaRule
	DropOnFail bool // if true, non-conforming events are dropped; otherwise passed through
}

// DefaultSchemaConfig returns a SchemaConfig with no rules and pass-through on failure.
func DefaultSchemaConfig() SchemaConfig {
	return SchemaConfig{
		Rules:      nil,
		DropOnFail: false,
	}
}

// NewSchemaValidator returns a middleware that validates events against the
// provided SchemaConfig. Events that fail validation are either dropped or
// passed through unchanged, depending on SchemaConfig.DropOnFail.
func NewSchemaValidator(in <-chan logevent.Event, cfg SchemaConfig) <-chan logevent.Event {
	out := make(chan logevent.Event, cap(in))
	go func() {
		defer close(out)
		for ev := range in {
			if validate(ev, cfg) {
				out <- ev
			} else if !cfg.DropOnFail {
				out <- ev
			}
		}
	}()
	return out
}

// validate checks an event against all rules in the config.
func validate(ev logevent.Event, cfg SchemaConfig) bool {
	for _, rule := range cfg.Rules {
		val := fieldValue(ev, rule.Field)
		if val == "" && rule.Required {
			return false
		}
		if rule.Pattern != nil && val != "" && !rule.Pattern.MatchString(val) {
			return false
		}
	}
	return true
}

// fieldValue extracts a named field from an event for validation.
func fieldValue(ev logevent.Event, field string) string {
	switch field {
	case "message":
		return ev.Message
	case "source":
		return string(ev.Source)
	case "severity":
		return fmt.Sprintf("%d", ev.Severity)
	default:
		return ""
	}
}
