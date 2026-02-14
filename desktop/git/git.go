package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository provides git operations for a repository
type Repository struct {
	path string
}

// NewRepository creates a new Repository instance
func NewRepository(path string) *Repository {
	return &Repository{path: path}
}

// IsGitRepo checks if the path is a git repository
func (r *Repository) IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = r.path
	err := cmd.Run()
	return err == nil
}

// GetCurrentBranch returns the current branch name
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetStatus returns the git status
func (r *Repository) GetStatus() (*Status, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	status := &Status{
		Modified:  []string{},
		Added:     []string{},
		Deleted:   []string{},
		Untracked: []string{},
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		file := strings.TrimSpace(line[3:])

		switch {
		case strings.Contains(statusCode, "M"):
			status.Modified = append(status.Modified, file)
		case strings.Contains(statusCode, "A"):
			status.Added = append(status.Added, file)
		case strings.Contains(statusCode, "D"):
			status.Deleted = append(status.Deleted, file)
		case statusCode == "??":
			status.Untracked = append(status.Untracked, file)
		}
	}

	return status, nil
}

// Status represents git status
type Status struct {
	Modified  []string `json:"modified"`
	Added     []string `json:"added"`
	Deleted   []string `json:"deleted"`
	Untracked []string `json:"untracked"`
}

// GetDiff returns the diff for a file
func (r *Repository) GetDiff(filePath string) (string, error) {
	var cmd *exec.Cmd
	if filePath == "" {
		// Get all diffs when no file path is specified
		cmd = exec.Command("git", "diff")
	} else {
		cmd = exec.Command("git", "diff", filePath)
	}
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetStagedDiff returns the diff for staged changes
func (r *Repository) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// StageFile stages a file
func (r *Repository) StageFile(filePath string) error {
	cmd := exec.Command("git", "add", filePath)
	cmd.Dir = r.path
	return cmd.Run()
}

// UnstageFile unstages a file
func (r *Repository) UnstageFile(filePath string) error {
	cmd := exec.Command("git", "reset", "HEAD", filePath)
	cmd.Dir = r.path
	return cmd.Run()
}

// Commit creates a commit with the given message
func (r *Repository) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.path
	return cmd.Run()
}

// GetCommitHistory returns recent commits
func (r *Repository) GetCommitHistory(limit int) ([]Commit, error) {
	format := "%H|%an|%ae|%at|%s"
	cmd := exec.Command("git", "log", "-n", string(rune(limit)), "--format="+format)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commits := []Commit{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, Commit{
			Hash:        parts[0],
			AuthorName:  parts[1],
			AuthorEmail: parts[2],
			Message:     parts[4],
		})
	}

	return commits, nil
}

// Commit represents a git commit
type Commit struct {
	Hash        string `json:"hash"`
	AuthorName  string `json:"authorName"`
	AuthorEmail string `json:"authorEmail"`
	Message     string `json:"message"`
}

// DiscardChanges discards changes to a file
func (r *Repository) DiscardChanges(filePath string) error {
	cmd := exec.Command("git", "checkout", "--", filePath)
	cmd.Dir = r.path
	return cmd.Run()
}

// GetFilePath returns the full path to a file in the repo
func (r *Repository) GetFilePath(relativePath string) string {
	return filepath.Join(r.path, relativePath)
}
