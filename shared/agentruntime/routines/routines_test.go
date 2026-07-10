package routines

import (
	"strings"
	"testing"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestDefaultRoutinesValidate(t *testing.T) {
	for _, routine := range DefaultRoutines() {
		t.Run(routine.ID, func(t *testing.T) {
			if err := Validate(routine); err != nil {
				t.Fatalf("Validate error: %v", err)
			}
		})
	}
}

func TestValuesAppliesDefaultsAndValidatesRequired(t *testing.T) {
	routine, ok := DefaultLibrary().Get(DatadogGQLSlowQueriesID)
	if !ok {
		t.Fatal("datadog routine missing")
	}
	values, err := Values(routine, map[string]string{"graph_area": "employer"})
	if err != nil {
		t.Fatalf("Values error: %v", err)
	}
	if values["top_n"] != "20" || values["lookback"] != "24h" || values["environment"] != "prod" {
		t.Fatalf("values = %#v, want defaults", values)
	}
	if _, err := Values(routine, nil); err == nil {
		t.Fatal("Values should require graph_area")
	}
	if _, err := Values(routine, map[string]string{"graph_area": "employer", "top_n": "nope"}); err == nil {
		t.Fatal("Values should validate integer parameters")
	}
	if _, err := Values(routine, map[string]string{"graph_area": "employer", "lookback": "7d"}); err != nil {
		t.Fatalf("Values should accept day lookbacks: %v", err)
	}
}

func TestRenderPrompt(t *testing.T) {
	routine, _ := DefaultLibrary().Get(DatadogGQLSlowQueriesID)
	prompt, err := RenderPrompt(routine, map[string]string{
		"graph_area":  "employer",
		"top_n":       "5",
		"lookback":    "12h",
		"environment": "prod",
		"service":     "employer-graphql",
	})
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}
	for _, want := range []string{"top 5", "employer", "12h", "employer-graphql", "Markdown report"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt should contain %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "{{") {
		t.Fatalf("prompt still contains template marker:\n%s", prompt)
	}
}

func TestBuildRequest(t *testing.T) {
	routine, _ := DefaultLibrary().Get(DatadogGQLSlowQueriesID)
	req, err := BuildRequest(routine, map[string]string{"graph_area": "employer"}, BuildOptions{
		RunID:    "routine-run",
		WorkDir:  "/repo",
		Provider: "claude-cli",
		MCPServers: []agentruntime.MCPServerRef{
			{Label: "datadog", Command: "npx"},
		},
	})
	if err != nil {
		t.Fatalf("BuildRequest error: %v", err)
	}
	if req.RunID != "routine-run" || req.Role != agentruntime.RoleRoutine || req.Profile != "routine.datadog-gql-slow-queries" {
		t.Fatalf("request = %#v, want routine request metadata", req)
	}
	if len(req.MCPServers) != 1 || req.MCPServers[0].Label != "datadog" {
		t.Fatalf("MCPServers = %#v, want datadog", req.MCPServers)
	}
	if req.Metadata["routineId"] != DatadogGQLSlowQueriesID || req.Metadata["phaseId"] != "routine" {
		t.Fatalf("Metadata = %#v, want routine metadata", req.Metadata)
	}
}

func TestNewRunID(t *testing.T) {
	got := NewRunID("datadog", time.Date(2026, 7, 10, 8, 9, 10, 0, time.UTC))
	if got != "datadog-20260710T080910Z" {
		t.Fatalf("NewRunID = %q", got)
	}
}
