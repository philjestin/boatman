package policy

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// PolicyGuard implements runner.Guard to enforce policy mid-run.
type PolicyGuard struct {
	engine    *Engine
	costStore storage.CostStore
	scope     storage.Scope
	bus       *eventbus.Bus
}

// Compile-time check that PolicyGuard implements runner.Guard.
var _ runner.Guard = (*PolicyGuard)(nil)

// NewPolicyGuard creates a guard backed by a policy engine.
func NewPolicyGuard(engine *Engine, costStore storage.CostStore, scope storage.Scope, bus *eventbus.Bus) *PolicyGuard {
	return &PolicyGuard{
		engine:    engine,
		costStore: costStore,
		scope:     scope,
		bus:       bus,
	}
}

// AllowStep checks the effective policy before each step.
func (g *PolicyGuard) AllowStep(ctx context.Context, step string, state *runner.GuardState) error {
	policy, err := g.engine.GetEffectivePolicy(ctx, g.scope)
	if err != nil {
		return nil // fail open if we can't read policy
	}
	if policy == nil {
		return nil
	}

	// Check cost budget
	if policy.MaxCostPerRun > 0 && state.TotalCostUSD > policy.MaxCostPerRun {
		reason := fmt.Sprintf("cost budget exceeded: $%.4f > $%.4f limit", state.TotalCostUSD, policy.MaxCostPerRun)
		g.publishViolation(ctx, step, reason)
		return fmt.Errorf("policy violation: %s", reason)
	}

	// Check files changed limit
	if policy.MaxFilesChanged > 0 && state.FilesChanged > policy.MaxFilesChanged {
		reason := fmt.Sprintf("files changed limit exceeded: %d > %d", state.FilesChanged, policy.MaxFilesChanged)
		g.publishViolation(ctx, step, reason)
		return fmt.Errorf("policy violation: %s", reason)
	}

	return nil
}

func (g *PolicyGuard) publishViolation(ctx context.Context, step, reason string) {
	if g.bus == nil {
		return
	}
	g.bus.Publish(ctx, &storage.Event{
		ID:      fmt.Sprintf("violation-%s-%s", step, reason),
		Scope:   g.scope,
		Type:    eventbus.SubjectPolicyViolation,
		Name:    "policy_violation",
		Message: reason,
		Data: map[string]any{
			"step":   step,
			"reason": reason,
		},
	})
}
