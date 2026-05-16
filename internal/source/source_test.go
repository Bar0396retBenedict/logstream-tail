package source_test

import (
	"testing"

	"github.com/user/logstream-tail/internal/logevent"
	"github.com/user/logstream-tail/internal/source"
)

func TestDefaultConfig(t *testing.T) {
	cfg := source.DefaultConfig()

	if cfg.Filter != logevent.SeverityDebug {
		t.Errorf("expected default filter SeverityDebug, got %v", cfg.Filter)
	}

	if cfg.MaxBackoffSeconds != 30 {
		t.Errorf("expected MaxBackoffSeconds=30, got %d", cfg.MaxBackoffSeconds)
	}
}

func TestDefaultConfig_FilterDropsLowerSeverity(t *testing.T) {
	cfg := source.DefaultConfig()

	// SeverityDebug is the lowest; nothing should be filtered by default.
	if cfg.Filter > logevent.SeverityDebug {
		t.Errorf("default config should not filter any events, but Filter=%v", cfg.Filter)
	}
}
