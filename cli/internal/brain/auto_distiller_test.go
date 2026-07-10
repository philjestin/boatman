package brain

import (
	"context"
	"os"
	"strings"
	"testing"

	harnessbrain "github.com/philjestin/boatman-ecosystem/harness/brain"
	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/memorydocs"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/cost"
)

type autoDistillerFakeProvider struct{}

func (autoDistillerFakeProvider) Name() string {
	return "fake"
}

func (autoDistillerFakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return agentruntime.Capabilities{}, nil
}

func (autoDistillerFakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (autoDistillerFakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (autoDistillerFakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestLLMDistillUsesRuntimeProvider(t *testing.T) {
	projectPath := t.TempDir()
	cfg := &config.Config{
		Claude: config.ClaudeConfig{
			Effort:              "low",
			EnablePromptCaching: true,
			Models: config.ModelConfig{
				Planner: "planner-model",
			},
		},
	}
	distiller := newAutoDistillerWithProvider(projectPath, cfg, autoDistillerFakeProvider{})
	if err := os.MkdirAll(distiller.outputDir, 0755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	var captured agentruntime.RunRequest
	distiller.runText = func(_ context.Context, provider agentruntime.Provider, req agentruntime.RunRequest, _ func(agentruntime.Event)) (string, *cost.Usage, error) {
		captured = req
		if provider.Name() != "fake" {
			t.Fatalf("provider = %q, want fake", provider.Name())
		}
		return `id: auto-payments
name: Payments Domain
version: 1
description: Payment patterns
confidence: 0.8
last_updated: 2026-06-18
triggers:
  keywords: [payments]
  entities: [Payment]
  file_patterns: [payments/]
sections:
  - title: Domain Model
    content: |
      Payment model notes.
references:
  - path: payments/service.go
    description: Payment service
`, &cost.Usage{InputTokens: 10}, nil
	}

	result, err := distiller.llmDistill(context.Background(), "payments", []harnessbrain.Signal{
		{
			Type:      harnessbrain.SignalReviewFailure,
			Domain:    "payments",
			Details:   "Agents miss idempotency checks",
			FilePaths: []string{"payments/service.go"},
			Count:     3,
		},
	}, map[string]string{"payments/service.go": "package payments\n"})
	if err != nil {
		t.Fatalf("llmDistill error: %v", err)
	}
	if result == nil || !result.UsedLLM || result.BrainID != "auto-payments" {
		t.Fatalf("result = %#v, want generated brain", result)
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Fatalf("expected generated brain file: %v", err)
	}
	if result.MemoryPath == "" {
		t.Fatalf("MemoryPath is empty, want generated memory document")
	}
	memoryStore := memorydocs.NewFileStore(memorydocs.DefaultDir(projectPath))
	memoryDoc, err := memoryStore.Read(context.Background(), "domains/payments")
	if err != nil {
		t.Fatalf("expected generated memory doc: %v", err)
	}
	if memoryDoc.SourceRunID != "brain-distiller-payments" || !strings.Contains(memoryDoc.Body, "idempotency") {
		t.Fatalf("memoryDoc = %#v, want provenance and signal details", memoryDoc)
	}
	if captured.RunID != "brain-distiller-payments" {
		t.Fatalf("RunID = %q, want brain-distiller-payments", captured.RunID)
	}
	if captured.Role != agentruntime.RoleMemory {
		t.Fatalf("Role = %q, want memory", captured.Role)
	}
	if captured.Profile != "brain-distiller" {
		t.Fatalf("Profile = %q, want brain-distiller", captured.Profile)
	}
	if captured.Provider != "fake" {
		t.Fatalf("Provider = %q, want fake", captured.Provider)
	}
	if captured.Model != "planner-model" {
		t.Fatalf("Model = %q, want planner-model", captured.Model)
	}
	if captured.WorkDir != projectPath {
		t.Fatalf("WorkDir = %q, want project path", captured.WorkDir)
	}
	if captured.ApprovalPolicy != agentruntime.ApprovalSuggest {
		t.Fatalf("ApprovalPolicy = %q, want suggest", captured.ApprovalPolicy)
	}
	if captured.Reasoning == nil || captured.Reasoning.Effort != "low" {
		t.Fatalf("Reasoning = %#v, want low effort", captured.Reasoning)
	}
	if captured.Metadata["phaseId"] != "brain-distiller" {
		t.Fatalf("phaseId = %q, want brain-distiller", captured.Metadata["phaseId"])
	}
	if captured.Metadata["enablePromptCaching"] != "true" {
		t.Fatalf("enablePromptCaching = %q, want true", captured.Metadata["enablePromptCaching"])
	}
}
