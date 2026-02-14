package project

import (
	"os"
	"path/filepath"
)

// WorkspaceInfo contains information about a workspace
type WorkspaceInfo struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	IsGitRepo  bool     `json:"isGitRepo"`
	HasPackage bool     `json:"hasPackage"`
	Languages  []string `json:"languages"`
}

// Workspace provides utilities for working with project workspaces
type Workspace struct {
	path string
}

// NewWorkspace creates a new workspace instance
func NewWorkspace(path string) *Workspace {
	return &Workspace{path: path}
}

// GetInfo returns information about the workspace
func (w *Workspace) GetInfo() (*WorkspaceInfo, error) {
	info := &WorkspaceInfo{
		Path:      w.path,
		Name:      filepath.Base(w.path),
		Languages: []string{},
	}

	// Check if it's a git repository
	if _, err := os.Stat(filepath.Join(w.path, ".git")); err == nil {
		info.IsGitRepo = true
	}

	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(w.path, "package.json")); err == nil {
		info.HasPackage = true
		info.Languages = append(info.Languages, "javascript")
	}

	// Check for go.mod (Go)
	if _, err := os.Stat(filepath.Join(w.path, "go.mod")); err == nil {
		info.Languages = append(info.Languages, "go")
	}

	// Check for Cargo.toml (Rust)
	if _, err := os.Stat(filepath.Join(w.path, "Cargo.toml")); err == nil {
		info.Languages = append(info.Languages, "rust")
	}

	// Check for pyproject.toml or requirements.txt (Python)
	if _, err := os.Stat(filepath.Join(w.path, "pyproject.toml")); err == nil {
		info.Languages = append(info.Languages, "python")
	} else if _, err := os.Stat(filepath.Join(w.path, "requirements.txt")); err == nil {
		info.Languages = append(info.Languages, "python")
	}

	return info, nil
}

// ListFiles returns files in the workspace (non-recursive)
func (w *Workspace) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(w.path)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return files, nil
}

// ReadFile reads a file from the workspace
func (w *Workspace) ReadFile(relativePath string) (string, error) {
	fullPath := filepath.Join(w.path, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes content to a file in the workspace
func (w *Workspace) WriteFile(relativePath string, content string) error {
	fullPath := filepath.Join(w.path, relativePath)
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// FileExists checks if a file exists in the workspace
func (w *Workspace) FileExists(relativePath string) bool {
	fullPath := filepath.Join(w.path, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetPath returns the workspace path
func (w *Workspace) GetPath() string {
	return w.path
}
