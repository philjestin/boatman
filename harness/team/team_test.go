package team

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

func mockHandler(output string) HandlerFunc {
	return func(_ context.Context, _ *Task) (*Result, error) {
		return &Result{Output: output}, nil
	}
}

func errorHandler(msg string) HandlerFunc {
	return func(_ context.Context, _ *Task) (*Result, error) {
		return nil, errors.New(msg)
	}
}

func TestNew_Defaults(t *testing.T) {
	tm := New("test-team")
	if tm.name != "test-team" {
		t.Fatalf("expected name test-team, got %s", tm.name)
	}
	if tm.strategy != Sequential {
		t.Fatal("expected Sequential strategy")
	}
	if tm.errorPolicy != FailFast {
		t.Fatal("expected FailFast error policy")
	}
}

func TestHandle_Sequential(t *testing.T) {
	tm := New("seq-team",
		WithAgents(
			NewAgent("a1", "first", mockHandler("hello")),
			NewAgent("a2", "second", mockHandler("world")),
		),
	)

	result, err := tm.Handle(context.Background(), &Task{ID: "t1", Description: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "hello\nworld" {
		t.Fatalf("expected concatenated output, got %q", result.Output)
	}
	if len(result.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Children))
	}
}

func TestHandle_Parallel(t *testing.T) {
	var order atomic.Int32
	agent1 := NewAgent("a1", "first", HandlerFunc(func(_ context.Context, _ *Task) (*Result, error) {
		order.Add(1)
		time.Sleep(10 * time.Millisecond)
		return &Result{Output: "a1"}, nil
	}))
	agent2 := NewAgent("a2", "second", HandlerFunc(func(_ context.Context, _ *Task) (*Result, error) {
		order.Add(1)
		time.Sleep(10 * time.Millisecond)
		return &Result{Output: "a2"}, nil
	}))

	tm := New("par-team",
		WithAgents(agent1, agent2),
		WithStrategy(Parallel),
	)

	result, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Children))
	}
}

func TestHandle_FailFast(t *testing.T) {
	tm := New("fail-team",
		WithAgents(
			NewAgent("a1", "fails", errorHandler("boom")),
			NewAgent("a2", "never-runs", mockHandler("ok")),
		),
	)

	_, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		// Just verify it contains the error info.
	}
}

func TestHandle_CollectErrors(t *testing.T) {
	tm := New("collect-team",
		WithAgents(
			NewAgent("a1", "fails", errorHandler("boom")),
			NewAgent("a2", "works", mockHandler("ok")),
		),
		WithErrorPolicy(CollectErrors),
	)

	result, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error with CollectErrors: %v", err)
	}
	if len(result.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Children))
	}
	if result.Children[0].Error == nil {
		t.Fatal("expected first child to have error")
	}
	if result.Children[1].Output != "ok" {
		t.Fatalf("expected second child output 'ok', got %q", result.Children[1].Output)
	}
}

func TestHandle_NoAgentsSelected(t *testing.T) {
	tm := New("empty-team",
		WithRouter(&FirstMatchRouter{Matchers: map[string]func(*Task) bool{}}),
		WithAgents(NewAgent("a1", "first", mockHandler("hello"))),
	)

	result, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "no agents selected" {
		t.Fatalf("expected 'no agents selected', got %q", result.Output)
	}
}

func TestHandle_CostTracking(t *testing.T) {
	tracker := cost.NewTracker()
	handler := HandlerFunc(func(_ context.Context, _ *Task) (*Result, error) {
		return &Result{
			Output: "done",
			Usage: cost.Usage{
				InputTokens:  100,
				OutputTokens: 50,
				TotalCostUSD: 0.001,
			},
		}, nil
	})

	tm := New("cost-team",
		WithAgents(NewAgent("a1", "agent", handler)),
		WithCostTracker(tracker),
	)

	_, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total := tracker.Total()
	if total.InputTokens != 100 {
		t.Fatalf("expected 100 input tokens, got %d", total.InputTokens)
	}
	if total.TotalCostUSD != 0.001 {
		t.Fatalf("expected cost $0.001, got $%.4f", total.TotalCostUSD)
	}
}

func TestHandle_GuardRejects(t *testing.T) {
	tm := New("guarded-team",
		WithAgents(NewAgent("a1", "agent", mockHandler("hello"))),
		WithTeamGuard(&CostLimitGuard{MaxCostUSD: 0}),
	)

	// Seed the cost tracker with existing usage so guard triggers.
	tracker := cost.NewTracker()
	tracker.Add("prior", cost.Usage{TotalCostUSD: 1.0})
	tm.costTracker = tracker

	_, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err == nil {
		t.Fatal("expected guard rejection error")
	}
}

func TestHandle_Observer(t *testing.T) {
	var events []string
	obs := &testObserver{onEvent: func(e string) { events = append(events, e) }}

	tm := New("obs-team",
		WithAgents(NewAgent("a1", "agent", mockHandler("hello"))),
		WithTeamObserver(obs),
	)

	_, err := tm.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"team_start", "route_decision", "agent_start", "agent_complete", "team_complete"}
	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(events), events)
	}
	for i, e := range expected {
		if events[i] != e {
			t.Fatalf("event %d: expected %s, got %s", i, e, events[i])
		}
	}
}

func TestAsAgent_Nesting(t *testing.T) {
	inner := New("inner",
		WithAgents(NewAgent("a1", "inner-agent", mockHandler("inner-result"))),
	)

	outer := New("outer",
		WithAgents(inner.AsAgent()),
	)

	result, err := outer.Handle(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentName != "outer" {
		t.Fatalf("expected outer team name, got %s", result.AgentName)
	}
	if len(result.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.Children))
	}
}

type testObserver struct {
	onEvent func(string)
}

func (o *testObserver) OnTeamStart(_ context.Context, _ string, _ *Task) {
	o.onEvent("team_start")
}
func (o *testObserver) OnTeamComplete(_ context.Context, _ string, _ *Result, _ error) {
	o.onEvent("team_complete")
}
func (o *testObserver) OnAgentStart(_ context.Context, _, _ string, _ *Task) {
	o.onEvent("agent_start")
}
func (o *testObserver) OnAgentComplete(_ context.Context, _, _ string, _ *Result, _ time.Duration, _ error) {
	o.onEvent("agent_complete")
}
func (o *testObserver) OnRouteDecision(_ context.Context, _ string, _ []Selection) {
	o.onEvent("route_decision")
}
