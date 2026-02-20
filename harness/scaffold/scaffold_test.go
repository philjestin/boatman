package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_RequiredFields(t *testing.T) {
	cfg := ScaffoldConfig{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty config")
	}
	if !strings.Contains(err.Error(), "project name is required") {
		t.Errorf("expected project name error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "output directory is required") {
		t.Errorf("expected output directory error, got: %v", err)
	}
}

func TestValidate_InvalidModulePath(t *testing.T) {
	cfg := ScaffoldConfig{
		ProjectName: "invalid path with spaces",
		OutputDir:   "/tmp/test",
		Provider:    ProviderGeneric,
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid module path")
	}
	if !strings.Contains(err.Error(), "invalid module path") {
		t.Errorf("expected module path error, got: %v", err)
	}
}

func TestValidate_InvalidProvider(t *testing.T) {
	cfg := ScaffoldConfig{
		ProjectName: "github.com/user/agent",
		OutputDir:   "/tmp/test",
		Provider:    "nonexistent",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid provider")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("expected provider error, got: %v", err)
	}
}

func TestValidate_InvalidLanguage(t *testing.T) {
	cfg := ScaffoldConfig{
		ProjectName: "github.com/user/agent",
		OutputDir:   "/tmp/test",
		Provider:    ProviderGeneric,
		ProjectLang: "cobol",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid language")
	}
	if !strings.Contains(err.Error(), "unknown language") {
		t.Errorf("expected language error, got: %v", err)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := ScaffoldConfig{
		ProjectName: "github.com/user/my-agent",
		OutputDir:   "/tmp/test",
		Provider:    ProviderClaude,
		ProjectLang: LangGo,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := ScaffoldConfig{
		ProjectName: "github.com/user/my-agent",
	}
	cfg.applyDefaults()

	if cfg.Provider != ProviderGeneric {
		t.Errorf("expected default provider %q, got %q", ProviderGeneric, cfg.Provider)
	}
	if cfg.ProjectLang != LangGeneric {
		t.Errorf("expected default language %q, got %q", LangGeneric, cfg.ProjectLang)
	}
	if cfg.MaxIterations != 3 {
		t.Errorf("expected default max iterations 3, got %d", cfg.MaxIterations)
	}
	if cfg.OutputDir != "my-agent" {
		t.Errorf("expected default output dir %q, got %q", "my-agent", cfg.OutputDir)
	}
}

func TestGenerate_MinimalProject(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "myagent")

	result, err := Generate(ScaffoldConfig{
		ProjectName: "github.com/test/myagent",
		OutputDir:   outDir,
		Provider:    ProviderGeneric,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.OutputDir != outDir {
		t.Errorf("expected output dir %q, got %q", outDir, result.OutputDir)
	}

	// Should have: go.mod, main.go, developer.go, reviewer.go, config.go
	expectedFiles := []string{"go.mod", "main.go", "developer.go", "reviewer.go", "config.go"}
	if len(result.FilesCreated) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d: %v", len(expectedFiles), len(result.FilesCreated), result.FilesCreated)
	}

	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// planner.go and tester.go should NOT exist
	for _, f := range []string{"planner.go", "tester.go"} {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("unexpected file %s exists", f)
		}
	}
}

func TestGenerate_WithOptionalRoles(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "full")

	result, err := Generate(ScaffoldConfig{
		ProjectName:    "github.com/test/full",
		OutputDir:      outDir,
		Provider:       ProviderClaude,
		IncludePlanner: true,
		IncludeTester:  true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		"go.mod", "main.go", "developer.go", "reviewer.go", "config.go",
		"planner.go", "tester.go",
	}
	if len(result.FilesCreated) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d: %v", len(expectedFiles), len(result.FilesCreated), result.FilesCreated)
	}

	for _, f := range expectedFiles {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestGenerate_AllProviders(t *testing.T) {
	for _, provider := range ValidProviders() {
		t.Run(string(provider), func(t *testing.T) {
			dir := t.TempDir()
			outDir := filepath.Join(dir, "proj")

			_, err := Generate(ScaffoldConfig{
				ProjectName: "github.com/test/proj",
				OutputDir:   outDir,
				Provider:    provider,
			})
			if err != nil {
				t.Fatalf("Generate failed for provider %s: %v", provider, err)
			}

			// Verify developer.go contains provider-specific hints
			devContent, err := os.ReadFile(filepath.Join(outDir, "developer.go"))
			if err != nil {
				t.Fatalf("read developer.go: %v", err)
			}

			meta := getProviderMeta(provider)
			if !strings.Contains(string(devContent), meta.Name) {
				t.Errorf("developer.go should mention provider name %q", meta.Name)
			}
		})
	}
}

func TestGenerate_GoModContent(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "proj")

	_, err := Generate(ScaffoldConfig{
		ProjectName: "github.com/user/my-agent",
		OutputDir:   outDir,
		Provider:    ProviderGeneric,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module github.com/user/my-agent") {
		t.Error("go.mod should contain correct module path")
	}
	if !strings.Contains(string(content), "github.com/philjestin/boatman-ecosystem/harness") {
		t.Error("go.mod should depend on harness module")
	}
}

func TestGenerate_CostTracking(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "proj")

	_, err := Generate(ScaffoldConfig{
		ProjectName:         "github.com/test/proj",
		OutputDir:           outDir,
		Provider:            ProviderGeneric,
		IncludeCostTracking: true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if !strings.Contains(string(content), "cost.NewTracker") {
		t.Error("main.go should contain cost tracking setup")
	}
}

func TestGenerate_LanguageTestHints(t *testing.T) {
	tests := []struct {
		lang     ProjectLanguage
		expected string
	}{
		{LangGo, "go test"},
		{LangTypeScript, "npm test"},
		{LangPython, "pytest"},
		{LangRuby, "rspec"},
		{LangGeneric, "make test"},
	}

	for _, tc := range tests {
		t.Run(string(tc.lang), func(t *testing.T) {
			dir := t.TempDir()
			outDir := filepath.Join(dir, "proj")

			_, err := Generate(ScaffoldConfig{
				ProjectName:   "github.com/test/proj",
				OutputDir:     outDir,
				Provider:      ProviderGeneric,
				ProjectLang:   tc.lang,
				IncludeTester: true,
			})
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			content, err := os.ReadFile(filepath.Join(outDir, "tester.go"))
			if err != nil {
				t.Fatalf("read tester.go: %v", err)
			}

			if !strings.Contains(string(content), tc.expected) {
				t.Errorf("tester.go should contain %q for language %s", tc.expected, tc.lang)
			}
		})
	}
}

func TestGenerate_MainIncludesPlanner(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "proj")

	_, err := Generate(ScaffoldConfig{
		ProjectName:    "github.com/test/proj",
		OutputDir:      outDir,
		Provider:       ProviderGeneric,
		IncludePlanner: true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if !strings.Contains(string(content), "WithPlanner") {
		t.Error("main.go should wire planner when IncludePlanner is true")
	}
}

func TestGenerate_MainExcludesOptionals(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "proj")

	_, err := Generate(ScaffoldConfig{
		ProjectName: "github.com/test/proj",
		OutputDir:   outDir,
		Provider:    ProviderGeneric,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if strings.Contains(string(content), "WithPlanner") {
		t.Error("main.go should NOT wire planner when IncludePlanner is false")
	}
	if strings.Contains(string(content), "WithTester") {
		t.Error("main.go should NOT wire tester when IncludeTester is false")
	}
	if strings.Contains(string(content), "cost.NewTracker") {
		t.Error("main.go should NOT include cost tracking when IncludeCostTracking is false")
	}
}
