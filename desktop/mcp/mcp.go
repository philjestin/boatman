package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	return []Server{
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
		{
			Name:        "datadog",
			Description: "Query Datadog logs, metrics, and monitors",
			Command:     "npx",
			Args:        []string{"-y", "@datadog/mcp-server"},
			Env: map[string]string{
				"DD_API_KEY": "",
				"DD_APP_KEY": "",
				"DD_SITE":    "datadoghq.com",
			},
			Enabled: false,
		},
		{
			Name:        "bugsnag",
			Description: "Investigate Bugsnag errors and exceptions",
			Command:     "npx",
			Args:        []string{"-y", "@bugsnag/mcp-server"},
			Env: map[string]string{
				"BUGSNAG_API_KEY": "",
			},
			Enabled: false,
		},
		{
			Name:        "linear",
			Description: "Linear project management and issue tracking",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-linear"},
			Env: map[string]string{
				"LINEAR_API_KEY": "",
			},
			Enabled: false,
		},
		{
			Name:        "slack",
			Description: "Slack workspace integration",
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-slack"},
			Env: map[string]string{
				"SLACK_BOT_TOKEN": "",
				"SLACK_TEAM_ID":   "",
			},
			Enabled: false,
		},
	}
}
