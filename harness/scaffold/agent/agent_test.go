package agent

import (
	"strings"
	"testing"

	"github.com/philjestin/boatman-ecosystem/harness/scaffold"
	scaffoldclaude "github.com/philjestin/boatman-ecosystem/harness/scaffold/agent/providers/claudecli"
	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestBuildEnhanceRunRequest(t *testing.T) {
	cfg := EnhanceConfig{
		ProjectDir:    "/repo",
		Provider:      scaffold.ProviderOpenAI,
		ProjectLang:   scaffold.LangGo,
		Model:         "claude-sonnet",
		ModelProvider: "claude-cli",
	}

	req := buildEnhanceRunRequest(cfg, "executor.go", "replace this stub")

	if req.RunID != "scaffold-enhance-executor" {
		t.Fatalf("RunID = %q, want scaffold-enhance-executor", req.RunID)
	}
	if req.Role != agentruntime.RoleExecutor {
		t.Fatalf("Role = %q, want executor", req.Role)
	}
	if req.Profile != "harness-scaffold-enhancer" {
		t.Fatalf("Profile = %q, want harness-scaffold-enhancer", req.Profile)
	}
	if req.Provider != "claude-cli" {
		t.Fatalf("Provider = %q, want claude-cli", req.Provider)
	}
	if req.Model != "claude-sonnet" {
		t.Fatalf("Model = %q, want claude-sonnet", req.Model)
	}
	if req.WorkDir != "/repo" {
		t.Fatalf("WorkDir = %q, want /repo", req.WorkDir)
	}
	if len(req.Messages) != 1 || req.Messages[0].Content != "replace this stub" {
		t.Fatalf("Messages = %#v, want stub prompt", req.Messages)
	}
	if req.ApprovalPolicy != agentruntime.ApprovalSuggest {
		t.Fatalf("ApprovalPolicy = %q, want suggest", req.ApprovalPolicy)
	}
	if req.OutputSchema == nil || req.OutputSchema.Name != "enhanced_source_file" {
		t.Fatalf("OutputSchema = %#v, want enhanced file schema", req.OutputSchema)
	}
	if req.Metadata["targetProvider"] != string(scaffold.ProviderOpenAI) {
		t.Fatalf("targetProvider = %q, want openai", req.Metadata["targetProvider"])
	}
	if req.Metadata["projectLang"] != string(scaffold.LangGo) {
		t.Fatalf("projectLang = %q, want go", req.Metadata["projectLang"])
	}
	if req.Metadata["filename"] != "executor.go" {
		t.Fatalf("filename = %q, want executor.go", req.Metadata["filename"])
	}
}

func TestClaudeTextArgsFromRuntimeRequest(t *testing.T) {
	req := agentruntime.RunRequest{
		Model:        "claude-sonnet",
		Instructions: "Return only code",
		Messages: []agentruntime.Message{
			{Role: "user", Content: "source prompt"},
		},
	}

	args := scaffoldclaude.BuildArgs(req)
	argString := strings.Join(args, "\x00")
	for _, want := range []string{"-p", "--output-format", "text", "--model", "claude-sonnet", "--system-prompt", "Return only code", "source prompt"} {
		if !strings.Contains(argString, want) {
			t.Fatalf("args = %v, missing %q", args, want)
		}
	}
}

func TestRunClaudeRejectsUnsupportedProvider(t *testing.T) {
	_, err := runClaude(nil, agentruntime.RunRequest{Provider: "openai-responses"})
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("err = %v, want unsupported provider error", err)
	}
}
