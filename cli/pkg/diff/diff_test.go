package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNew tests creating a new diff analyzer
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{"simple path", "/path/to/repo"},
		{"empty path", ""},
		{"relative path", "./repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := New(tt.repoPath)
			if analyzer == nil {
				t.Fatal("New returned nil")
			}

			if analyzer.repoPath != tt.repoPath {
				t.Errorf("Expected repoPath %s, got %s", tt.repoPath, analyzer.repoPath)
			}
		})
	}
}

// TestAnalyzer_GetDiff tests getting unstaged diff
func TestAnalyzer_GetDiff(t *testing.T) {
	// Create a temporary git repository
	tempDir, cleanup := setupTestRepo(t)
	defer cleanup()

	analyzer := New(tempDir)

	// Initially no diff
	diff, err := analyzer.GetDiff()
	if err != nil {
		t.Fatalf("GetDiff failed: %v", err)
	}

	if diff != "" {
		t.Log("Note: diff is not empty (might have uncommitted changes from test setup)")
	}

	// Make a change
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("new content\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get diff again
	diff, err = analyzer.GetDiff()
	if err != nil {
		t.Fatalf("GetDiff failed after change: %v", err)
	}

	if diff == "" {
		t.Error("Expected non-empty diff after making changes")
	}
}

// TestAnalyzer_GetStagedDiff tests getting staged diff
func TestAnalyzer_GetStagedDiff(t *testing.T) {
	tempDir, cleanup := setupTestRepo(t)
	defer cleanup()

	analyzer := New(tempDir)

	// Initially no staged diff
	diff, err := analyzer.GetStagedDiff()
	if err != nil {
		t.Fatalf("GetStagedDiff failed: %v", err)
	}

	if diff != "" {
		t.Logf("Note: staged diff is not empty: %s", diff)
	}

	// Make a change and stage it
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("staged content\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Get staged diff
	diff, err = analyzer.GetStagedDiff()
	if err != nil {
		t.Fatalf("GetStagedDiff failed after staging: %v", err)
	}

	if diff == "" {
		t.Error("Expected non-empty diff after staging changes")
	}
}

// TestAnalyzer_GetDiff_InvalidRepo tests error handling with invalid repo
func TestAnalyzer_GetDiff_InvalidRepo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "diff-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Not a git repository
	analyzer := New(tempDir)
	_, err = analyzer.GetDiff()
	if err == nil {
		t.Error("Expected error for non-git repository, got nil")
	}
}

// TestParseDiff tests diff parsing
func TestParseDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected *DiffStats
	}{
		{
			name: "empty diff",
			diff: "",
			expected: &DiffStats{
				FilesChanged: 0,
				FilesAdded:   0,
				FilesDeleted: 0,
				LinesAdded:   0,
				LinesDeleted: 0,
				Files:        nil,
			},
		},
		{
			name: "simple addition",
			diff: `diff --git a/file.txt b/file.txt
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/file.txt
@@ -0,0 +1,3 @@
+line 1
+line 2
+line 3`,
			expected: &DiffStats{
				FilesChanged: 1,
				FilesAdded:   1,
				FilesDeleted: 0,
				LinesAdded:   3,
				LinesDeleted: 0,
				Files:        []string{"file.txt"},
			},
		},
		{
			name: "modification",
			diff: `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,8 @@ func main() {
 	fmt.Println("old line")
-	fmt.Println("removed")
+	fmt.Println("added")
+	fmt.Println("another add")`,
			expected: &DiffStats{
				FilesChanged: 1,
				FilesAdded:   0,
				FilesDeleted: 0,
				LinesAdded:   2,
				LinesDeleted: 1,
				Files:        nil,
			},
		},
		{
			name: "file deletion",
			diff: `diff --git a/old.txt b/old.txt
deleted file mode 100644
index abc123..0000000
--- a/old.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-line 1
-line 2`,
			expected: &DiffStats{
				FilesChanged: 1,
				FilesAdded:   0,
				FilesDeleted: 1,
				LinesAdded:   0,
				LinesDeleted: 2,
				Files:        nil,
			},
		},
		{
			name: "multiple files",
			diff: `diff --git a/file1.txt b/file1.txt
index abc123..def456 100644
--- a/file1.txt
+++ b/file1.txt
@@ -1,3 +1,4 @@
 line 1
+added line
 line 2
diff --git a/file2.txt b/file2.txt
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1,2 @@
+new file line 1
+new file line 2`,
			expected: &DiffStats{
				FilesChanged: 2,
				FilesAdded:   1,
				FilesDeleted: 0,
				LinesAdded:   3,
				LinesDeleted: 0,
				Files:        []string{"file2.txt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := ParseDiff(tt.diff)

			if stats.FilesChanged != tt.expected.FilesChanged {
				t.Errorf("FilesChanged = %d, want %d", stats.FilesChanged, tt.expected.FilesChanged)
			}

			if stats.FilesAdded != tt.expected.FilesAdded {
				t.Errorf("FilesAdded = %d, want %d", stats.FilesAdded, tt.expected.FilesAdded)
			}

			if stats.FilesDeleted != tt.expected.FilesDeleted {
				t.Errorf("FilesDeleted = %d, want %d", stats.FilesDeleted, tt.expected.FilesDeleted)
			}

			if stats.LinesAdded != tt.expected.LinesAdded {
				t.Errorf("LinesAdded = %d, want %d", stats.LinesAdded, tt.expected.LinesAdded)
			}

			if stats.LinesDeleted != tt.expected.LinesDeleted {
				t.Errorf("LinesDeleted = %d, want %d", stats.LinesDeleted, tt.expected.LinesDeleted)
			}

			if len(stats.Files) != len(tt.expected.Files) {
				t.Errorf("Files length = %d, want %d", len(stats.Files), len(tt.expected.Files))
			}
		})
	}
}

// TestDiffStats_Total tests the Total method
func TestDiffStats_Total(t *testing.T) {
	tests := []struct {
		name     string
		stats    *DiffStats
		expected int
	}{
		{
			name: "zero changes",
			stats: &DiffStats{
				LinesAdded:   0,
				LinesDeleted: 0,
			},
			expected: 0,
		},
		{
			name: "only additions",
			stats: &DiffStats{
				LinesAdded:   10,
				LinesDeleted: 0,
			},
			expected: 10,
		},
		{
			name: "only deletions",
			stats: &DiffStats{
				LinesAdded:   0,
				LinesDeleted: 5,
			},
			expected: 5,
		},
		{
			name: "mixed changes",
			stats: &DiffStats{
				LinesAdded:   15,
				LinesDeleted: 8,
			},
			expected: 23,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.stats.Total()
			if total != tt.expected {
				t.Errorf("Total() = %d, want %d", total, tt.expected)
			}
		})
	}
}

// TestDiffStats_Summary tests the Summary method
func TestDiffStats_Summary(t *testing.T) {
	tests := []struct {
		name     string
		stats    *DiffStats
		expected string
	}{
		{
			name: "no changes",
			stats: &DiffStats{
				FilesChanged: 0,
				LinesAdded:   0,
				LinesDeleted: 0,
			},
			expected: "0 files changed, +0/-0 lines",
		},
		{
			name: "additions only",
			stats: &DiffStats{
				FilesChanged: 1,
				LinesAdded:   5,
				LinesDeleted: 0,
			},
			expected: "1 files changed, +5/-0 lines",
		},
		{
			name: "mixed changes",
			stats: &DiffStats{
				FilesChanged: 3,
				LinesAdded:   10,
				LinesDeleted: 7,
			},
			expected: "3 files changed, +10/-7 lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.stats.Summary()
			if summary != tt.expected {
				t.Errorf("Summary() = %q, want %q", summary, tt.expected)
			}
		})
	}
}

// Test helper functions

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "diff-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	// Create initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to stage file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// Benchmark tests
func BenchmarkParseDiff(b *testing.B) {
	diff := `diff --git a/file1.txt b/file1.txt
index abc123..def456 100644
--- a/file1.txt
+++ b/file1.txt
@@ -1,10 +1,12 @@
 line 1
+added line 1
 line 2
 line 3
-removed line
 line 4
+added line 2
 line 5`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseDiff(diff)
	}
}

func BenchmarkDiffStats_Total(b *testing.B) {
	stats := &DiffStats{
		LinesAdded:   100,
		LinesDeleted: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.Total()
	}
}

func BenchmarkDiffStats_Summary(b *testing.B) {
	stats := &DiffStats{
		FilesChanged: 5,
		LinesAdded:   100,
		LinesDeleted: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.Summary()
	}
}
