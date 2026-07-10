// Package toolbroker provides provider-neutral local tool descriptions and
// execution for model adapters.
package toolbroker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// Invocation describes one local tool call requested by a provider.
type Invocation struct {
	ID             string                      `json:"id,omitempty"`
	Name           string                      `json:"name"`
	WorkDir        string                      `json:"workDir,omitempty"`
	Input          json.RawMessage             `json:"input,omitempty"`
	ApprovalPolicy agentruntime.ApprovalPolicy `json:"approvalPolicy,omitempty"`
	Metadata       map[string]string           `json:"metadata,omitempty"`
}

// Result is the normalized output from a local tool.
type Result struct {
	ID      string          `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Output  json.RawMessage `json:"output,omitempty"`
	IsError bool            `json:"isError,omitempty"`
}

// Tool is implemented by provider-neutral local tools.
type Tool interface {
	Name() string
	Ref() agentruntime.ToolRef
	Invoke(ctx context.Context, inv Invocation) (Result, error)
}

// Broker stores executable local tools.
type Broker struct {
	tools map[string]Tool
}

// New creates a broker with the supplied tools.
func New(tools ...Tool) *Broker {
	b := &Broker{tools: make(map[string]Tool)}
	for _, tool := range tools {
		b.Register(tool)
	}
	return b
}

// NewLocal creates a broker with Boatman's built-in local tools.
func NewLocal() *Broker {
	return New(
		ReadTool{},
		WriteTool{},
		EditTool{},
		BashTool{},
		GrepTool{},
		GlobTool{},
	)
}

// Register adds or replaces a tool.
func (b *Broker) Register(tool Tool) {
	if tool == nil {
		return
	}
	b.tools[tool.Name()] = tool
}

// Get returns a tool by exact name.
func (b *Broker) Get(name string) (Tool, bool) {
	tool, ok := b.tools[name]
	return tool, ok
}

// Invoke executes a registered tool.
func (b *Broker) Invoke(ctx context.Context, inv Invocation) (Result, error) {
	tool, ok := b.Get(inv.Name)
	if !ok {
		return Result{ID: inv.ID, Name: inv.Name, IsError: true}, fmt.Errorf("tool %q is not registered", inv.Name)
	}
	result, err := tool.Invoke(ctx, inv)
	if result.ID == "" {
		result.ID = inv.ID
	}
	if result.Name == "" {
		result.Name = inv.Name
	}
	if err != nil {
		result.IsError = true
	}
	return result, err
}

// Refs returns tool references in the requested order.
func (b *Broker) Refs(names ...string) []agentruntime.ToolRef {
	refs := make([]agentruntime.ToolRef, 0, len(names))
	for _, name := range names {
		if tool, ok := b.Get(name); ok {
			refs = append(refs, tool.Ref())
		}
	}
	return refs
}

// Names returns registered tool names.
func (b *Broker) Names() []string {
	names := make([]string, 0, len(b.tools))
	for name := range b.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PlannerRefs returns the safe read/search tools used by planning workflows.
func PlannerRefs() []agentruntime.ToolRef {
	return NewLocal().Refs("Read", "Grep", "Glob")
}

// ExecutorRefs returns the local tools used by execution/refactor workflows.
func ExecutorRefs() []agentruntime.ToolRef {
	return NewLocal().Refs("Read", "Write", "Edit", "Bash", "Grep", "Glob")
}

// AutoEditRefs returns the edit-only tools exposed by desktop auto-edit mode.
func AutoEditRefs() []agentruntime.ToolRef {
	return NewLocal().Refs("Edit", "Write")
}

func rawJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return data
}

func schema(properties map[string]any, required ...string) json.RawMessage {
	return rawJSON(map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             required,
		"properties":           properties,
	})
}

func decodeInput[T any](raw json.RawMessage) (T, error) {
	var input T
	if len(raw) == 0 {
		return input, nil
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return input, fmt.Errorf("decode tool input: %w", err)
	}
	return input, nil
}

func textProperty(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolProperty(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func numberProperty(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}

func normalizeToolPath(path string) string {
	return strings.TrimSpace(path)
}
