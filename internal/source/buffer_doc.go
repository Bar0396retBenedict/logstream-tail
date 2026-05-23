// Package source provides log-source adapters and pipeline utilities for
// logstream-tail.
//
// # Buffer
//
// Buffer sits between a noisy upstream channel and downstream consumers.
// It accumulates incoming [logevent.Event] values and forwards them in two
// situations:
//
//  1. The internal batch reaches the configured Size threshold.
//  2. The FlushInterval ticker fires.
//
// This smooths out micro-bursts from CloudWatch or GCP Logging without
// introducing visible latency under normal load. Typical usage:
//
//	cfg := source.DefaultBufferConfig()
//	buf, out := source.NewBuffer(upstream, cfg)
//	go buf.Run(ctx)
//	// consume from out …
//
// The output channel is closed automatically when Run returns, making it
// safe to range over.
package source
