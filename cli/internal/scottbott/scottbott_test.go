package scottbott

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
)

type scottBottFakeProvider struct{}

func (scottBottFakeProvider) Name() string {
	return "fake"
}

func (scottBottFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (scottBottFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (scottBottFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (scottBottFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestReviewOutputSchemaIsValidJSON(t *testing.T) {
	schema := reviewOutputSchema()
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

func TestRunReviewProviderBuildsSkillRuntimeRequest(t *testing.T) {
	cfg := &config.Config{
		ReviewSkill: "peer-review",
		Claude: config.ClaudeConfig{
			Effort:              "medium",
			EnablePromptCaching: true,
			Models: config.ModelConfig{
				Reviewer: "reviewer-model",
			},
		},
	}
	reviewer := newScottBottWithProvider("/repo", "reviewer-2", "peer-review", cfg, scottBottFakeProvider{})

	var captured agentruntime.RunRequest
	reviewer.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest, onEvent func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		if onEvent == nil {
			t.Fatal("expected provider raw event observer")
		}
		return `{"passed":true,"score":95,"summary":"looks good","issues":[],"praise":["nice"],"guidance":""}`, &cost.Usage{InputTokens: 9}, nil
	}

	response, usage, err := reviewer.runReviewProvider(context.Background(), "", "review this", true)
	if err != nil {
		t.Fatalf("runReviewProvider error: %v", err)
	}
	if response == "" {
		t.Fatal("response should not be empty")
	}
	if usage == nil || usage.InputTokens != 9 {
		t.Fatalf("usage = %#v, want input tokens", usage)
	}
	if captured.RunID != "reviewer-2" {
		t.Fatalf("RunID = %q, want reviewer-2", captured.RunID)
	}
	if captured.Role != agentruntime.RoleReviewer {
		t.Fatalf("Role = %q, want reviewer", captured.Role)
	}
	if captured.Profile != "peer-review" {
		t.Fatalf("Profile = %q, want peer-review", captured.Profile)
	}
	if captured.Provider != "fake" {
		t.Fatalf("Provider = %q, want fake", captured.Provider)
	}
	if captured.Model != "reviewer-model" {
		t.Fatalf("Model = %q, want reviewer-model", captured.Model)
	}
	if captured.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", captured.WorkDir)
	}
	if captured.ApprovalPolicy != agentruntime.ApprovalSuggest {
		t.Fatalf("ApprovalPolicy = %q, want suggest", captured.ApprovalPolicy)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "medium" {
		t.Fatalf("Reasoning = %#v, want medium effort", captured.Reasoning)
	}
	if captured.OutputSchema == nil || captured.OutputSchema.Name != "review_result" {
		t.Fatalf("OutputSchema = %#v, want review schema", captured.OutputSchema)
	}
	if captured.Metadata["phaseId"] != "reviewer-2" {
		t.Fatalf("phaseId = %q, want reviewer-2", captured.Metadata["phaseId"])
	}
	if captured.Metadata["claudeAgent"] != "peer-review" {
		t.Fatalf("claudeAgent = %q, want peer-review", captured.Metadata["claudeAgent"])
	}
	if captured.Metadata["outputFormat"] != "text" {
		t.Fatalf("outputFormat = %q, want text", captured.Metadata["outputFormat"])
	}
	if captured.Metadata["enablePromptCaching"] != "true" {
		t.Fatalf("enablePromptCaching = %q, want true", captured.Metadata["enablePromptCaching"])
	}
}

func TestReviewFallsBackThroughRuntimeProvider(t *testing.T) {
	cfg := &config.Config{
		ReviewSkill: "missing-skill",
		Claude:      config.ClaudeConfig{},
	}
	reviewer := newScottBottWithProvider("/repo", "reviewer-1", "missing-skill", cfg, scottBottFakeProvider{})
	reviewer.outputDir = t.TempDir()

	var requests []agentruntime.RunRequest
	reviewer.runText = func(_ context.Context, _ agentruntime.Provider, req agentruntime.RunRequest, _ func(agentruntime.Event)) (string, *cost.Usage, error) {
		requests = append(requests, req)
		if len(requests) == 1 {
			return "", nil, errors.New("agent missing")
		}
		return `{"passed":true,"score":90,"summary":"fallback passed","issues":[],"praise":[],"guidance":""}`, &cost.Usage{OutputTokens: 12}, nil
	}

	result, usage, err := reviewer.Review(context.Background(), "ticket", "diff")
	if err != nil {
		t.Fatalf("Review error: %v", err)
	}
	if result == nil || !result.Passed || result.Summary != "fallback passed" {
		t.Fatalf("result = %#v, want fallback pass", result)
	}
	if usage == nil || usage.OutputTokens != 12 {
		t.Fatalf("usage = %#v, want fallback usage", usage)
	}
	if len(requests) != 2 {
		t.Fatalf("captured %d requests, want skill plus fallback", len(requests))
	}
	if requests[0].Metadata["claudeAgent"] != "missing-skill" {
		t.Fatalf("first request metadata = %#v, want skill agent", requests[0].Metadata)
	}
	if _, ok := requests[1].Metadata["claudeAgent"]; ok {
		t.Fatalf("fallback request metadata = %#v, should not include skill agent", requests[1].Metadata)
	}
	if requests[1].Instructions == "" {
		t.Fatal("fallback request should include system prompt")
	}
}
