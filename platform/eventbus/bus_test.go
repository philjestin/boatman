package eventbus_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
	"github.com/philjestin/boatman-ecosystem/platform/storage/sqlite"
)

func newTestBus(t *testing.T) (*eventbus.Bus, storage.EventStore) {
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

	return bus, store.Events()
}

func TestBusStartStop(t *testing.T) {
	bus, _ := newTestBus(t)
	if bus.ClientURL() == "" {
		t.Error("expected non-empty client URL")
	}
}

func TestPubSubRoundtrip(t *testing.T) {
	bus, _ := newTestBus(t)
	ctx := context.Background()

	// Subscribe
	subject := eventbus.BuildSubject("org1", "team1", "run.started")
	ch, cancel, err := bus.Subscribe(ctx, subject)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer cancel()

	// Publish
	event := &storage.Event{
		ID:    "evt-1",
		RunID: "run-1",
		Scope: storage.Scope{OrgID: "org1", TeamID: "team1"},
		Type:  "run.started",
		Name:  "test run",
	}
	if err := bus.Publish(ctx, event); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Receive with timeout
	select {
	case got := <-ch:
		if got.ID != "evt-1" {
			t.Errorf("expected event ID evt-1, got %s", got.ID)
		}
		if got.Type != "run.started" {
			t.Errorf("expected type run.started, got %s", got.Type)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestWildcardSubscription(t *testing.T) {
	bus, _ := newTestBus(t)
	ctx := context.Background()

	// Subscribe to all org events
	ch, cancel, err := bus.Subscribe(ctx, eventbus.OrgWildcard("org1"))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer cancel()

	// Publish to different teams
	for i, team := range []string{"team1", "team2"} {
		event := &storage.Event{
			ID:    fmt.Sprintf("evt-%d", i),
			Scope: storage.Scope{OrgID: "org1", TeamID: team},
			Type:  "run.started",
		}
		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}

	// Should receive both
	received := 0
	timeout := time.After(5 * time.Second)
	for received < 2 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout: received only %d/2 events", received)
		}
	}
}

func TestReplay(t *testing.T) {
	bus, _ := newTestBus(t)
	ctx := context.Background()
	scope := storage.Scope{OrgID: "org1", TeamID: "team1"}

	// Publish 10 events
	for i := range 10 {
		event := &storage.Event{
			ID:    fmt.Sprintf("replay-evt-%d", i),
			RunID: "run-1",
			Scope: scope,
			Type:  "step.complete",
			Name:  fmt.Sprintf("step_%d", i),
		}
		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Publish %d: %v", i, err)
		}
	}

	// Replay all events for the run
	replayCh, err := bus.Replay(ctx, storage.EventFilter{
		RunID: "run-1",
	})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	var replayed []*storage.Event
	for e := range replayCh {
		replayed = append(replayed, e)
	}

	if len(replayed) != 10 {
		t.Errorf("expected 10 replayed events, got %d", len(replayed))
	}

	// Verify ordering (should be ascending by created_at)
	for i := 1; i < len(replayed); i++ {
		if replayed[i].CreatedAt.Before(replayed[i-1].CreatedAt) {
			t.Errorf("events not in order at index %d", i)
		}
	}
}
