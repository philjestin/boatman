package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// generate orchestrates the full scaffold generation.
func generate(cfg *ScaffoldConfig) (*ScaffoldResult, error) {
	data := newTemplateData(cfg)

	// Create output directory.
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	var filesCreated []string

	// Render each template in the registry.
	for _, entry := range templateRegistry {
		if entry.optional && (entry.condition == nil || !entry.condition(data)) {
			continue
		}

		content, err := renderTemplate(entry.tmplName, data)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", entry.tmplName, err)
		}

		outPath := filepath.Join(cfg.OutputDir, entry.outName)
		if err := os.WriteFile(outPath, content, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", outPath, err)
		}

		filesCreated = append(filesCreated, entry.outName)
	}

	return &ScaffoldResult{
		OutputDir:    cfg.OutputDir,
		FilesCreated: filesCreated,
	}, nil
}
