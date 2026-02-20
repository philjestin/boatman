package runner

import (
	"context"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/harness/review"
)

// Hooks allows callers to observe Runner progress without subclassing.
// All fields are optional; nil hooks are silently ignored.
type Hooks struct {
	OnPlanComplete      func(ctx context.Context, plan *Plan, err error)
	OnExecuteComplete   func(ctx context.Context, result *ExecuteResult, err error)
	OnTestComplete      func(ctx context.Context, result *TestResult, iteration int)
	OnReviewComplete    func(ctx context.Context, result *review.ReviewResult, iteration int)
	OnRefactorComplete  func(ctx context.Context, result *RefactorResult, iteration int)
	OnIterationComplete func(ctx context.Context, iteration int, passed bool)
	OnCostUpdate        func(stepName string, usage cost.Usage)
	OnStepStart         func(stepName string)
	OnStepEnd           func(stepName string, duration time.Duration, err error)
}

func (h *Hooks) callOnPlanComplete(ctx context.Context, plan *Plan, err error) {
	if h != nil && h.OnPlanComplete != nil {
		h.OnPlanComplete(ctx, plan, err)
	}
}

func (h *Hooks) callOnExecuteComplete(ctx context.Context, result *ExecuteResult, err error) {
	if h != nil && h.OnExecuteComplete != nil {
		h.OnExecuteComplete(ctx, result, err)
	}
}

func (h *Hooks) callOnTestComplete(ctx context.Context, result *TestResult, iteration int) {
	if h != nil && h.OnTestComplete != nil {
		h.OnTestComplete(ctx, result, iteration)
	}
}

func (h *Hooks) callOnReviewComplete(ctx context.Context, result *review.ReviewResult, iteration int) {
	if h != nil && h.OnReviewComplete != nil {
		h.OnReviewComplete(ctx, result, iteration)
	}
}

func (h *Hooks) callOnRefactorComplete(ctx context.Context, result *RefactorResult, iteration int) {
	if h != nil && h.OnRefactorComplete != nil {
		h.OnRefactorComplete(ctx, result, iteration)
	}
}

func (h *Hooks) callOnIterationComplete(ctx context.Context, iteration int, passed bool) {
	if h != nil && h.OnIterationComplete != nil {
		h.OnIterationComplete(ctx, iteration, passed)
	}
}

func (h *Hooks) callOnCostUpdate(stepName string, usage cost.Usage) {
	if h != nil && h.OnCostUpdate != nil {
		h.OnCostUpdate(stepName, usage)
	}
}

func (h *Hooks) callOnStepStart(name string) {
	if h != nil && h.OnStepStart != nil {
		h.OnStepStart(name)
	}
}

func (h *Hooks) callOnStepEnd(name string, duration time.Duration, err error) {
	if h != nil && h.OnStepEnd != nil {
		h.OnStepEnd(name, duration, err)
	}
}
