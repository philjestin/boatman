// Package agentruntime defines the provider-neutral contracts shared by Boatman
// clients, runtimes, and model/provider adapters.
package agentruntime

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// ProtocolVersion is incremented when the runtime event contract changes in a
// backwards-incompatible way.
const ProtocolVersion = 1

// EventType identifies a normalized runtime event.
type EventType string

const (
	EventRunStarted       EventType = "run.started"
	EventRunCompleted     EventType = "run.completed"
	EventRunFailed        EventType = "run.failed"
	EventPhaseStarted     EventType = "phase.started"
	EventPhaseCompleted   EventType = "phase.completed"
	EventTaskCreated      EventType = "task.created"
	EventTaskUpdated      EventType = "task.updated"
	EventMessageDelta     EventType = "message.delta"
	EventMessageCompleted EventType = "message.completed"
	EventToolCall         EventType = "tool.call"
	EventToolResult       EventType = "tool.result"
	EventApprovalRequest  EventType = "approval.requested"
	EventApprovalResolved EventType = "approval.resolved"
	EventUsageUpdated     EventType = "usage.updated"
	EventArtifactChanged  EventType = "artifact.changed"
	EventProviderRaw      EventType = "provider.raw"
	EventLogMessage       EventType = "log.message"
	EventSchemaResult     EventType = "schema.result"
	EventMemoryLoaded     EventType = "memory.loaded"
	EventIntegrationState EventType = "integration.state"
)

// Status describes the normalized lifecycle status for runs, phases, and tasks.
type Status string

const (
	StatusStarted    Status = "started"
	StatusRunning    Status = "running"
	StatusWaiting    Status = "waiting"
	StatusCompleted  Status = "completed"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
	StatusSkipped    Status = "skipped"
	StatusInProgress Status = "in_progress"
)

// Role identifies the job a model/provider is performing.
type Role string

const (
	RolePlanner     Role = "planner"
	RoleExecutor    Role = "executor"
	RoleReviewer    Role = "reviewer"
	RoleRefactorer  Role = "refactorer"
	RoleTester      Role = "tester"
	RoleScorer      Role = "scorer"
	RoleTriage      Role = "triage"
	RoleChat        Role = "chat"
	RoleFirefight   Role = "firefight"
	RoleMemory      Role = "memory"
	RoleIntegration Role = "integration"
	RoleRoutine     Role = "routine"
)

// ModelProfile describes phase-specific model routing for multi-step agent
// workflows. Empty fields are resolved by callers from their selected default
// model.
type ModelProfile struct {
	// Plan is used for planning phases such as /plan.
	Plan string `json:"plan,omitempty"`
	// Implementation is used for code changes and remediation/refactor work.
	Implementation string `json:"implementation,omitempty"`
	// Skills is used for skill-backed review or other skill execution phases.
	Skills string `json:"skills,omitempty"`
}

// WithDefault returns a copy with empty phase fields filled by defaultModel.
func (m ModelProfile) WithDefault(defaultModel string) ModelProfile {
	defaultModel = strings.TrimSpace(defaultModel)
	out := ModelProfile{
		Plan:           strings.TrimSpace(m.Plan),
		Implementation: strings.TrimSpace(m.Implementation),
		Skills:         strings.TrimSpace(m.Skills),
	}
	if out.Plan == "" {
		out.Plan = defaultModel
	}
	if out.Implementation == "" {
		out.Implementation = defaultModel
	}
	if out.Skills == "" {
		out.Skills = defaultModel
	}
	return out
}

// IsZero reports whether no model overrides are set.
func (m ModelProfile) IsZero() bool {
	return strings.TrimSpace(m.Plan) == "" &&
		strings.TrimSpace(m.Implementation) == "" &&
		strings.TrimSpace(m.Skills) == ""
}

// Event is the normalized runtime stream item. It intentionally keeps provider
// details in optional fields so CLI, desktop, and future services can consume a
// stable contract while still retaining raw provider logs for debugging.
type Event struct {
	Version     int             `json:"version"`
	Type        EventType       `json:"type"`
	RunID       string          `json:"runId,omitempty"`
	PhaseID     string          `json:"phaseId,omitempty"`
	TaskID      string          `json:"taskId,omitempty"`
	Provider    string          `json:"provider,omitempty"`
	Model       string          `json:"model,omitempty"`
	Role        Role            `json:"role,omitempty"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Status      Status          `json:"status,omitempty"`
	Message     string          `json:"message,omitempty"`
	Usage       *Usage          `json:"usage,omitempty"`
	Tool        *ToolEvent      `json:"tool,omitempty"`
	Approval    *ApprovalEvent  `json:"approval,omitempty"`
	Artifact    *ArtifactEvent  `json:"artifact,omitempty"`
	Schema      *SchemaEvent    `json:"schema,omitempty"`
	Raw         json.RawMessage `json:"raw,omitempty"`
	Data        map[string]any  `json:"data,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// NewEvent creates a normalized event with the current protocol version.
func NewEvent(eventType EventType) Event {
	return Event{
		Version:   ProtocolVersion,
		Type:      eventType,
		Timestamp: time.Now().UTC(),
	}
}

// EventStream is the transport returned by provider adapters.
type EventStream <-chan Event

// Provider describes a model backend such as Claude CLI, Anthropic Messages,
// Claude Agent SDK, OpenAI Responses, OpenAI Agents, or a local model runtime.
type Provider interface {
	Name() string
	Capabilities(ctx context.Context) (Capabilities, error)
	StartRun(ctx context.Context, req RunRequest) (EventStream, error)
	ResumeRun(ctx context.Context, runID string, input RunInput) (EventStream, error)
	CancelRun(ctx context.Context, runID string) error
}

// Capabilities lets the orchestrator choose a provider by what it can do, not
// by hard-coded model names or one-off CLI flags.
type Capabilities struct {
	Provider              string   `json:"provider"`
	Models                []string `json:"models,omitempty"`
	SupportsStreaming     bool     `json:"supportsStreaming"`
	SupportsBackground    bool     `json:"supportsBackground"`
	SupportsResume        bool     `json:"supportsResume"`
	SupportsToolCalls     bool     `json:"supportsToolCalls"`
	SupportsMCP           bool     `json:"supportsMCP"`
	SupportsApprovals     bool     `json:"supportsApprovals"`
	SupportsStructuredOut bool     `json:"supportsStructuredOutput"`
	SupportsArtifacts     bool     `json:"supportsArtifacts"`
	SupportsUsage         bool     `json:"supportsUsage"`
	SupportsVision        bool     `json:"supportsVision"`
	SupportsAudio         bool     `json:"supportsAudio"`
	SupportsComputerUse   bool     `json:"supportsComputerUse"`
	Experimental          []string `json:"experimental,omitempty"`
}

// RunRequest describes provider-neutral work for a model adapter.
type RunRequest struct {
	RunID          string            `json:"runId,omitempty"`
	Role           Role              `json:"role"`
	Profile        string            `json:"profile,omitempty"`
	Provider       string            `json:"provider,omitempty"`
	Model          string            `json:"model,omitempty"`
	WorkDir        string            `json:"workDir,omitempty"`
	Instructions   string            `json:"instructions,omitempty"`
	Messages       []Message         `json:"messages,omitempty"`
	Tools          []ToolRef         `json:"tools,omitempty"`
	MCPServers     []MCPServerRef    `json:"mcpServers,omitempty"`
	OutputSchema   *OutputSchema     `json:"outputSchema,omitempty"`
	ApprovalPolicy ApprovalPolicy    `json:"approvalPolicy,omitempty"`
	Reasoning      *ReasoningOptions `json:"reasoning,omitempty"`
	Background     bool              `json:"background,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// RunInput is used to continue a provider run after user input or approval.
type RunInput struct {
	Messages   []Message         `json:"messages,omitempty"`
	Approval   *ApprovalDecision `json:"approval,omitempty"`
	ToolResult *ToolResult       `json:"toolResult,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Message is provider-neutral chat or task input.
type Message struct {
	Role    string         `json:"role"`
	Content string         `json:"content,omitempty"`
	Blocks  []ContentBlock `json:"blocks,omitempty"`
}

// ContentBlock carries multimodal input without binding callers to one vendor
// shape.
type ContentBlock struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	MimeType string          `json:"mimeType,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// ToolRef declares a local or hosted tool available to a run.
type ToolRef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Kind        string          `json:"kind,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Strict      bool            `json:"strict,omitempty"`
}

// MCPServerRef describes a tool server in a provider-neutral way. Providers can
// translate this to Claude MCP config, OpenAI remote MCP, or a local broker.
type MCPServerRef struct {
	Label       string            `json:"label"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	URL         string            `json:"url,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
}

// ApprovalPolicy controls whether a run may execute side-effecting actions.
type ApprovalPolicy string

const (
	ApprovalSuggest  ApprovalPolicy = "suggest"
	ApprovalAutoEdit ApprovalPolicy = "auto_edit"
	ApprovalFullAuto ApprovalPolicy = "full_auto"
)

// ReasoningOptions capture reasoning controls without coupling callers to a
// provider's parameter names.
type ReasoningOptions struct {
	Effort      string `json:"effort,omitempty"`
	MaxTokens   int    `json:"maxTokens,omitempty"`
	TokenBudget int    `json:"tokenBudget,omitempty"`
}

// OutputSchema requests typed output from a provider.
type OutputSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema"`
	Strict      bool            `json:"strict"`
}

// Usage normalizes token and cost reporting across providers.
type Usage struct {
	InputTokens      int     `json:"inputTokens,omitempty"`
	OutputTokens     int     `json:"outputTokens,omitempty"`
	CacheReadTokens  int     `json:"cacheReadTokens,omitempty"`
	CacheWriteTokens int     `json:"cacheWriteTokens,omitempty"`
	TotalCostUSD     float64 `json:"totalCostUsd,omitempty"`
}

// ToolEvent describes a model tool call or result.
type ToolEvent struct {
	ID      string          `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Output  json.RawMessage `json:"output,omitempty"`
	IsError bool            `json:"isError,omitempty"`
}

// ToolResult is supplied when resuming a run after local tool execution.
type ToolResult struct {
	ID      string          `json:"id"`
	Output  json.RawMessage `json:"output,omitempty"`
	IsError bool            `json:"isError,omitempty"`
}

// ApprovalEvent describes a pending or resolved approval.
type ApprovalEvent struct {
	ID       string          `json:"id,omitempty"`
	Action   string          `json:"action,omitempty"`
	Reason   string          `json:"reason,omitempty"`
	Payload  json.RawMessage `json:"payload,omitempty"`
	Decision string          `json:"decision,omitempty"`
}

// ApprovalDecision supplies a user's answer to a pending approval request.
type ApprovalDecision struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// ArtifactEvent tracks files, diffs, PRs, notebooks, skill docs, and other
// durable outputs produced by a run.
type ArtifactEvent struct {
	Kind string `json:"kind"`
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
	Diff string `json:"diff,omitempty"`
}

// SchemaEvent carries parsed structured output and validation state.
type SchemaEvent struct {
	Name   string          `json:"name,omitempty"`
	Valid  bool            `json:"valid"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
