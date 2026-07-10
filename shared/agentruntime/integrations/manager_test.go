package integrations

import (
	"context"
	"testing"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestManagerConnectCachesHandle(t *testing.T) {
	manager := NewManager(DefaultCatalog())
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	manager.Now = func() time.Time { return now }

	conn, err := manager.Connect(context.Background(), "linear", ResolveOptions{
		Enabled: true,
		Env:     map[string]string{"LINEAR_API_KEY": "secret"},
	})
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	if conn.State != StateConnected || conn.Status.State != StateConnected {
		t.Fatalf("conn = %#v, want connected", conn)
	}
	if conn.HandleID == "" || conn.ConnectedAt == nil {
		t.Fatalf("conn = %#v, want handle ID and connected time", conn)
	}
	if conn.MCP == nil || conn.MCP.Label != "linear" || conn.MCP.Env["LINEAR_API_KEY"] != "secret" {
		t.Fatalf("MCP = %#v, want configured linear ref", conn.MCP)
	}

	again, err := manager.Connect(context.Background(), "linear", ResolveOptions{
		Enabled: true,
		Env:     map[string]string{"LINEAR_API_KEY": "secret"},
	})
	if err != nil {
		t.Fatalf("second Connect error: %v", err)
	}
	if again.HandleID != conn.HandleID {
		t.Fatalf("HandleID = %q, want cached %q", again.HandleID, conn.HandleID)
	}
}

func TestManagerConnectNeedsConfiguration(t *testing.T) {
	manager := NewManager(DefaultCatalog())
	conn, err := manager.Connect(context.Background(), "datadog", ResolveOptions{Enabled: true})
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	if conn.State != StateNeedsConfiguration || conn.MCP != nil {
		t.Fatalf("conn = %#v, want needs_configuration without MCP ref", conn)
	}
}

func TestManagerDisconnectAndEvent(t *testing.T) {
	manager := NewManager(DefaultCatalog())
	conn, err := manager.Connect(context.Background(), "linear", ResolveOptions{
		Enabled: true,
		Env:     map[string]string{"LINEAR_API_KEY": "secret"},
	})
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	event := manager.Event(conn)
	if event.Type != agentruntime.EventIntegrationState || event.Status != agentruntime.StatusSucceeded {
		t.Fatalf("event = %#v, want connected integration.state", event)
	}

	disconnected, err := manager.Disconnect(context.Background(), "linear")
	if err != nil {
		t.Fatalf("Disconnect error: %v", err)
	}
	if disconnected.State != StateDisabled || disconnected.MCP != nil {
		t.Fatalf("disconnected = %#v, want disabled without MCP", disconnected)
	}
	if got := manager.List(); len(got) != 1 || got[0].State != StateDisabled {
		t.Fatalf("List = %#v, want disabled connection record", got)
	}
}
