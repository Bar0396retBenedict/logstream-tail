// Package source — Throttler
//
// The Throttler pipeline stage limits how frequently events from a single
// source are forwarded downstream. When a burst of events arrives from the
// same source within the configured Interval, only the first event in each
// window is passed through; subsequent events are silently dropped until the
// interval has elapsed.
//
// This is useful when a noisy log group emits hundreds of identical (or
// near-identical) lines per second and you only need to see one representative
// event rather than every repetition.
//
// Usage:
//
//	cfg := source.NewThrottleConfig(
//		source.WithThrottleInterval(200 * time.Millisecond),
//	)
//	throttled := source.NewThrottlerContext(ctx, cfg, upstream)
//
// The Throttler is keyed on logevent.Source so that independent sources are
// throttled independently — a busy CloudWatch stream will not suppress events
// arriving from a quiet GCP stream.
package source
