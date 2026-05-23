package source

import (
	"testing"
)

func TestNewSchemaConfig_Defaults(t *testing.T) {
	cfg := NewSchemaConfig()
	if cfg.DropOnFail {
		t.Error("expected DropOnFail=false by default")
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(cfg.Rules))
	}
}

func TestNewSchemaConfig_WithRequiredField(t *testing.T) {
	cfg := NewSchemaConfig(WithRequiredField("message"))
	if len(cfg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Field != "message" {
		t.Errorf("expected field 'message', got %q", cfg.Rules[0].Field)
	}
	if !cfg.Rules[0].Required {
		t.Error("expected Required=true")
	}
}

func TestNewSchemaConfig_WithFieldPattern(t *testing.T) {
	cfg := NewSchemaConfig(WithFieldPattern("source", `^gcp$`))
	if len(cfg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.Rules))
	}
	if cfg.Rules[0].Pattern == nil {
		t.Error("expected non-nil pattern")
	}
}

func TestNewSchemaConfig_MultipleOptions(t *testing.T) {
	cfg := NewSchemaConfig(
		WithRequiredField("message"),
		WithFieldPattern("source", `^cloudwatch$`),
		WithDropOnFail(),
	)
	if len(cfg.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
	}
	if !cfg.DropOnFail {
		t.Error("expected DropOnFail=true")
	}
}
