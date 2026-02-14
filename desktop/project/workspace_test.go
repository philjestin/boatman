package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorkspace(t *testing.T) {
	path := "/test/path"
	ws := NewWorkspace(path)

	if ws == nil {
		t.Fatal("NewWorkspace() returned nil")
	}

	if ws.GetPath() != path {
		t.Errorf("Expected path '%s', got '%s'", path, ws.GetPath())
	}
}

func TestGetPath(t *testing.T) {
	paths := []string{
		"/test/path",
		"/another/path",
		"relative/path",
		"",
	}

	for _, path := range paths {
		ws := NewWorkspace(path)
		if ws.GetPath() != path {
			t.Errorf("Expected path '%s', got '%s'", path, ws.GetPath())
		}
	}
}

func TestGetInfo_EmptyWorkspace(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if info.Path != tmpDir {
		t.Errorf("Expected path '%s', got '%s'", tmpDir, info.Path)
	}

	expectedName := filepath.Base(tmpDir)
	if info.Name != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, info.Name)
	}

	if info.IsGitRepo {
		t.Error("Expected IsGitRepo to be false for empty workspace")
	}

	if info.HasPackage {
		t.Error("Expected HasPackage to be false for empty workspace")
	}

	if len(info.Languages) != 0 {
		t.Errorf("Expected no languages, got %v", info.Languages)
	}
}

func TestGetInfo_GitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !info.IsGitRepo {
		t.Error("Expected IsGitRepo to be true when .git directory exists")
	}
}

func TestGetInfo_NodeJsProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !info.HasPackage {
		t.Error("Expected HasPackage to be true when package.json exists")
	}

	if !containsString(info.Languages, "javascript") {
		t.Errorf("Expected languages to contain 'javascript', got %v", info.Languages)
	}
}

func TestGetInfo_GoProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !containsString(info.Languages, "go") {
		t.Errorf("Expected languages to contain 'go', got %v", info.Languages)
	}
}

func TestGetInfo_RustProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Cargo.toml
	cargoToml := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoToml, []byte("[package]"), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !containsString(info.Languages, "rust") {
		t.Errorf("Expected languages to contain 'rust', got %v", info.Languages)
	}
}

func TestGetInfo_PythonProject_PyprojectToml(t *testing.T) {
	tmpDir := t.TempDir()

	// Create pyproject.toml
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	if err := os.WriteFile(pyproject, []byte("[tool.poetry]"), 0644); err != nil {
		t.Fatalf("Failed to create pyproject.toml: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !containsString(info.Languages, "python") {
		t.Errorf("Expected languages to contain 'python', got %v", info.Languages)
	}
}

func TestGetInfo_PythonProject_RequirementsTxt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create requirements.txt
	requirements := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(requirements, []byte("requests==2.0.0"), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	if !containsString(info.Languages, "python") {
		t.Errorf("Expected languages to contain 'python', got %v", info.Languages)
	}
}

func TestGetInfo_MultiLanguageProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple language markers
	files := map[string]string{
		"package.json":     "{}",
		"go.mod":           "module test",
		"Cargo.toml":       "[package]",
		"requirements.txt": "requests==2.0.0",
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	ws := NewWorkspace(tmpDir)
	info, err := ws.GetInfo()

	if err != nil {
		t.Fatalf("GetInfo() failed: %v", err)
	}

	expectedLanguages := []string{"javascript", "go", "rust", "python"}
	if len(info.Languages) != len(expectedLanguages) {
		t.Errorf("Expected %d languages, got %d: %v", len(expectedLanguages), len(info.Languages), info.Languages)
	}

	for _, lang := range expectedLanguages {
		if !containsString(info.Languages, lang) {
			t.Errorf("Expected languages to contain '%s', got %v", lang, info.Languages)
		}
	}
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	testFiles := []string{"file1.txt", "file2.go", "file3.md"}
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	files, err := ws.ListFiles()

	if err != nil {
		t.Fatalf("ListFiles() failed: %v", err)
	}

	// Should include files and subdirectory
	expectedCount := len(testFiles) + 1
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d: %v", expectedCount, len(files), files)
	}

	for _, expectedFile := range testFiles {
		if !containsString(files, expectedFile) {
			t.Errorf("Expected files to contain '%s', got %v", expectedFile, files)
		}
	}

	if !containsString(files, "subdir") {
		t.Errorf("Expected files to contain 'subdir', got %v", files)
	}
}

func TestListFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	ws := NewWorkspace(tmpDir)
	files, err := ws.ListFiles()

	if err != nil {
		t.Fatalf("ListFiles() failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected no files in empty directory, got %v", files)
	}
}

func TestListFiles_NonexistentDirectory(t *testing.T) {
	ws := NewWorkspace("/nonexistent/directory/xyz123")
	_, err := ws.ListFiles()

	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testContent := "Hello, World!\nThis is a test file."
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	content, err := ws.ReadFile("test.txt")

	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}
}

func TestReadFile_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	ws := NewWorkspace(tmpDir)
	_, err := ws.ReadFile("nonexistent.txt")

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestReadFile_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory and file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testContent := "Nested file content"
	testFile := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	content, err := ws.ReadFile("subdir/nested.txt")

	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	ws := NewWorkspace(tmpDir)
	testContent := "Test content to write"

	err := ws.WriteFile("output.txt", testContent)
	if err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Verify file was created with correct content
	fullPath := filepath.Join(tmpDir, "output.txt")
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected written content '%s', got '%s'", testContent, string(content))
	}
}

func TestWriteFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial file
	initialContent := "Initial content"
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Overwrite with new content
	ws := NewWorkspace(tmpDir)
	newContent := "New content"
	err := ws.WriteFile("test.txt", newContent)

	if err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Verify content was overwritten
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, string(content))
	}
}

func TestWriteFile_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	testContent := "Nested file content"

	err := ws.WriteFile("subdir/nested.txt", testContent)
	if err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	// Verify file was created
	fullPath := filepath.Join(tmpDir, "subdir", "nested.txt")
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected written content '%s', got '%s'", testContent, string(content))
	}
}

func TestFileExists_Exists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	if !ws.FileExists("exists.txt") {
		t.Error("FileExists() returned false for existing file")
	}
}

func TestFileExists_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	ws := NewWorkspace(tmpDir)
	if ws.FileExists("nonexistent.txt") {
		t.Error("FileExists() returned true for nonexistent file")
	}
}

func TestFileExists_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	if !ws.FileExists("subdir") {
		t.Error("FileExists() returned false for existing directory")
	}
}

func TestFileExists_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory and file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFile := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	ws := NewWorkspace(tmpDir)
	if !ws.FileExists("subdir/nested.txt") {
		t.Error("FileExists() returned false for existing nested file")
	}

	if ws.FileExists("subdir/nonexistent.txt") {
		t.Error("FileExists() returned true for nonexistent nested file")
	}
}

// Helper function to check if a slice contains a string
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
