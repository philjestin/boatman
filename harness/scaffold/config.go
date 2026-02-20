// Package scaffold generates compilable Go projects that implement the
// harness/runner role interfaces (Developer, Reviewer, Tester, Planner).
package scaffold

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// LLMProvider identifies the LLM backend the generated stubs target.
type LLMProvider string

const (
	ProviderClaude  LLMProvider = "claude"
	ProviderOpenAI  LLMProvider = "openai"
	ProviderOllama  LLMProvider = "ollama"
	ProviderGeneric LLMProvider = "generic"
)

// ValidProviders returns all recognised provider values.
func ValidProviders() []LLMProvider {
	return []LLMProvider{ProviderClaude, ProviderOpenAI, ProviderOllama, ProviderGeneric}
}

// ProjectLanguage is the language of the project being developed by the agent.
// It affects the test runner hints in the generated Tester stub.
type ProjectLanguage string

const (
	LangGo         ProjectLanguage = "go"
	LangTypeScript ProjectLanguage = "typescript"
	LangPython     ProjectLanguage = "python"
	LangRuby       ProjectLanguage = "ruby"
	LangGeneric    ProjectLanguage = "generic"
)

// ValidLanguages returns all recognised language values.
func ValidLanguages() []ProjectLanguage {
	return []ProjectLanguage{LangGo, LangTypeScript, LangPython, LangRuby, LangGeneric}
}

// ScaffoldConfig describes what to generate.
type ScaffoldConfig struct {
	// ProjectName is the Go module path (e.g. "github.com/user/my-agent").
	ProjectName string

	// OutputDir is the directory where files will be written.
	OutputDir string

	// Provider selects the LLM backend for generated comments/hints.
	Provider LLMProvider

	// ProjectLang is the target project language (affects test runner hints).
	ProjectLang ProjectLanguage

	// IncludePlanner generates a Planner implementation stub.
	IncludePlanner bool

	// IncludeTester generates a Tester implementation stub.
	IncludeTester bool

	// IncludeCostTracking wires up cost.Tracker in the generated main.go.
	IncludeCostTracking bool

	// MaxIterations is the default max review iterations in generated code.
	MaxIterations int

	// ReviewCriteria is an optional description for reviewer focus.
	ReviewCriteria string
}

// modulePathRe matches a valid Go module path.
var modulePathRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$`)

// Validate checks the configuration and returns an error describing any problems.
func (c *ScaffoldConfig) Validate() error {
	var errs []string

	if c.ProjectName == "" {
		errs = append(errs, "project name is required")
	} else if !modulePathRe.MatchString(c.ProjectName) {
		errs = append(errs, fmt.Sprintf("invalid module path: %q", c.ProjectName))
	}

	if c.OutputDir == "" {
		errs = append(errs, "output directory is required")
	}

	if !isValidProvider(c.Provider) {
		errs = append(errs, fmt.Sprintf("unknown provider: %q (valid: %s)", c.Provider, joinProviders()))
	}

	if c.ProjectLang != "" && !isValidLang(c.ProjectLang) {
		errs = append(errs, fmt.Sprintf("unknown language: %q (valid: %s)", c.ProjectLang, joinLanguages()))
	}

	if c.MaxIterations < 0 {
		errs = append(errs, "max-iterations must be >= 0")
	}

	if len(errs) > 0 {
		return fmt.Errorf("scaffold config: %s", strings.Join(errs, "; "))
	}
	return nil
}

// applyDefaults fills in zero-valued fields with sensible defaults.
func (c *ScaffoldConfig) applyDefaults() {
	if c.Provider == "" {
		c.Provider = ProviderGeneric
	}
	if c.ProjectLang == "" {
		c.ProjectLang = LangGeneric
	}
	if c.MaxIterations == 0 {
		c.MaxIterations = 3
	}
	if c.OutputDir == "" && c.ProjectName != "" {
		// Use the last segment of the module path.
		c.OutputDir = filepath.Base(c.ProjectName)
	}
}

func isValidProvider(p LLMProvider) bool {
	for _, v := range ValidProviders() {
		if v == p {
			return true
		}
	}
	return false
}

func isValidLang(l ProjectLanguage) bool {
	for _, v := range ValidLanguages() {
		if v == l {
			return true
		}
	}
	return false
}

func joinProviders() string {
	pp := ValidProviders()
	ss := make([]string, len(pp))
	for i, p := range pp {
		ss[i] = string(p)
	}
	return strings.Join(ss, ", ")
}

func joinLanguages() string {
	ll := ValidLanguages()
	ss := make([]string, len(ll))
	for i, l := range ll {
		ss[i] = string(l)
	}
	return strings.Join(ss, ", ")
}
