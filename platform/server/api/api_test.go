package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/server/api"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func setupTest(t *testing.T) (*http.ServeMux, storage.Store) {
	t.Helper()

	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	bus, err := eventbus.New(eventbus.WithEventStore(store.Events()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		bus.Close()
		store.Close()
	})

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, store, bus)
	return mux, store
}

func TestHealthEndpoint(t *testing.T) {
	mux, _ := setupTest(t)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

func TestRunsCRUD(t *testing.T) {
	mux, _ := setupTest(t)

	// Create run
	run := storage.Run{
		ID:     "test-run-1",
		UserID: "user1",
		Status: storage.RunStatusPending,
		Prompt: "add hello endpoint",
	}
	body, _ := json.Marshal(run)
	req := httptest.NewRequest("POST", "/api/v1/runs", bytes.NewReader(body))
	req.Header.Set("X-Boatman-Org", "org1")
	req.Header.Set("X-Boatman-Team", "team1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Get run
	req = httptest.NewRequest("GET", "/api/v1/runs/test-run-1", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var gotRun storage.Run
	json.NewDecoder(w.Body).Decode(&gotRun)
	if gotRun.ID != "test-run-1" {
		t.Errorf("expected run ID test-run-1, got %q", gotRun.ID)
	}

	// List runs
	req = httptest.NewRequest("GET", "/api/v1/runs", nil)
	req.Header.Set("X-Boatman-Org", "org1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("list: expected 200, got %d", w.Code)
	}
}

func TestPoliciesAPI(t *testing.T) {
	mux, _ := setupTest(t)

	// Set policy
	policy := storage.Policy{
		ID:            "pol-1",
		MaxIterations: 5,
		RequireTests:  true,
	}
	body, _ := json.Marshal(policy)
	req := httptest.NewRequest("PUT", "/api/v1/policies", bytes.NewReader(body))
	req.Header.Set("X-Boatman-Org", "org1")
	req.Header.Set("X-Boatman-Team", "team1")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("set policy: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Get policy
	req = httptest.NewRequest("GET", "/api/v1/policies", nil)
	req.Header.Set("X-Boatman-Org", "org1")
	req.Header.Set("X-Boatman-Team", "team1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get policy: expected 200, got %d", w.Code)
	}

	// Get effective policy
	req = httptest.NewRequest("GET", "/api/v1/policies/effective", nil)
	req.Header.Set("X-Boatman-Org", "org1")
	req.Header.Set("X-Boatman-Team", "team1")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get effective: expected 200, got %d", w.Code)
	}
}
