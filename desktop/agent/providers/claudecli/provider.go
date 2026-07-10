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
	cmd.Env = commandEnv(cmd.Environ(), req.Metadata)

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

func commandEnv(base []string, metadata map[string]string) []string {
	env := append([]string{}, base...)
	if metadata["authMethod"] == "google-cloud" {
		if project := metadata["gcpProjectId"]; project != "" {
			env = append(env, "CLOUD_ML_PROJECT_ID="+project)
		}
		if region := metadata["gcpRegion"]; region != "" {
			env = append(env, "CLOUD_ML_REGION="+region)
		}
		return env
	}
	if key := metadata["anthropicApiKey"]; key != "" {
		env = append(env, "ANTHROPIC_API_KEY="+key)
	}
	return env
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
