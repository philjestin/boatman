package triage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TriageOptions configures a triage pipeline run from the desktop.
type TriageOptions struct {
	Teams        []string `json:"teams"`
	States       []string `json:"states"`
	Limit        int      `json:"limit"`
	TicketIDs    []string `json:"ticketIds"`
	PostComments bool     `json:"postComments"`
	DryRun       bool     `json:"dryRun"`
	OutputDir    string   `json:"outputDir"`
	Concurrency   int      `json:"concurrency"`
	GeneratePlans bool     `json:"generatePlans"`
	RepoPath      string   `json:"repoPath"`
}

// TriageEvent represents a structured event from the triage CLI.
type TriageEvent struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Status  string                 `json:"status,omitempty"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Integration provides triage functionality via subprocess calls to the boatman CLI.
type Integration struct {
	boatmanPath  string
	repoPath     string
	linearAPIKey string
	claudeAPIKey string
}

// NewIntegration creates a new triage integration.
func NewIntegration(linearAPIKey, claudeAPIKey, repoPath string) (*Integration, error) {
	boatmanPath, err := exec.LookPath("boatman")
	if err != nil {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = "/Users/pmiddleton"
		}
		monorepoPath := filepath.Join(homeDir, "workspace/personal/boatman-ecosystem/cli/boatman")
		standalonePath := filepath.Join(homeDir, "workspace/handshake/boatmanmode/boatman")

		if _, err := os.Stat(monorepoPath); err == nil {
			boatmanPath = monorepoPath
		} else if _, err := os.Stat(standalonePath); err == nil {
			boatmanPath = standalonePath
		} else {
			return nil, fmt.Errorf("boatman binary not found in PATH, monorepo (%s), or standalone (%s)", monorepoPath, standalonePath)
		}
	}

	return &Integration{
		boatmanPath:  boatmanPath,
		repoPath:     repoPath,
		linearAPIKey: linearAPIKey,
		claudeAPIKey: claudeAPIKey,
	}, nil
}

// StreamTriageExecution runs `boatman triage --emit-events` as a subprocess,
// parsing JSON events from stdout and emitting them as Wails events.
// The onEvent callback is called for each parsed event, allowing the caller
// to handle events (like storing triage_complete results) synchronously.
func (i *Integration) StreamTriageExecution(ctx context.Context, sessionID string, opts TriageOptions, outputChan chan<- string, onEvent func(TriageEvent)) error {
	args := []string{"triage", "--emit-events"}

	if len(opts.TicketIDs) > 0 {
		for _, id := range opts.TicketIDs {
			args = append(args, "--ticket-ids", id)
		}
	} else if len(opts.Teams) > 0 {
		for _, team := range opts.Teams {
			args = append(args, "--teams", team)
		}
	}

	if len(opts.States) > 0 {
		for _, state := range opts.States {
			args = append(args, "--states", state)
		}
	}

	if opts.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", opts.Limit))
	}

	if opts.PostComments {
		args = append(args, "--post-comments")
	}

	if opts.DryRun {
		args = append(args, "--dry-run")
	}

	if opts.OutputDir != "" {
		args = append(args, "--output-dir", opts.OutputDir)
	}

	if opts.Concurrency > 0 {
		args = append(args, "--concurrency", fmt.Sprintf("%d", opts.Concurrency))
	}

	if opts.GeneratePlans {
		args = append(args, "--generate-plans")
		repoPath := opts.RepoPath
		if repoPath == "" {
			repoPath = i.repoPath
		}
		args = append(args, "--repo-path", repoPath)
	}

	cmd := exec.CommandContext(ctx, i.boatmanPath, args...)
	cmd.Dir = i.repoPath

	cmd.Env = os.Environ()
	if i.linearAPIKey != "" {
		cmd.Env = append(cmd.Env, "LINEAR_API_KEY="+i.linearAPIKey)
	}
	if i.claudeAPIKey != "" {
		cmd.Env = append(cmd.Env, "ANTHROPIC_API_KEY="+i.claudeAPIKey)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	fmt.Printf("[triage] Starting command: %s %v in directory: %s\n", i.boatmanPath, args, i.repoPath)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start triage at %s: %w", i.boatmanPath, err)
	}

	fmt.Printf("[triage] Command started, PID: %d\n", cmd.Process.Pid)

	// Parse JSON events from stdout
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)

		scanner := bufio.NewScanner(stdout)
		scanBuf := make([]byte, 0, 64*1024)
		scanner.Buffer(scanBuf, 10*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			var event TriageEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil && event.Type != "" {
				// Call the onEvent callback first (stores result synchronously)
				if onEvent != nil {
					onEvent(event)
				}

				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("[triage] Failed to emit event: %v\n", r)
						}
					}()
					runtime.EventsEmit(ctx, "triage:event", map[string]interface{}{
						"sessionId": sessionID,
						"event":     event,
					})
				}()

				outputChan <- fmt.Sprintf("[%s] %s\n", event.Type, event.Message)
			} else {
				outputChan <- line + "\n"
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("[triage] Scanner error: %v\n", err)
		}
	}()

	// Wait for stdout reader to finish BEFORE cmd.Wait(), because
	// cmd.Wait() closes the stdout pipe (per Go docs for StdoutPipe).
	<-readerDone
	waitErr := cmd.Wait()

	// Always log stderr — it contains slog output from the CLI with
	// scoring errors, API failures, and other diagnostics.
	stderr := stderrBuf.String()
	if stderr != "" {
		fmt.Printf("[triage] Subprocess stderr:\n%s\n", stderr)
	}

	if waitErr != nil {
		if stderr != "" {
			return fmt.Errorf("triage execution failed: %w\nStderr: %s", waitErr, stderr)
		}
		return fmt.Errorf("triage execution failed: %w", waitErr)
	}

	return nil
}
