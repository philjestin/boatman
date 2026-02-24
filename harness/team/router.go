package team

import (
	"context"
	"strings"
	"sync/atomic"
)

// Router selects which agents should handle a given task.
type Router interface {
	Select(ctx context.Context, task *Task, agents []Agent) ([]Selection, error)
}

// Selection pairs an agent with the task it should handle.
type Selection struct {
	Agent Agent
	Task  *Task // Original or derived sub-task.
}

// AllRouter routes every task to all agents.
type AllRouter struct{}

// Select returns a selection for every agent with the original task.
func (AllRouter) Select(_ context.Context, task *Task, agents []Agent) ([]Selection, error) {
	sels := make([]Selection, len(agents))
	for i, a := range agents {
		sels[i] = Selection{Agent: a, Task: task}
	}
	return sels, nil
}

// FirstMatchRouter routes to the first agent whose matcher returns true.
// Agents without a matcher entry are skipped.
type FirstMatchRouter struct {
	Matchers map[string]func(*Task) bool
}

// Select returns a single-element selection for the first matching agent.
func (r *FirstMatchRouter) Select(_ context.Context, task *Task, agents []Agent) ([]Selection, error) {
	for _, a := range agents {
		if fn, ok := r.Matchers[a.Name]; ok && fn(task) {
			return []Selection{{Agent: a, Task: task}}, nil
		}
	}
	return nil, nil
}

// RoundRobinRouter distributes tasks across agents in round-robin order.
// It is safe for concurrent use.
type RoundRobinRouter struct {
	counter atomic.Uint64
}

// Select returns a single-element selection for the next agent in rotation.
func (r *RoundRobinRouter) Select(_ context.Context, task *Task, agents []Agent) ([]Selection, error) {
	if len(agents) == 0 {
		return nil, nil
	}
	idx := r.counter.Add(1) - 1
	a := agents[idx%uint64(len(agents))]
	return []Selection{{Agent: a, Task: task}}, nil
}

// DescriptionRouter selects agents whose description contains any keyword
// found in the task description. Matching is case-insensitive.
type DescriptionRouter struct{}

// Select returns selections for all agents with keyword overlap.
func (DescriptionRouter) Select(_ context.Context, task *Task, agents []Agent) ([]Selection, error) {
	if task.Description == "" {
		return nil, nil
	}
	taskWords := strings.Fields(strings.ToLower(task.Description))
	var sels []Selection
	for _, a := range agents {
		desc := strings.ToLower(a.Description)
		for _, w := range taskWords {
			if len(w) > 2 && strings.Contains(desc, w) {
				sels = append(sels, Selection{Agent: a, Task: task})
				break
			}
		}
	}
	return sels, nil
}
