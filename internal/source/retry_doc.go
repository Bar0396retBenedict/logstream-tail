// Package source provides log-source adapters and supporting utilities for
// logstream-tail.
//
// # Retry
//
// The [Retryer] type wraps any fallible operation with configurable
// exponential back-off.  It is used internally by [NewCloudWatchSource] and
// [NewGCPSource] to reconnect after transient API errors without flooding the
// upstream service.
//
// Basic usage:
//
//	r := source.NewRetryer(source.DefaultRetryConfig())
//	err := r.Do(ctx, func() error {
//	    return doSomethingFallible()
//	})
//
// [RetryConfig.MaxAttempts] of 0 means unlimited retries; the loop will only
// stop when the context is cancelled or the operation succeeds.
package source
