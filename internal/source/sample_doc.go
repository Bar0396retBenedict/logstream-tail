// Package source — Sampler
//
// Sampler is a pipeline stage that reduces event volume by forwarding only
// every Nth event and silently discarding the rest.
//
// # When to use
//
// Use Sampler when a log stream produces far more events than the terminal
// operator can meaningfully read. A sampling rate of 10 reduces throughput
// to 10 % of the original while preserving the overall shape of the stream.
//
// # Configuration
//
//	cfg := source.NewSampleConfig(source.WithSampleRate(10))
//	sampled := source.NewSampler(ctx, upstream, cfg)
//
// A Rate of 1 (the default) is a no-op — every event is forwarded.
package source
