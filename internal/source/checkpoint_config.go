package source

// CheckpointConfig controls where checkpoint state is persisted.
type CheckpointConfig struct {
	// FilePath is the path on disk where cursor state is written.
	// If empty, checkpointing is disabled and each run starts from the
	// beginning of the configured look-back window.
	FilePath string
}

// DefaultCheckpointConfig returns a CheckpointConfig with checkpointing
// disabled. Callers that want persistence must supply a non-empty FilePath.
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{}
}

// CheckpointOption is a functional option for CheckpointConfig.
type CheckpointOption func(*CheckpointConfig)

// WithCheckpointFile sets the file path used for cursor persistence.
func WithCheckpointFile(path string) CheckpointOption {
	return func(cfg *CheckpointConfig) {
		cfg.FilePath = path
	}
}

// NewCheckpointConfig builds a CheckpointConfig from the supplied options.
func NewCheckpointConfig(opts ...CheckpointOption) CheckpointConfig {
	cfg := DefaultCheckpointConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
