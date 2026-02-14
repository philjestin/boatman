package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestProjectManager creates a ProjectManager instance with a temporary storage path
func setupTestProjectManager(t *testing.T) (*ProjectManager, string) {
	t.Helper()

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "boatman-project-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	pm := &ProjectManager{
		projects:    []Project{},
		recentLimit: 10,
		storagePath: filepath.Join(tempDir, "projects.json"),
	}

	return pm, tempDir
}

// createTestDir creates a temporary directory for testing
func createTestDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "boatman-testproject-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return dir
}

func TestNewProjectManager(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "boatman-home-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set temporary HOME
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create new project manager
	pm, err := NewProjectManager()
	if err != nil {
		t.Fatalf("NewProjectManager() error = %v", err)
	}

	// Verify config directory was created
	configDir := filepath.Join(tempHome, ".boatman")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %v", err)
	}

	// Verify storage path is correct
	expectedPath := filepath.Join(configDir, "projects.json")
	if pm.storagePath != expectedPath {
		t.Errorf("Expected storage path = %v, got %v", expectedPath, pm.storagePath)
	}

	// Verify default values
	if pm.recentLimit != 10 {
		t.Errorf("Expected recentLimit = 10, got %d", pm.recentLimit)
	}

	if pm.projects == nil {
		t.Errorf("Expected projects to be initialized")
	}

	if len(pm.projects) != 0 {
		t.Errorf("Expected projects to be empty, got %d projects", len(pm.projects))
	}
}

func TestAddProject_NewProject(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create a test directory
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	// Add the project
	project, err := pm.AddProject(projectDir)
	if err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}

	// Verify project details
	if project.Path != projectDir {
		t.Errorf("Expected Path = %v, got %v", projectDir, project.Path)
	}

	expectedName := filepath.Base(projectDir)
	if project.Name != expectedName {
		t.Errorf("Expected Name = %v, got %v", expectedName, project.Name)
	}

	if project.ID == "" {
		t.Errorf("Expected non-empty ID")
	}

	if project.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt time")
	}

	if project.LastOpened.IsZero() {
		t.Errorf("Expected non-zero LastOpened time")
	}

	// Verify project was added to list
	if len(pm.projects) != 1 {
		t.Errorf("Expected 1 project in list, got %d", len(pm.projects))
	}
}

func TestAddProject_ExistingProject(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	// Add project first time
	project1, err := pm.AddProject(projectDir)
	if err != nil {
		t.Fatalf("AddProject() first call error = %v", err)
	}

	originalID := project1.ID
	originalCreatedAt := project1.CreatedAt
	originalLastOpened := project1.LastOpened

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Add same project again
	project2, err := pm.AddProject(projectDir)
	if err != nil {
		t.Fatalf("AddProject() second call error = %v", err)
	}

	// Verify ID and CreatedAt didn't change
	if project2.ID != originalID {
		t.Errorf("Expected ID to remain %v, got %v", originalID, project2.ID)
	}

	if !project2.CreatedAt.Equal(originalCreatedAt) {
		t.Errorf("Expected CreatedAt to remain unchanged")
	}

	// Verify LastOpened was updated
	if !project2.LastOpened.After(originalLastOpened) {
		t.Errorf("Expected LastOpened to be updated")
	}

	// Verify only one project in list
	if len(pm.projects) != 1 {
		t.Errorf("Expected 1 project in list, got %d", len(pm.projects))
	}
}

func TestAddProject_InvalidPath(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to add non-existent path
	_, err := pm.AddProject("/nonexistent/path/12345")
	if err == nil {
		t.Errorf("Expected error for non-existent path, got nil")
	}
}

func TestAddProject_FileNotDirectory(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create a file instead of directory
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to add file as project
	_, err := pm.AddProject(testFile)
	if err != os.ErrInvalid {
		t.Errorf("Expected os.ErrInvalid for file path, got %v", err)
	}
}

func TestRemoveProject(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add some projects
	dir1 := createTestDir(t)
	defer os.RemoveAll(dir1)
	dir2 := createTestDir(t)
	defer os.RemoveAll(dir2)

	project1, _ := pm.AddProject(dir1)
	project2, _ := pm.AddProject(dir2)

	// Remove first project
	err := pm.RemoveProject(project1.ID)
	if err != nil {
		t.Fatalf("RemoveProject() error = %v", err)
	}

	// Verify project was removed
	if len(pm.projects) != 1 {
		t.Errorf("Expected 1 project remaining, got %d", len(pm.projects))
	}

	// Verify correct project remains
	if pm.projects[0].ID != project2.ID {
		t.Errorf("Wrong project was removed")
	}

	// Verify storage was updated
	data, err := os.ReadFile(pm.storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	var saved []Project
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved projects: %v", err)
	}

	if len(saved) != 1 {
		t.Errorf("Expected 1 saved project, got %d", len(saved))
	}
}

func TestRemoveProject_NotFound(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to remove non-existent project
	err := pm.RemoveProject("nonexistent-id")
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestGetProject_Exists(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add a project
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	added, _ := pm.AddProject(projectDir)

	// Get the project
	project, err := pm.GetProject(added.ID)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	// Verify project details
	if project.ID != added.ID {
		t.Errorf("Expected ID = %v, got %v", added.ID, project.ID)
	}
	if project.Path != added.Path {
		t.Errorf("Expected Path = %v, got %v", added.Path, project.Path)
	}
}

func TestGetProject_NotExists(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to get non-existent project
	_, err := pm.GetProject("nonexistent-id")
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestGetProjectByPath_Exists(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add a project
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	added, _ := pm.AddProject(projectDir)

	// Get project by path
	project, err := pm.GetProjectByPath(projectDir)
	if err != nil {
		t.Fatalf("GetProjectByPath() error = %v", err)
	}

	if project.ID != added.ID {
		t.Errorf("Expected ID = %v, got %v", added.ID, project.ID)
	}
}

func TestGetProjectByPath_NotExists(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to get non-existent project
	_, err := pm.GetProjectByPath("/nonexistent/path")
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestListProjects(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Initially empty
	projects := pm.ListProjects()
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// Add some projects
	dir1 := createTestDir(t)
	defer os.RemoveAll(dir1)
	dir2 := createTestDir(t)
	defer os.RemoveAll(dir2)
	dir3 := createTestDir(t)
	defer os.RemoveAll(dir3)

	pm.AddProject(dir1)
	pm.AddProject(dir2)
	pm.AddProject(dir3)

	// List projects
	projects = pm.ListProjects()
	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect original)
	projects[0].Name = "modified"
	if pm.projects[0].Name == "modified" {
		t.Errorf("ListProjects() should return a copy, not a reference")
	}
}

func TestGetRecentProjects_WithLimit(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add 5 projects
	for i := 0; i < 5; i++ {
		dir := createTestDir(t)
		defer os.RemoveAll(dir)
		pm.AddProject(dir)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Get recent projects with limit
	recent := pm.GetRecentProjects(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent projects, got %d", len(recent))
	}

	// Verify it's a copy
	recent[0].Name = "modified"
	if pm.projects[0].Name == "modified" {
		t.Errorf("GetRecentProjects() should return a copy, not a reference")
	}
}

func TestGetRecentProjects_NegativeLimit(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add 3 projects
	for i := 0; i < 3; i++ {
		dir := createTestDir(t)
		defer os.RemoveAll(dir)
		pm.AddProject(dir)
	}

	// Get with negative limit (should return all)
	recent := pm.GetRecentProjects(-1)
	if len(recent) != 3 {
		t.Errorf("Expected 3 projects with negative limit, got %d", len(recent))
	}
}

func TestGetRecentProjects_ExceedsAvailable(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add 2 projects
	for i := 0; i < 2; i++ {
		dir := createTestDir(t)
		defer os.RemoveAll(dir)
		pm.AddProject(dir)
	}

	// Request more than available
	recent := pm.GetRecentProjects(10)
	if len(recent) != 2 {
		t.Errorf("Expected 2 projects when requesting 10, got %d", len(recent))
	}
}

func TestGetRecentProjects_Sorting(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add projects with different times
	dir1 := createTestDir(t)
	defer os.RemoveAll(dir1)
	dir2 := createTestDir(t)
	defer os.RemoveAll(dir2)
	dir3 := createTestDir(t)
	defer os.RemoveAll(dir3)

	pm.AddProject(dir1)
	time.Sleep(10 * time.Millisecond)
	pm.AddProject(dir2)
	time.Sleep(10 * time.Millisecond)
	pm.AddProject(dir3)

	// Get recent projects
	recent := pm.GetRecentProjects(3)

	// Verify they're in reverse chronological order (most recent first)
	// dir3 should be first since it was added last
	if recent[0].Path != dir3 {
		t.Errorf("Expected most recent project first, got %v", recent[0].Path)
	}
	if recent[2].Path != dir1 {
		t.Errorf("Expected oldest project last, got %v", recent[2].Path)
	}
}

func TestRecentProjectsLimit(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add more projects than the limit (10)
	for i := 0; i < 15; i++ {
		dir := createTestDir(t)
		defer os.RemoveAll(dir)
		pm.AddProject(dir)
	}

	// Verify only recentLimit projects are kept
	if len(pm.projects) != 10 {
		t.Errorf("Expected projects to be limited to %d, got %d", pm.recentLimit, len(pm.projects))
	}
}

func TestUpdateProject(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add a project
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	original, _ := pm.AddProject(projectDir)

	// Update project metadata
	updated := Project{
		ID:          original.ID,
		Name:        "Updated Name",
		Path:        original.Path,
		Description: "Test description",
		LastOpened:  time.Now(),
		CreatedAt:   original.CreatedAt,
	}

	err := pm.UpdateProject(updated)
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// Verify update
	project, _ := pm.GetProject(original.ID)
	if project.Name != "Updated Name" {
		t.Errorf("Expected Name = 'Updated Name', got %v", project.Name)
	}
	if project.Description != "Test description" {
		t.Errorf("Expected Description = 'Test description', got %v", project.Description)
	}

	// Verify it was saved to disk
	data, err := os.ReadFile(pm.storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	var saved []Project
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved projects: %v", err)
	}

	if len(saved) != 1 {
		t.Fatalf("Expected 1 saved project, got %d", len(saved))
	}

	if saved[0].Name != "Updated Name" {
		t.Errorf("Expected saved Name = 'Updated Name', got %v", saved[0].Name)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to update non-existent project
	project := Project{
		ID:   "nonexistent-id",
		Name: "Test",
		Path: "/test/path",
	}

	err := pm.UpdateProject(project)
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestValidatePath_ValidDirectory(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create a test directory
	testDir := createTestDir(t)
	defer os.RemoveAll(testDir)

	// Validate path
	if !pm.ValidatePath(testDir) {
		t.Errorf("Expected ValidatePath() = true for valid directory")
	}
}

func TestValidatePath_InvalidPath(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Test non-existent path
	if pm.ValidatePath("/nonexistent/path/12345") {
		t.Errorf("Expected ValidatePath() = false for non-existent path")
	}
}

func TestValidatePath_File(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create a file
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Validate file path (should be false)
	if pm.ValidatePath(testFile) {
		t.Errorf("Expected ValidatePath() = false for file path")
	}
}

func TestProjectMetadata(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	// Add project
	project, err := pm.AddProject(projectDir)
	if err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}

	// Verify metadata fields
	if project.ID == "" {
		t.Errorf("Expected non-empty ID")
	}

	expectedName := filepath.Base(projectDir)
	if project.Name != expectedName {
		t.Errorf("Expected Name = %v, got %v", expectedName, project.Name)
	}

	if project.Path != projectDir {
		t.Errorf("Expected Path = %v, got %v", projectDir, project.Path)
	}

	if project.Description != "" {
		t.Errorf("Expected empty Description, got %v", project.Description)
	}

	if project.LastOpened.IsZero() {
		t.Errorf("Expected non-zero LastOpened")
	}

	if project.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}

	// Verify timestamps are reasonable (within last minute)
	now := time.Now()
	if now.Sub(project.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt timestamp seems incorrect: %v", project.CreatedAt)
	}
	if now.Sub(project.LastOpened) > time.Minute {
		t.Errorf("LastOpened timestamp seems incorrect: %v", project.LastOpened)
	}
}

func TestProjectPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "boatman-persist-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "projects.json")

	// Create first project manager and add projects
	pm1 := &ProjectManager{
		projects:    []Project{},
		recentLimit: 10,
		storagePath: storagePath,
	}

	dir1 := createTestDir(t)
	defer os.RemoveAll(dir1)
	dir2 := createTestDir(t)
	defer os.RemoveAll(dir2)

	project1, _ := pm1.AddProject(dir1)
	project2, _ := pm1.AddProject(dir2)

	// Create second project manager (simulating app restart)
	pm2 := &ProjectManager{
		projects:    []Project{},
		recentLimit: 10,
		storagePath: storagePath,
	}

	// Load projects
	if err := pm2.load(); err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}

	// Verify projects were loaded
	if len(pm2.projects) != 2 {
		t.Errorf("Expected 2 loaded projects, got %d", len(pm2.projects))
	}

	// Verify project details
	loaded1, _ := pm2.GetProject(project1.ID)
	if loaded1.Path != project1.Path {
		t.Errorf("Expected Path = %v, got %v", project1.Path, loaded1.Path)
	}

	loaded2, _ := pm2.GetProject(project2.ID)
	if loaded2.Path != project2.Path {
		t.Errorf("Expected Path = %v, got %v", project2.Path, loaded2.Path)
	}
}

func TestLoadProjects_FileNotExists(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Try to load non-existent file
	err := pm.load()
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got %v", err)
	}
}

func TestLoadProjects_CorruptedJSON(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Write corrupted JSON
	corruptedData := []byte(`[{"id": "test", invalid json}]`)
	if err := os.WriteFile(pm.storagePath, corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted data: %v", err)
	}

	// Try to load
	err := pm.load()
	if err == nil {
		t.Errorf("Expected error when loading corrupted JSON, got nil")
	}
}

func TestSaveProjects_Success(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Add a project
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	pm.AddProject(projectDir)

	// Verify file was created
	if _, err := os.Stat(pm.storagePath); os.IsNotExist(err) {
		t.Errorf("Storage file was not created")
	}

	// Read and verify JSON format
	data, err := os.ReadFile(pm.storagePath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		t.Fatalf("Failed to unmarshal projects: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 saved project, got %d", len(projects))
	}
}

func TestProjectConcurrency(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create test directory
	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	pm.AddProject(projectDir)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			project, _ := pm.GetProject(pm.projects[0].ID)
			project.Name = "concurrent-test"
			pm.UpdateProject(*project)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			pm.ListProjects()
			pm.GetRecentProjects(5)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without panic, the test passes
}

func TestProjectIDFormat(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	projectDir := createTestDir(t)
	defer os.RemoveAll(projectDir)

	project, _ := pm.AddProject(projectDir)

	// Verify ID format (should be basename-timestamp)
	expectedPrefix := filepath.Base(projectDir) + "-"
	if len(project.ID) <= len(expectedPrefix) {
		t.Errorf("Project ID seems too short: %v", project.ID)
	}

	// ID should start with directory basename
	if project.ID[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected ID to start with %v, got %v", expectedPrefix, project.ID)
	}
}

func TestMultipleProjectsSameName(t *testing.T) {
	pm, tempDir := setupTestProjectManager(t)
	defer os.RemoveAll(tempDir)

	// Create two directories with the same basename in different locations
	dir1 := filepath.Join(tempDir, "test1", "myproject")
	dir2 := filepath.Join(tempDir, "test2", "myproject")

	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Add both projects with enough delay to ensure different timestamps
	project1, _ := pm.AddProject(dir1)
	time.Sleep(1 * time.Second) // Ensure different timestamps (format is down to seconds)
	project2, _ := pm.AddProject(dir2)

	// Both should have the same name but different IDs (unless timestamp resolution is too coarse)
	if project1.Name != project2.Name {
		t.Errorf("Expected same names, got %v and %v", project1.Name, project2.Name)
	}

	if project1.ID == project2.ID {
		t.Errorf("Expected different IDs, got same: %v", project1.ID)
	}

	// Both should be in the list
	if len(pm.projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(pm.projects))
	}
}

func TestProjectManagerStoragePath(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "boatman-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set temporary HOME
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create project manager
	pm, err := NewProjectManager()
	if err != nil {
		t.Fatalf("NewProjectManager() error = %v", err)
	}

	expectedPath := filepath.Join(tempHome, ".boatman", "projects.json")
	if pm.storagePath != expectedPath {
		t.Errorf("Expected storage path = %v, got %v", expectedPath, pm.storagePath)
	}
}
