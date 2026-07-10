package planner

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
	"github.com/philjestin/boatmanmode/internal/task"
)

type plannerFakeProvider struct{}

func (plannerFakeProvider) Name() string {
	return "fake"
}

func (plannerFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (plannerFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (plannerFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (plannerFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestPlannerOutputSchemaIsValidJSON(t *testing.T) {
	schema := plannerOutputSchema()
	if schema == nil {
		t.Fatal("schema should not be nil")
	}
	if !schema.Strict {
		t.Fatal("schema should be strict")
	}
	if !json.Valid(schema.Schema) {
		t.Fatalf("schema is not valid JSON: %s", schema.Schema)
	}
	if err := agentruntime.ValidateOutputSchema(schema); err != nil {
		t.Fatalf("schema should pass runtime validation: %v", err)
	}
}

func TestPlannerToolsFollowToolFlag(t *testing.T) {
	if tools := plannerTools(false); tools != nil {
		t.Fatalf("plannerTools(false) = %#v, want nil", tools)
	}

	tools := plannerTools(true)
	if len(tools) != 3 {
		t.Fatalf("plannerTools(true) returned %d tools, want 3", len(tools))
	}

	names := []string{tools[0].Name, tools[1].Name, tools[2].Name}
	if strings.Join(names, ",") != "Read,Grep,Glob" {
		t.Fatalf("tool names = %v, want Read/Grep/Glob", names)
	}
	for _, tool := range tools {
		if tool.Kind != "local" {
			t.Fatalf("tool %s kind = %q, want local", tool.Name, tool.Kind)
		}
	}
}

func TestAnalyzeUsesRuntimeProviderRequest(t *testing.T) {
	cfg := &config.Config{
		EnableTools: true,
		Claude: config.ClaudeConfig{
			Effort: "high",
			Models: config.ModelConfig{
				Planner: "planner-model",
			},
		},
	}
	planner := newPlannerWithProvider("/repo", cfg, plannerFakeProvider{})

	var captured agentruntime.RunRequest
	planner.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest, onEvent func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		if onEvent == nil {
			t.Fatal("expected provider raw event observer")
		}
		return `{"summary":"plan summary","approach":["read code"],"relevant_files":["cli/main.go"],"relevant_dirs":["cli/"],"existing_patterns":["follow existing command setup"],"test_strategy":"go test ./...","warnings":[]}`, &cost.Usage{InputTokens: 11}, nil
	}

	promptTask := task.NewPromptTask("Add support for new provider features", "Provider task", "provider-task")
	plan, usage, err := planner.Analyze(context.Background(), promptTask)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if plan.Summary != "plan summary" {
		t.Fatalf("Summary = %q, want plan summary", plan.Summary)
	}
	if usage == nil || usage.InputTokens != 11 {
		t.Fatalf("usage = %#v, want input tokens", usage)
	}
	if captured.RunID != "planner-"+promptTask.GetID() {
		t.Fatalf("RunID = %q, want planner task ID", captured.RunID)
	}
	if captured.Role != agentruntime.RolePlanner {
		t.Fatalf("Role = %q, want %q", captured.Role, agentruntime.RolePlanner)
	}
	if captured.Profile != "planner" {
		t.Fatalf("Profile = %q, want planner", captured.Profile)
	}
	if captured.Provider != "fake" {
		t.Fatalf("Provider = %q, want fake", captured.Provider)
	}
	if captured.Model != "planner-model" {
		t.Fatalf("Model = %q, want planner-model", captured.Model)
	}
	if captured.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", captured.WorkDir)
	}
	if captured.ApprovalPolicy != agentruntime.ApprovalFullAuto {
		t.Fatalf("ApprovalPolicy = %q, want %q", captured.ApprovalPolicy, agentruntime.ApprovalFullAuto)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "high" {
		t.Fatalf("Reasoning = %#v, want high effort", captured.Reasoning)
	}
	if len(captured.Tools) != 3 {
		t.Fatalf("Tools = %#v, want 3 planner tools", captured.Tools)
	}
	if captured.OutputSchema == nil || captured.OutputSchema.Name != "work_planner_plan" {
		t.Fatalf("OutputSchema = %#v, want work planner schema", captured.OutputSchema)
	}
	if captured.Metadata["phaseId"] != "planner" {
		t.Fatalf("phaseId = %q, want planner", captured.Metadata["phaseId"])
	}
	if !strings.Contains(captured.Instructions, "senior software architect") {
		t.Fatal("instructions should contain planner system prompt")
	}
	if len(captured.Messages) != 1 || !strings.Contains(captured.Messages[0].Content, "Provider task") {
		t.Fatalf("Messages = %#v, want task prompt", captured.Messages)
	}
}
