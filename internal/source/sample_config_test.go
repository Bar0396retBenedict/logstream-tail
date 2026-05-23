package source_test

import (
	"testing"

	"github.com/your-org/logstream-tail/internal/source"
)

func TestWithSampleRate_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for rate 0")
		}
	}()
	source.WithSampleRate(0)
}

func TestNewSampleConfig_Defaults(t *testing.T) {
	cfg := source.NewSampleConfig()
	if cfg.Rate != 1 {
		t.Fatalf("expected default rate 1, got %d", cfg.Rate)
	}
}

func TestNewSampleConfig_MultipleOptions(t *testing.T) {
	// last write wins — both options apply in order
	cfg := source.NewSampleConfig(
		source.WithSampleRate(3),
		source.WithSampleRate(9),
	)
	if cfg.Rate != 9 {
		t.Fatalf("expected rate 9, got %d", cfg.Rate)
	}
}
