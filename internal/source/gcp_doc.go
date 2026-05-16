// Package source provides streaming log source adapters for logstream-tail.
//
// # GCP Logging Source
//
// GCPSource polls the Google Cloud Logging API at a configurable interval and
// fans log entries into a logevent.Event channel. It mirrors the interface
// exposed by CloudWatchSource so both can be composed via FanIn.
//
// Basic usage:
//
//	client := gcplogging.NewRealClient(ctx, projectID)
//	src := source.NewGCPSource(client, projectID, logName, source.DefaultConfig)
//	ch, err := src.Stream(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for event := range ch {
//		fmt.Println(event)
//	}
//
// The GCPLoggingClient interface allows injecting a mock during tests without
// requiring live GCP credentials.
package source
