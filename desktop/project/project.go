package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Project represents a project/workspace
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description,omitempty"`
	LastOpened  time.Time `json:"lastOpened"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ProjectManager manages projects and workspaces
type ProjectManager struct {
	mu           sync.RWMutex
	projects     []Project
	recentLimit  int
	storagePath  string
}

// NewProjectManager creates a new project manager
func NewProjectManager() (*ProjectManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".boatman")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	pm := &ProjectManager{
		projects:    []Project{},
		recentLimit: 10,
		storagePath: filepath.Join(configDir, "projects.json"),
	}

	// Load existing projects
	if err := pm.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return pm, nil
}

// load reads projects from disk
func (pm *ProjectManager) load() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.storagePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pm.projects)
}

// save writes projects to disk
func (pm *ProjectManager) save() error {
	data, err := json.MarshalIndent(pm.projects, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.storagePath, data, 0644)
}

// AddProject adds or updates a project
func (pm *ProjectManager) AddProject(path string) (*Project, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	// Check if project already exists
	for i, p := range pm.projects {
		if p.Path == path {
			pm.projects[i].LastOpened = time.Now()
			pm.save()
			return &pm.projects[i], nil
		}
	}

	// Create new project
	project := Project{
		ID:         filepath.Base(path) + "-" + time.Now().Format("20060102150405"),
		Name:       filepath.Base(path),
		Path:       path,
		LastOpened: time.Now(),
		CreatedAt:  time.Now(),
	}

	pm.projects = append([]Project{project}, pm.projects...)

	// Limit recent projects
	if len(pm.projects) > pm.recentLimit {
		pm.projects = pm.projects[:pm.recentLimit]
	}

	pm.save()
	return &project, nil
}

// RemoveProject removes a project
func (pm *ProjectManager) RemoveProject(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, p := range pm.projects {
		if p.ID == id {
			pm.projects = append(pm.projects[:i], pm.projects[i+1:]...)
			return pm.save()
		}
	}

	return os.ErrNotExist
}

// GetProject returns a project by ID
func (pm *ProjectManager) GetProject(id string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.projects {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, os.ErrNotExist
}

// GetProjectByPath returns a project by path
func (pm *ProjectManager) GetProjectByPath(path string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.projects {
		if p.Path == path {
			return &p, nil
		}
	}

	return nil, os.ErrNotExist
}

// ListProjects returns all projects
func (pm *ProjectManager) ListProjects() []Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	projects := make([]Project, len(pm.projects))
	copy(projects, pm.projects)
	return projects
}

// GetRecentProjects returns recently opened projects
func (pm *ProjectManager) GetRecentProjects(limit int) []Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if limit <= 0 || limit > len(pm.projects) {
		limit = len(pm.projects)
	}

	projects := make([]Project, limit)
	copy(projects, pm.projects[:limit])
	return projects
}

// UpdateProject updates project metadata
func (pm *ProjectManager) UpdateProject(project Project) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, p := range pm.projects {
		if p.ID == project.ID {
			pm.projects[i] = project
			return pm.save()
		}
	}

	return os.ErrNotExist
}

// ValidatePath checks if a path is a valid project directory
func (pm *ProjectManager) ValidatePath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
