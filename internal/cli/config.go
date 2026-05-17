package cli

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/user/logstream-tail/internal/formatter"
	"github.com/user/logstream-tail/internal/source"
)

// Config holds all runtime configuration parsed from CLI flags.
type Config struct {
	// CloudWatch options
	LogGroup  string
	LogStream string
	Region    string

	// GCP options
	GCPProject string
	GCPFilter  string

	// Common options
	PollInterval time.Duration
	MaxEventsPS  int
	Style        string
	HideSource   bool
	NoColor      bool

	// Retry options
	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
}

// ParseFlags parses os.Args using the supplied FlagSet and returns a Config.
func ParseFlags(fs *flag.FlagSet, args []string) (*Config, error) {
	cfg := &Config{}

	fs.StringVar(&cfg.LogGroup, "log-group", "", "CloudWatch log group name")
	fs.StringVar(&cfg.LogStream, "log-stream", "", "CloudWatch log stream prefix")
	fs.StringVar(&cfg.Region, "region", "us-east-1", "AWS region")

	fs.StringVar(&cfg.GCPProject, "gcp-project", "", "GCP project ID")
	fs.StringVar(&cfg.GCPFilter, "gcp-filter", "", "GCP logging filter expression")

	fs.DurationVar(&cfg.PollInterval, "poll-interval", source.DefaultConfig().PollInterval, "polling interval")
	fs.IntVar(&cfg.MaxEventsPS, "max-events-ps", 0, "max log events per second (0 = unlimited)")
	fs.StringVar(&cfg.Style, "style", "plain", "output style: plain|json|color")
	fs.BoolVar(&cfg.HideSource, "hide-source", false, "omit source label from output")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "disable ANSI colour even in color style")

	fs.IntVar(&cfg.RetryMaxAttempts, "retry-max-attempts", 0, "max reconnect attempts (0 = unlimited)")
	fs.DurationVar(&cfg.RetryBaseDelay, "retry-base-delay", 500*time.Millisecond, "initial retry back-off delay")
	fs.DurationVar(&cfg.RetryMaxDelay, "retry-max-delay", 30*time.Second, "maximum retry back-off delay")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate returns an error if the config is not self-consistent.
func (c *Config) Validate() error {
	hasCW := c.LogGroup != ""
	hasGCP := c.GCPProject != ""

	if !hasCW && !hasGCP {
		return errors.New("at least one source must be configured (--log-group or --gcp-project)")
	}
	if hasCW && c.Region == "" {
		return errors.New("--region is required when --log-group is set")
	}
	if _, err := formatter.StyleFromString(c.Style); err != nil {
		return fmt.Errorf("invalid --style: %w", err)
	}
	if c.RetryBaseDelay <= 0 {
		return errors.New("--retry-base-delay must be positive")
	}
	if c.RetryMaxDelay < c.RetryBaseDelay {
		return errors.New("--retry-max-delay must be >= --retry-base-delay")
	}
	return nil
}

// RetryConfig converts the CLI retry flags into a source.RetryConfig.
func (c *Config) RetryConfig() source.RetryConfig {
	return source.RetryConfig{
		MaxAttempts: c.RetryMaxAttempts,
		BaseDelay:   c.RetryBaseDelay,
		MaxDelay:    c.RetryMaxDelay,
	}
}
