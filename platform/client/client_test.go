package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philjestin/boatman-ecosystem/platform/client"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

func TestPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	c := client.New(srv.URL, storage.Scope{OrgID: "org1"})
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPingFail(t *testing.T) {
	c := client.New("http://localhost:1", storage.Scope{})
	if err := c.Ping(context.Background()); err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestCreateRun(t *testing.T) {
	var receivedOrg string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedOrg = r.Header.Get("X-Boatman-Org")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "run-1"})
	}))
	defer srv.Close()

	c := client.New(srv.URL, storage.Scope{OrgID: "org1", TeamID: "team1"})
	err := c.CreateRun(context.Background(), &storage.Run{
		ID:     "run-1",
		Status: storage.RunStatusPending,
	})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	if receivedOrg != "org1" {
		t.Errorf("expected org header org1, got %q", receivedOrg)
	}
}

func TestGetEffectivePolicy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(storage.Policy{
			ID:            "pol-1",
			MaxIterations: 3,
			RequireTests:  true,
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL, storage.Scope{OrgID: "org1"})
	policy, err := c.GetEffectivePolicy(context.Background())
	if err != nil {
		t.Fatalf("GetEffectivePolicy: %v", err)
	}

	if policy.MaxIterations != 3 {
		t.Errorf("expected MaxIterations=3, got %d", policy.MaxIterations)
	}
	if !policy.RequireTests {
		t.Error("expected RequireTests=true")
	}
}
