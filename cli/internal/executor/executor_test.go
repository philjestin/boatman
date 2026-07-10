package executor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/planner"
	"github.com/philjestin/boatmanmode/internal/task"
)

type executorFakeProvider struct{}

func (executorFakeProvider) Name() string {
	return "fake"
}

func (executorFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (executorFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (executorFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (executorFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestExecutorToolsFollowToolFlag(t *testing.T) {
	if tools := executorTools(false); tools != nil {
		t.Fatalf("executorTools(false) = %#v, want nil", tools)
	}

	tools := executorTools(true)
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
		if tool.Kind != "local" {
			t.Fatalf("tool %s kind = %q, want local", tool.Name, tool.Kind)
		}
	}

	want := "Read,Write,Edit,Bash,Grep,Glob"
	if strings.Join(names, ",") != want {
		t.Fatalf("tools = %v, want %s", names, want)
	}
}

func TestRunModelBuildsRuntimeProviderRequest(t *testing.T) {
	cfg := &config.Config{
		EnableTools: true,
		Claude: config.ClaudeConfig{
			Effort:              "high",
			EnablePromptCaching: true,
			Models: config.ModelConfig{
				Executor: "executor-model",
			},
		},
	}
	executor := newExecutorWithProvider("/repo", cfg, executorFakeProvider{}, "executor", "executor", cfg.Claude.Models.Executor)

	var captured agentruntime.RunRequest
	executor.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest, onEvent func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		if onEvent == nil {
			t.Fatal("expected provider raw event observer")
		}
		return "done", &cost.Usage{InputTokens: 5}, nil
	}

	response, usage, err := executor.runModel(context.Background(), agentruntime.RoleExecutor, "executor", "executor", "run-1", "system", "prompt")
	if err != nil {
		t.Fatalf("runModel error: %v", err)
	}
	if response != "done" {
		t.Fatalf("response = %q, want done", response)
	}
	if usage == nil || usage.InputTokens != 5 {
		t.Fatalf("usage = %#v, want input tokens", usage)
	}
	if captured.RunID != "run-1" {
		t.Fatalf("RunID = %q, want run-1", captured.RunID)
	}
	if captured.Role != agentruntime.RoleExecutor {
		t.Fatalf("Role = %q, want executor", captured.Role)
	}
	if captured.Profile != "executor" {
		t.Fatalf("Profile = %q, want executor", captured.Profile)
	}
	if captured.Provider != "fake" {
		t.Fatalf("Provider = %q, want fake", captured.Provider)
	}
	if captured.Model != "executor-model" {
		t.Fatalf("Model = %q, want executor-model", captured.Model)
	}
	if captured.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", captured.WorkDir)
	}
	if captured.Instructions != "system" {
		t.Fatalf("Instructions = %q, want system", captured.Instructions)
	}
	if len(captured.Messages) != 1 || captured.Messages[0].Content != "prompt" {
		t.Fatalf("Messages = %#v, want user prompt", captured.Messages)
	}
	if captured.ApprovalPolicy != agentruntime.ApprovalFullAuto {
		t.Fatalf("ApprovalPolicy = %q, want full_auto", captured.ApprovalPolicy)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "high" {
		t.Fatalf("Reasoning = %#v, want high effort", captured.Reasoning)
	}
	if len(captured.Tools) != 6 {
		t.Fatalf("Tools = %#v, want executor toolset", captured.Tools)
	}
	if captured.Metadata["phaseId"] != "executor" {
		t.Fatalf("phaseId = %q, want executor", captured.Metadata["phaseId"])
	}
	if captured.Metadata["useTmux"] != "true" {
		t.Fatalf("useTmux = %q, want true", captured.Metadata["useTmux"])
	}
	if captured.Metadata["enablePromptCaching"] != "true" {
		t.Fatalf("enablePromptCaching = %q, want true", captured.Metadata["enablePromptCaching"])
	}
}

func TestExecuteWithPlanUsesRuntimeProviderAndDetectsFiles(t *testing.T) {
	worktree := t.TempDir()
	initGitRepo(t, worktree)

	cfg := &config.Config{
		EnableTools: true,
		Claude: config.ClaudeConfig{
			Models: config.ModelConfig{Executor: "executor-model"},
		},
	}
	executor := newExecutorWithProvider(worktree, cfg, executorFakeProvider{}, "executor", "executor", cfg.Claude.Models.Executor)

	var captured agentruntime.RunRequest
	executor.runText = func(_ context.Context, _ agentruntime.Provider, req agentruntime.RunRequest, _ func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if err := os.WriteFile(filepath.Join(worktree, "new.txt"), []byte("hello\n"), 0644); err != nil {
			t.Fatalf("write provider output: %v", err)
		}
		return "## Analysis\nImplemented the task\n## Result\nDone", &cost.Usage{OutputTokens: 8}, nil
	}

	promptTask := task.NewPromptTask("Create a file", "Create file", "create-file")
	result, usage, err := executor.ExecuteWithPlan(context.Background(), promptTask, &planner.Plan{
		Summary:       "Create a file",
		Approach:      []string{"write new.txt"},
		RelevantFiles: []string{"new.txt"},
	})
	if err != nil {
		t.Fatalf("ExecuteWithPlan error: %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("result = %#v, want success", result)
	}
	if len(result.FilesChanged) != 1 || result.FilesChanged[0] != "new.txt" {
		t.Fatalf("FilesChanged = %v, want new.txt", result.FilesChanged)
	}
	if result.Summary != "Implemented the task" {
		t.Fatalf("Summary = %q, want implemented summary", result.Summary)
	}
	if usage == nil || usage.OutputTokens != 8 {
		t.Fatalf("usage = %#v, want output tokens", usage)
	}
	if captured.Role != agentruntime.RoleExecutor {
		t.Fatalf("Role = %q, want executor", captured.Role)
	}
	if captured.RunID != "executor-"+promptTask.GetID() {
		t.Fatalf("RunID = %q, want executor task ID", captured.RunID)
	}
	if !strings.Contains(captured.Messages[0].Content, "# Execution Plan") {
		t.Fatal("prompt should include planner handoff")
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}
}
