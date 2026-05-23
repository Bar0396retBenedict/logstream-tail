package source

import (
	"testing"
	"time"
)

func TestDefaultDedupeConfig(t *testing.T) {
	cfg := DefaultDedupeConfig()

	if cfg.Window != 5*time.Second {
		t.Errorf("expected Window=5s, got %v", cfg.Window)
	}
	if cfg.MaxTracked != 1024 {
		t.Errorf("expected MaxTracked=1024, got %d", cfg.MaxTracked)
	}
}

func TestDedupeConfig_CustomValues(t *testing.T) {
	cfg := DedupeConfig{
		Window:     30 * time.Second,
		MaxTracked: 512,
	}
	if cfg.Window != 30*time.Second {
		t.Errorf("unexpected Window: %v", cfg.Window)
	}
	if cfg.MaxTracked != 512 {
		t.Errorf("unexpected MaxTracked: %d", cfg.MaxTracked)
	}
}
