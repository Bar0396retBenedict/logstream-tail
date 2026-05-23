package source

import (
	"fmt"
	"time"
)

// ThrottleOption is a functional option for ThrottleConfig.
type ThrottleOption func(*ThrottleConfig)

// WithThrottleInterval sets the minimum interval between forwarded events
// for a given source. Panics if d is zero or negative.
func WithThrottleInterval(d time.Duration) ThrottleOption {
	if d <= 0 {
		panic(fmt.Sprintf("throttle: interval must be positive, got %v", d))
	}
	return func(c *ThrottleConfig) {
		c.Interval = d
	}
}

// NewThrottleConfig builds a ThrottleConfig starting from defaults and
// applying any supplied options.
func NewThrottleConfig(opts ...ThrottleOption) ThrottleConfig {
	cfg := DefaultThrottleConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
