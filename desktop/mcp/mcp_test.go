package mcp

// Comprehensive test suite for the MCP manager module.
// This test file covers:
// - Manager initialization and configuration paths
// - CRUD operations on MCP servers (Get, Add, Remove, Update)
// - Config file reading/writing (~/.claude/claude_mcp_config.json)
// - Server validation and preset servers
// - JSON marshaling/unmarshaling
// - Error handling (file permissions, corrupted JSON, invalid data)
// - Edge cases (empty configs, special characters, concurrent access, large configs)

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestNewManager tests the NewManager initialization
func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	if manager.configPath == "" {
		t.Error("NewManager() returned manager with empty configPath")
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".claude", "claude_mcp_config.json")
	if manager.configPath != expectedPath {
		t.Errorf("NewManager() configPath = %v, want %v", manager.configPath, expectedPath)
	}
}

// TestGetServers_Empty tests GetServers with no config file
func TestGetServers_Empty(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() on non-existent config failed: %v", err)
	}

	if servers == nil {
		t.Fatal("GetServers() returned nil servers slice")
	}

	if len(servers) != 0 {
		t.Errorf("GetServers() on non-existent config returned %d servers, want 0", len(servers))
	}
}

// TestGetServers_Populated tests GetServers with populated config
func TestGetServers_Populated(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Create test config
	testConfig := Config{
		McpServers: map[string]ServerDef{
			"test-server-1": {
				Command: "npx",
				Args:    []string{"-y", "@test/server1"},
				Env: map[string]string{
					"API_KEY": "test123",
				},
			},
			"test-server-2": {
				Command: "python",
				Args:    []string{"-m", "server2"},
			},
		},
	}

	// Write test config
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	manager := &Manager{
		configPath: configPath,
	}

	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 2 {
		t.Errorf("GetServers() returned %d servers, want 2", len(servers))
	}

	// Verify server details
	serverMap := make(map[string]Server)
	for _, s := range servers {
		serverMap[s.Name] = s
	}

	server1, ok := serverMap["test-server-1"]
	if !ok {
		t.Error("GetServers() missing test-server-1")
	} else {
		if server1.Command != "npx" {
			t.Errorf("test-server-1 Command = %v, want npx", server1.Command)
		}
		if len(server1.Args) != 2 {
			t.Errorf("test-server-1 Args length = %d, want 2", len(server1.Args))
		}
		if server1.Env["API_KEY"] != "test123" {
			t.Errorf("test-server-1 Env[API_KEY] = %v, want test123", server1.Env["API_KEY"])
		}
		if !server1.Enabled {
			t.Error("test-server-1 Enabled = false, want true")
		}
	}
}

// TestAddServer_Success tests adding a new server successfully
func TestAddServer_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	newServer := Server{
		Name:        "new-server",
		Description: "Test server",
		Command:     "node",
		Args:        []string{"server.js"},
		Env: map[string]string{
			"PORT": "3000",
		},
		Enabled: true,
	}

	err := manager.AddServer(newServer)
	if err != nil {
		t.Fatalf("AddServer() failed: %v", err)
	}

	// Verify server was added
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() after AddServer failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers after AddServer, want 1", len(servers))
	}

	if servers[0].Name != "new-server" {
		t.Errorf("Added server name = %v, want new-server", servers[0].Name)
	}
	if servers[0].Command != "node" {
		t.Errorf("Added server command = %v, want node", servers[0].Command)
	}
}

// TestAddServer_Duplicate tests adding a server with duplicate name (should overwrite)
func TestAddServer_Duplicate(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add first server
	server1 := Server{
		Name:    "test-server",
		Command: "python",
		Args:    []string{"server1.py"},
	}

	err := manager.AddServer(server1)
	if err != nil {
		t.Fatalf("AddServer() first call failed: %v", err)
	}

	// Add second server with same name but different command
	server2 := Server{
		Name:    "test-server",
		Command: "node",
		Args:    []string{"server2.js"},
	}

	err = manager.AddServer(server2)
	if err != nil {
		t.Fatalf("AddServer() second call failed: %v", err)
	}

	// Verify the server was overwritten
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers, want 1 (duplicate should overwrite)", len(servers))
	}

	if servers[0].Command != "node" {
		t.Errorf("Server command = %v, want node (should be overwritten)", servers[0].Command)
	}
}

// TestRemoveServer_Exists tests removing an existing server
func TestRemoveServer_Exists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add servers
	server1 := Server{Name: "server1", Command: "cmd1"}
	server2 := Server{Name: "server2", Command: "cmd2"}

	manager.AddServer(server1)
	manager.AddServer(server2)

	// Remove server1
	err := manager.RemoveServer("server1")
	if err != nil {
		t.Fatalf("RemoveServer() failed: %v", err)
	}

	// Verify server1 was removed
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() after RemoveServer failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers after RemoveServer, want 1", len(servers))
	}

	if servers[0].Name != "server2" {
		t.Errorf("Remaining server name = %v, want server2", servers[0].Name)
	}
}

// TestRemoveServer_NotExists tests removing a non-existent server
func TestRemoveServer_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add a server
	server := Server{Name: "server1", Command: "cmd1"}
	manager.AddServer(server)

	// Try to remove non-existent server (should not error, just no-op)
	err := manager.RemoveServer("non-existent")
	if err != nil {
		t.Fatalf("RemoveServer() on non-existent server failed: %v", err)
	}

	// Verify original server still exists
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers, want 1", len(servers))
	}
}

// TestRemoveServer_EmptyConfig tests removing from non-existent config
func TestRemoveServer_EmptyConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non_existent_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Try to remove from non-existent config
	err := manager.RemoveServer("server1")
	if err == nil {
		t.Error("RemoveServer() on non-existent config should return error")
	}
}

// TestUpdateServer_Exists tests updating an existing server
func TestUpdateServer_Exists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add initial server
	originalServer := Server{
		Name:    "test-server",
		Command: "python",
		Args:    []string{"old.py"},
	}
	manager.AddServer(originalServer)

	// Update server
	updatedServer := Server{
		Name:    "test-server",
		Command: "node",
		Args:    []string{"new.js"},
		Env: map[string]string{
			"NEW_VAR": "value",
		},
	}

	err := manager.UpdateServer(updatedServer)
	if err != nil {
		t.Fatalf("UpdateServer() failed: %v", err)
	}

	// Verify update
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers, want 1", len(servers))
	}

	server := servers[0]
	if server.Command != "node" {
		t.Errorf("Updated server command = %v, want node", server.Command)
	}
	if server.Env["NEW_VAR"] != "value" {
		t.Errorf("Updated server Env[NEW_VAR] = %v, want value", server.Env["NEW_VAR"])
	}
}

// TestUpdateServer_NotExists tests updating a non-existent server (creates new)
func TestUpdateServer_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Update non-existent server (should create it)
	newServer := Server{
		Name:    "new-server",
		Command: "go",
		Args:    []string{"run", "main.go"},
	}

	err := manager.UpdateServer(newServer)
	if err != nil {
		t.Fatalf("UpdateServer() on non-existent server failed: %v", err)
	}

	// Verify server was created
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers, want 1", len(servers))
	}

	if servers[0].Name != "new-server" {
		t.Errorf("Created server name = %v, want new-server", servers[0].Name)
	}
}

// TestGetPresets tests the GetPresetServers function
func TestGetPresets(t *testing.T) {
	presets := GetPresetServers()

	if presets == nil {
		t.Fatal("GetPresetServers() returned nil")
	}

	if len(presets) == 0 {
		t.Error("GetPresetServers() returned empty slice")
	}

	// Verify common presets exist
	presetNames := make(map[string]bool)
	for _, preset := range presets {
		presetNames[preset.Name] = true

		// Verify all presets are disabled by default
		if preset.Enabled {
			t.Errorf("Preset %s is enabled by default, want disabled", preset.Name)
		}

		// Verify required fields
		if preset.Name == "" {
			t.Error("Preset has empty name")
		}
		if preset.Command == "" {
			t.Error("Preset has empty command")
		}
	}

	// Check for expected presets
	expectedPresets := []string{"filesystem", "github", "postgres"}
	for _, name := range expectedPresets {
		if !presetNames[name] {
			t.Errorf("GetPresetServers() missing expected preset: %s", name)
		}
	}
}

// TestConfigFileReading tests reading config from file
func TestConfigFileReading(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".claude", "claude_mcp_config.json")

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write test config
	testConfig := `{
  "mcpServers": {
    "test-server": {
      "command": "npx",
      "args": ["-y", "@test/server"],
      "env": {
        "TEST_VAR": "test_value"
      }
    }
  }
}`

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	manager := &Manager{
		configPath: configPath,
	}

	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("GetServers() returned %d servers, want 1", len(servers))
	}

	if servers[0].Name != "test-server" {
		t.Errorf("Server name = %v, want test-server", servers[0].Name)
	}
}

// TestConfigFileWriting tests writing config to file
func TestConfigFileWriting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".claude", "claude_mcp_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add server (should create directory and file)
	server := Server{
		Name:    "write-test",
		Command: "test-cmd",
		Args:    []string{"arg1", "arg2"},
		Env: map[string]string{
			"KEY": "value",
		},
	}

	err := manager.AddServer(server)
	if err != nil {
		t.Fatalf("AddServer() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if _, ok := config.McpServers["write-test"]; !ok {
		t.Error("Config file does not contain written server")
	}
}

// TestServerValidation_Valid tests ValidateServer with valid servers
func TestServerValidation_Valid(t *testing.T) {
	manager := &Manager{}

	validServers := []Server{
		{Name: "test", Command: "cmd"},
		{Name: "server", Command: "python", Args: []string{"-m", "server"}},
		{Name: "env-server", Command: "node", Env: map[string]string{"KEY": "val"}},
	}

	for _, server := range validServers {
		err := manager.ValidateServer(server)
		if err != nil {
			t.Errorf("ValidateServer(%+v) returned error: %v", server, err)
		}
	}
}

// TestServerValidation_Invalid tests ValidateServer with invalid servers
func TestServerValidation_Invalid(t *testing.T) {
	manager := &Manager{}

	invalidServers := []struct {
		server Server
		desc   string
	}{
		{Server{Name: "", Command: "cmd"}, "empty name"},
		{Server{Name: "test", Command: ""}, "empty command"},
		{Server{Name: "", Command: ""}, "empty name and command"},
	}

	for _, tc := range invalidServers {
		err := manager.ValidateServer(tc.server)
		if err == nil {
			t.Errorf("ValidateServer() with %s should return error", tc.desc)
		}
		if err != os.ErrInvalid {
			t.Errorf("ValidateServer() with %s returned %v, want os.ErrInvalid", tc.desc, err)
		}
	}
}

// TestJSONMarshaling tests JSON marshaling of Config and Server
func TestJSONMarshaling(t *testing.T) {
	testConfig := Config{
		McpServers: map[string]ServerDef{
			"test-server": {
				Command: "npx",
				Args:    []string{"-y", "@test/server"},
				Env: map[string]string{
					"API_KEY": "secret",
				},
			},
		},
	}

	// Marshal
	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Unmarshal
	var unmarshaledConfig Config
	if err := json.Unmarshal(data, &unmarshaledConfig); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Compare
	if !reflect.DeepEqual(testConfig.McpServers, unmarshaledConfig.McpServers) {
		t.Errorf("Unmarshaled config does not match original")
	}
}

// TestJSONUnmarshaling tests JSON unmarshaling edge cases
func TestJSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid config",
			json:    `{"mcpServers": {"test": {"command": "cmd"}}}`,
			wantErr: false,
		},
		{
			name:    "empty mcpServers",
			json:    `{"mcpServers": {}}`,
			wantErr: false,
		},
		{
			name:    "missing mcpServers field",
			json:    `{}`,
			wantErr: false, // Should default to empty map
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "incomplete json",
			json:    `{"mcpServers": {`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			err := json.Unmarshal([]byte(tt.json), &config)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestErrorHandling_FilePermissions tests handling of permission errors
func TestErrorHandling_FilePermissions(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "restricted", "config.json")

	// Create directory with no write permissions
	restrictedDir := filepath.Join(tempDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0555); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	manager := &Manager{
		configPath: configPath,
	}

	server := Server{Name: "test", Command: "cmd"}
	err := manager.AddServer(server)
	if err == nil {
		t.Error("AddServer() with read-only directory should return error")
	}
}

// TestErrorHandling_CorruptedJSON tests handling of corrupted JSON files
func TestErrorHandling_CorruptedJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "corrupted.json")

	// Write corrupted JSON
	corruptedData := `{
  "mcpServers": {
    "test": {
      "command": "cmd",
      "args": [unclosed array
    }
  }
`
	if err := os.WriteFile(configPath, []byte(corruptedData), 0644); err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	manager := &Manager{
		configPath: configPath,
	}

	_, err := manager.GetServers()
	if err == nil {
		t.Error("GetServers() with corrupted JSON should return error")
	}
}

// TestErrorHandling_InvalidJSON tests handling of various invalid JSON scenarios
func TestErrorHandling_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty file", ""},
		{"only whitespace", "   \n\t  "},
		{"invalid json", "not json at all"},
		{"incomplete object", `{"mcpServers":`},
		{"wrong type", `"string instead of object"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "invalid.json")

			if err := os.WriteFile(configPath, []byte(tt.data), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			manager := &Manager{
				configPath: configPath,
			}

			_, err := manager.GetServers()
			if err == nil {
				t.Errorf("GetServers() with %s should return error", tt.name)
			}
		})
	}
}

// TestGetConfigPath tests the GetConfigPath method
func TestGetConfigPath(t *testing.T) {
	expectedPath := "/test/path/config.json"
	manager := &Manager{
		configPath: expectedPath,
	}

	path := manager.GetConfigPath()
	if path != expectedPath {
		t.Errorf("GetConfigPath() = %v, want %v", path, expectedPath)
	}
}

// TestMultipleOperations tests a sequence of operations
func TestMultipleOperations(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Start with empty config
	servers, err := manager.GetServers()
	if err != nil || len(servers) != 0 {
		t.Fatalf("Initial GetServers() failed or returned servers: err=%v, len=%d", err, len(servers))
	}

	// Add first server
	server1 := Server{Name: "server1", Command: "cmd1", Args: []string{"arg1"}}
	if err := manager.AddServer(server1); err != nil {
		t.Fatalf("AddServer(server1) failed: %v", err)
	}

	// Add second server
	server2 := Server{Name: "server2", Command: "cmd2", Env: map[string]string{"KEY": "val"}}
	if err := manager.AddServer(server2); err != nil {
		t.Fatalf("AddServer(server2) failed: %v", err)
	}

	// Verify two servers
	servers, err = manager.GetServers()
	if err != nil || len(servers) != 2 {
		t.Fatalf("GetServers() after two adds: err=%v, len=%d, want len=2", err, len(servers))
	}

	// Update server1
	updatedServer1 := Server{Name: "server1", Command: "updated-cmd1"}
	if err := manager.UpdateServer(updatedServer1); err != nil {
		t.Fatalf("UpdateServer(server1) failed: %v", err)
	}

	// Verify update
	servers, err = manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() after update failed: %v", err)
	}
	found := false
	for _, s := range servers {
		if s.Name == "server1" && s.Command == "updated-cmd1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Updated server1 not found with correct command")
	}

	// Remove server1
	if err := manager.RemoveServer("server1"); err != nil {
		t.Fatalf("RemoveServer(server1) failed: %v", err)
	}

	// Verify only server2 remains
	servers, err = manager.GetServers()
	if err != nil || len(servers) != 1 {
		t.Fatalf("GetServers() after remove: err=%v, len=%d, want len=1", err, len(servers))
	}
	if servers[0].Name != "server2" {
		t.Errorf("Remaining server name = %v, want server2", servers[0].Name)
	}

	// Remove server2
	if err := manager.RemoveServer("server2"); err != nil {
		t.Fatalf("RemoveServer(server2) failed: %v", err)
	}

	// Verify empty
	servers, err = manager.GetServers()
	if err != nil || len(servers) != 0 {
		t.Fatalf("GetServers() after removing all: err=%v, len=%d, want len=0", err, len(servers))
	}
}

// TestConcurrentAccess tests basic concurrent access patterns
func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "concurrent_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add initial server
	server := Server{Name: "initial", Command: "cmd"}
	if err := manager.AddServer(server); err != nil {
		t.Fatalf("AddServer() failed: %v", err)
	}

	// Test concurrent reads (should be safe)
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := manager.GetServers()
			if err != nil {
				t.Errorf("Concurrent GetServers() failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestEmptyServerFields tests handling of servers with empty optional fields
func TestEmptyServerFields(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add server with minimal fields
	server := Server{
		Name:    "minimal",
		Command: "cmd",
		// No Args, Env, or Description
	}

	if err := manager.AddServer(server); err != nil {
		t.Fatalf("AddServer() with minimal fields failed: %v", err)
	}

	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("GetServers() returned %d servers, want 1", len(servers))
	}

	s := servers[0]
	if s.Name != "minimal" {
		t.Errorf("Server Name = %v, want minimal", s.Name)
	}
	if s.Command != "cmd" {
		t.Errorf("Server Command = %v, want cmd", s.Command)
	}
	// Args and Env can be nil or empty, both are acceptable
}

// TestLargeConfig tests handling of config with many servers
func TestLargeConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "large_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	// Add 100 servers
	numServers := 100
	for i := 0; i < numServers; i++ {
		server := Server{
			Name:    fmt.Sprintf("server-%d", i),
			Command: "cmd",
			Args:    []string{"arg1", "arg2", "arg3"},
		}
		if err := manager.AddServer(server); err != nil {
			t.Fatalf("AddServer() failed at index %d: %v", i, err)
		}
	}

	// Verify all servers were added
	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != numServers {
		t.Errorf("GetServers() returned %d servers, want %d", len(servers), numServers)
	}
}

// TestSpecialCharactersInServerData tests handling of special characters
func TestSpecialCharactersInServerData(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	manager := &Manager{
		configPath: configPath,
	}

	server := Server{
		Name:        "test-server",
		Description: "Server with \"quotes\" and \nnewlines\t and \\ backslashes",
		Command:     "cmd",
		Args:        []string{"arg with spaces", "arg\"with\"quotes"},
		Env: map[string]string{
			"PATH":        "/usr/bin:/bin",
			"SPECIAL":     "value with = and ; chars",
			"JSON_STRING": `{"key": "value"}`,
		},
	}

	if err := manager.AddServer(server); err != nil {
		t.Fatalf("AddServer() with special characters failed: %v", err)
	}

	servers, err := manager.GetServers()
	if err != nil {
		t.Fatalf("GetServers() failed: %v", err)
	}

	if len(servers) != 1 {
		t.Fatalf("GetServers() returned %d servers, want 1", len(servers))
	}

	s := servers[0]
	if s.Env["JSON_STRING"] != `{"key": "value"}` {
		t.Errorf("Env[JSON_STRING] not preserved correctly: %v", s.Env["JSON_STRING"])
	}
}
