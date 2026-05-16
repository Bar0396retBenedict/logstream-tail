package source

import (
	"context"

	"github.com/user/logstream-tail/internal/logevent"
)

// Source represents a log source that can stream log events.
type Source interface {
	// Name returns the human-readable name of this source.
	Name() string

	// Start begins streaming log events into the provided channel.
	// It blocks until the context is cancelled or a fatal error occurs.
	Start(ctx context.Context, out chan<- logevent.Event) error

	// Close releases any resources held by the source.
	Close() error
}

// Config holds common configuration shared across log sources.
type Config struct {
	// Filter is an optional log level filter; events below this severity are dropped.
	Filter logevent.Severity

	// MaxBackoff is the maximum duration to wait between reconnection attempts.
	MaxBackoffSeconds int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Filter:            logevent.SeverityDebug,
		MaxBackoffSeconds: 30,
	}
}
