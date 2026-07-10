package integrations

import (
	"context"
	"testing"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestBrokerStatusNeedsConfiguration(t *testing.T) {
	broker := NewBroker(DefaultCatalog())
	broker.Now = func() time.Time { return time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC) }

	status, err := broker.Status(context.Background(), "datadog", ResolveOptions{Enabled: true})
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if status.State != StateNeedsConfiguration {
		t.Fatalf("State = %q, want needs_configuration", status.State)
	}
	if len(status.MissingEnv) != 2 || status.MissingEnv[0] != "DD_API_KEY" || status.MissingEnv[1] != "DD_APP_KEY" {
		t.Fatalf("MissingEnv = %#v, want DD_API_KEY/DD_APP_KEY", status.MissingEnv)
	}
	if status.LastChecked.IsZero() {
		t.Fatal("LastChecked should be set")
	}
}

func TestBrokerResolveMCPReady(t *testing.T) {
	broker := NewBroker(DefaultCatalog())
	ref, status, err := broker.ResolveMCP(context.Background(), "linear", ResolveOptions{
		Enabled: true,
		Env: map[string]string{
			"LINEAR_API_KEY": "secret",
		},
	})
	if err != nil {
		t.Fatalf("ResolveMCP error: %v", err)
	}
	if status.State != StateReady {
		t.Fatalf("State = %q, want ready", status.State)
	}
	if ref.Label != "linear" || ref.Command != "npx" || ref.Env["LINEAR_API_KEY"] != "secret" {
		t.Fatalf("ref = %#v, want configured linear MCP ref", ref)
	}
}

func TestBrokerResolveMCPDisabled(t *testing.T) {
	broker := NewBroker(DefaultCatalog())
	ref, status, err := broker.ResolveMCP(context.Background(), "linear", ResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveMCP error: %v", err)
	}
	if status.State != StateDisabled {
		t.Fatalf("State = %q, want disabled", status.State)
	}
	if ref.Label != "" {
		t.Fatalf("ref = %#v, want empty ref for disabled integration", ref)
	}
}

func TestBrokerResolveRemoteMCP(t *testing.T) {
	broker := NewBroker(DefaultCatalog())
	ref, status, err := broker.ResolveMCP(context.Background(), "slack", ResolveOptions{
		Enabled: true,
		URL:     "https://mcp.example.com/slack",
		Env: map[string]string{
			"SLACK_BOT_TOKEN": "bot",
			"SLACK_TEAM_ID":   "team",
		},
	})
	if err != nil {
		t.Fatalf("ResolveMCP error: %v", err)
	}
	if status.State != StateReady {
		t.Fatalf("State = %q, want ready", status.State)
	}
	if ref.URL != "https://mcp.example.com/slack" || ref.Command != "" || len(ref.Args) != 0 {
		t.Fatalf("ref = %#v, want remote MCP ref", ref)
	}
}

func TestBrokerEvent(t *testing.T) {
	broker := NewBroker(DefaultCatalog())
	status := Status{
		Name:        "linear",
		State:       StateNeedsConfiguration,
		Message:     "missing config",
		MissingEnv:  []string{"LINEAR_API_KEY"},
		LastChecked: time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
	}

	event := broker.Event(status)
	if event.Type != agentruntime.EventIntegrationState || event.Name != "linear" {
		t.Fatalf("event = %#v, want integration.state for linear", event)
	}
	if event.Status != agentruntime.StatusWaiting {
		t.Fatalf("event.Status = %q, want waiting", event.Status)
	}
	if event.Data["state"] != StateNeedsConfiguration {
		t.Fatalf("event.Data = %#v, want state", event.Data)
	}
}
