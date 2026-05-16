package cli

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func newFS() (*pflag.FlagSet, *Config) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	cfg := ParseFlags(fs)
	return fs, cfg
}

func TestParseFlags_Defaults(t *testing.T) {
	fs, cfg := newFS()
	if err := fs.Parse([]string{"--log-group", "/app/prod"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if cfg.MinSeverity != "INFO" {
		t.Errorf("expected default MinSeverity INFO, got %q", cfg.MinSeverity)
	}
	if cfg.OutputStyle != "color" {
		t.Errorf("expected default OutputStyle color, got %q", cfg.OutputStyle)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected default PollInterval 5s, got %v", cfg.PollInterval)
	}
	if cfg.HideSource {
		t.Error("expected HideSource to default to false")
	}
}

func TestValidate_CloudWatchHappyPath(t *testing.T) {
	fs, cfg := newFS()
	if err := fs.Parse([]string{"--source", "cloudwatch", "--log-group", "/app/prod"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := cfg.Validate(fs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0] != SourceCloudWatch {
		t.Errorf("expected [cloudwatch], got %v", cfg.Sources)
	}
}

func TestValidate_GCPHappyPath(t *testing.T) {
	fs, cfg := newFS()
	if err := fs.Parse([]string{"--source", "gcp", "--gcp-project", "my-project"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := cfg.Validate(fs); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidate_MissingLogGroup(t *testing.T) {
	fs, cfg := newFS()
	_ = fs.Parse([]string{"--source", "cloudwatch"})
	if err := cfg.Validate(fs); err == nil {
		t.Error("expected error for missing --log-group")
	}
}

func TestValidate_MissingGCPProject(t *testing.T) {
	fs, cfg := newFS()
	_ = fs.Parse([]string{"--source", "gcp"})
	if err := cfg.Validate(fs); err == nil {
		t.Error("expected error for missing --gcp-project")
	}
}

func TestValidate_UnknownSource(t *testing.T) {
	fs, cfg := newFS()
	_ = fs.Parse([]string{"--source", "splunk"})
	if err := cfg.Validate(fs); err == nil {
		t.Error("expected error for unknown source")
	}
}

func TestValidate_MultiSource(t *testing.T) {
	fs, cfg := newFS()
	_ = fs.Parse([]string{
		"--source", "cloudwatch,gcp",
		"--log-group", "/app/prod",
		"--gcp-project", "my-project",
	})
	if err := cfg.Validate(fs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(cfg.Sources))
	}
}
