// Package source — schema validator middleware.
//
// # Schema Validation
//
// NewSchemaValidator wraps an event channel and applies field-level validation
// rules defined in a SchemaConfig. Rules can require that a field is non-empty
// or that its value matches a regular expression.
//
// Events that fail validation are either silently dropped (DropOnFail: true) or
// forwarded unchanged (DropOnFail: false, the default). The pass-through default
// lets operators observe malformed events during development before enforcing
// strict schemas in production.
//
// # Example
//
//	cfg := source.NewSchemaConfig(
//		source.WithRequiredField("message"),
//		source.WithFieldPattern("source", `^(cloudwatch|gcp)$`),
//		source.WithDropOnFail(),
//	)
//	validated := source.NewSchemaValidator(raw, cfg)
package source
