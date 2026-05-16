package source

import (
	"context"
	"fmt"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// GCPLoggingClient is an interface for fetching GCP log entries, allowing
// easy substitution of a mock in tests.
type GCPLoggingClient interface {
	ListLogEntries(ctx context.Context, projectID, filter string, since time.Time) ([]GCPLogEntry, error)
}

// GCPLogEntry represents a raw entry returned by the GCP Logging API.
type GCPLogEntry struct {
	Timestamp time.Time
	Severity  string
	Payload   string
	LogName   string
}

// GCPSource streams log events from a GCP Logging log name.
type GCPSource struct {
	client    GCPLoggingClient
	projectID string
	logName   string
	cfg       Config
}

// NewGCPSource constructs a GCPSource with the given client and configuration.
func NewGCPSource(client GCPLoggingClient, projectID, logName string, cfg Config) *GCPSource {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = DefaultConfig.PollInterval
	}
	if cfg.MinSeverity == 0 {
		cfg.MinSeverity = DefaultConfig.MinSeverity
	}
	return &GCPSource{
		client:    client,
		projectID: projectID,
		logName:   logName,
		cfg:       cfg,
	}
}

// Stream begins polling GCP Logging and emits matching events to the returned
// channel. The channel is closed when ctx is cancelled.
func (g *GCPSource) Stream(ctx context.Context) (<-chan logevent.Event, error) {
	ch := make(chan logevent.Event, 64)
	go func() {
		defer close(ch)
		since := time.Now().Add(-g.cfg.LookbackWindow)
		ticker := time.NewTicker(g.cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case tick := <-ticker.C:
				entries, err := g.client.ListLogEntries(ctx, g.projectID, g.logName, since)
				if err != nil {
					continue
				}
				for _, e := range entries {
					sev := logevent.ParseSeverity(e.Severity)
					if sev < g.cfg.MinSeverity {
						continue
					}
					ch <- logevent.Event{
						Timestamp: e.Timestamp,
						Severity:  sev,
						Message:   e.Payload,
						Source:    fmt.Sprintf("gcp/%s/%s", g.projectID, g.logName),
					}
				}
				since = tick
			}
		}
	}()
	return ch, nil
}
