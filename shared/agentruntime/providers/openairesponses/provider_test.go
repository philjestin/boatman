package openairesponses

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/conformance"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"
)

func TestProviderImplementsRuntimeProvider(t *testing.T) {
	var _ agentruntime.Provider = New()
}

func TestCapabilities(t *testing.T) {
	caps, err := New().Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities error: %v", err)
	}
	if caps.Provider != providerName {
		t.Fatalf("Provider = %q, want %q", caps.Provider, providerName)
	}
	if !caps.SupportsStructuredOut {
		t.Fatal("SupportsStructuredOut should be true")
	}
	if !caps.SupportsMCP {
		t.Fatal("SupportsMCP should be true")
	}
}

func TestProviderRunConformance(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id":"resp_123",
			"status":"completed",
			"output_text":"done",
			"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}
		}`))
	})
	provider := New(WithAPIKey("key"), WithBaseURL(server.URL), WithDefaultModel("gpt-5.5"))

	conformance.AssertProviderRun(t, provider, agentruntime.RunRequest{
		RunID:    "run-conformance",
		Role:     agentruntime.RoleScorer,
		Messages: []agentruntime.Message{{Role: "user", Content: "score it"}},
	}, conformance.RunOptions{
		RequireMessage:     true,
		RequireUsage:       true,
		RequireProviderRaw: true,
	})
}

func TestStartRunPostsResponsesRequest(t *testing.T) {
	var captured map[string]any
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want bearer test-key", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id":"resp_123",
			"status":"completed",
			"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"{\"ok\":true}"}]}],
			"usage":{"input_tokens":42,"output_tokens":7,"total_tokens":49,"input_tokens_details":{"cached_tokens":4}}
		}`))
	})
	provider := New(WithAPIKey("test-key"), WithBaseURL(server.URL))

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID:        "run-1",
		Role:         agentruntime.RoleScorer,
		Profile:      "triage-scorer",
		Model:        "gpt-5.5",
		Instructions: "score tickets",
		Messages: []agentruntime.Message{
			{Role: "user", Content: "ticket text"},
		},
		OutputSchema: &agentruntime.OutputSchema{
			Name:   "score",
			Strict: true,
			Schema: json.RawMessage(`{"type":"object","properties":{"ok":{"type":"boolean"}},"required":["ok"],"additionalProperties":false}`),
		},
		Reasoning: &agentruntime.ReasoningOptions{Effort: "low", MaxTokens: 100},
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	assertEventTypes(t, events,
		agentruntime.EventRunStarted,
		agentruntime.EventPhaseStarted,
		agentruntime.EventProviderRaw,
		agentruntime.EventUsageUpdated,
		agentruntime.EventMessageCompleted,
		agentruntime.EventPhaseCompleted,
		agentruntime.EventRunCompleted,
	)

	if events[4].Message != `{"ok":true}` {
		t.Fatalf("message = %q, want structured JSON", events[4].Message)
	}
	if events[3].Usage.InputTokens != 42 || events[3].Usage.CacheReadTokens != 4 {
		t.Fatalf("usage = %#v, want response token details", events[3].Usage)
	}
	if captured["model"] != "gpt-5.5" {
		t.Fatalf("model = %#v, want gpt-5.5", captured["model"])
	}
	if captured["store"] != false {
		t.Fatalf("store = %#v, want false", captured["store"])
	}
	text, ok := captured["text"].(map[string]any)
	if !ok || text["format"] == nil {
		t.Fatalf("text format missing from request: %#v", captured["text"])
	}
	reasoning, ok := captured["reasoning"].(map[string]any)
	if !ok || reasoning["effort"] != "low" {
		t.Fatalf("reasoning = %#v, want low effort", captured["reasoning"])
	}
}

func TestStartRunFailsForLocalTools(t *testing.T) {
	provider := New(WithAPIKey("test-key"), WithDefaultModel("gpt-5.5"))

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID: "run-local-tools",
		Role:  agentruntime.RolePlanner,
		Tools: []agentruntime.ToolRef{{Name: "Read", Kind: "local"}},
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	if events[len(events)-1].Type != agentruntime.EventRunFailed {
		t.Fatalf("last event = %q, want run.failed", events[len(events)-1].Type)
	}
	if !strings.Contains(events[len(events)-1].Message, "cannot run local tool") {
		t.Fatalf("failure message = %q, want local tool guidance", events[len(events)-1].Message)
	}
}

func TestStartRunExecutesLocalToolLoop(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("broker notes"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var requests []map[string]any
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var captured map[string]any
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, captured)

		w.Header().Set("Content-Type", "application/json")
		if len(requests) == 1 {
			w.Write([]byte(`{
				"id":"resp_tools_1",
				"status":"completed",
				"output":[
					{"id":"fc_1","call_id":"call_1","type":"function_call","name":"Read","arguments":"{\"file_path\":\"notes.txt\"}"}
				]
			}`))
			return
		}
		w.Write([]byte(`{
			"id":"resp_tools_2",
			"status":"completed",
			"output_text":"read complete"
		}`))
	})
	provider := New(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
		WithDefaultModel("gpt-5.5"),
		WithToolBroker(toolbroker.NewLocal()),
	)

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID:          "run-tools",
		Role:           agentruntime.RolePlanner,
		WorkDir:        root,
		ApprovalPolicy: agentruntime.ApprovalFullAuto,
		Messages:       []agentruntime.Message{{Role: "user", Content: "read notes"}},
		Tools:          toolbroker.PlannerRefs(),
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	if len(requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(requests))
	}
	if events[len(events)-3].Type != agentruntime.EventMessageCompleted || events[len(events)-3].Message != "read complete" {
		t.Fatalf("events = %#v, want final completed message", events)
	}

	var sawCall, sawResult bool
	for _, event := range events {
		if event.Type == agentruntime.EventToolCall && event.Tool != nil && event.Tool.Name == "Read" {
			sawCall = true
		}
		if event.Type == agentruntime.EventToolResult && event.Tool != nil && strings.Contains(string(event.Tool.Output), "broker notes") {
			sawResult = true
		}
	}
	if !sawCall || !sawResult {
		t.Fatalf("tool events missing call=%v result=%v: %#v", sawCall, sawResult, events)
	}

	input, ok := requests[1]["input"].([]any)
	if !ok {
		t.Fatalf("second input = %#v, want array", requests[1]["input"])
	}
	var sawFunctionOutput bool
	for _, item := range input {
		entry, ok := item.(map[string]any)
		if ok && entry["type"] == "function_call_output" && strings.Contains(fmt.Sprint(entry["output"]), "broker notes") {
			sawFunctionOutput = true
		}
	}
	if !sawFunctionOutput {
		t.Fatalf("second input missing function_call_output with tool result: %#v", input)
	}
}

func TestStartRunFailsWithoutModel(t *testing.T) {
	provider := New(WithAPIKey("test-key"))

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{
		RunID:    "run-missing-model",
		Role:     agentruntime.RoleScorer,
		Messages: []agentruntime.Message{{Role: "user", Content: "score"}},
	})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	events := collect(stream)
	if events[len(events)-1].Type != agentruntime.EventRunFailed {
		t.Fatalf("last event = %q, want run.failed", events[len(events)-1].Type)
	}
	if !strings.Contains(events[len(events)-1].Message, "requires RunRequest.Model") {
		t.Fatalf("failure message = %q, want missing model guidance", events[len(events)-1].Message)
	}
}

func TestOpenAIToolsTranslatesRemoteMCP(t *testing.T) {
	tools, err := openAITools(nil, []agentruntime.MCPServerRef{
		{Label: "linear", URL: "https://mcp.example.com/linear", Description: "Linear tickets"},
	}, agentruntime.ApprovalSuggest, false)
	if err != nil {
		t.Fatalf("openAITools error: %v", err)
	}

	tool := tools[0].(map[string]any)
	if tool["type"] != "mcp" || tool["server_label"] != "linear" || tool["require_approval"] != "always" {
		t.Fatalf("mcp tool = %#v, want remote MCP descriptor", tool)
	}
}

func TestOpenAIToolsTranslatesLocalToolsWithBroker(t *testing.T) {
	tools, err := openAITools(toolbroker.PlannerRefs(), nil, agentruntime.ApprovalFullAuto, true)
	if err != nil {
		t.Fatalf("openAITools error: %v", err)
	}
	tool := tools[0].(map[string]any)
	if tool["type"] != "function" || tool["name"] != "Read" || tool["parameters"] == nil {
		t.Fatalf("local tool = %#v, want function descriptor", tool)
	}
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", r.URL.Path)
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func collect(stream agentruntime.EventStream) []agentruntime.Event {
	var events []agentruntime.Event
	for event := range stream {
		events = append(events, event)
	}
	return events
}

func assertEventTypes(t *testing.T, events []agentruntime.Event, want ...agentruntime.EventType) {
	t.Helper()
	if len(events) != len(want) {
		t.Fatalf("event count = %d, want %d: %#v", len(events), len(want), events)
	}
	for i := range want {
		if events[i].Type != want[i] {
			t.Fatalf("events[%d].Type = %q, want %q", i, events[i].Type, want[i])
		}
	}
}
