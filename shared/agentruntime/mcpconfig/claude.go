// Package mcpconfig translates provider-neutral MCP references into provider
// configuration formats.
package mcpconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

type claudeConfig struct {
	MCPServers map[string]claudeServer `json:"mcpServers"`
}

type claudeServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
}

// ClaudeJSON renders MCP refs into the JSON shape accepted by Claude Code's
// --mcp-config flag.
func ClaudeJSON(refs []agentruntime.MCPServerRef) (string, error) {
	config := claudeConfig{MCPServers: map[string]claudeServer{}}
	for _, ref := range refs {
		label := strings.TrimSpace(ref.Label)
		if label == "" {
			return "", fmt.Errorf("MCP server label is required")
		}
		if strings.TrimSpace(ref.Command) == "" && strings.TrimSpace(ref.URL) == "" {
			return "", fmt.Errorf("MCP server %q requires command or URL", label)
		}
		config.MCPServers[label] = claudeServer{
			Command: strings.TrimSpace(ref.Command),
			Args:    append([]string(nil), ref.Args...),
			Env:     cloneStringMap(ref.Env),
			URL:     strings.TrimSpace(ref.URL),
		}
	}
	if len(config.MCPServers) == 0 {
		return "", nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
