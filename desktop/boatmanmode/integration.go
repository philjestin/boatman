package boatmanmode

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

// Integration provides boatmanmode functionality via subprocess calls
type Integration struct {
	boatmanmodePath string
	repoPath        string
	linearAPIKey    string
	claudeAPIKey    string
}

// NewIntegration creates a new boatmanmode integration
func NewIntegration(linearAPIKey, claudeAPIKey, repoPath string) (*Integration, error) {
	// Find boatmanmode binary in PATH or use default location
	boatmanmodePath, err := exec.LookPath("boatman")
	if err != nil {
		// Try monorepo location first
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = "/Users/pmiddleton"
		}
		monorepoPath := filepath.Join(homeDir, "workspace/personal/boatman-ecosystem/cli/boatman")

		// Fall back to old standalone location
		standalonePath := filepath.Join(homeDir, "workspace/handshake/boatmanmode/boatman")

		// Check monorepo location first
		if _, err := os.Stat(monorepoPath); err == nil {
			boatmanmodePath = monorepoPath
		} else if _, err := os.Stat(standalonePath); err == nil {
			boatmanmodePath = standalonePath
		} else {
			return nil, fmt.Errorf("boatman binary not found in PATH, monorepo (%s), or standalone (%s). Please build: cd ~/workspace/personal/boatman-ecosystem && make build-cli", monorepoPath, standalonePath)
		}
	}

	return &Integration{
		boatmanmodePath: boatmanmodePath,
		repoPath:        repoPath,
		linearAPIKey:    linearAPIKey,
		claudeAPIKey:    claudeAPIKey,
	}, nil
}

// ExecuteTicket runs the full boatmanmode workflow for a Linear ticket
func (i *Integration) ExecuteTicket(ctx context.Context, ticketID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, i.boatmanmodePath,
		"work", ticketID,
	)
	cmd.Dir = i.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("boatmanmode execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, return raw output
		return map[string]interface{}{
			"success": true,
			"output":  string(output),
		}, nil
	}

	return result, nil
}

// ExecutePrompt runs the boatmanmode workflow with a custom prompt
func (i *Integration) ExecutePrompt(ctx context.Context, prompt string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, i.boatmanmodePath,
		"work",
		"--prompt", prompt,
	)
	cmd.Dir = i.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("boatmanmode execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, return raw output
		return map[string]interface{}{
			"success": true,
			"output":  string(output),
		}, nil
	}

	return result, nil
}

// BoatmanEvent represents a structured event from boatmanmode
type BoatmanEvent struct {
	Type        string                 `json:"type"`
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// MessageCallback is called for non-JSON output lines to route them as session messages
type MessageCallback func(role, content string)

// Config contains configuration for boatmanmode execution
type Config struct {
	MaxIterations int
	BaseBranch    string
	AutoPR        bool
	ReviewSkill   string
	Timeout       int
}

// StreamExecution runs the workflow with live streaming output
// It parses structured JSON events for agent/task tracking and emits them via Wails runtime
// mode can be "ticket" or "prompt"
// onMessage, if non-nil, receives non-JSON output lines as messages for the session
func (i *Integration) StreamExecution(ctx context.Context, sessionID string, input string, mode string, outputChan chan<- string, onMessage MessageCallback, config *Config) (map[string]interface{}, error) {
	// Build command arguments
	args := []string{"work"}

	if mode == "ticket" {
		args = append(args, input)
	} else {
		// prompt mode
		args = append(args, "--prompt", input)
	}

	// Add configuration flags if provided
	if config != nil {
		if config.MaxIterations > 0 {
			args = append(args, "--max-iterations", fmt.Sprintf("%d", config.MaxIterations))
		}
		if config.BaseBranch != "" {
			args = append(args, "--base-branch", config.BaseBranch)
		}
		if !config.AutoPR {
			args = append(args, "--auto-pr=false")
		}
		if config.ReviewSkill != "" {
			args = append(args, "--review-skill", config.ReviewSkill)
		}
		if config.Timeout > 0 {
			args = append(args, "--timeout", fmt.Sprintf("%d", config.Timeout))
		}
	}

	cmd := exec.CommandContext(ctx, i.boatmanmodePath, args...)
	cmd.Dir = i.repoPath

	// Set environment variables for authentication
	// Start with parent environment so PATH, HOME, etc. are available
	cmd.Env = os.Environ()
	if i.linearAPIKey != "" {
		cmd.Env = append(cmd.Env, "LINEAR_API_KEY="+i.linearAPIKey)
	}
	if i.claudeAPIKey != "" {
		cmd.Env = append(cmd.Env, "ANTHROPIC_API_KEY="+i.claudeAPIKey)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr to a buffer so we can include it in error messages
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	fmt.Printf("[boatmanmode] Starting command: %s %v in directory: %s\n", i.boatmanmodePath, cmd.Args[1:], i.repoPath)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start boatmanmode at %s: %w", i.boatmanmodePath, err)
	}

	fmt.Printf("[boatmanmode] Command started successfully, PID: %d\n", cmd.Process.Pid)

	// Stream stdout and parse JSON events
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[boatmanmode] Recovered from panic in stdout reader: %v\n", r)
			}
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// Try to parse as structured event
			var event BoatmanEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil && event.Type != "" {
				// Emit structured event via Wails runtime (with panic recovery)
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("[boatmanmode] Failed to emit event: %v\n", r)
						}
					}()
					runtime.EventsEmit(ctx, "boatmanmode:event", map[string]interface{}{
						"sessionId": sessionID,
						"event":     event,
					})
				}()

				// Also send formatted output to channel
				switch event.Type {
				case "agent_started":
					outputChan <- fmt.Sprintf("Agent started: %s\n", event.Name)
				case "agent_completed":
					outputChan <- fmt.Sprintf("Agent completed: %s (status: %s)\n", event.Name, event.Status)
				case "task_created":
					outputChan <- fmt.Sprintf("Task created: %s\n", event.Name)
				case "task_updated":
					outputChan <- fmt.Sprintf("Task updated: %s (status: %s)\n", event.Name, event.Status)
				case "progress":
					outputChan <- fmt.Sprintf("%s\n", event.Message)
				case "claude_stream":
					// Forward raw Claude stream line to the session for parsing
					if onMessage != nil && event.Message != "" {
						onMessage("claude_stream", event.Message)
					}
				default:
					outputChan <- line + "\n"
				}
			} else {
				// Regular output line (not JSON event)
				outputChan <- line + "\n"
				if onMessage != nil {
					onMessage("assistant", line)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("[boatmanmode] Scanner error: %v\n", err)
		}
	}()

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		stderr := stderrBuf.String()
		if stderr != "" {
			return nil, fmt.Errorf("boatmanmode execution failed: %w\nStderr: %s", err, stderr)
		}
		return nil, fmt.Errorf("boatmanmode execution failed: %w\nCommand: %s %v", err, i.boatmanmodePath, cmd.Args[1:])
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// FetchTickets retrieves tickets from Linear (via boatmanmode CLI)
func (i *Integration) FetchTickets(ctx context.Context, filters map[string]string) ([]map[string]interface{}, error) {
	args := []string{"list-tickets"}

	// Add filters as flags
	if labels, ok := filters["labels"]; ok {
		args = append(args, "--labels", labels)
	}

	cmd := exec.CommandContext(ctx, i.boatmanmodePath, args...)
	cmd.Dir = i.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickets: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var tickets []map[string]interface{}
	if err := json.Unmarshal(output, &tickets); err != nil {
		return nil, fmt.Errorf("failed to parse tickets: %w", err)
	}

	return tickets, nil
}
