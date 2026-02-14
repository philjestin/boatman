package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a temporary git repository for testing
func createTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to configure git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to configure git name: %v", err)
	}

	return tmpDir, cleanup
}

// Helper function to create a file in the test repo
func createFile(t *testing.T, dir, filename, content string) {
	t.Helper()

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", filename, err)
	}
}

// Helper function to commit changes
func commitChanges(t *testing.T, dir, message string) {
	t.Helper()

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage files: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository("/test/path")
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}
	if repo.path != "/test/path" {
		t.Errorf("Expected path '/test/path', got '%s'", repo.path)
	}
}

func TestIsGitRepo(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (string, func())
		expected bool
	}{
		{
			name: "valid git repository",
			setup: func() (string, func()) {
				return createTestRepo(t)
			},
			expected: true,
		},
		{
			name: "non-git directory",
			setup: func() (string, func()) {
				tmpDir, err := os.MkdirTemp("", "non-git-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				cleanup := func() { os.RemoveAll(tmpDir) }
				return tmpDir, cleanup
			},
			expected: false,
		},
		{
			name: "non-existent directory",
			setup: func() (string, func()) {
				return "/non/existent/path", func() {}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setup()
			defer cleanup()

			repo := NewRepository(path)
			result := repo.IsGitRepo()

			if result != tt.expected {
				t.Errorf("IsGitRepo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCurrentBranch(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test Repo")
	commitChanges(t, repoPath, "Initial commit")

	repo := NewRepository(repoPath)

	t.Run("default branch", func(t *testing.T) {
		branch, err := repo.GetCurrentBranch()
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}

		// Default branch can be 'master' or 'main' depending on git configuration
		if branch != "master" && branch != "main" {
			t.Errorf("GetCurrentBranch() = %s, want 'master' or 'main'", branch)
		}
	})

	t.Run("custom branch", func(t *testing.T) {
		// Create and checkout a new branch
		cmd := exec.Command("git", "checkout", "-b", "feature-branch")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		branch, err := repo.GetCurrentBranch()
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}

		if branch != "feature-branch" {
			t.Errorf("GetCurrentBranch() = %s, want 'feature-branch'", branch)
		}
	})

	t.Run("error on non-git repo", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "non-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		repo := NewRepository(tmpDir)
		_, err = repo.GetCurrentBranch()
		if err == nil {
			t.Error("GetCurrentBranch() expected error for non-git repo, got nil")
		}
	})
}

func TestGetStatus_CleanRepo(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit initial file
	createFile(t, repoPath, "README.md", "# Test Repo")
	commitChanges(t, repoPath, "Initial commit")

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Modified) != 0 {
		t.Errorf("Expected 0 modified files, got %d", len(status.Modified))
	}
	if len(status.Added) != 0 {
		t.Errorf("Expected 0 added files, got %d", len(status.Added))
	}
	if len(status.Deleted) != 0 {
		t.Errorf("Expected 0 deleted files, got %d", len(status.Deleted))
	}
	if len(status.Untracked) != 0 {
		t.Errorf("Expected 0 untracked files, got %d", len(status.Untracked))
	}
}

func TestGetStatus_ModifiedFiles(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit initial file
	createFile(t, repoPath, "file.txt", "original content")
	commitChanges(t, repoPath, "Initial commit")

	// Modify the file
	createFile(t, repoPath, "file.txt", "modified content")

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Modified) != 1 {
		t.Fatalf("Expected 1 modified file, got %d", len(status.Modified))
	}

	if status.Modified[0] != "file.txt" {
		t.Errorf("Expected modified file 'file.txt', got '%s'", status.Modified[0])
	}
}

func TestGetStatus_UntrackedFiles(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create untracked files
	createFile(t, repoPath, "untracked.txt", "new file")
	createFile(t, repoPath, "another.txt", "another new file")

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Untracked) != 2 {
		t.Fatalf("Expected 2 untracked files, got %d", len(status.Untracked))
	}

	untrackedMap := make(map[string]bool)
	for _, file := range status.Untracked {
		untrackedMap[file] = true
	}

	if !untrackedMap["untracked.txt"] || !untrackedMap["another.txt"] {
		t.Errorf("Expected untracked files 'untracked.txt' and 'another.txt', got %v", status.Untracked)
	}
}

func TestGetStatus_AddedFiles(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create and stage a new file
	createFile(t, repoPath, "new.txt", "new content")
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Added) != 1 {
		t.Fatalf("Expected 1 added file, got %d", len(status.Added))
	}

	if status.Added[0] != "new.txt" {
		t.Errorf("Expected added file 'new.txt', got '%s'", status.Added[0])
	}
}

func TestGetStatus_DeletedFiles(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit file
	createFile(t, repoPath, "to_delete.txt", "content")
	commitChanges(t, repoPath, "Initial commit")

	// Delete the file
	filePath := filepath.Join(repoPath, "to_delete.txt")
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Deleted) != 1 {
		t.Fatalf("Expected 1 deleted file, got %d", len(status.Deleted))
	}

	if status.Deleted[0] != "to_delete.txt" {
		t.Errorf("Expected deleted file 'to_delete.txt', got '%s'", status.Deleted[0])
	}
}

func TestGetStatus_MixedChanges(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit initial files
	createFile(t, repoPath, "existing.txt", "content")
	createFile(t, repoPath, "to_delete.txt", "content")
	commitChanges(t, repoPath, "Initial commit")

	// Modify existing file
	createFile(t, repoPath, "existing.txt", "modified content")

	// Delete file
	os.Remove(filepath.Join(repoPath, "to_delete.txt"))

	// Create untracked file
	createFile(t, repoPath, "untracked.txt", "new content")

	// Create and stage new file
	createFile(t, repoPath, "staged.txt", "staged content")
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = repoPath
	cmd.Run()

	repo := NewRepository(repoPath)
	status, err := repo.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if len(status.Modified) != 1 || status.Modified[0] != "existing.txt" {
		t.Errorf("Expected modified file 'existing.txt', got %v", status.Modified)
	}

	if len(status.Deleted) != 1 || status.Deleted[0] != "to_delete.txt" {
		t.Errorf("Expected deleted file 'to_delete.txt', got %v", status.Deleted)
	}

	if len(status.Untracked) != 1 || status.Untracked[0] != "untracked.txt" {
		t.Errorf("Expected untracked file 'untracked.txt', got %v", status.Untracked)
	}

	if len(status.Added) != 1 || status.Added[0] != "staged.txt" {
		t.Errorf("Expected added file 'staged.txt', got %v", status.Added)
	}
}

func TestGetDiff_ModifiedFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit initial file
	createFile(t, repoPath, "file.txt", "line 1\nline 2\nline 3\n")
	commitChanges(t, repoPath, "Initial commit")

	// Modify the file
	createFile(t, repoPath, "file.txt", "line 1\nmodified line 2\nline 3\n")

	repo := NewRepository(repoPath)
	diff, err := repo.GetDiff("file.txt")

	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	if diff == "" {
		t.Error("GetDiff() returned empty diff for modified file")
	}

	// Check for expected diff markers
	if !strings.Contains(diff, "diff --git") {
		t.Error("Diff should contain 'diff --git' header")
	}

	if !strings.Contains(diff, "@@") {
		t.Error("Diff should contain hunk header '@@'")
	}

	if !strings.Contains(diff, "-line 2") && !strings.Contains(diff, "+modified line 2") {
		t.Error("Diff should show the line modification")
	}
}

func TestGetDiff_NewFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create new file (untracked)
	createFile(t, repoPath, "new.txt", "new content")

	repo := NewRepository(repoPath)
	diff, err := repo.GetDiff("new.txt")

	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	// Untracked files typically have no diff output
	if diff != "" {
		t.Log("Note: GetDiff for untracked file returned:", diff)
	}
}

func TestGetDiff_DeletedFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit file
	createFile(t, repoPath, "to_delete.txt", "content to delete\n")
	commitChanges(t, repoPath, "Initial commit")

	// Delete the file
	os.Remove(filepath.Join(repoPath, "to_delete.txt"))

	repo := NewRepository(repoPath)
	diff, err := repo.GetDiff("to_delete.txt")

	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	if diff == "" {
		t.Error("GetDiff() returned empty diff for deleted file")
	}

	if !strings.Contains(diff, "-content to delete") {
		t.Error("Diff should show deleted content")
	}
}

func TestGetDiff_UnchangedFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit file
	createFile(t, repoPath, "unchanged.txt", "content")
	commitChanges(t, repoPath, "Initial commit")

	repo := NewRepository(repoPath)
	diff, err := repo.GetDiff("unchanged.txt")

	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	if diff != "" {
		t.Errorf("GetDiff() should return empty diff for unchanged file, got: %s", diff)
	}
}

func TestGetStagedDiff(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "file.txt", "original\n")
	commitChanges(t, repoPath, "Initial commit")

	// Modify and stage file
	createFile(t, repoPath, "file.txt", "modified\n")
	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	repo := NewRepository(repoPath)
	diff, err := repo.GetStagedDiff()

	if err != nil {
		t.Fatalf("GetStagedDiff() error = %v", err)
	}

	if diff == "" {
		t.Error("GetStagedDiff() returned empty diff for staged changes")
	}

	if !strings.Contains(diff, "-original") || !strings.Contains(diff, "+modified") {
		t.Error("Staged diff should show the modification")
	}
}

func TestStageFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create new file
	createFile(t, repoPath, "new.txt", "content")

	repo := NewRepository(repoPath)
	err := repo.StageFile("new.txt")

	if err != nil {
		t.Fatalf("StageFile() error = %v", err)
	}

	// Verify file is staged
	status, _ := repo.GetStatus()
	if len(status.Added) != 1 || status.Added[0] != "new.txt" {
		t.Error("File was not staged successfully")
	}
}

func TestUnstageFile(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create and stage new file
	createFile(t, repoPath, "new.txt", "content")
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = repoPath
	cmd.Run()

	repo := NewRepository(repoPath)

	// Unstage the file
	err := repo.UnstageFile("new.txt")
	if err != nil {
		t.Fatalf("UnstageFile() error = %v", err)
	}

	// Verify file is unstaged
	status, _ := repo.GetStatus()
	if len(status.Added) != 0 {
		t.Error("File was not unstaged successfully")
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "new.txt" {
		t.Error("File should be untracked after unstaging")
	}
}

func TestCommit(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create initial commit
	createFile(t, repoPath, "README.md", "# Test")
	commitChanges(t, repoPath, "Initial commit")

	// Create and stage new file
	createFile(t, repoPath, "file.txt", "content")
	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = repoPath
	cmd.Run()

	repo := NewRepository(repoPath)
	err := repo.Commit("Add new file")

	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify commit was created
	cmd = exec.Command("git", "log", "--oneline", "-n", "1")
	cmd.Dir = repoPath
	output, _ := cmd.Output()
	if !strings.Contains(string(output), "Add new file") {
		t.Error("Commit message not found in git log")
	}
}

func TestGetCommitHistory(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create multiple commits
	createFile(t, repoPath, "file1.txt", "content1")
	commitChanges(t, repoPath, "First commit")

	createFile(t, repoPath, "file2.txt", "content2")
	commitChanges(t, repoPath, "Second commit")

	createFile(t, repoPath, "file3.txt", "content3")
	commitChanges(t, repoPath, "Third commit")

	repo := NewRepository(repoPath)
	commits, err := repo.GetCommitHistory(5)

	if err != nil {
		t.Fatalf("GetCommitHistory() error = %v", err)
	}

	if len(commits) != 3 {
		t.Fatalf("Expected 3 commits, got %d", len(commits))
	}

	// Commits should be in reverse chronological order
	if commits[0].Message != "Third commit" {
		t.Errorf("Expected first commit message 'Third commit', got '%s'", commits[0].Message)
	}

	if commits[2].Message != "First commit" {
		t.Errorf("Expected last commit message 'First commit', got '%s'", commits[2].Message)
	}

	// Verify commit fields
	if commits[0].AuthorName != "Test User" {
		t.Errorf("Expected author 'Test User', got '%s'", commits[0].AuthorName)
	}

	if commits[0].AuthorEmail != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", commits[0].AuthorEmail)
	}

	if commits[0].Hash == "" {
		t.Error("Commit hash should not be empty")
	}
}

func TestDiscardChanges(t *testing.T) {
	repoPath, cleanup := createTestRepo(t)
	defer cleanup()

	// Create and commit file
	createFile(t, repoPath, "file.txt", "original content")
	commitChanges(t, repoPath, "Initial commit")

	// Modify the file
	createFile(t, repoPath, "file.txt", "modified content")

	// Verify file is modified
	content, _ := os.ReadFile(filepath.Join(repoPath, "file.txt"))
	if string(content) != "modified content" {
		t.Fatal("File was not modified")
	}

	repo := NewRepository(repoPath)
	err := repo.DiscardChanges("file.txt")

	if err != nil {
		t.Fatalf("DiscardChanges() error = %v", err)
	}

	// Verify file is restored
	content, _ = os.ReadFile(filepath.Join(repoPath, "file.txt"))
	if string(content) != "original content" {
		t.Error("File changes were not discarded")
	}
}

func TestGetFilePath(t *testing.T) {
	repo := NewRepository("/test/repo")

	tests := []struct {
		name         string
		relativePath string
		expected     string
	}{
		{
			name:         "simple file",
			relativePath: "file.txt",
			expected:     "/test/repo/file.txt",
		},
		{
			name:         "nested file",
			relativePath: "subdir/file.txt",
			expected:     "/test/repo/subdir/file.txt",
		},
		{
			name:         "empty path",
			relativePath: "",
			expected:     "/test/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.GetFilePath(tt.relativePath)
			if result != tt.expected {
				t.Errorf("GetFilePath(%s) = %s, want %s", tt.relativePath, result, tt.expected)
			}
		})
	}
}

func TestGetStatus_ErrorHandling(t *testing.T) {
	t.Run("non-git directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "non-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		repo := NewRepository(tmpDir)
		_, err = repo.GetStatus()

		if err == nil {
			t.Error("GetStatus() expected error for non-git directory, got nil")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		repo := NewRepository("/non/existent/path")
		_, err := repo.GetStatus()

		if err == nil {
			t.Error("GetStatus() expected error for non-existent directory, got nil")
		}
	})
}

func TestGetDiff_ErrorHandling(t *testing.T) {
	t.Run("non-git directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "non-git-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		repo := NewRepository(tmpDir)
		_, err = repo.GetDiff("file.txt")

		if err == nil {
			t.Error("GetDiff() expected error for non-git directory, got nil")
		}
	})
}
