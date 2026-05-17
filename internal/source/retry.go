package source

import (
	"context"
	"math"
	"time"
)

// RetryConfig controls exponential back-off behaviour for source reconnects.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retries (0 = unlimited).
	MaxAttempts int
	// BaseDelay is the initial back-off duration.
	BaseDelay time.Duration
	// MaxDelay caps the back-off duration.
	MaxDelay time.Duration
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 0,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    30 * time.Second,
	}
}

// Retryer executes fn with exponential back-off until fn returns nil,
// the context is cancelled, or MaxAttempts is exhausted.
type Retryer struct {
	cfg RetryConfig
}

// NewRetryer creates a Retryer with the given config.
func NewRetryer(cfg RetryConfig) *Retryer {
	return &Retryer{cfg: cfg}
}

// Do calls fn repeatedly, sleeping between failures.
// It returns the first nil error from fn, the context error, or the last
// fn error when MaxAttempts is reached.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; ; attempt++ {
		if r.cfg.MaxAttempts > 0 && attempt >= r.cfg.MaxAttempts {
			return lastErr
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		delay := r.backoff(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (r *Retryer) backoff(attempt int) time.Duration {
	delay := float64(r.cfg.BaseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(r.cfg.MaxDelay) {
		delay = float64(r.cfg.MaxDelay)
	}
	return time.Duration(delay)
}
