package routines

import (
	"os"
	"path/filepath"
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
	if values["top_n"] != "20" || values["lookback"] != "24h" || values["environment"] != "prod" || values["max_remediations"] != "3" {
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
		"graph_area":       "employer",
		"top_n":            "5",
		"lookback":         "12h",
		"environment":      "prod",
		"service":          "employer-graphql",
		"max_remediations": "2",
	})
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}
	for _, want := range []string{"top 5", "employer", "12h", "employer-graphql", "Markdown report", "at most 2", "boatman_remediation_candidates", "should_remediate", "prompt"} {
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

func TestProjectLibraryLoadsExtendingRoutineDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".boatman"), 0755); err != nil {
		t.Fatal(err)
	}
	body := `{
  "routines": [
    {
      "id": "daily-employer-gql",
      "extends": "datadog-gql-slow-queries",
      "name": "Daily Employer GraphQL",
      "schedule": "0 8 * * *",
      "defaults": {
        "graph_area": "employer",
        "service": "employer-graphql",
        "top_n": "10"
      },
      "models": {
        "plan": "opus",
        "implementation": "sonnet",
        "skills": "haiku"
      }
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(dir, ".boatman", "routines.json"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	library, err := ProjectLibrary(dir)
	if err != nil {
		t.Fatalf("ProjectLibrary error: %v", err)
	}
	routine, ok := library.Get("daily-employer-gql")
	if !ok {
		t.Fatal("project routine missing")
	}
	if routine.Profile != "routine.daily-employer-gql" || routine.Output.DefaultPath != filepath.Join(".boatman", "routines", "daily-employer-gql") {
		t.Fatalf("routine defaults not normalized: %#v", routine)
	}
	if routine.Models.Plan != "opus" || routine.Models.Implementation != "sonnet" || routine.Models.Skills != "haiku" {
		t.Fatalf("routine models = %#v, want project overrides", routine.Models)
	}
	values, err := Values(routine, nil)
	if err != nil {
		t.Fatalf("Values error: %v", err)
	}
	if values["graph_area"] != "employer" || values["service"] != "employer-graphql" || values["top_n"] != "10" {
		t.Fatalf("values = %#v, want project defaults", values)
	}
	prompt, err := RenderPrompt(routine, nil)
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}
	if !strings.Contains(prompt, "employer-graphql") {
		t.Fatalf("prompt should contain service default:\n%s", prompt)
	}
}

func TestProjectLibraryLoadsSplitRoutineFile(t *testing.T) {
	dir := t.TempDir()
	routineDir := filepath.Join(dir, ".boatman", "routines")
	if err := os.MkdirAll(routineDir, 0755); err != nil {
		t.Fatal(err)
	}
	body := `{
  "id": "custom-check",
  "name": "Custom Check",
  "parameters": [
    {"name": "scope", "type": "string", "required": true}
  ],
  "instructions": "Read only.",
  "promptTemplate": "Inspect {{scope}}."
}`
	if err := os.WriteFile(filepath.Join(routineDir, "custom-check.json"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	library, err := ProjectLibrary(dir)
	if err != nil {
		t.Fatalf("ProjectLibrary error: %v", err)
	}
	routine, ok := library.Get("custom-check")
	if !ok {
		t.Fatal("custom routine missing")
	}
	if routine.Role != agentruntime.RoleRoutine || routine.Output.Format != "markdown" {
		t.Fatalf("routine defaults not applied: %#v", routine)
	}
	prompt, err := RenderPrompt(routine, map[string]string{"scope": "deploys"})
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}
	if prompt != "Inspect deploys." {
		t.Fatalf("prompt = %q", prompt)
	}
}

func TestProjectLibraryValidatesProjectDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".boatman"), 0755); err != nil {
		t.Fatal(err)
	}
	body := `{
  "id": "bad-default",
  "extends": "datadog-gql-slow-queries",
  "defaults": {
    "top_n": "many"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, ".boatman", "routines.json"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := ProjectLibrary(dir); err == nil {
		t.Fatal("ProjectLibrary should reject invalid project defaults")
	}
}
