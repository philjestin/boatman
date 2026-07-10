package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"boatman/config"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/routines"
)

func TestDesktopRoutineIntegrationRefsUsesPreferencesAndSanitizesSecrets(t *testing.T) {
	t.Setenv("DD_API_KEY", "")
	t.Setenv("DD_APP_KEY", "")
	t.Setenv("DD_SITE", "")

	routine, _ := routines.DefaultLibrary().Get(routines.DatadogGQLSlowQueriesID)
	statuses, refs, secretEnv, err := desktopRoutineIntegrationRefs(context.Background(), routine, config.UserPreferences{
		DatadogAPIKey: "secret-api",
		DatadogAppKey: "secret-app",
		DatadogSite:   "datadoghq.eu",
	}, false)
	if err != nil {
		t.Fatalf("desktopRoutineIntegrationRefs error: %v", err)
	}
	if len(statuses) != 1 || statuses[0].State != integrations.StateConnected {
		t.Fatalf("statuses = %#v, want connected datadog", statuses)
	}
	if len(refs) != 1 || refs[0].Label != "datadog" {
		t.Fatalf("refs = %#v, want datadog ref", refs)
	}
	if refs[0].Env["DD_API_KEY"] != "" || refs[0].Env["DD_APP_KEY"] != "" {
		t.Fatalf("serialized MCP env should not include secrets: %#v", refs[0].Env)
	}
	if refs[0].Env["DD_SITE"] != "datadoghq.eu" {
		t.Fatalf("serialized MCP env DD_SITE = %q, want datadoghq.eu", refs[0].Env["DD_SITE"])
	}
	if secretEnv["DD_API_KEY"] != "secret-api" || secretEnv["DD_APP_KEY"] != "secret-app" {
		t.Fatalf("secretEnv = %#v, want Datadog keys for process env", secretEnv)
	}
}

func TestListProjectRoutinesLoadsProjectDefinitions(t *testing.T) {
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
      "defaults": {
        "graph_area": "employer",
        "service": "employer-graphql"
      }
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(dir, ".boatman", "routines.json"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	app := &App{}
	items, err := app.ListProjectRoutines(dir)
	if err != nil {
		t.Fatalf("ListProjectRoutines error: %v", err)
	}
	var found *DesktopRoutine
	for i := range items {
		if items[i].ID == "daily-employer-gql" {
			found = &items[i]
		}
	}
	if found == nil {
		t.Fatalf("project routine missing from %#v", items)
	}
	if found.Output.DefaultPath != filepath.Join(".boatman", "routines", "daily-employer-gql") {
		t.Fatalf("default output path = %q", found.Output.DefaultPath)
	}
	defaults := map[string]string{}
	for _, param := range found.Parameters {
		defaults[param.Name] = param.Default
	}
	if defaults["graph_area"] != "employer" || defaults["service"] != "employer-graphql" {
		t.Fatalf("parameter defaults = %#v", defaults)
	}
}

func TestDesktopRoutineDryRunAllowsMissingDatadogConfig(t *testing.T) {
	t.Setenv("DD_API_KEY", "")
	t.Setenv("DD_APP_KEY", "")
	t.Setenv("DD_SITE", "")

	routine, _ := routines.DefaultLibrary().Get(routines.DatadogGQLSlowQueriesID)
	statuses, refs, _, err := desktopRoutineIntegrationRefs(context.Background(), routine, config.UserPreferences{}, true)
	if err != nil {
		t.Fatalf("dry run should not fail on missing config: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("refs = %#v, want none when config missing", refs)
	}
	if len(statuses) != 1 || statuses[0].State != integrations.StateNeedsConfiguration {
		t.Fatalf("statuses = %#v, want needs_configuration", statuses)
	}

	_, _, _, err = desktopRoutineIntegrationRefs(context.Background(), routine, config.UserPreferences{}, false)
	if err == nil {
		t.Fatal("non-dry run should fail when Datadog config is missing")
	}
}

func TestRoutineStreamCollectorUsesNormalizedEvents(t *testing.T) {
	collector := routineStreamCollector{}
	collector.Observe(agentruntime.Event{
		Type:    agentruntime.EventMessageCompleted,
		Message: "final report",
	})
	collector.Observe(agentruntime.Event{
		Type: agentruntime.EventUsageUpdated,
		Usage: &agentruntime.Usage{
			InputTokens:     10,
			OutputTokens:    20,
			CacheReadTokens: 3,
			TotalCostUSD:    0.25,
		},
	})

	if got := collector.Report(); got != "final report" {
		t.Fatalf("Report() = %q, want final report", got)
	}
	if collector.usage == nil || collector.usage.InputTokens != 10 || collector.usage.OutputTokens != 20 || collector.usage.CacheReadTokens != 3 || collector.usage.TotalCostUSD != 0.25 {
		t.Fatalf("usage = %#v, want parsed usage", collector.usage)
	}
}

func TestRoutineRequestPreviewSummarizesRequest(t *testing.T) {
	preview := routineRequestPreview(agentruntime.RunRequest{
		RunID:        "run-1",
		Role:         agentruntime.RoleRoutine,
		Profile:      "routine.datadog-gql-slow-queries",
		Provider:     "claude-cli",
		Model:        "sonnet",
		WorkDir:      "/repo",
		Instructions: "Be careful",
		Messages:     []agentruntime.Message{{Role: "user", Content: "Investigate slow queries"}},
		MCPServers: []agentruntime.MCPServerRef{
			{Label: "datadog", Env: map[string]string{"DD_SITE": "datadoghq.com"}},
		},
		ApprovalPolicy: agentruntime.ApprovalSuggest,
		Reasoning:      &agentruntime.ReasoningOptions{Effort: "high"},
		Metadata:       map[string]string{"routineId": "datadog-gql-slow-queries"},
	})

	if preview.RunID != "run-1" || preview.Role != "routine" || preview.ReasoningEffort != "high" {
		t.Fatalf("preview = %#v", preview)
	}
	if len(preview.MCPServerLabels) != 1 || preview.MCPServerLabels[0] != "datadog" {
		t.Fatalf("MCPServerLabels = %#v, want datadog", preview.MCPServerLabels)
	}
}
