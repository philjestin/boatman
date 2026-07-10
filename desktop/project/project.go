package project

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProjectRepository represents a repository that belongs to a Boatman project.
type ProjectRepository struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsPrimary bool   `json:"isPrimary,omitempty"`
}

// Project represents a project/workspace
type Project struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Path         string              `json:"path"`
	Description  string              `json:"description,omitempty"`
	Repositories []ProjectRepository `json:"repositories,omitempty"`
	LastOpened   time.Time           `json:"lastOpened"`
	CreatedAt    time.Time           `json:"createdAt"`
}

// ProjectManager manages projects and workspaces
type ProjectManager struct {
	mu          sync.RWMutex
	projects    []Project
	recentLimit int
	storagePath string
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

	if err := json.Unmarshal(data, &pm.projects); err != nil {
		return err
	}
	for i := range pm.projects {
		pm.projects[i] = normalizeProject(pm.projects[i])
	}
	return nil
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
			pm.projects[i] = normalizeProject(pm.projects[i])
			pm.save()
			return &pm.projects[i], nil
		}
	}

	// Create new project
	project := Project{
		ID:           filepath.Base(path) + "-" + time.Now().Format("20060102150405"),
		Name:         filepath.Base(path),
		Path:         path,
		Repositories: []ProjectRepository{newProjectRepository(path, true)},
		LastOpened:   time.Now(),
		CreatedAt:    time.Now(),
	}

	pm.projects = append([]Project{project}, pm.projects...)

	// Limit recent projects
	if len(pm.projects) > pm.recentLimit {
		pm.projects = pm.projects[:pm.recentLimit]
	}

	pm.save()
	return &project, nil
}

// AddMultiRepoProject adds or updates a project backed by multiple repositories.
func (pm *ProjectManager) AddMultiRepoProject(name string, repositoryPaths []string) (*Project, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	repositories, err := buildProjectRepositories(repositoryPaths)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = filepath.Base(repositories[0].Path)
		if len(repositories) > 1 {
			name += " workspace"
		}
	}

	now := time.Now()
	for i, p := range pm.projects {
		project := normalizeProject(p)
		if sameRepositorySet(project.Repositories, repositories) {
			pm.projects[i].Name = name
			pm.projects[i].Path = repositories[0].Path
			pm.projects[i].Repositories = repositories
			pm.projects[i].LastOpened = now
			pm.save()
			return &pm.projects[i], nil
		}
	}

	project := Project{
		ID:           projectID(name),
		Name:         name,
		Path:         repositories[0].Path,
		Repositories: repositories,
		LastOpened:   now,
		CreatedAt:    now,
	}
	pm.projects = append([]Project{project}, pm.projects...)
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
			project := cloneProject(normalizeProject(p))
			return &project, nil
		}
	}

	return nil, os.ErrNotExist
}

// GetProjectByPath returns a project by path
func (pm *ProjectManager) GetProjectByPath(path string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	cleanPath := filepath.Clean(path)
	for _, p := range pm.projects {
		project := normalizeProject(p)
		if filepath.Clean(project.Path) == cleanPath {
			cloned := cloneProject(project)
			return &cloned, nil
		}
		for _, repo := range project.Repositories {
			if filepath.Clean(repo.Path) == cleanPath {
				cloned := cloneProject(project)
				return &cloned, nil
			}
		}
	}

	return nil, os.ErrNotExist
}

// ListProjects returns all projects
func (pm *ProjectManager) ListProjects() []Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return cloneProjects(pm.projects)
}

// GetRecentProjects returns recently opened projects
func (pm *ProjectManager) GetRecentProjects(limit int) []Project {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if limit <= 0 || limit > len(pm.projects) {
		limit = len(pm.projects)
	}

	return cloneProjects(pm.projects[:limit])
}

// UpdateProject updates project metadata
func (pm *ProjectManager) UpdateProject(project Project) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, p := range pm.projects {
		if p.ID == project.ID {
			pm.projects[i] = normalizeProject(project)
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

func normalizeProject(project Project) Project {
	if len(project.Repositories) == 0 && project.Path != "" {
		project.Repositories = []ProjectRepository{newProjectRepository(project.Path, true)}
	}
	if project.Path == "" && len(project.Repositories) > 0 {
		project.Path = project.Repositories[0].Path
	}
	hasPrimary := false
	for i := range project.Repositories {
		if strings.TrimSpace(project.Repositories[i].ID) == "" {
			project.Repositories[i].ID = repositoryID(project.Repositories[i].Path)
		}
		if strings.TrimSpace(project.Repositories[i].Name) == "" {
			project.Repositories[i].Name = filepath.Base(project.Repositories[i].Path)
		}
		if project.Repositories[i].IsPrimary {
			hasPrimary = true
		}
	}
	if len(project.Repositories) > 0 && !hasPrimary {
		project.Repositories[0].IsPrimary = true
	}
	return project
}

func cloneProjects(projects []Project) []Project {
	cloned := make([]Project, len(projects))
	for i, project := range projects {
		cloned[i] = cloneProject(normalizeProject(project))
	}
	return cloned
}

func cloneProject(project Project) Project {
	if project.Repositories != nil {
		repositories := make([]ProjectRepository, len(project.Repositories))
		copy(repositories, project.Repositories)
		project.Repositories = repositories
	}
	return project
}

func buildProjectRepositories(paths []string) ([]ProjectRepository, error) {
	seen := make(map[string]bool, len(paths))
	repositories := make([]ProjectRepository, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		path = filepath.Clean(path)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, os.ErrInvalid
		}
		key := filepath.Clean(path)
		if seen[key] {
			continue
		}
		seen[key] = true
		repositories = append(repositories, newProjectRepository(path, len(repositories) == 0))
	}
	if len(repositories) == 0 {
		return nil, os.ErrInvalid
	}
	return repositories, nil
}

func newProjectRepository(path string, primary bool) ProjectRepository {
	return ProjectRepository{
		ID:        repositoryID(path),
		Name:      filepath.Base(path),
		Path:      path,
		IsPrimary: primary,
	}
}

func sameRepositorySet(left, right []ProjectRepository) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]bool, len(left))
	for _, repo := range left {
		seen[filepath.Clean(repo.Path)] = true
	}
	for _, repo := range right {
		if !seen[filepath.Clean(repo.Path)] {
			return false
		}
	}
	return true
}

func projectID(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	base = strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-").Replace(base)
	base = strings.Trim(base, "-")
	if base == "" {
		base = "project"
	}
	return base + "-" + time.Now().Format("20060102150405")
}

func repositoryID(path string) string {
	base := strings.ToLower(filepath.Base(path))
	base = strings.NewReplacer(" ", "-", "_", "-").Replace(base)
	base = strings.Trim(base, "-")
	if base == "" {
		base = "repo"
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(filepath.Clean(path)))
	return fmt.Sprintf("%s-%08x", base, hash.Sum32())
}
