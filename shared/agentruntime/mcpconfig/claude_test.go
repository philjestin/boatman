package mcpconfig

import (
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestClaudeJSONRendersMCPServers(t *testing.T) {
	rendered, err := ClaudeJSON([]agentruntime.MCPServerRef{
		{
			Label:   "datadog",
			Command: "npx",
			Args:    []string{"-y", "@datadog/mcp-server"},
			Env:     map[string]string{"DD_SITE": "datadoghq.com"},
		},
	})
	if err != nil {
		t.Fatalf("ClaudeJSON error: %v", err)
	}
	var decoded map[string]map[string]map[string]any
	if err := json.Unmarshal([]byte(rendered), &decoded); err != nil {
		t.Fatalf("rendered JSON did not decode: %v", err)
	}
	server := decoded["mcpServers"]["datadog"]
	if server["command"] != "npx" {
		t.Fatalf("command = %#v, want npx", server["command"])
	}
	if env := server["env"].(map[string]any); env["DD_SITE"] != "datadoghq.com" {
		t.Fatalf("env = %#v, want DD_SITE", env)
	}
}

func TestClaudeJSONRejectsInvalidRefs(t *testing.T) {
	if _, err := ClaudeJSON([]agentruntime.MCPServerRef{{Command: "npx"}}); err == nil {
		t.Fatal("ClaudeJSON should require labels")
	}
	_, err := ClaudeJSON([]agentruntime.MCPServerRef{{Label: "empty"}})
	if err == nil || !strings.Contains(err.Error(), "requires command or URL") {
		t.Fatalf("err = %v, want missing command/url error", err)
	}
}

func TestClaudeJSONEmptyRefs(t *testing.T) {
	rendered, err := ClaudeJSON(nil)
	if err != nil {
		t.Fatalf("ClaudeJSON error: %v", err)
	}
	if rendered != "" {
		t.Fatalf("rendered = %q, want empty", rendered)
	}
}
