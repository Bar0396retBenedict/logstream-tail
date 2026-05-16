package source

import (
	"context"
	"fmt"
	"time"

	"github.com/user/logstream-tail/internal/logevent"
)

// CloudWatchConfig holds configuration for a CloudWatch log source.
type CloudWatchConfig struct {
	Config
	Region    string
	LogGroup  string
	LogStream string // optional, empty means all streams
}

// CloudWatchSource polls CloudWatch Logs and emits events onto a channel.
type CloudWatchSource struct {
	cfg    CloudWatchConfig
	client cloudWatchLogsClient
}

// cloudWatchLogsClient abstracts the AWS SDK call so we can inject a fake in tests.
type cloudWatchLogsClient interface {
	FilterLogEvents(ctx context.Context, group, stream string, startTime time.Time) ([]rawCWEvent, error)
}

type rawCWEvent struct {
	Timestamp int64
	Message   string
	Stream    string
}

// NewCloudWatchSource constructs a CloudWatchSource with the given config and client.
func NewCloudWatchSource(cfg CloudWatchConfig, client cloudWatchLogsClient) *CloudWatchSource {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = DefaultConfig.PollInterval
	}
	return &CloudWatchSource{cfg: cfg, client: client}
}

// Stream implements the Source interface. It polls CloudWatch at PollInterval
// and writes matching log events to out until ctx is cancelled.
func (s *CloudWatchSource) Stream(ctx context.Context, out chan<- logevent.Event) error {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	cursor := time.Now().Add(-s.cfg.PollInterval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tick := <-ticker.C:
			events, err := s.client.FilterLogEvents(ctx, s.cfg.LogGroup, s.cfg.LogStream, cursor)
			if err != nil {
				// non-fatal: log and retry next tick
				_ = fmt.Errorf("cloudwatch poll error: %w", err)
				continue
			}
			for _, raw := range events {
				ev := logevent.Event{
					Timestamp: time.UnixMilli(raw.Timestamp),
					Message:   raw.Message,
					Source:    logevent.SourceCloudWatch,
					Severity:  logevent.SeverityInfo,
					Labels: map[string]string{
						"log_group":  s.cfg.LogGroup,
						"log_stream": raw.Stream,
						"region":     s.cfg.Region,
					},
				}
				if s.cfg.Filter == nil || s.cfg.Filter(ev) {
					select {
					case out <- ev:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
			cursor = tick
		}
	}
}
