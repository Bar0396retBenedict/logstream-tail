// Package cli provides configuration loading and validation for the
// logstream-tail command-line interface.
package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// SourceType identifies which cloud logging backend to stream from.
type SourceType string

const (
	SourceCloudWatch SourceType = "cloudwatch"
	SourceGCP        SourceType = "gcp"
)

// Config holds all runtime options parsed from flags / environment.
type Config struct {
	Sources       []SourceType
	LogGroups     []string   // CloudWatch log group names
	GCPProject    string     // GCP project ID
	GCPFilter     string     // GCP advanced log filter
	MinSeverity   string     // minimum severity to display
	OutputStyle   string     // plain | json | color
	PollInterval  time.Duration
	HideSource    bool
}

// ParseFlags registers flags on fs and returns a Config populated from them.
// Call fs.Parse(os.Args[1:]) before using the returned Config.
func ParseFlags(fs *pflag.FlagSet) *Config {
	cfg := &Config{}

	var sources string
	fs.StringVarP(&sources, "source", "s", "cloudwatch",
		"comma-separated list of sources to stream (cloudwatch, gcp)")
	fs.StringArrayVar(&cfg.LogGroups, "log-group", nil,
		"CloudWatch log group name (repeatable)")
	fs.StringVar(&cfg.GCPProject, "gcp-project", "",
		"GCP project ID")
	fs.StringVar(&cfg.GCPFilter, "gcp-filter", "",
		"GCP advanced log filter expression")
	fs.StringVar(&cfg.MinSeverity, "min-severity", "INFO",
		"minimum severity level to display (DEBUG, INFO, WARN, ERROR)")
	fs.StringVar(&cfg.OutputStyle, "style", "color",
		"output style: plain, json, or color")
	fs.DurationVar(&cfg.PollInterval, "poll-interval", 5*time.Second,
		"how often to poll each source for new log entries")
	fs.BoolVar(&cfg.HideSource, "hide-source", false,
		"omit the source label from output lines")

	// Resolve sources after Parse via a post-parse hook stored in cfg.
	fs.ParseErrorsWhitelist.UnknownFlags = true
	_ = fs.Parse(nil) // no-op; caller calls Parse

	// Lazy evaluation: store raw string, resolve in Validate.
	cfg.Sources = nil
	_ = sources // resolved in Validate via fs lookup

	return cfg
}

// Validate checks that the Config is self-consistent and resolves any
// fields that depend on other flags.
func (c *Config) Validate(fs *pflag.FlagSet) error {
	raw, err := fs.GetString("source")
	if err != nil {
		return fmt.Errorf("reading --source flag: %w", err)
	}
	for _, part := range strings.Split(raw, ",") {
		st := SourceType(strings.TrimSpace(strings.ToLower(part)))
		switch st {
		case SourceCloudWatch, SourceGCP:
			c.Sources = append(c.Sources, st)
		default:
			return fmt.Errorf("unknown source %q; valid values: cloudwatch, gcp", part)
		}
	}
	if len(c.Sources) == 0 {
		return errors.New("at least one --source is required")
	}
	for _, st := range c.Sources {
		switch st {
		case SourceCloudWatch:
			if len(c.LogGroups) == 0 {
				return errors.New("--log-group is required when using the cloudwatch source")
			}
		case SourceGCP:
			if c.GCPProject == "" {
				return errors.New("--gcp-project is required when using the gcp source")
			}
		}
	}
	return nil
}
