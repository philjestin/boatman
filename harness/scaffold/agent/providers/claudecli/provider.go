// Package claudecli adapts Claude CLI text output for scaffold enhancement.
package claudecli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

const providerName = "claude-cli"

type runFunc func(context.Context, agentruntime.RunRequest) (string, error)

// Provider runs scaffold enhancement prompts through Claude CLI.
type Provider struct {
	command string
	run     runFunc
}

// Option configures Provider.
type Option func(*Provider)

// WithCommand overrides the executable used for Claude CLI calls.
func WithCommand(command string) Option {
	return func(p *Provider) {
		p.command = command
	}
}

// New creates a scaffold Claude CLI provider.
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

// Capabilities returns the scaffold adapter capabilities.
func (p *Provider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{
		Provider:              providerName,
		SupportsStreaming:     false,
		SupportsBackground:    false,
		SupportsResume:        false,
		SupportsToolCalls:     false,
		SupportsMCP:           false,
		SupportsApprovals:     false,
		SupportsStructuredOut: false,
		SupportsUsage:         false,
	}, nil
}

// StartRun starts a one-shot text generation request.
func (p *Provider) StartRun(ctx context.Context, req agentruntime.RunRequest) (agentruntime.EventStream, error) {
	if p.run == nil {
		p.run = p.runClaude
	}
	if req.OutputSchema != nil {
		if err := agentruntime.ValidateOutputSchema(req.OutputSchema); err != nil {
			return nil, err
		}
	}

	events := make(chan agentruntime.Event, 8)
	runID := req.RunID
	if runID == "" {
		runID = fmt.Sprintf("scaffold-run-%d", time.Now().UnixNano())
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

		response, err := p.run(ctx, req)
		if err != nil {
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

		if response != "" {
			message := agentruntime.NewEvent(agentruntime.EventMessageCompleted)
			message.Message = response
			emit(message)
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

// ResumeRun is not supported by the scaffold adapter.
func (p *Provider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, fmt.Errorf("%s does not support resume", providerName)
}

// CancelRun is handled by canceling the StartRun context.
func (p *Provider) CancelRun(context.Context, string) error {
	return nil
}

func (p *Provider) runClaude(ctx context.Context, req agentruntime.RunRequest) (string, error) {
	cmd := exec.CommandContext(ctx, p.command, BuildArgs(req)...)
	if req.WorkDir != "" {
		cmd.Dir = req.WorkDir
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("claude exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// BuildArgs converts a runtime request into Claude CLI text arguments.
func BuildArgs(req agentruntime.RunRequest) []string {
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[0].Content
	}

	args := []string{
		"-p",
		"--output-format", "text",
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.Instructions != "" {
		args = append(args, "--system-prompt", req.Instructions)
	}
	args = append(args, prompt)
	return args
}

func phaseID(req agentruntime.RunRequest) string {
	if phaseID, ok := req.Metadata["phaseId"]; ok && phaseID != "" {
		return phaseID
	}
	if req.Profile != "" {
		return req.Profile
	}
	return "scaffold-enhance"
}
