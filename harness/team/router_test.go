package team

import (
	"context"
	"testing"
)

func TestAllRouter(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "first", mockHandler("1")),
		NewAgent("a2", "second", mockHandler("2")),
		NewAgent("a3", "third", mockHandler("3")),
	}

	sels, err := AllRouter{}.Select(context.Background(), &Task{ID: "t1"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 3 {
		t.Fatalf("expected 3 selections, got %d", len(sels))
	}
	for i, s := range sels {
		if s.Agent.Name != agents[i].Name {
			t.Fatalf("selection %d: expected agent %s, got %s", i, agents[i].Name, s.Agent.Name)
		}
	}
}

func TestFirstMatchRouter(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "first", mockHandler("1")),
		NewAgent("a2", "second", mockHandler("2")),
	}

	r := &FirstMatchRouter{
		Matchers: map[string]func(*Task) bool{
			"a1": func(task *Task) bool { return task.Description == "match-a1" },
			"a2": func(*Task) bool { return true },
		},
	}

	// Should match a2 (a1 doesn't match).
	sels, err := r.Select(context.Background(), &Task{Description: "something"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 1 || sels[0].Agent.Name != "a2" {
		t.Fatalf("expected a2 selected, got %v", sels)
	}

	// Should match a1.
	sels, err = r.Select(context.Background(), &Task{Description: "match-a1"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 1 || sels[0].Agent.Name != "a1" {
		t.Fatalf("expected a1 selected, got %v", sels)
	}
}

func TestFirstMatchRouter_NoMatch(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "first", mockHandler("1")),
	}

	r := &FirstMatchRouter{
		Matchers: map[string]func(*Task) bool{
			"a1": func(*Task) bool { return false },
		},
	}

	sels, err := r.Select(context.Background(), &Task{ID: "t1"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 0 {
		t.Fatalf("expected no selections, got %d", len(sels))
	}
}

func TestRoundRobinRouter(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "first", mockHandler("1")),
		NewAgent("a2", "second", mockHandler("2")),
		NewAgent("a3", "third", mockHandler("3")),
	}

	r := &RoundRobinRouter{}
	task := &Task{ID: "t1"}

	expected := []string{"a1", "a2", "a3", "a1", "a2"}
	for i, exp := range expected {
		sels, err := r.Select(context.Background(), task, agents)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if len(sels) != 1 || sels[0].Agent.Name != exp {
			t.Fatalf("call %d: expected %s, got %v", i, exp, sels)
		}
	}
}

func TestRoundRobinRouter_Empty(t *testing.T) {
	r := &RoundRobinRouter{}
	sels, err := r.Select(context.Background(), &Task{ID: "t1"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 0 {
		t.Fatalf("expected no selections, got %d", len(sels))
	}
}

func TestDescriptionRouter(t *testing.T) {
	agents := []Agent{
		NewAgent("frontend", "handles CSS and React components", mockHandler("1")),
		NewAgent("backend", "handles API and database queries", mockHandler("2")),
		NewAgent("infra", "handles deployment and kubernetes", mockHandler("3")),
	}

	r := DescriptionRouter{}

	// Should match frontend and backend (CSS matches frontend, API matches backend).
	sels, err := r.Select(context.Background(), &Task{Description: "fix CSS API"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 2 {
		t.Fatalf("expected 2 selections, got %d", len(sels))
	}
	if sels[0].Agent.Name != "frontend" || sels[1].Agent.Name != "backend" {
		t.Fatalf("expected frontend and backend, got %v and %v", sels[0].Agent.Name, sels[1].Agent.Name)
	}
}

func TestDescriptionRouter_NoMatch(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "handles frontend", mockHandler("1")),
	}

	r := DescriptionRouter{}
	sels, err := r.Select(context.Background(), &Task{Description: "xy"}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 0 {
		t.Fatalf("expected no selections (short words filtered), got %d", len(sels))
	}
}

func TestDescriptionRouter_EmptyDescription(t *testing.T) {
	agents := []Agent{
		NewAgent("a1", "handles frontend", mockHandler("1")),
	}

	r := DescriptionRouter{}
	sels, err := r.Select(context.Background(), &Task{Description: ""}, agents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sels) != 0 {
		t.Fatalf("expected no selections, got %d", len(sels))
	}
}
