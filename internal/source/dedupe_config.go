package source

import "time"

// DedupeConfig controls the behaviour of the Deduplicator.
type DedupeConfig struct {
	// Window is the duration during which an identical event is considered a
	// duplicate and will be suppressed. Defaults to 5 seconds.
	Window time.Duration

	// MaxTracked is the maximum number of distinct event fingerprints kept in
	// memory at once. Oldest entries are evicted when the limit is reached.
	// Defaults to 1024.
	MaxTracked int
}

// DefaultDedupeConfig returns a DedupeConfig populated with sensible defaults.
func DefaultDedupeConfig() DedupeConfig {
	return DedupeConfig{
		Window:     5 * time.Second,
		MaxTracked: 1024,
	}
}
