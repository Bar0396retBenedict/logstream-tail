package source

// SampleConfigOption is a functional option for SampleConfig.
type SampleConfigOption func(*SampleConfig)

// WithSampleRate sets the 1-in-N sampling rate.
// A rate of 1 keeps every event; a rate of 10 keeps every 10th event.
// Panics if rate is zero.
func WithSampleRate(n uint64) SampleConfigOption {
	if n == 0 {
		panic("sample rate must be >= 1")
	}
	return func(c *SampleConfig) {
		c.Rate = n
	}
}

// NewSampleConfig builds a SampleConfig from the default values, then applies
// any supplied options.
func NewSampleConfig(opts ...SampleConfigOption) SampleConfig {
	cfg := DefaultSampleConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
