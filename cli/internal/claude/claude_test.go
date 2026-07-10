package claude

import (
	"testing"
)

func TestNewWithTools(t *testing.T) {
	tests := []struct {
		name          string
		tools         []string
		expectEnabled bool
	}{
		{
			name:          "All tools allowed (nil)",
			tools:         nil,
			expectEnabled: true,
		},
		{
			name:          "Specific tools allowed",
			tools:         []string{"Read", "Grep", "Glob"},
			expectEnabled: true,
		},
		{
			name:          "All tools disabled (empty slice)",
			tools:         []string{},
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewWithTools("/tmp", "test", tt.tools)

			if client.EnableTools != tt.expectEnabled {
				t.Errorf("EnableTools = %v, want %v", client.EnableTools, tt.expectEnabled)
			}

			if len(tt.tools) > 0 {
				if len(client.AllowedTools) != len(tt.tools) {
					t.Errorf("AllowedTools length = %d, want %d", len(client.AllowedTools), len(tt.tools))
				}
			}
		})
	}
}

func TestNewWithTmux_BackwardCompat(t *testing.T) {
	client := NewWithTmux("/tmp", "test")

	if client.EnableTools {
		t.Error("NewWithTmux should disable tools for backward compatibility")
	}

	if client.AllowedTools != nil {
		t.Error("NewWithTmux should have nil AllowedTools")
	}
}

func TestStreamingArgsIncludesMCPConfig(t *testing.T) {
	client := NewWithWorkDir("/tmp")
	client.EnableTools = true
	client.MCPConfigs = []string{`{"mcpServers":{"datadog":{"command":"npx"}}}`}

	args := client.streamingArgs()
	if !containsArg(args, "--mcp-config") || !containsArg(args, client.MCPConfigs[0]) {
		t.Fatalf("args = %#v, want MCP config", args)
	}
	if containsArg(args, "--tools") {
		t.Fatalf("args = %#v, should not disable tools when MCP config is present", args)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
