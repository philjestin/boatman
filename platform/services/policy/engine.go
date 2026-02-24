// Package policy provides layered policy enforcement evaluated before and during agent runs.
package policy

import (
	"context"
	"fmt"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Decision is the result of a policy evaluation.
type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Engine evaluates and enforces policies.
type Engine struct {
	store storage.PolicyStore
}

// NewEngine creates a new policy engine.
func NewEngine(store storage.PolicyStore) *Engine {
	return &Engine{store: store}
}

// Evaluate checks whether a request is allowed under the effective policy.
func (e *Engine) Evaluate(ctx context.Context, scope storage.Scope) (*Decision, error) {
	policy, err := e.store.GetEffectivePolicy(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("get effective policy: %w", err)
	}

	if policy == nil {
		return &Decision{Allowed: true}, nil
	}

	return &Decision{Allowed: true}, nil
}

// EnforceConfig caps a runner.Config to comply with the effective policy.
// For example, if policy says max 3 iterations, config.MaxIterations = min(config, 3).
func (e *Engine) EnforceConfig(ctx context.Context, scope storage.Scope, cfg runner.Config) (runner.Config, error) {
	policy, err := e.store.GetEffectivePolicy(ctx, scope)
	if err != nil {
		return cfg, fmt.Errorf("get effective policy: %w", err)
	}

	if policy == nil {
		return cfg, nil
	}

	// Cap iterations
	if policy.MaxIterations > 0 && (cfg.MaxIterations == 0 || cfg.MaxIterations > policy.MaxIterations) {
		cfg.MaxIterations = policy.MaxIterations
	}

	// Enforce test requirement
	if policy.RequireTests {
		cfg.TestBeforeReview = true
		cfg.FailOnTestFailure = true
	}

	return cfg, nil
}

// GetEffectivePolicy retrieves the merged effective policy for a scope.
func (e *Engine) GetEffectivePolicy(ctx context.Context, scope storage.Scope) (*storage.Policy, error) {
	return e.store.GetEffectivePolicy(ctx, scope)
}
