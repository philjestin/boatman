package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
)

// Server represents an MCP server configuration
type Server struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Enabled     bool              `json:"enabled"`
}

// IntegrationStatus is the broker-visible status for a known service integration.
type IntegrationStatus = integrations.Status

// Config represents the MCP configuration file structure
type Config struct {
	McpServers map[string]ServerDef `json:"mcpServers"`
}

// ServerDef represents the Claude MCP server definition format
type ServerDef struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Manager manages MCP server configurations
type Manager struct {
	configPath string
}

// NewManager creates a new MCP manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Claude stores MCP config in ~/.claude/claude_mcp_config.json
	configPath := filepath.Join(homeDir, ".claude", "claude_mcp_config.json")

	return &Manager{
		configPath: configPath,
	}, nil
}

// GetServers returns configured MCP servers
func (m *Manager) GetServers() ([]Server, error) {
	config, err := m.loadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return []Server{}, nil
		}
		return nil, err
	}

	servers := make([]Server, 0, len(config.McpServers))
	for name, def := range config.McpServers {
		servers = append(servers, Server{
			Name:    name,
			Command: def.Command,
			Args:    def.Args,
			Env:     def.Env,
			Enabled: true,
		})
	}

	return servers, nil
}

// AddServer adds a new MCP server
func (m *Manager) AddServer(server Server) error {
	config, err := m.loadConfig()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if config == nil {
		config = &Config{
			McpServers: make(map[string]ServerDef),
		}
	}

	config.McpServers[server.Name] = ServerDef{
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
	}

	return m.saveConfig(config)
}

// RemoveServer removes an MCP server
func (m *Manager) RemoveServer(name string) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	delete(config.McpServers, name)
	return m.saveConfig(config)
}

// UpdateServer updates an MCP server configuration
func (m *Manager) UpdateServer(server Server) error {
	return m.AddServer(server)
}

// loadConfig loads the MCP configuration file
func (m *Manager) loadConfig() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig saves the MCP configuration file
func (m *Manager) saveConfig(config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfigPath returns the path to the MCP config file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// ValidateServer validates server configuration
func (m *Manager) ValidateServer(server Server) error {
	if server.Name == "" {
		return os.ErrInvalid
	}
	if server.Command == "" {
		return os.ErrInvalid
	}
	return nil
}

// GetPresetServers returns common MCP server presets
func GetPresetServers() []Server {
	servers := []Server{
		{
			Name:        "filesystem",
			Description: "File system access for Claude",
			Command:     "npx",
			Args:        []string{"-y", "@anthropic/mcp-server-filesystem"},
			Enabled:     false,
		},
		{
			Name:        "github",
			Description: "GitHub integration",
			Command:     "npx",
			Args:        []string{"-y", "@anthropic/mcp-server-github"},
			Enabled:     false,
		},
		{
			Name:        "postgres",
			Description: "PostgreSQL database access",
			Command:     "npx",
			Args:        []string{"-y", "@anthropic/mcp-server-postgres"},
			Enabled:     false,
		},
	}
	catalog := integrations.DefaultCatalog()
	for _, name := range []string{"datadog", "bugsnag", "linear", "slack"} {
		if item, ok := catalog.Lookup(name); ok {
			if ref, ok := item.MCPRef(); ok {
				servers = append(servers, Server{
					Name:        ref.Label,
					Description: ref.Description,
					Command:     ref.Command,
					Args:        ref.Args,
					Env:         ref.Env,
					Enabled:     false,
				})
			}
		}
	}
	servers = append(servers,
		Server{
			Name:        "figma",
			Description: "Figma design file access and manipulation",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-figma"},
			Env: map[string]string{
				"FIGMA_PERSONAL_ACCESS_TOKEN": "",
			},
			Enabled: false,
		},
	)
	return servers
}

// GetIntegrationStatuses reports descriptor-level health for known service
// integrations based on configured MCP servers.
func GetIntegrationStatuses(servers []Server) []IntegrationStatus {
	catalog := integrations.DefaultCatalog()
	manager := integrations.NewManager(catalog)
	configured := make(map[string]Server, len(servers))
	for _, server := range servers {
		configured[strings.ToLower(strings.TrimSpace(server.Name))] = server
	}

	statuses := make([]IntegrationStatus, 0, len(catalog.List()))
	for _, item := range catalog.List() {
		server, ok := configured[item.Name]
		opts := integrations.ResolveOptions{}
		if ok {
			opts.Enabled = server.Enabled
			opts.Env = server.Env
		}
		conn, err := manager.Connect(context.Background(), item.Name, opts)
		status := conn.Status
		if err != nil {
			status = integrations.Status{
				Name:    item.Name,
				State:   integrations.StateFailed,
				Message: err.Error(),
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}
