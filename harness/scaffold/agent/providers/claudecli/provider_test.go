package claudecli

import (
	"context"
	"errors"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/conformance"
)

func TestProviderRunConformance(t *testing.T) {
	provider := New()
	provider.run = func(context.Context, agentruntime.RunRequest) (string, error) {
		return "enhanced source", nil
	}

	conformance.AssertProviderRun(t, provider, agentruntime.RunRequest{
		RunID:    "scaffold-run",
		Role:     agentruntime.RoleExecutor,
		Profile:  "harness-scaffold-enhancer",
		Provider: providerName,
		Model:    "sonnet",
		WorkDir:  "/repo",
		Messages: []agentruntime.Message{{Role: "user", Content: "enhance"}},
	}, conformance.RunOptions{RequireMessage: true})
}

func TestStartRunFailure(t *testing.T) {
	provider := New()
	provider.run = func(context.Context, agentruntime.RunRequest) (string, error) {
		return "", errors.New("boom")
	}

	stream, err := provider.StartRun(context.Background(), agentruntime.RunRequest{RunID: "run-1"})
	if err != nil {
		t.Fatalf("StartRun error: %v", err)
	}

	var failed agentruntime.Event
	for event := range stream {
		if event.Type == agentruntime.EventRunFailed {
			failed = event
		}
	}
	if failed.Message != "boom" {
		t.Fatalf("run failed message = %q, want boom", failed.Message)
	}
}

func TestBuildArgs(t *testing.T) {
	args := BuildArgs(agentruntime.RunRequest{
		Model:        "sonnet",
		Instructions: "Return code only",
		Messages: []agentruntime.Message{
			{Role: "user", Content: "enhance this"},
		},
	})
	argString := strings.Join(args, "\x00")
	for _, want := range []string{"-p", "--output-format", "text", "--model", "sonnet", "--system-prompt", "Return code only", "enhance this"} {
		if !strings.Contains(argString, want) {
			t.Fatalf("args = %v, missing %q", args, want)
		}
	}
}
