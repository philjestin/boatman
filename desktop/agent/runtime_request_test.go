package agent

import (
	"strings"
	"testing"

	runtimeproviders "boatman/agent/providers"
	desktopclaude "boatman/agent/providers/claudecli"
	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestBuildRuntimeRequestStandardAutoEdit(t *testing.T) {
	session := NewSession("session-1", "/repo")
	session.Model = "claude-sonnet"
	session.ReasoningEffort = "high"
	session.systemPrompt = "Be useful"
	session.conversationID = "conversation-1"

	req := session.buildRuntimeRequest("hello", AuthConfig{
		Method:       "anthropic-api",
		ApprovalMode: "auto-edit",
	})

	if req.RunID != "session-1" {
		t.Fatalf("RunID = %q, want session-1", req.RunID)
	}
	if req.Role != agentruntime.RoleChat {
		t.Fatalf("Role = %q, want chat", req.Role)
	}
	if req.Profile != "desktop-chat" {
		t.Fatalf("Profile = %q, want desktop-chat", req.Profile)
	}
	if req.Provider != runtimeproviders.DefaultProvider {
		t.Fatalf("Provider = %q, want %s", req.Provider, runtimeproviders.DefaultProvider)
	}
	if req.Model != "claude-sonnet" {
		t.Fatalf("Model = %q, want claude-sonnet", req.Model)
	}
	if req.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", req.WorkDir)
	}
	if req.Instructions != "Be useful" {
		t.Fatalf("Instructions = %q, want system prompt", req.Instructions)
	}
	if len(req.Messages) != 1 || req.Messages[0].Content != "hello" {
		t.Fatalf("Messages = %#v, want prompt", req.Messages)
	}
	if req.ApprovalPolicy != agentruntime.ApprovalAutoEdit {
		t.Fatalf("ApprovalPolicy = %q, want auto_edit", req.ApprovalPolicy)
	}
	if len(req.Tools) != 2 || req.Tools[0].Name != "Edit" || req.Tools[1].Name != "Write" {
		t.Fatalf("Tools = %#v, want Edit/Write", req.Tools)
	}
	if req.Reasoning == nil || req.Reasoning.Effort != "high" {
		t.Fatalf("Reasoning = %#v, want high effort", req.Reasoning)
	}
	if req.Metadata["conversationId"] != "conversation-1" {
		t.Fatalf("conversationId = %q, want conversation-1", req.Metadata["conversationId"])
	}
	if req.Metadata["outputFormat"] != "stream-json" || req.Metadata["verbose"] != "true" {
		t.Fatalf("metadata = %#v, want stream-json verbose", req.Metadata)
	}

	args := desktopclaude.BuildArgs(req)
	argString := strings.Join(args, "\x00")
	for _, want := range []string{"-p", "hello", "--output-format", "stream-json", "--verbose", "--system-prompt", "Be useful", "-r", "conversation-1", "--model", "claude-sonnet", "--effort", "high", "--allowedTools", "Edit,Write"} {
		if !strings.Contains(argString, want) {
			t.Fatalf("args = %v, missing %q", args, want)
		}
	}
}

func TestBuildRuntimeRequestProviderOverride(t *testing.T) {
	session := NewSession("session-provider", "/repo")
	session.Model = "gpt-5.5"
	session.ModeConfig = map[string]interface{}{
		"provider": "openai-responses",
	}

	req := session.buildRuntimeRequest("hello", AuthConfig{ApprovalMode: "suggest"})

	if req.Provider != "openai-responses" {
		t.Fatalf("Provider = %q, want openai-responses", req.Provider)
	}
	if len(req.Tools) != 0 {
		t.Fatalf("Tools = %#v, want no local edit tools for suggest mode", req.Tools)
	}
}

func TestBuildRuntimeRequestFirefighterFullAuto(t *testing.T) {
	session := NewSession("session-2", "/repo")
	session.Mode = "firefighter"
	session.ModeConfig = map[string]interface{}{
		"scope":      "payments",
		"mcpServers": []interface{}{"datadog", "linear"},
	}
	session.Messages = []Message{{Role: "user", Content: "Investigate"}}

	req := session.buildRuntimeRequest("investigate this alert", AuthConfig{
		Method:       "google-cloud",
		ApprovalMode: "suggest",
	})

	if req.Role != agentruntime.RoleFirefight {
		t.Fatalf("Role = %q, want firefight", req.Role)
	}
	if req.Profile != "desktop-firefighter" {
		t.Fatalf("Profile = %q, want desktop-firefighter", req.Profile)
	}
	if req.ApprovalPolicy != agentruntime.ApprovalFullAuto {
		t.Fatalf("ApprovalPolicy = %q, want full_auto", req.ApprovalPolicy)
	}
	if len(req.Tools) != 0 {
		t.Fatalf("Tools = %#v, want unrestricted tool metadata", req.Tools)
	}
	if len(req.MCPServers) != 2 || req.MCPServers[0].Label != "datadog" || req.MCPServers[1].Label != "linear" {
		t.Fatalf("MCPServers = %#v, want datadog and linear refs", req.MCPServers)
	}
	if len(req.Messages) != 1 || !strings.Contains(req.Messages[0].Content, "payments") || !strings.Contains(req.Messages[0].Content, "investigate this alert") {
		t.Fatalf("firefighter prompt = %#v, want scoped prompt", req.Messages)
	}
	if req.Metadata["authMethod"] != "google-cloud" {
		t.Fatalf("authMethod = %q, want google-cloud", req.Metadata["authMethod"])
	}

	args := desktopclaude.BuildArgs(req)
	argString := strings.Join(args, "\x00")
	if !strings.Contains(argString, "--dangerously-skip-permissions") {
		t.Fatalf("args = %v, want full-auto permission flag", args)
	}
	if strings.Contains(argString, "--allowedTools") {
		t.Fatalf("args = %v, should not include auto-edit tool allowlist", args)
	}
}
