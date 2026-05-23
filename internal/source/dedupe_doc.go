// Package source provides log source implementations for logstream-tail.
//
// # Deduplicator
//
// NewDeduplicator wraps any upstream channel and suppresses duplicate log
// events within a configurable time window. Two events are considered
// duplicates when they share the same source, severity, and message text.
//
// Example usage:
//
//	cfg := source.DefaultDedupeConfig()
//	cfg.Window = 10 * time.Second
//
//	out := source.NewDeduplicator(ctx, upstream, cfg)
//	for ev := range out {
//	    // ev is deduplicated
//	}
//
// The deduplicator uses a fixed-size LRU-style ring to bound memory usage
// regardless of event volume.
package source
