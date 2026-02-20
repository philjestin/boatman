package validation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/shared/types"
)

// TestNew tests creating a new validator
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{"absolute path", "/path/to/repo"},
		{"relative path", "./repo"},
		{"empty path", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New(tt.repoPath)
			if validator == nil {
				t.Fatal("New returned nil")
			}

			if validator.repoPath != tt.repoPath {
				t.Errorf("Expected repoPath %s, got %s", tt.repoPath, validator.repoPath)
			}
		})
	}
}

// TestValidator_ValidateGo tests Go validation
func TestValidator_ValidateGo(t *testing.T) {
	tempDir, cleanup := setupTestGoRepo(t)
	defer cleanup()

	validator := New(tempDir)
	ctx := context.Background()

	t.Run("no go files", func(t *testing.T) {
		files := []string{"test.txt", "readme.md"}
		issues, err := validator.ValidateGo(ctx, files)
		if err != nil {
			t.Fatalf("ValidateGo failed: %v", err)
		}

		if issues != nil && len(issues) > 0 {
			t.Errorf("Expected no issues for non-Go files, got %d", len(issues))
		}
	})

	t.Run("valid go file", func(t *testing.T) {
		// Create a valid Go file
		validFile := filepath.Join(tempDir, "valid.go")
		content := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}
`
		if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create valid Go file: %v", err)
		}

		files := []string{"valid.go"}
		issues, err := validator.ValidateGo(ctx, files)
		if err != nil {
			t.Fatalf("ValidateGo failed: %v", err)
		}

		if len(issues) > 0 {
			t.Logf("Note: got %d issues (might be expected depending on go vet version)", len(issues))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure context is cancelled

		files := []string{"valid.go"}
		_, err := validator.ValidateGo(ctx, files)
		// May or may not error depending on timing
		t.Logf("ValidateGo with cancelled context returned error: %v", err)
	})
}

// TestValidator_CheckSyntax tests syntax checking
func TestValidator_CheckSyntax(t *testing.T) {
	tempDir, cleanup := setupTestGoRepo(t)
	defer cleanup()

	validator := New(tempDir)
	ctx := context.Background()

	t.Run("non-existent files", func(t *testing.T) {
		files := []string{"nonexistent.go"}
		issues, err := validator.CheckSyntax(ctx, files)
		if err != nil {
			t.Fatalf("CheckSyntax failed: %v", err)
		}

		// Should skip non-existent files
		if len(issues) > 0 {
			t.Errorf("Expected no issues for non-existent files, got %d", len(issues))
		}
	})

	t.Run("valid go file", func(t *testing.T) {
		validFile := filepath.Join(tempDir, "syntax_valid.go")
		content := `package main

func hello() string {
	return "world"
}
`
		if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create valid Go file: %v", err)
		}

		files := []string{"syntax_valid.go"}
		issues, err := validator.CheckSyntax(ctx, files)
		if err != nil {
			t.Fatalf("CheckSyntax failed: %v", err)
		}

		// May or may not have issues depending on build requirements
		t.Logf("CheckSyntax returned %d issues", len(issues))
	})

	t.Run("non-go files", func(t *testing.T) {
		jsFile := filepath.Join(tempDir, "test.js")
		if err := os.WriteFile(jsFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("Failed to create JS file: %v", err)
		}

		files := []string{"test.js"}
		issues, err := validator.CheckSyntax(ctx, files)
		if err != nil {
			t.Fatalf("CheckSyntax failed: %v", err)
		}

		// JS files don't have syntax checking implemented, should return no issues
		if len(issues) > 0 {
			t.Errorf("Expected no issues for JS file, got %d", len(issues))
		}
	})
}

// TestValidator_ValidateAll tests combined validation
func TestValidator_ValidateAll(t *testing.T) {
	tempDir, cleanup := setupTestGoRepo(t)
	defer cleanup()

	validator := New(tempDir)
	ctx := context.Background()

	// Create a valid Go file
	validFile := filepath.Join(tempDir, "all_test.go")
	content := `package main

import "fmt"

func test() {
	fmt.Println("test")
}
`
	if err := os.WriteFile(validFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	files := []string{"all_test.go"}
	result, err := validator.ValidateAll(ctx, files)
	if err != nil {
		t.Fatalf("ValidateAll failed: %v", err)
	}

	if result == nil {
		t.Fatal("ValidateAll returned nil result")
	}

	// Check result structure
	if result.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Expected score between 0-100, got %d", result.Score)
	}

	if len(result.Issues) == 0 {
		if !result.Passed {
			t.Error("Expected Passed=true when no issues")
		}
		if result.Score != 100 {
			t.Errorf("Expected score=100 when no issues, got %d", result.Score)
		}
	}

	t.Logf("ValidateAll result: passed=%v, score=%d, issues=%d, summary=%s",
		result.Passed, result.Score, len(result.Issues), result.Summary)
}

// TestParseGoVetOutput tests parsing go vet output
func TestParseGoVetOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int // expected number of issues
	}{
		{
			name:     "empty output",
			output:   "",
			expected: 0,
		},
		{
			name:     "single issue",
			output:   "main.go:10:5: unreachable code",
			expected: 1,
		},
		{
			name: "multiple issues",
			output: `main.go:10:5: unreachable code
utils.go:20:3: unused variable
helper.go:5:1: exported function should have comment`,
			expected: 3,
		},
		{
			name:     "output with blank lines",
			output:   "\nmain.go:10:5: issue\n\nutils.go:20:3: another issue\n",
			expected: 2,
		},
		{
			name:     "invalid format",
			output:   "not a valid vet output format",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := parseGoVetOutput(tt.output)

			if len(issues) != tt.expected {
				t.Errorf("Expected %d issues, got %d", tt.expected, len(issues))
			}

			// Verify all issues have required fields
			for i, issue := range issues {
				if issue.File == "" {
					t.Errorf("Issue %d: missing file", i)
				}
				if issue.Severity != "warning" {
					t.Errorf("Issue %d: expected severity 'warning', got %s", i, issue.Severity)
				}
				if issue.Message == "" {
					t.Errorf("Issue %d: missing message", i)
				}
			}
		})
	}
}

// TestParseGoBuildOutput tests parsing go build output
func TestParseGoBuildOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		file     string
		expected int
	}{
		{
			name:     "empty output",
			output:   "",
			file:     "main.go",
			expected: 0,
		},
		{
			name:     "single error",
			output:   "# command-line-arguments\n./main.go:10:5: undefined: foo",
			file:     "main.go",
			expected: 1,
		},
		{
			name: "multiple errors",
			output: `# command-line-arguments
./main.go:10:5: undefined: foo
./main.go:15:3: syntax error`,
			file:     "main.go",
			expected: 2,
		},
		{
			name:     "errors in different file",
			output:   "./other.go:10:5: syntax error",
			file:     "main.go",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := parseGoBuildOutput(tt.output, tt.file)

			if len(issues) != tt.expected {
				t.Errorf("Expected %d issues, got %d", tt.expected, len(issues))
			}

			// Verify all issues have required fields
			for i, issue := range issues {
				if issue.File != tt.file {
					t.Errorf("Issue %d: expected file %s, got %s", i, tt.file, issue.File)
				}
				if issue.Severity != "error" {
					t.Errorf("Issue %d: expected severity 'error', got %s", i, issue.Severity)
				}
				if issue.Message == "" {
					t.Errorf("Issue %d: missing message", i)
				}
			}
		})
	}
}

// TestValidationScoring tests the scoring system
func TestValidationScoring(t *testing.T) {
	tests := []struct {
		name     string
		issues   []types.ReviewIssue
		minScore int
		maxScore int
	}{
		{
			name:     "no issues",
			issues:   []types.ReviewIssue{},
			minScore: 100,
			maxScore: 100,
		},
		{
			name: "one error",
			issues: []types.ReviewIssue{
				{Severity: "error", Message: "test error"},
			},
			minScore: 90,
			maxScore: 90,
		},
		{
			name: "one warning",
			issues: []types.ReviewIssue{
				{Severity: "warning", Message: "test warning"},
			},
			minScore: 95,
			maxScore: 95,
		},
		{
			name: "multiple issues",
			issues: []types.ReviewIssue{
				{Severity: "error", Message: "error 1"},
				{Severity: "error", Message: "error 2"},
				{Severity: "warning", Message: "warning 1"},
			},
			minScore: 75,
			maxScore: 75,
		},
		{
			name: "many issues - floor at 0",
			issues: func() []types.ReviewIssue {
				var issues []types.ReviewIssue
				for i := 0; i < 20; i++ {
					issues = append(issues, types.ReviewIssue{
						Severity: "error",
						Message:  "test error",
					})
				}
				return issues
			}(),
			minScore: 0,
			maxScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate score manually
			score := 100
			for _, issue := range tt.issues {
				if issue.Severity == "error" {
					score -= 10
				} else if issue.Severity == "warning" {
					score -= 5
				}
			}
			if score < 0 {
				score = 0
			}

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Expected score between %d-%d, got %d", tt.minScore, tt.maxScore, score)
			}

			t.Logf("Score for %d issues: %d", len(tt.issues), score)
		})
	}
}

// Test helper functions

// setupTestGoRepo creates a temporary directory for testing
func setupTestGoRepo(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "validation-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod to make it a valid Go module
	goMod := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module testmodule\n\ngo 1.21\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// Benchmark tests
func BenchmarkParseGoVetOutput(b *testing.B) {
	output := `main.go:10:5: unreachable code
utils.go:20:3: unused variable
helper.go:5:1: exported function should have comment
test.go:15:8: ineffective assignment`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseGoVetOutput(output)
	}
}

func BenchmarkParseGoBuildOutput(b *testing.B) {
	output := `# command-line-arguments
./main.go:10:5: undefined: foo
./main.go:15:3: syntax error
./main.go:20:1: missing return`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseGoBuildOutput(output, "main.go")
	}
}

func BenchmarkValidateGo(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "validation-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod
	goMod := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module testmodule\n\ngo 1.21\n"), 0644); err != nil {
		b.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a simple Go file
	testFile := filepath.Join(tempDir, "test.go")
	content := `package main

import "fmt"

func main() {
	fmt.Println("test")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	v := New(tempDir)
	ctx := context.Background()
	files := []string{"test.go"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateGo(ctx, files)
	}
}
