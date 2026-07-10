package cli

import (
	"bytes"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/routines"
)

func TestWriteRoutineListText(t *testing.T) {
	var out bytes.Buffer
	err := writeRoutineList(&out, routines.DefaultLibrary().List(), false)
	if err != nil {
		t.Fatalf("writeRoutineList error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"datadog-gql-slow-queries", "0 8 * * *", "datadog"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteRoutineShowText(t *testing.T) {
	routine, _ := routines.DefaultLibrary().Get(routines.DatadogGQLSlowQueriesID)
	var out bytes.Buffer
	err := writeRoutineShow(&out, routine, false)
	if err != nil {
		t.Fatalf("writeRoutineShow error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"GraphQL Slow Queries", "graph_area", "lookback", "routine.datadog"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteRoutineDryRunText(t *testing.T) {
	routine, _ := routines.DefaultLibrary().Get(routines.DatadogGQLSlowQueriesID)
	req := agentruntime.RunRequest{
		RunID:    "dry",
		Provider: "claude-cli",
		Profile:  routine.Profile,
		WorkDir:  "/repo",
		MCPServers: []agentruntime.MCPServerRef{
			{Label: "datadog", Command: "npx"},
		},
	}
	var out bytes.Buffer
	err := writeRoutineDryRun(&out, routineDryRun{
		Routine: routine,
		Values:  map[string]string{"graph_area": "employer", "top_n": "10"},
		Request: req,
		Integrations: []integrations.Status{
			{Name: "datadog", State: integrations.StateConnected, Message: "ready"},
		},
		ReportPath: "/repo/.boatman/routines/report.md",
	}, false)
	if err != nil {
		t.Fatalf("writeRoutineDryRun error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"Routine:", "MCP refs:  1", "graph_area=employer", "datadog"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestRoutineReportPathDefault(t *testing.T) {
	cmd := routinesRunCmd
	path, err := routineReportPath(cmd, routines.Routine{
		ID: "routine",
		Output: routines.Output{
			DefaultPath: ".boatman/routines/routine",
		},
	}, "run-1", "/repo")
	if err != nil {
		t.Fatalf("routineReportPath error: %v", err)
	}
	if path != "/repo/.boatman/routines/routine/run-1.md" {
		t.Fatalf("path = %q", path)
	}
}

func TestSanitizeMCPEnvRemovesRequiredSecretValues(t *testing.T) {
	item, ok := integrations.DefaultCatalog().Lookup("datadog")
	if !ok {
		t.Fatal("datadog integration missing")
	}
	env := sanitizeMCPEnv(map[string]string{
		"DD_API_KEY": "secret-api",
		"DD_APP_KEY": "secret-app",
		"DD_SITE":    "datadoghq.com",
	}, item)
	if _, ok := env["DD_API_KEY"]; ok {
		t.Fatalf("env should not include DD_API_KEY: %#v", env)
	}
	if _, ok := env["DD_APP_KEY"]; ok {
		t.Fatalf("env should not include DD_APP_KEY: %#v", env)
	}
	if env["DD_SITE"] != "datadoghq.com" {
		t.Fatalf("env = %#v, want DD_SITE retained", env)
	}
}
