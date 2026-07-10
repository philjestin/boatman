// Package integrations defines provider-neutral integration descriptors.
package integrations

import (
	"fmt"
	"sort"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// Integration describes a service connection Boatman can expose to model
// providers through MCP or future broker handles.
type Integration struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	MCP         *MCPDescriptor    `json:"mcp,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MCPDescriptor is a provider-neutral local or remote MCP server description.
type MCPDescriptor struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Catalog stores known integration descriptors by normalized name.
type Catalog struct {
	byName map[string]Integration
}

// NewCatalog creates a catalog from integration descriptors.
func NewCatalog(items ...Integration) Catalog {
	byName := make(map[string]Integration, len(items))
	for _, item := range items {
		name := normalizeName(item.Name)
		if name == "" {
			continue
		}
		item.Name = name
		byName[name] = item
	}
	return Catalog{byName: byName}
}

// DefaultCatalog returns built-in integration descriptors shared by CLI,
// desktop, and provider adapters.
func DefaultCatalog() Catalog {
	return NewCatalog(
		Integration{
			Name:        "datadog",
			DisplayName: "Datadog",
			Description: "Query Datadog logs, metrics, monitors, and incidents",
			Tags:        []string{"observability", "incident"},
			MCP: &MCPDescriptor{
				Command: "npx",
				Args:    []string{"-y", "@datadog/mcp-server"},
				Env: map[string]string{
					"DD_API_KEY": "",
					"DD_APP_KEY": "",
					"DD_SITE":    "datadoghq.com",
				},
			},
		},
		Integration{
			Name:        "bugsnag",
			DisplayName: "Bugsnag",
			Description: "Investigate Bugsnag errors, projects, and events",
			Tags:        []string{"observability", "incident"},
			MCP: &MCPDescriptor{
				Command: "npx",
				Args:    []string{"-y", "@bugsnag/mcp-server"},
				Env: map[string]string{
					"BUGSNAG_API_KEY": "",
				},
			},
		},
		Integration{
			Name:        "linear",
			DisplayName: "Linear",
			Description: "Query and update Linear projects and issues",
			Tags:        []string{"planning", "ticketing"},
			MCP: &MCPDescriptor{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-linear"},
				Env: map[string]string{
					"LINEAR_API_KEY": "",
				},
			},
		},
		Integration{
			Name:        "slack",
			DisplayName: "Slack",
			Description: "Read Slack context and reply in alert threads",
			Tags:        []string{"communication", "incident"},
			MCP: &MCPDescriptor{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-slack"},
				Env: map[string]string{
					"SLACK_BOT_TOKEN": "",
					"SLACK_TEAM_ID":   "",
				},
			},
		},
	)
}

// List returns integrations sorted by name.
func (c Catalog) List() []Integration {
	names := make([]string, 0, len(c.byName))
	for name := range c.byName {
		names = append(names, name)
	}
	sort.Strings(names)
	items := make([]Integration, 0, len(names))
	for _, name := range names {
		items = append(items, c.byName[name])
	}
	return items
}

// Lookup finds an integration by normalized name.
func (c Catalog) Lookup(name string) (Integration, bool) {
	item, ok := c.byName[normalizeName(name)]
	return item, ok
}

// MCPRefs returns MCP references for known integration names.
func (c Catalog) MCPRefs(names ...string) ([]agentruntime.MCPServerRef, error) {
	refs := make([]agentruntime.MCPServerRef, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		name = normalizeName(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		item, ok := c.Lookup(name)
		if !ok {
			return nil, fmt.Errorf("unknown integration %q", name)
		}
		ref, ok := item.MCPRef()
		if !ok {
			return nil, fmt.Errorf("integration %q does not expose MCP", name)
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// KnownMCPRefs returns MCP refs for known integrations and silently skips
// custom names. Use this when another layer, such as Claude's local MCP config,
// may still resolve unknown server labels.
func (c Catalog) KnownMCPRefs(names ...string) []agentruntime.MCPServerRef {
	refs := make([]agentruntime.MCPServerRef, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		name = normalizeName(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		item, ok := c.Lookup(name)
		if !ok {
			continue
		}
		ref, ok := item.MCPRef()
		if ok {
			refs = append(refs, ref)
		}
	}
	return refs
}

// MCPRef converts an integration descriptor to a runtime MCP server reference.
func (i Integration) MCPRef() (agentruntime.MCPServerRef, bool) {
	if i.MCP == nil {
		return agentruntime.MCPServerRef{}, false
	}
	label := normalizeName(i.Name)
	if label == "" {
		return agentruntime.MCPServerRef{}, false
	}
	return agentruntime.MCPServerRef{
		Label:       label,
		Command:     i.MCP.Command,
		Args:        append([]string(nil), i.MCP.Args...),
		URL:         i.MCP.URL,
		Env:         cloneStringMap(i.MCP.Env),
		Description: i.Description,
	}, true
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
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
