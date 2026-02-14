package agent

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestNewClaudeCLI(t *testing.T) {
	cli := NewClaudeCLI()
	if cli == nil {
		t.Fatal("NewClaudeCLI() returned nil")
	}
	if cli.path != "claude" {
		t.Errorf("Expected default path 'claude', got '%s'", cli.path)
	}
}

func TestIsInstalled(t *testing.T) {
	cli := NewClaudeCLI()

	// Test with a command that should exist (e.g., "go")
	cli.SetPath("go")
	if !cli.IsInstalled() {
		t.Error("IsInstalled() should return true for 'go' command")
	}

	// Test with a command that should not exist
	cli.SetPath("nonexistent-command-xyz-12345")
	if cli.IsInstalled() {
		t.Error("IsInstalled() should return false for nonexistent command")
	}
}

func TestGetVersion(t *testing.T) {
	cli := NewClaudeCLI()

	// Test with 'go' command as a proxy (since claude may not be installed)
	cli.SetPath("go")
	version, err := cli.GetVersion()
	if err != nil {
		t.Errorf("GetVersion() failed for 'go' command: %v", err)
	}
	if version == "" {
		t.Error("GetVersion() returned empty string")
	}

	// Test with nonexistent command
	cli.SetPath("nonexistent-command-xyz-12345")
	_, err = cli.GetVersion()
	if err == nil {
		t.Error("GetVersion() should return error for nonexistent command")
	}
}

func TestGetDefaultPath(t *testing.T) {
	cli := NewClaudeCLI()
	defaultPath := cli.GetDefaultPath()

	switch runtime.GOOS {
	case "darwin", "linux":
		expected := "/usr/local/bin/claude"
		if defaultPath != expected {
			t.Errorf("Expected default path '%s' for %s, got '%s'", expected, runtime.GOOS, defaultPath)
		}
	case "windows":
		expected := "C:\\Program Files\\Claude\\claude.exe"
		if defaultPath != expected {
			t.Errorf("Expected default path '%s' for windows, got '%s'", expected, defaultPath)
		}
	default:
		if defaultPath != "claude" {
			t.Errorf("Expected default path 'claude' for unknown OS, got '%s'", defaultPath)
		}
	}
}

func TestSetPathAndGetPath(t *testing.T) {
	cli := NewClaudeCLI()

	// Test initial path
	if cli.GetPath() != "claude" {
		t.Errorf("Expected initial path 'claude', got '%s'", cli.GetPath())
	}

	// Test setting custom path
	customPath := "/custom/path/to/claude"
	cli.SetPath(customPath)
	if cli.GetPath() != customPath {
		t.Errorf("Expected path '%s', got '%s'", customPath, cli.GetPath())
	}

	// Test setting empty path
	cli.SetPath("")
	if cli.GetPath() != "" {
		t.Errorf("Expected empty path, got '%s'", cli.GetPath())
	}
}

func TestBuildArgs_JSONOutput(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		JSONOutput: true,
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{"--print", "json", "--output-format", "stream-json"}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_Model(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		Model: "claude-opus-4",
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{"--model", "claude-opus-4"}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_ApprovalMode_AutoEdit(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		ApprovalMode: "auto-edit",
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{"--allowedTools", "Edit,Write"}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_ApprovalMode_FullAuto(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		ApprovalMode: "full-auto",
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{"--dangerously-skip-permissions"}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_ApprovalMode_Suggest(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		ApprovalMode: "suggest",
	}

	args := cli.BuildArgs(opts)

	// "suggest" mode should not add any approval-related args
	if len(args) != 0 {
		t.Errorf("Expected no args for 'suggest' mode, got %d args", len(args))
	}
}

func TestBuildArgs_MCPServers(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		MCPServers: []MCPServerConfig{
			{Name: "filesystem", Enabled: true},
			{Name: "github", Enabled: true},
			{Name: "disabled", Enabled: false},
		},
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{"--mcp", "filesystem", "--mcp", "github"}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_Combined(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{
		JSONOutput:   true,
		Model:        "claude-sonnet-4",
		ApprovalMode: "auto-edit",
		MCPServers: []MCPServerConfig{
			{Name: "filesystem", Enabled: true},
			{Name: "postgres", Enabled: true},
		},
	}

	args := cli.BuildArgs(opts)

	expectedArgs := []string{
		"--print", "json",
		"--output-format", "stream-json",
		"--model", "claude-sonnet-4",
		"--allowedTools", "Edit,Write",
		"--mcp", "filesystem",
		"--mcp", "postgres",
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d: %v", len(expectedArgs), len(args), args)
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

func TestBuildArgs_EmptyOptions(t *testing.T) {
	cli := NewClaudeCLI()

	opts := SessionOptions{}

	args := cli.BuildArgs(opts)

	if len(args) != 0 {
		t.Errorf("Expected no args for empty options, got %d args: %v", len(args), args)
	}
}

// TestIsInstalledIntegration tests actual claude CLI installation if available
func TestIsInstalledIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cli := NewClaudeCLI()
	_, err := exec.LookPath("claude")
	expectedInstalled := (err == nil)

	if cli.IsInstalled() != expectedInstalled {
		if expectedInstalled {
			t.Error("IsInstalled() returned false but claude was found in PATH")
		} else {
			t.Error("IsInstalled() returned true but claude was not found in PATH")
		}
	}
}

// TestGetVersionIntegration tests actual claude CLI version if available
func TestGetVersionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("CLAUDE_CLI_INSTALLED") != "true" {
		t.Skip("Skipping integration test: set CLAUDE_CLI_INSTALLED=true to run")
	}

	cli := NewClaudeCLI()
	if !cli.IsInstalled() {
		t.Skip("Claude CLI not installed, skipping integration test")
	}

	version, err := cli.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion() failed: %v", err)
	}

	if version == "" {
		t.Error("GetVersion() returned empty string")
	}
}
