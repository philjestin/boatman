package cli

import (
	"bytes"
	"strings"
	"testing"

	runtimeworkflows "github.com/philjestin/boatman-ecosystem/shared/agentruntime/workflows"
)

func TestWriteWorkflowListText(t *testing.T) {
	var out bytes.Buffer
	err := writeWorkflowList(&out, runtimeworkflows.DefaultLibrary().List(), false)
	if err != nil {
		t.Fatalf("writeWorkflowList error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"ID", "feature", "firefighter", "human:"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteWorkflowShowJSON(t *testing.T) {
	template, ok := runtimeworkflows.DefaultLibrary().Get("feature")
	if !ok {
		t.Fatal("feature template missing")
	}
	var out bytes.Buffer
	err := writeWorkflowShow(&out, template, true)
	if err != nil {
		t.Fatalf("writeWorkflowShow error: %v", err)
	}
	output := out.String()
	for _, want := range []string{`"id": "feature"`, `"kind": "planning"`, `"gate": "human"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWorkflowGateSummary(t *testing.T) {
	template, ok := runtimeworkflows.DefaultLibrary().Get("research")
	if !ok {
		t.Fatal("research template missing")
	}
	if got := workflowGateSummary(template); got != "none" {
		t.Fatalf("workflowGateSummary = %q, want none", got)
	}
}
