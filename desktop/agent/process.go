package agent

import (
	"os/exec"
	"runtime"
)

// ClaudeCLI provides utilities for interacting with the Claude CLI
type ClaudeCLI struct {
	path string
}

// NewClaudeCLI creates a new Claude CLI wrapper
func NewClaudeCLI() *ClaudeCLI {
	return &ClaudeCLI{
		path: "claude",
	}
}

// IsInstalled checks if the Claude CLI is available
func (c *ClaudeCLI) IsInstalled() bool {
	_, err := exec.LookPath(c.path)
	return err == nil
}

// GetVersion returns the installed Claude CLI version
func (c *ClaudeCLI) GetVersion() (string, error) {
	cmd := exec.Command(c.path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetDefaultPath returns the expected path for Claude CLI
func (c *ClaudeCLI) GetDefaultPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/local/bin/claude"
	case "linux":
		return "/usr/local/bin/claude"
	case "windows":
		return "C:\\Program Files\\Claude\\claude.exe"
	default:
		return "claude"
	}
}

// SetPath sets a custom path for the Claude CLI
func (c *ClaudeCLI) SetPath(path string) {
	c.path = path
}

// GetPath returns the current CLI path
func (c *ClaudeCLI) GetPath() string {
	return c.path
}

// BuildArgs constructs command line arguments for the CLI
func (c *ClaudeCLI) BuildArgs(opts SessionOptions) []string {
	args := []string{}

	// Use print mode for JSON output
	if opts.JSONOutput {
		args = append(args, "--print", "json")
		args = append(args, "--output-format", "stream-json")
	}

	// Set model if specified
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	// Set approval mode
	switch opts.ApprovalMode {
	case "auto-edit":
		args = append(args, "--allowedTools", "Edit,Write")
	case "full-auto":
		args = append(args, "--dangerously-skip-permissions")
	}

	// Add MCP configuration if provided
	for _, mcp := range opts.MCPServers {
		if mcp.Enabled {
			args = append(args, "--mcp", mcp.Name)
		}
	}

	return args
}

// SessionOptions configures a Claude CLI session
type SessionOptions struct {
	Model        string
	ApprovalMode string
	MCPServers   []MCPServerConfig
	JSONOutput   bool
	WorkingDir   string
}

// MCPServerConfig represents MCP server configuration for CLI
type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Enabled bool
}
