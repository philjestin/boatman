package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/server"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func TestServerLifecycle(t *testing.T) {
	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	bus, err := eventbus.New(eventbus.WithEventStore(store.Events()))
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()

	srv := server.New(server.Config{
		Port:  0, // random port
		Store: store,
		Bus:   bus,
	})

	// Start in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)

	// Check health endpoint
	resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/health", srv.Addr()))
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var health map[string]string
	json.NewDecoder(resp.Body).Decode(&health)
	if health["status"] != "ok" {
		t.Errorf("expected status ok, got %q", health["status"])
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("stop: %v", err)
	}
}
