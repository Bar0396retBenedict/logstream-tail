package cli

import (
	"flag"
	"testing"
	"time"
)

func newFS() *flag.FlagSet {
	return flag.NewFlagSet("test", flag.ContinueOnError)
}

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := ParseFlags(newFS(), []string{"--log-group", "my-group"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("expected default region us-east-1, got %s", cfg.Region)
	}
	if cfg.Style != "plain" {
		t.Errorf("expected default style plain, got %s", cfg.Style)
	}
	if cfg.RetryBaseDelay != 500*time.Millisecond {
		t.Errorf("unexpected RetryBaseDelay: %v", cfg.RetryBaseDelay)
	}
	if cfg.RetryMaxDelay != 30*time.Second {
		t.Errorf("unexpected RetryMaxDelay: %v", cfg.RetryMaxDelay)
	}
}

func TestValidate_CloudWatchHappyPath(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{"--log-group", "grp", "--region", "eu-west-1"})
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_GCPHappyPath(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{"--gcp-project", "my-project"})
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_MissingLogGroup(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{})
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing sources")
	}
}

func TestValidate_InvalidStyle(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{"--log-group", "g", "--style", "neon"})
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid style")
	}
}

func TestValidate_RetryDelayValidation(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{
		"--log-group", "g",
		"--retry-base-delay", "10s",
		"--retry-max-delay", "1s",
	})
	if err := cfg.Validate(); err == nil {
		t.Error("expected error when max-delay < base-delay")
	}
}

func TestConfig_RetryConfig(t *testing.T) {
	cfg, _ := ParseFlags(newFS(), []string{
		"--log-group", "g",
		"--retry-max-attempts", "5",
		"--retry-base-delay", "200ms",
		"--retry-max-delay", "60s",
	})
	rc := cfg.RetryConfig()
	if rc.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts 5, got %d", rc.MaxAttempts)
	}
	if rc.BaseDelay != 200*time.Millisecond {
		t.Errorf("unexpected BaseDelay: %v", rc.BaseDelay)
	}
	if rc.MaxDelay != 60*time.Second {
		t.Errorf("unexpected MaxDelay: %v", rc.MaxDelay)
	}
}
