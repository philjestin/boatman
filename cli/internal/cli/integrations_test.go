package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/integrations"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/runstore"
)

func TestWriteIntegrationsText(t *testing.T) {
	var buf bytes.Buffer
	err := writeIntegrations(&buf, []integrations.Integration{
		{
			Name:        "linear",
			DisplayName: "Linear",
			Description: "Tickets",
			MCP: &integrations.MCPDescriptor{
				Env: map[string]string{"LINEAR_API_KEY": ""},
			},
		},
	}, false)
	if err != nil {
		t.Fatalf("writeIntegrations error: %v", err)
	}
	output := buf.String()
	for _, want := range []string{"NAME", "linear", "LINEAR_API_KEY", "Tickets"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteIntegrationStatusesJSON(t *testing.T) {
	var buf bytes.Buffer
	err := writeIntegrationStatuses(&buf, []integrations.Status{
		{
			Name:       "datadog",
			State:      integrations.StateNeedsConfiguration,
			MissingEnv: []string{"DD_API_KEY"},
		},
	}, true)
	if err != nil {
		t.Fatalf("writeIntegrationStatuses error: %v", err)
	}

	var decoded []integrations.Status
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json output did not decode: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Name != "datadog" || decoded[0].MissingEnv[0] != "DD_API_KEY" {
		t.Fatalf("decoded = %#v, want datadog missing DD_API_KEY", decoded)
	}
}

func TestEnvForIntegrationUsesEnvironment(t *testing.T) {
	t.Setenv("LINEAR_API_KEY", "secret")
	env := envForIntegration(integrations.Integration{
		Name: "linear",
		MCP: &integrations.MCPDescriptor{
			Env: map[string]string{"LINEAR_API_KEY": ""},
		},
	})
	if env["LINEAR_API_KEY"] != "secret" {
		t.Fatalf("env = %#v, want LINEAR_API_KEY from environment", env)
	}
}

func TestIntegrationCheckEvents(t *testing.T) {
	events, err := integrationCheckEvents([]integrations.Status{
		{
			Name:       "linear",
			State:      integrations.StateNeedsConfiguration,
			Message:    "missing",
			MissingEnv: []string{"LINEAR_API_KEY"},
		},
	}, "check-run")
	if err != nil {
		t.Fatalf("integrationCheckEvents error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %#v, want started, integration.state, completed", events)
	}
	if events[1].Type != agentruntime.EventIntegrationState || events[1].RunID != "check-run" || events[1].Provider != "boatman-cli" {
		t.Fatalf("integration event = %#v, want normalized integration.state", events[1])
	}
	if events[2].Type != agentruntime.EventRunCompleted || events[2].Status != agentruntime.StatusWaiting {
		t.Fatalf("terminal event = %#v, want waiting run.completed", events[2])
	}
}

func TestRecordIntegrationEvents(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BOATMAN_RUNTIME_STORE_DIR", dir)
	events, err := integrationCheckEvents([]integrations.Status{
		{Name: "linear", State: integrations.StateReady, Message: "ready"},
	}, "recorded-check")
	if err != nil {
		t.Fatalf("integrationCheckEvents error: %v", err)
	}
	if err := recordIntegrationEvents(context.Background(), events); err != nil {
		t.Fatalf("recordIntegrationEvents error: %v", err)
	}

	store := runstore.NewFileStore(dir)
	metadata, recorded, err := store.LoadRun(context.Background(), "recorded-check")
	if err != nil {
		t.Fatalf("LoadRun error: %v", err)
	}
	if metadata.Provider != "boatman-cli" || metadata.Role != agentruntime.RoleIntegration {
		t.Fatalf("metadata = %#v, want integration metadata", metadata)
	}
	if len(recorded) != 3 || recorded[1].Type != agentruntime.EventIntegrationState {
		t.Fatalf("recorded = %#v, want integration.state event", recorded)
	}
}
