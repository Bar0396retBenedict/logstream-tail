package source

import "regexp"

// SchemaOption is a functional option for SchemaConfig.
type SchemaOption func(*SchemaConfig)

// WithRequiredField adds a rule that the named field must be non-empty.
func WithRequiredField(field string) SchemaOption {
	return func(cfg *SchemaConfig) {
		cfg.Rules = append(cfg.Rules, SchemaRule{
			Field:    field,
			Required: true,
		})
	}
}

// WithFieldPattern adds a rule that the named field must match the given regex.
// Panics if the pattern is invalid.
func WithFieldPattern(field, pattern string) SchemaOption {
	re := regexp.MustCompile(pattern)
	return func(cfg *SchemaConfig) {
		cfg.Rules = append(cfg.Rules, SchemaRule{
			Field:   field,
			Pattern: re,
		})
	}
}

// WithDropOnFail configures the validator to drop events that fail validation.
func WithDropOnFail() SchemaOption {
	return func(cfg *SchemaConfig) {
		cfg.DropOnFail = true
	}
}

// NewSchemaConfig constructs a SchemaConfig from the default and applies options.
func NewSchemaConfig(opts ...SchemaOption) SchemaConfig {
	cfg := DefaultSchemaConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
