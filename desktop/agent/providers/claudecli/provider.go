// Package claudecli adapts Claude CLI streaming for desktop agent sessions.
package claudecli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/mcpconfig"
)

const providerName = "claude-cli"

type runFunc func(context.Context, agentruntime.RunRequest, func(agentruntime.Event) bool) error

// Provider streams Claude CLI output as normalized runtime events.
type Provider struct {
	command string
	env     map[string]string
	run     runFunc
}

// Option configures Provider.
type Option func(*Provider)

// WithCommand overrides the Claude executable.
func WithCommand(command string) Option {
	return func(p *Provider) {
		p.command = command
	}
}

// WithExtraEnv appends process environment variables when invoking Claude.
// These values are intentionally provider-local so callers can pass secrets to
// local MCP subprocesses without serializing them into RunRequest metadata.
func WithExtraEnv(env map[string]string) Option {
	return func(p *Provider) {
		p.env = cloneStringMap(env)
	}
}

// New creates a desktop Claude CLI provider.
func New(opts ...Option) *Provider {
	p := &Provider{command: "claude"}
	for _, opt := range opts {
		opt(p)
	}
	p.run = p.runClaude
	return p
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return providerName
}

// Capabilities returns the desktop adapter capabilities.
func (p *Provider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{
		Provider:           providerName,
		SupportsStreaming:  true,
		SupportsToolCalls:  true,
		SupportsMCP:        true,
		SupportsApprovals:  false,
		SupportsUsage:      true,
		SupportsArtifacts:  true,
		SupportsResume:     true,
		SupportsBackground: false,
	}, nil
}

// StartRun starts the command and streams normalized runtime events.
func (p *Provider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if p.run == nil {
		p.run = p.runClaude
	}
	if req.OutputSchema != nil {
		if err := agentruntime.ValidateOutputSchema(req.OutputSchema); err != nil {
			return nil, err
		}
	}

	events := make(chan agentruntime.Event, 32)
	runID := req.RunID
	if runID == "" {
		runID = fmt.Sprintf("desktop-run-%d", time.Now().UnixNano())
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

		if err := p.run(ctx, req, emit); err != nil {
			phaseFailed := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
			phaseFailed.Status = agentruntime.StatusFailed
			phaseFailed.Message = err.Error()
			emit(phaseFailed)

			runFailed := agentruntime.NewEvent(agentruntime.EventRunFailed)
			runFailed.Status = agentruntime.StatusFailed
			runFailed.Message = err.Error()
			emit(runFailed)
			return
		}

		phaseDone := agentruntime.NewEvent(agentruntime.EventPhaseCompleted)
		phaseDone.Status = agentruntime.StatusSucceeded
		emit(phaseDone)

		runDone := agentruntime.NewEvent(agentruntime.EventRunCompleted)
		runDone.Status = agentruntime.StatusSucceeded
		emit(runDone)
	}()

	return events, nil
}

// ResumeRun is represented through conversationId metadata on a new StartRun.
func (p *Provider) ResumeRun(ctx context.Context, runID string, input agentruntime.RunInput) (agentruntime.EventStream, error) {
	req := agentruntime.RunRequest{
		RunID:    runID,
		Role:     agentruntime.RoleChat,
		Provider: providerName,
		Messages: input.Messages,
		Metadata: input.Metadata,
	}
	return p.StartRun(ctx, req)
}

// CancelRun is handled by canceling the context passed to StartRun.
func (p *Provider) CancelRun(context.Context, string) error {
	return nil
}

func (p *Provider) runClaude(ctx context.Context, req agentruntime.RunRequest, emit func(agentruntime.Event) bool) error {
	cmd := exec.CommandContext(ctx, p.command, BuildArgs(req)...)
	cmd.Dir = req.WorkDir
	cmd.Env = commandEnv(cmd.Environ(), req.Metadata, p.env)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			event := agentruntime.NewEvent(agentruntime.EventLogMessage)
			event.Message = line
			if !emit(event) {
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			event := agentruntime.NewEvent(agentruntime.EventProviderRaw)
			if raw := json.RawMessage(line); json.Valid(raw) {
				event.Raw = raw
			} else {
				event.Message = line
			}
			if !emit(event) {
				return
			}
			if truthy(req.Metadata["normalizeClaudeResult"]) {
				for _, normalized := range normalizedClaudeEvents(line) {
					if !emit(normalized) {
						return
					}
				}
			}
		}
	}()

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude command failed: %w", err)
	}
	return nil
}

// BuildArgs converts a runtime request into Claude CLI stream arguments.
func BuildArgs(req agentruntime.RunRequest) []string {
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[0].Content
	}

	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
	}
	if req.Instructions != "" {
		args = append(args, "--system-prompt", req.Instructions)
	}
	if conversationID := req.Metadata["conversationId"]; conversationID != "" {
		args = append(args, "-r", conversationID)
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.Reasoning != nil && req.Reasoning.Effort != "" {
		args = append(args, "--effort", req.Reasoning.Effort)
	}
	if len(req.MCPServers) > 0 {
		if config, err := mcpconfig.ClaudeJSON(req.MCPServers); err == nil && config != "" {
			args = append(args, "--mcp-config", config)
		}
	}

	switch req.ApprovalPolicy {
	case agentruntime.ApprovalAutoEdit:
		args = append(args, "--allowedTools", strings.Join(toolNames(req.Tools), ","))
	case agentruntime.ApprovalFullAuto:
		args = append(args, "--dangerously-skip-permissions")
	case agentruntime.ApprovalSuggest:
		// Current desktop streaming mode cannot answer interactive approvals.
		args = append(args, "--dangerously-skip-permissions")
	}
	return args
}

func commandEnv(base []string, metadata map[string]string, extra map[string]string) []string {
	env := append([]string{}, base...)
	if metadata["authMethod"] == "google-cloud" {
		if project := metadata["gcpProjectId"]; project != "" {
			env = append(env, "CLOUD_ML_PROJECT_ID="+project)
		}
		if region := metadata["gcpRegion"]; region != "" {
			env = append(env, "CLOUD_ML_REGION="+region)
		}
	} else if key := metadata["anthropicApiKey"]; key != "" {
		env = append(env, "ANTHROPIC_API_KEY="+key)
	}
	for key, value := range extra {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func normalizedClaudeEvents(line string) []agentruntime.Event {
	var payload struct {
		Type         string  `json:"type"`
		Result       any     `json:"result"`
		TotalCostUSD float64 `json:"total_cost_usd"`
		CostUSD      float64 `json:"cost_usd"`
		Usage        struct {
			InputTokens        int `json:"input_tokens"`
			OutputTokens       int `json:"output_tokens"`
			CacheReadTokens    int `json:"cache_read_input_tokens"`
			CacheCreatedTokens int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return nil
	}
	var events []agentruntime.Event
	if payload.Type == "result" {
		if text, ok := payload.Result.(string); ok && strings.TrimSpace(text) != "" {
			event := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
			event.Status = agentruntime.StatusCompleted
			event.Message = text
			events = append(events, event)
		}
	}
	usage := agentruntime.Usage{
		InputTokens:      payload.Usage.InputTokens,
		OutputTokens:     payload.Usage.OutputTokens,
		CacheReadTokens:  payload.Usage.CacheReadTokens,
		CacheWriteTokens: payload.Usage.CacheCreatedTokens,
		TotalCostUSD:     payload.TotalCostUSD,
	}
	if usage.TotalCostUSD == 0 {
		usage.TotalCostUSD = payload.CostUSD
	}
	if usage.InputTokens != 0 || usage.OutputTokens != 0 || usage.CacheReadTokens != 0 || usage.CacheWriteTokens != 0 || usage.TotalCostUSD != 0 {
		event := agentruntime.NewEvent(agentruntime.EventUsageUpdated)
		event.Status = agentruntime.StatusCompleted
		event.Usage = &usage
		events = append(events, event)
	}
	return events
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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

func phaseID(req agentruntime.RunRequest) string {
	if phaseID, ok := req.Metadata["phaseId"]; ok && phaseID != "" {
		return phaseID
	}
	if req.Profile != "" {
		return req.Profile
	}
	return "desktop-agent"
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
