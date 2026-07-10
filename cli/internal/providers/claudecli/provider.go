// Package claudecli adapts the existing Claude CLI integration to the
// provider-neutral agentruntime.Provider contract.
package claudecli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/claude"
	"github.com/philjestin/boatmanmode/internal/cost"
)

const providerName = "claude-cli"

type invokeFunc func(context.Context, agentruntime.RunRequest, func(string)) (string, *cost.Usage, error)

// Provider wraps the Claude CLI behind the normalized runtime contract.
type Provider struct {
	command string
	invoke  invokeFunc
}

// Option configures a Provider.
type Option func(*Provider)

// WithCommand overrides the executable used for Claude CLI calls.
func WithCommand(command string) Option {
	return func(p *Provider) {
		p.command = command
	}
}

// New creates a Claude CLI provider adapter.
func New(opts ...Option) *Provider {
	p := &Provider{
		command: "claude",
	}
	for _, opt := range opts {
		opt(p)
	}
	p.invoke = p.invokeClaude
	return p
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return providerName
}

// Capabilities reports the features available through the Claude CLI adapter.
func (p *Provider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{
		Provider:              providerName,
		SupportsStreaming:     true,
		SupportsBackground:    false,
		SupportsResume:        false,
		SupportsToolCalls:     true,
		SupportsMCP:           true,
		SupportsApprovals:     false,
		SupportsStructuredOut: false,
		SupportsArtifacts:     true,
		SupportsUsage:         true,
	}, nil
}

// StartRun starts a Claude CLI invocation and returns normalized runtime events.
func (p *Provider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if p.invoke == nil {
		p.invoke = p.invokeClaude
	}
	if req.OutputSchema != nil {
		if err := agentruntime.ValidateOutputSchema(req.OutputSchema); err != nil {
			return nil, err
		}
	}

	events := make(chan agentruntime.Event, 32)
	runID := req.RunID
	if runID == "" {
		runID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	phaseID := phaseID(req)

	go func() {
		defer close(events)

		emit := func(event agentruntime.Event) bool {
			event.RunID = runID
			if event.PhaseID == "" {
				event.PhaseID = phaseID
			}
			event.Provider = providerName
			event.Model = req.Model
			event.Role = req.Role
			select {
			case <-ctx.Done():
				return false
			case events <- event:
				return true
			}
		}

		started := agentruntime.NewEvent(agentruntime.EventRunStarted)
		started.Status = agentruntime.StatusStarted
		started.Name = req.Profile
		emit(started)

		phaseStarted := agentruntime.NewEvent(agentruntime.EventPhaseStarted)
		phaseStarted.Status = agentruntime.StatusStarted
		phaseStarted.Name = string(req.Role)
		emit(phaseStarted)

		response, usage, err := p.invoke(ctx, req, func(rawLine string) {
			rawEvent := agentruntime.NewEvent(agentruntime.EventProviderRaw)
			if raw := json.RawMessage(rawLine); json.Valid(raw) {
				rawEvent.Raw = raw
			} else {
				rawEvent.Message = rawLine
			}
			emit(rawEvent)
		})
		if err != nil {
			failed := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
			failed.Status = agentruntime.StatusFailed
			failed.Message = err.Error()
			emit(failed)

			runFailed := agentruntime.NewEvent(agentruntime.EventRunFailed)
			runFailed.Status = agentruntime.StatusFailed
			runFailed.Message = err.Error()
			emit(runFailed)
			return
		}

		if usage != nil && !usage.IsEmpty() {
			usageEvent := agentruntime.NewEvent(agentruntime.EventUsageUpdated)
			usageEvent.Usage = normalizeUsage(usage)
			emit(usageEvent)
		}

		if response != "" {
			message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
			message.Message = response
			emit(message)
		}

		completed := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
		completed.Status = agentruntime.StatusSucceeded
		emit(completed)

		runCompleted := agentruntime.NewEvent(agentruntime.EventRunCompleted)
		runCompleted.Status = agentruntime.StatusSucceeded
		emit(runCompleted)
	}()

	return events, nil
}

// ResumeRun is not available through the current one-shot Claude CLI adapter.
func (p *Provider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, fmt.Errorf("%s does not support runtime resume yet", providerName)
}

// CancelRun is handled by canceling the context passed to StartRun.
func (p *Provider) CancelRun(context.Context, string) error {
	return nil
}

func (p *Provider) invokeClaude(ctx context.Context, req agentruntime.RunRequest, forward func(string)) (string, *cost.Usage, error) {
	client := claude.NewWithWorkDir(req.WorkDir)
	client.Command = p.command
	client.Model = req.Model
	client.EventForwarder = forward
	client.SkipPermissions = req.ApprovalPolicy == agentruntime.ApprovalFullAuto
	client.SessionName = phaseID(req)

	if agent := strings.TrimSpace(req.Metadata["claudeAgent"]); agent != "" {
		client.Agent = agent
	}
	if strings.EqualFold(strings.TrimSpace(req.Metadata["outputFormat"]), "text") {
		client.Stream = false
	}
	if truthy(req.Metadata["useTmux"]) {
		client.UseTmux = true
	}
	if truthy(req.Metadata["enablePromptCaching"]) {
		client.EnablePromptCaching = true
	}

	if req.Reasoning != nil {
		client.Effort = req.Reasoning.Effort
	}
	if req.Tools != nil {
		client.EnableTools = true
		client.AllowedTools = toolNames(req.Tools)
	} else {
		client.EnableTools = false
	}

	return client.Message(ctx, req.Instructions, promptFromMessages(req.Messages))
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func phaseID(req agentruntime.RunRequest) string {
	if phaseID, ok := req.Metadata["phaseId"]; ok && phaseID != "" {
		return phaseID
	}
	if req.Role != "" {
		return string(req.Role)
	}
	return "run"
}

func promptFromMessages(messages []agentruntime.Message) string {
	if len(messages) == 0 {
		return ""
	}
	if len(messages) == 1 {
		return messages[0].Content
	}
	var prompt string
	for i, msg := range messages {
		if i > 0 {
			prompt += "\n\n"
		}
		if msg.Role != "" {
			prompt += msg.Role + ": "
		}
		prompt += msg.Content
	}
	return prompt
}

func toolNames(tools []agentruntime.ToolRef) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool.Name != "" {
			names = append(names, tool.Name)
		}
	}
	return names
}

func normalizeUsage(usage *cost.Usage) *agentruntime.Usage {
	return &agentruntime.Usage{
		InputTokens:      usage.InputTokens,
		OutputTokens:     usage.OutputTokens,
		CacheReadTokens:  usage.CacheReadTokens,
		CacheWriteTokens: usage.CacheWriteTokens,
		TotalCostUSD:     usage.TotalCostUSD,
	}
}
