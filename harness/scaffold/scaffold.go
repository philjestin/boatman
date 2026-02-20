package scaffold

// ScaffoldResult describes the output of a successful scaffold generation.
type ScaffoldResult struct {
	// OutputDir is the directory where files were written.
	OutputDir string

	// FilesCreated lists the filenames that were generated (relative to OutputDir).
	FilesCreated []string
}

// Generate creates a new Go project with stub implementations of the harness
// runner role interfaces. The generated project compiles and runs immediately;
// stubs return placeholder values with provider-specific guidance in comments.
func Generate(cfg ScaffoldConfig) (*ScaffoldResult, error) {
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return generate(&cfg)
}
