// Package source provides streaming log source implementations for
// logstream-tail.
//
// # CloudWatch source
//
// CloudWatchSource polls an AWS CloudWatch Logs log group at a configurable
// interval and emits [logevent.Event] values onto a shared channel that is
// consumed by a [FanIn] aggregator.
//
// Usage:
//
//	cfg := source.CloudWatchConfig{
//		Config: source.Config{
//			PollInterval: 5 * time.Second,
//			Filter: func(e logevent.Event) bool {
//				return e.Severity >= logevent.SeverityWarn
//			},
//		},
//		Region:   "us-east-1",
//		LogGroup: "/my-service/prod",
//	}
//
//	client := newRealAWSClient(cfg.Region) // your AWS SDK wrapper
//	src := source.NewCloudWatchSource(cfg, client)
//
// The real AWS SDK adapter must implement [cloudWatchLogsClient]; because that
// interface is unexported the adapter lives in the same package (or a
// dedicated adapter sub-package) and is wired up in main.
package source
