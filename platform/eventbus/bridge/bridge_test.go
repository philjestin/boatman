package bridge_test

import (
	"context"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus/bridge"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func newTestBus(t *testing.T) *eventbus.Bus {
	t.Helper()
	store, err := sqlite.New(sqlite.WithInMemory())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	bus, err := eventbus.New(
		eventbus.WithEventStore(store.Events()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		bus.Close()
		store.Close()
	})

	return bus
}

func TestHooksAdapterPublishesEvents(t *testing.T) {
	bus := newTestBus(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	// Subscribe to all org events
	ch, cancel, err := bus.Subscribe(ctx, eventbus.OrgWildcard("org1"))
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	hooks := bridge.HooksAdapter(bus, "run-1", scope)

	// Trigger hooks
	hooks.OnStepStart("execute")
	hooks.OnStepEnd("execute", 5*time.Second, nil)

	// Verify events received
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 2 {
		select {
		case evt := <-ch:
			if evt.RunID != "run-1" {
				t.Errorf("expected runID run-1, got %s", evt.RunID)
			}
			received++
		case <-timeout:
			t.Fatalf("timeout: received only %d/2 events", received)
		}
	}
}

func TestObserverAdapterPublishesEvents(t *testing.T) {
	bus := newTestBus(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	// Subscribe
	ch, cancel, err := bus.Subscribe(ctx, eventbus.OrgWildcard("org1"))
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	obs := bridge.ObserverAdapter(bus, "run-2", scope)

	// Trigger observer events
	obs.OnRunStart(ctx, &runner.Request{Description: "test"})
	obs.OnStepStart(ctx, "plan")
	obs.OnStepComplete(ctx, "plan", 2*time.Second, nil)
	obs.OnRunComplete(ctx, &runner.Result{})

	// Should receive 4 events
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 4 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout: received only %d/4 events", received)
		}
	}
}
