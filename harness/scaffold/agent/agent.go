package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/philjestin/boatman-ecosystem/harness/scaffold"
)

// EnhanceConfig configures the Claude enhancement agent.
type EnhanceConfig struct {
	// ProjectDir is the directory containing the scaffolded project.
	ProjectDir string

	// Provider is the LLM provider the stubs target.
	Provider scaffold.LLMProvider

	// ProjectLang is the target project language.
	ProjectLang scaffold.ProjectLanguage

	// Model is the Claude model to use (default: claude-sonnet-4-20250514).
	Model string
}

// Enhance uses the Claude CLI to replace stub implementations with real
// LLM-calling code. It processes each role file that exists in the project
// directory, sends its content to Claude with a role-specific prompt, and
// writes the enhanced code back.
//
// Requires the claude CLI to be installed and available on PATH.
func Enhance(ctx context.Context, cfg EnhanceConfig) error {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.Provider == "" {
		cfg.Provider = scaffold.ProviderGeneric
	}
	if cfg.ProjectLang == "" {
		cfg.ProjectLang = scaffold.LangGeneric
	}

	// Verify claude CLI is available.
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found on PATH: install it from https://claude.ai/claude-code\n%w", err)
	}

	prompts := rolePrompts(&cfg)
	var enhanced []string

	for _, rp := range prompts {
		filePath := filepath.Join(cfg.ProjectDir, rp.filename)

		// Skip files that don't exist (e.g. planner.go when not generated).
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		stubCode, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", rp.filename, err)
		}

		prompt := rp.buildPrompt(&cfg, string(stubCode))

		result, err := runClaude(ctx, cfg.Model, prompt)
		if err != nil {
			return fmt.Errorf("enhance %s: %w", rp.filename, err)
		}

		// Clean up any markdown fences the model might include.
		result = stripCodeFences(result)

		if err := os.WriteFile(filePath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rp.filename, err)
		}

		enhanced = append(enhanced, rp.filename)
	}

	if len(enhanced) == 0 {
		return fmt.Errorf("no stub files found in %s", cfg.ProjectDir)
	}

	// Verify the project still compiles.
	if err := verifyBuild(ctx, cfg.ProjectDir); err != nil {
		return fmt.Errorf("enhanced project does not compile: %w", err)
	}

	fmt.Printf("Enhanced %d files: %s\n", len(enhanced), strings.Join(enhanced, ", "))
	return nil
}

// runClaude shells out to the claude CLI with the given prompt.
func runClaude(ctx context.Context, model, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "claude",
		"-p",
		"--output-format", "text",
		"--model", model,
		prompt,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// verifyBuild runs go build in the project directory.
func verifyBuild(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%w", string(output), err)
	}
	return nil
}

// stripCodeFences removes markdown code fences if the model wrapped its output.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```go") {
		s = strings.TrimPrefix(s, "```go")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
