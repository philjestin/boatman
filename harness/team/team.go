package team

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

// Team orchestrates multiple agents with routing, execution strategy,
// and result aggregation.
type Team struct {
	name        string
	description string
	agents      []Agent
	router      Router
	strategy    Strategy
	aggregator  Aggregator
	observer    TeamObserver
	guard       TeamGuard
	costTracker *cost.Tracker
	errorPolicy ErrorPolicy
}

// New creates a Team with the given name and options.
// Defaults: AllRouter, Sequential, ConcatAggregator, NopTeamObserver, NopTeamGuard, FailFast.
func New(name string, opts ...Option) *Team {
	t := &Team{
		name:        name,
		router:      AllRouter{},
		strategy:    Sequential,
		aggregator:  ConcatAggregator{},
		observer:    NopTeamObserver{},
		guard:       NopTeamGuard{},
		errorPolicy: FailFast,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Handle executes the team orchestration: route → execute → aggregate.
// It implements the Handler interface.
func (t *Team) Handle(ctx context.Context, task *Task) (*Result, error) {
	t.observer.OnTeamStart(ctx, t.name, task)

	selections, err := t.router.Select(ctx, task, t.agents)
	if err != nil {
		t.observer.OnTeamComplete(ctx, t.name, nil, err)
		return nil, fmt.Errorf("team %s: routing failed: %w", t.name, err)
	}

	t.observer.OnRouteDecision(ctx, t.name, selections)

	if len(selections) == 0 {
		result := &Result{AgentName: t.name, Output: "no agents selected"}
		t.observer.OnTeamComplete(ctx, t.name, result, nil)
		return result, nil
	}

	var results []Result
	var execErr error

	switch t.strategy {
	case Sequential:
		results, execErr = t.executeSequential(ctx, selections)
	case Parallel:
		results, execErr = t.executeParallel(ctx, selections)
	}

	if execErr != nil && t.errorPolicy == FailFast {
		t.observer.OnTeamComplete(ctx, t.name, nil, execErr)
		return nil, execErr
	}

	aggregated, aggErr := t.aggregator.Aggregate(ctx, results)
	if aggErr != nil {
		t.observer.OnTeamComplete(ctx, t.name, nil, aggErr)
		return nil, fmt.Errorf("team %s: aggregation failed: %w", t.name, aggErr)
	}
	aggregated.AgentName = t.name

	t.observer.OnTeamComplete(ctx, t.name, aggregated, nil)
	return aggregated, nil
}

// AsAgent returns the team as an Agent, enabling nesting within other teams.
func (t *Team) AsAgent() Agent {
	return Agent{
		Name:        t.name,
		Description: t.description,
		Handler:     t,
	}
}

func (t *Team) executeSequential(ctx context.Context, selections []Selection) ([]Result, error) {
	start := time.Now()
	results := make([]Result, 0, len(selections))
	completed := 0

	for i, sel := range selections {
		state := &TeamGuardState{
			TeamName:        t.name,
			AgentsCompleted: completed,
			AgentsRemaining: len(selections) - i,
			ElapsedTime:     time.Since(start),
		}
		if t.costTracker != nil {
			state.TotalUsage = t.costTracker.Total()
		}

		if err := t.guard.AllowAgent(ctx, sel.Agent.Name, sel.Task, state); err != nil {
			if t.errorPolicy == FailFast {
				return results, fmt.Errorf("team %s: guard rejected agent %s: %w", t.name, sel.Agent.Name, err)
			}
			results = append(results, Result{AgentName: sel.Agent.Name, Error: err})
			continue
		}

		r, err := t.executeAgent(ctx, sel)
		if err != nil {
			if t.errorPolicy == FailFast {
				return results, err
			}
			results = append(results, Result{AgentName: sel.Agent.Name, Error: err})
			continue
		}
		results = append(results, *r)
		completed++
	}

	return results, nil
}

func (t *Team) executeParallel(ctx context.Context, selections []Selection) ([]Result, error) {
	start := time.Now()

	// Pre-flight guard check for all agents.
	for i, sel := range selections {
		state := &TeamGuardState{
			TeamName:        t.name,
			AgentsCompleted: 0,
			AgentsRemaining: len(selections) - i,
			ElapsedTime:     time.Since(start),
		}
		if t.costTracker != nil {
			state.TotalUsage = t.costTracker.Total()
		}
		if err := t.guard.AllowAgent(ctx, sel.Agent.Name, sel.Task, state); err != nil {
			return nil, fmt.Errorf("team %s: guard rejected agent %s: %w", t.name, sel.Agent.Name, err)
		}
	}

	results := make([]Result, len(selections))
	var wg sync.WaitGroup

	for i, sel := range selections {
		wg.Add(1)
		go func(idx int, s Selection) {
			defer wg.Done()
			r, err := t.executeAgent(ctx, s)
			if err != nil {
				results[idx] = Result{AgentName: s.Agent.Name, Error: err}
				return
			}
			results[idx] = *r
		}(i, sel)
	}

	wg.Wait()
	return results, nil
}

func (t *Team) executeAgent(ctx context.Context, sel Selection) (*Result, error) {
	t.observer.OnAgentStart(ctx, t.name, sel.Agent.Name, sel.Task)
	agentStart := time.Now()

	result, err := sel.Agent.Handler.Handle(ctx, sel.Task)
	duration := time.Since(agentStart)

	if err != nil {
		t.observer.OnAgentComplete(ctx, t.name, sel.Agent.Name, nil, duration, err)
		return nil, fmt.Errorf("team %s: agent %s failed: %w", t.name, sel.Agent.Name, err)
	}

	result.AgentName = sel.Agent.Name

	if t.costTracker != nil && !result.Usage.IsEmpty() {
		t.costTracker.Add(fmt.Sprintf("%s/%s", t.name, sel.Agent.Name), result.Usage)
	}

	t.observer.OnAgentComplete(ctx, t.name, sel.Agent.Name, result, duration, nil)
	return result, nil
}
