package claudecli

import (
	"context"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/conformance"
)

func TestProviderRunConformance(t *testing.T) {
	provider := New()
	provider.run = func(_ context.Context, _ agentruntime.RunRequest, emit func(agentruntime.Event) bool) error {
		raw := agentruntime.NewEvent(agentruntime.EventProviderRaw)
		raw.Message = `{"type":"content_block_delta","delta":{"text":"hello"}}`
		emit(raw)
		log := agentruntime.NewEvent(agentruntime.EventLogMessage)
		log.Message = "token usage unavailable"
		emit(log)
		return nil
	}

	events := conformance.AssertProviderRun(t, provider, agentruntime.RunRequest{
		RunID:    "desktop-run",
		Role:     agentruntime.RoleChat,
		Profile:  "desktop-chat",
		Provider: providerName,
		Model:    "sonnet",
		WorkDir:  "/repo",
		Messages: []agentruntime.Message{{Role: "user", Content: "hello"}},
	}, conformance.RunOptions{RequireProviderRaw: true})

	if !hasEvent(events, agentruntime.EventLogMessage) {
		t.Fatalf("events should include %q", agentruntime.EventLogMessage)
	}
}

func TestBuildArgs(t *testing.T) {
	args := BuildArgs(agentruntime.RunRequest{
		Model:        "sonnet",
		Instructions: "Be concise",
		Messages: []agentruntime.Message{
			{Role: "user", Content: "hello"},
		},
		Tools: []agentruntime.ToolRef{
			{Name: "Edit"},
			{Name: "Write"},
		},
		ApprovalPolicy: agentruntime.ApprovalAutoEdit,
		Reasoning:      &agentruntime.ReasoningOptions{Effort: "high"},
		MCPServers: []agentruntime.MCPServerRef{
			{Label: "datadog", Command: "npx", Args: []string{"-y", "@datadog/mcp-server"}},
		},
		Metadata: map[string]string{
			"conversationId": "abc123",
		},
	})

	argString := strings.Join(args, "\x00")
	for _, want := range []string{"-p", "hello", "--output-format", "stream-json", "--verbose", "--system-prompt", "Be concise", "-r", "abc123", "--model", "sonnet", "--effort", "high", "--mcp-config", "datadog", "--allowedTools", "Edit,Write"} {
		if !strings.Contains(argString, want) {
			t.Fatalf("args = %v, missing %q", args, want)
		}
	}
}

func TestCommandEnv(t *testing.T) {
	anthropic := commandEnv([]string{"PATH=/bin"}, map[string]string{
		"authMethod":      "anthropic-api",
		"anthropicApiKey": "secret",
	}, map[string]string{"DD_API_KEY": "datadog"})
	if !containsEnv(anthropic, "ANTHROPIC_API_KEY=secret") {
		t.Fatalf("anthropic env = %v, want API key", anthropic)
	}
	if !containsEnv(anthropic, "DD_API_KEY=datadog") {
		t.Fatalf("anthropic env = %v, want extra env", anthropic)
	}

	gcp := commandEnv([]string{"PATH=/bin"}, map[string]string{
		"authMethod":   "google-cloud",
		"gcpProjectId": "project",
		"gcpRegion":    "us-central1",
	}, nil)
	if !containsEnv(gcp, "CLOUD_ML_PROJECT_ID=project") || !containsEnv(gcp, "CLOUD_ML_REGION=us-central1") {
		t.Fatalf("gcp env = %v, want project and region", gcp)
	}
}

func TestNormalizedClaudeEvents(t *testing.T) {
	events := normalizedClaudeEvents(`{"type":"result","result":"final report","total_cost_usd":0.25,"usage":{"input_tokens":10,"output_tokens":20,"cache_read_input_tokens":3}}`)
	if len(events) != 2 {
		t.Fatalf("events len = %d, want message and usage", len(events))
	}
	if events[0].Type != agentruntime.EventMessageCompleted || events[0].Message != "final report" {
		t.Fatalf("first event = %#v, want completed message", events[0])
	}
	if events[1].Type != agentruntime.EventUsageUpdated || events[1].Usage == nil || events[1].Usage.InputTokens != 10 || events[1].Usage.OutputTokens != 20 || events[1].Usage.CacheReadTokens != 3 || events[1].Usage.TotalCostUSD != 0.25 {
		t.Fatalf("second event = %#v, want parsed usage", events[1])
	}
}

func hasEvent(events []agentruntime.Event, eventType agentruntime.EventType) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func containsEnv(env []string, want string) bool {
	for _, value := range env {
		if value == want {
			return true
		}
	}
	return false
}
