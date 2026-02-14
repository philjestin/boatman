package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupTestConfig creates a Config instance with a temporary config path
func setupTestConfig(t *testing.T) (*Config, string) {
	t.Helper()

	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "boatman-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	config := &Config{
		configPath: filepath.Join(tempDir, "config.json"),
		preferences: UserPreferences{
			AuthMethod:           AuthMethodAnthropicAPI,
			GCPRegion:            "us-east5",
			ApprovalMode:         ApprovalModeSuggest,
			DefaultModel:         "sonnet",
			Theme:                ThemeDark,
			NotificationsEnabled: true,
			MCPServers:           []MCPServer{},
			OnboardingCompleted:  false,
		},
		projects: make(map[string]ProjectPreferences),
	}

	return config, tempDir
}

func TestNewConfig(t *testing.T) {
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

	// Create new config
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	// Verify config directory was created
	configDir := filepath.Join(tempHome, ".boatman")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %v", err)
	}

	// Verify default values
	if cfg.preferences.AuthMethod != AuthMethodAnthropicAPI {
		t.Errorf("Expected default AuthMethod = %v, got %v", AuthMethodAnthropicAPI, cfg.preferences.AuthMethod)
	}
	if cfg.preferences.GCPRegion != "us-east5" {
		t.Errorf("Expected default GCPRegion = us-east5, got %v", cfg.preferences.GCPRegion)
	}
	if cfg.preferences.ApprovalMode != ApprovalModeSuggest {
		t.Errorf("Expected default ApprovalMode = %v, got %v", ApprovalModeSuggest, cfg.preferences.ApprovalMode)
	}
	if cfg.preferences.DefaultModel != "sonnet" {
		t.Errorf("Expected default DefaultModel = sonnet, got %v", cfg.preferences.DefaultModel)
	}
	if cfg.preferences.Theme != ThemeDark {
		t.Errorf("Expected default Theme = %v, got %v", ThemeDark, cfg.preferences.Theme)
	}
	if !cfg.preferences.NotificationsEnabled {
		t.Errorf("Expected default NotificationsEnabled = true, got false")
	}
	if cfg.preferences.OnboardingCompleted {
		t.Errorf("Expected default OnboardingCompleted = false, got true")
	}
	if cfg.preferences.MCPServers == nil {
		t.Errorf("Expected MCPServers to be initialized, got nil")
	}
}

func TestLoadPreferences_FileExists(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Create a config file with test data
	testPrefs := UserPreferences{
		APIKey:               "test-api-key",
		AuthMethod:           AuthMethodGoogleCloud,
		GCPProjectID:         "test-project",
		GCPRegion:            "us-west1",
		ApprovalMode:         ApprovalModeFullAuto,
		DefaultModel:         "opus",
		Theme:                ThemeLight,
		NotificationsEnabled: false,
		MCPServers: []MCPServer{
			{
				Name:    "test-server",
				Command: "test-command",
				Args:    []string{"arg1", "arg2"},
				Env:     map[string]string{"KEY": "value"},
				Enabled: true,
			},
		},
		OnboardingCompleted: true,
	}

	testProjects := map[string]ProjectPreferences{
		"/test/path": {
			ProjectPath:  "/test/path",
			ApprovalMode: ApprovalModeAutoEdit,
			Model:        "haiku",
		},
	}

	data, _ := json.MarshalIndent(struct {
		Preferences UserPreferences               `json:"preferences"`
		Projects    map[string]ProjectPreferences `json:"projects"`
	}{
		Preferences: testPrefs,
		Projects:    testProjects,
	}, "", "  ")

	if err := os.WriteFile(cfg.configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	if err := cfg.load(); err != nil {
		t.Fatalf("load() error = %v", err)
	}

	// Verify loaded preferences
	if cfg.preferences.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey = test-api-key, got %v", cfg.preferences.APIKey)
	}
	if cfg.preferences.AuthMethod != AuthMethodGoogleCloud {
		t.Errorf("Expected AuthMethod = %v, got %v", AuthMethodGoogleCloud, cfg.preferences.AuthMethod)
	}
	if cfg.preferences.GCPProjectID != "test-project" {
		t.Errorf("Expected GCPProjectID = test-project, got %v", cfg.preferences.GCPProjectID)
	}
	if cfg.preferences.OnboardingCompleted != true {
		t.Errorf("Expected OnboardingCompleted = true, got false")
	}

	// Verify MCP servers
	if len(cfg.preferences.MCPServers) != 1 {
		t.Errorf("Expected 1 MCP server, got %d", len(cfg.preferences.MCPServers))
	} else {
		server := cfg.preferences.MCPServers[0]
		if server.Name != "test-server" {
			t.Errorf("Expected server name = test-server, got %v", server.Name)
		}
	}

	// Verify projects
	if len(cfg.projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(cfg.projects))
	}
	proj, ok := cfg.projects["/test/path"]
	if !ok {
		t.Errorf("Expected project at /test/path not found")
	} else {
		if proj.ApprovalMode != ApprovalModeAutoEdit {
			t.Errorf("Expected project ApprovalMode = %v, got %v", ApprovalModeAutoEdit, proj.ApprovalMode)
		}
	}
}

func TestLoadPreferences_FileNotExists(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Try to load non-existent file
	err := cfg.load()
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got %v", err)
	}
}

func TestLoadPreferences_CorruptedJSON(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Write corrupted JSON
	corruptedData := []byte(`{"preferences": {"apiKey": "test", invalid json}`)
	if err := os.WriteFile(cfg.configPath, corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	// Try to load corrupted file
	err := cfg.load()
	if err == nil {
		t.Errorf("Expected error when loading corrupted JSON, got nil")
	}

	// Verify it's a JSON error
	if _, ok := err.(*json.SyntaxError); !ok {
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			t.Logf("Got error type: %T", err)
		}
	}
}

func TestSavePreferences_Success(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Modify preferences
	cfg.preferences.APIKey = "new-api-key"
	cfg.preferences.OnboardingCompleted = true

	// Save
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cfg.configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Read and verify contents
	data, err := os.ReadFile(cfg.configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var saved struct {
		Preferences UserPreferences               `json:"preferences"`
		Projects    map[string]ProjectPreferences `json:"projects"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if saved.Preferences.APIKey != "new-api-key" {
		t.Errorf("Expected saved APIKey = new-api-key, got %v", saved.Preferences.APIKey)
	}
	if saved.Preferences.OnboardingCompleted != true {
		t.Errorf("Expected saved OnboardingCompleted = true, got false")
	}
}

func TestSavePreferences_PermissionError(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Make directory read-only to cause permission error
	if err := os.Chmod(tempDir, 0444); err != nil {
		t.Skipf("Cannot set directory permissions: %v", err)
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	// Try to save
	err := cfg.Save()
	if err == nil {
		t.Errorf("Expected permission error, got nil")
	}
}

func TestGetPreferences(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Set some preferences
	cfg.preferences.APIKey = "test-key"
	cfg.preferences.Theme = ThemeLight

	// Get preferences
	prefs := cfg.GetPreferences()

	// Verify values
	if prefs.APIKey != "test-key" {
		t.Errorf("Expected APIKey = test-key, got %v", prefs.APIKey)
	}
	if prefs.Theme != ThemeLight {
		t.Errorf("Expected Theme = %v, got %v", ThemeLight, prefs.Theme)
	}

	// Verify it's a copy (modify returned value shouldn't affect original)
	prefs.APIKey = "modified"
	if cfg.preferences.APIKey == "modified" {
		t.Errorf("GetPreferences() should return a copy, not a reference")
	}
}

func TestSetPreferences(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Create new preferences
	newPrefs := UserPreferences{
		APIKey:               "updated-key",
		AuthMethod:           AuthMethodGoogleCloud,
		GCPProjectID:         "gcp-project",
		ApprovalMode:         ApprovalModeFullAuto,
		DefaultModel:         "opus",
		Theme:                ThemeLight,
		NotificationsEnabled: false,
		MCPServers:           []MCPServer{},
		OnboardingCompleted:  true,
	}

	// Set preferences
	if err := cfg.SetPreferences(newPrefs); err != nil {
		t.Fatalf("SetPreferences() error = %v", err)
	}

	// Verify preferences were updated
	if cfg.preferences.APIKey != "updated-key" {
		t.Errorf("Expected APIKey = updated-key, got %v", cfg.preferences.APIKey)
	}
	if cfg.preferences.AuthMethod != AuthMethodGoogleCloud {
		t.Errorf("Expected AuthMethod = %v, got %v", AuthMethodGoogleCloud, cfg.preferences.AuthMethod)
	}

	// Verify preferences were saved to disk
	data, err := os.ReadFile(cfg.configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var saved struct {
		Preferences UserPreferences `json:"preferences"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if saved.Preferences.APIKey != "updated-key" {
		t.Errorf("Expected saved APIKey = updated-key, got %v", saved.Preferences.APIKey)
	}
}

func TestIsOnboardingCompleted(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Initially should be false
	if cfg.IsOnboardingCompleted() {
		t.Errorf("Expected IsOnboardingCompleted() = false, got true")
	}

	// Set to true
	cfg.preferences.OnboardingCompleted = true
	if !cfg.IsOnboardingCompleted() {
		t.Errorf("Expected IsOnboardingCompleted() = true, got false")
	}
}

func TestCompleteOnboarding(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Complete onboarding
	if err := cfg.CompleteOnboarding(); err != nil {
		t.Fatalf("CompleteOnboarding() error = %v", err)
	}

	// Verify flag is set
	if !cfg.preferences.OnboardingCompleted {
		t.Errorf("Expected OnboardingCompleted = true, got false")
	}

	// Verify it was saved to disk
	data, err := os.ReadFile(cfg.configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var saved struct {
		Preferences UserPreferences `json:"preferences"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if !saved.Preferences.OnboardingCompleted {
		t.Errorf("Expected saved OnboardingCompleted = true, got false")
	}
}

func TestGetAPIKey(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	testKey := "test-api-key-12345"
	cfg.preferences.APIKey = testKey

	result := cfg.GetAPIKey()
	if result != testKey {
		t.Errorf("Expected GetAPIKey() = %v, got %v", testKey, result)
	}
}

func TestGetAuthMethod(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Test default
	if cfg.GetAuthMethod() != AuthMethodAnthropicAPI {
		t.Errorf("Expected default AuthMethod = %v, got %v", AuthMethodAnthropicAPI, cfg.GetAuthMethod())
	}

	// Test updated value
	cfg.preferences.AuthMethod = AuthMethodGoogleCloud
	if cfg.GetAuthMethod() != AuthMethodGoogleCloud {
		t.Errorf("Expected AuthMethod = %v, got %v", AuthMethodGoogleCloud, cfg.GetAuthMethod())
	}
}

func TestGetGCPConfig(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	cfg.preferences.GCPProjectID = "test-project"
	cfg.preferences.GCPRegion = "us-west1"

	projectID, region := cfg.GetGCPConfig()

	if projectID != "test-project" {
		t.Errorf("Expected GCPProjectID = test-project, got %v", projectID)
	}
	if region != "us-west1" {
		t.Errorf("Expected GCPRegion = us-west1, got %v", region)
	}
}

func TestGetMCPServers(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	testServers := []MCPServer{
		{
			Name:    "server1",
			Command: "cmd1",
			Args:    []string{"arg1"},
			Env:     map[string]string{"KEY1": "val1"},
			Enabled: true,
		},
		{
			Name:    "server2",
			Command: "cmd2",
			Args:    []string{"arg2", "arg3"},
			Enabled: false,
		},
	}

	cfg.preferences.MCPServers = testServers

	result := cfg.GetMCPServers()

	if len(result) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(result))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect original)
	result[0].Name = "modified"
	if cfg.preferences.MCPServers[0].Name == "modified" {
		t.Errorf("GetMCPServers() should return a copy, not a reference")
	}
}

func TestSetMCPServers(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	testServers := []MCPServer{
		{
			Name:    "test-server",
			Command: "test-cmd",
			Args:    []string{"--arg"},
			Enabled: true,
		},
	}

	if err := cfg.SetMCPServers(testServers); err != nil {
		t.Fatalf("SetMCPServers() error = %v", err)
	}

	// Verify servers were updated
	if len(cfg.preferences.MCPServers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(cfg.preferences.MCPServers))
	}

	// Verify it was saved to disk
	data, err := os.ReadFile(cfg.configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var saved struct {
		Preferences UserPreferences `json:"preferences"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if len(saved.Preferences.MCPServers) != 1 {
		t.Errorf("Expected 1 saved server, got %d", len(saved.Preferences.MCPServers))
	}
}

func TestGetProjectPreferences(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	testPath := "/test/project/path"
	testPrefs := ProjectPreferences{
		ProjectPath:  testPath,
		ApprovalMode: ApprovalModeAutoEdit,
		Model:        "opus",
	}

	cfg.projects[testPath] = testPrefs

	result := cfg.GetProjectPreferences(testPath)

	if result.ProjectPath != testPath {
		t.Errorf("Expected ProjectPath = %v, got %v", testPath, result.ProjectPath)
	}
	if result.ApprovalMode != ApprovalModeAutoEdit {
		t.Errorf("Expected ApprovalMode = %v, got %v", ApprovalModeAutoEdit, result.ApprovalMode)
	}
	if result.Model != "opus" {
		t.Errorf("Expected Model = opus, got %v", result.Model)
	}

	// Test non-existent project (should return zero value)
	result = cfg.GetProjectPreferences("/nonexistent")
	if result.ProjectPath != "" {
		t.Errorf("Expected empty ProjectPreferences for non-existent project")
	}
}

func TestSetProjectPreferences(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	testPrefs := ProjectPreferences{
		ProjectPath:  "/test/path",
		ApprovalMode: ApprovalModeFullAuto,
		Model:        "haiku",
	}

	if err := cfg.SetProjectPreferences(testPrefs); err != nil {
		t.Fatalf("SetProjectPreferences() error = %v", err)
	}

	// Verify preferences were set
	result := cfg.projects[testPrefs.ProjectPath]
	if result.ApprovalMode != ApprovalModeFullAuto {
		t.Errorf("Expected ApprovalMode = %v, got %v", ApprovalModeFullAuto, result.ApprovalMode)
	}

	// Verify it was saved to disk
	data, err := os.ReadFile(cfg.configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var saved struct {
		Projects map[string]ProjectPreferences `json:"projects"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	savedProj, ok := saved.Projects[testPrefs.ProjectPath]
	if !ok {
		t.Errorf("Expected project at %v not found in saved config", testPrefs.ProjectPath)
	} else {
		if savedProj.Model != "haiku" {
			t.Errorf("Expected saved Model = haiku, got %v", savedProj.Model)
		}
	}
}

func TestConfigFilePath(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "boatman-filepath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set temporary HOME
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create config
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	expectedPath := filepath.Join(tempHome, ".boatman", "config.json")
	if cfg.configPath != expectedPath {
		t.Errorf("Expected config path = %v, got %v", expectedPath, cfg.configPath)
	}
}

func TestConfigDefaultValues(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"AuthMethod", cfg.preferences.AuthMethod, AuthMethodAnthropicAPI},
		{"GCPRegion", cfg.preferences.GCPRegion, "us-east5"},
		{"ApprovalMode", cfg.preferences.ApprovalMode, ApprovalModeSuggest},
		{"DefaultModel", cfg.preferences.DefaultModel, "sonnet"},
		{"Theme", cfg.preferences.Theme, ThemeDark},
		{"NotificationsEnabled", cfg.preferences.NotificationsEnabled, true},
		{"OnboardingCompleted", cfg.preferences.OnboardingCompleted, false},
		{"APIKey", cfg.preferences.APIKey, ""},
		{"GCPProjectID", cfg.preferences.GCPProjectID, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Default %s = %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Test MCPServers is initialized but empty
	if cfg.preferences.MCPServers == nil {
		t.Errorf("Expected MCPServers to be initialized (non-nil)")
	}
	if len(cfg.preferences.MCPServers) != 0 {
		t.Errorf("Expected MCPServers to be empty, got %d servers", len(cfg.preferences.MCPServers))
	}

	// Test projects map is initialized
	if cfg.projects == nil {
		t.Errorf("Expected projects map to be initialized (non-nil)")
	}
}

func TestConfigConcurrency(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cfg.SetPreferences(UserPreferences{
				APIKey:       "concurrent-test",
				AuthMethod:   AuthMethodAnthropicAPI,
				ApprovalMode: ApprovalModeSuggest,
				DefaultModel: "sonnet",
				Theme:        ThemeDark,
				MCPServers:   []MCPServer{},
			})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cfg.GetPreferences()
			cfg.IsOnboardingCompleted()
			cfg.GetAPIKey()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without panic, the test passes
}

func TestLoadPreferences_NilProjects(t *testing.T) {
	cfg, tempDir := setupTestConfig(t)
	defer os.RemoveAll(tempDir)

	// Create config with null projects
	data := []byte(`{
		"preferences": {
			"apiKey": "",
			"authMethod": "anthropic-api",
			"approvalMode": "suggest",
			"defaultModel": "sonnet",
			"theme": "dark",
			"notificationsEnabled": true,
			"mcpServers": [],
			"onboardingCompleted": false
		},
		"projects": null
	}`)

	if err := os.WriteFile(cfg.configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	if err := cfg.load(); err != nil {
		t.Fatalf("load() error = %v", err)
	}

	// Verify projects map is initialized
	if cfg.projects == nil {
		t.Errorf("Expected projects map to be initialized after loading null projects")
	}
}
