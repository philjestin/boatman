// Package cost provides budget tracking, limit enforcement, and alerts.
package cost

import (
	"context"
	"fmt"
	"time"

	harnesscost "github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// BudgetStatus reports current spend vs limits.
type BudgetStatus struct {
	Scope        storage.Scope `json:"scope"`
	Budget       *storage.Budget `json:"budget,omitempty"`
	DailySpend   float64       `json:"daily_spend"`
	MonthlySpend float64       `json:"monthly_spend"`
	AtLimit      bool          `json:"at_limit"`
	AlertTriggered bool        `json:"alert_triggered"`
}

// Governor manages cost tracking, budget enforcement, and alerts.
type Governor struct {
	costStore storage.CostStore
	bus       *eventbus.Bus
}

// NewGovernor creates a new cost governor.
func NewGovernor(costStore storage.CostStore, bus *eventbus.Bus) *Governor {
	return &Governor{
		costStore: costStore,
		bus:       bus,
	}
}

// RecordStep records token usage from a harness/cost.Usage and checks the budget.
func (g *Governor) RecordStep(ctx context.Context, runID, step string, usage harnesscost.Usage, scope storage.Scope) error {
	record := &storage.UsageRecord{
		ID:               fmt.Sprintf("%s-%s-%d", runID, step, time.Now().UnixNano()),
		RunID:            runID,
		Scope:            scope,
		Step:             step,
		InputTokens:      usage.InputTokens,
		OutputTokens:     usage.OutputTokens,
		CacheReadTokens:  usage.CacheReadTokens,
		CacheWriteTokens: usage.CacheWriteTokens,
		TotalCostUSD:     usage.TotalCostUSD,
	}

	if err := g.costStore.RecordUsage(ctx, record); err != nil {
		return fmt.Errorf("record usage: %w", err)
	}

	// Publish cost event
	if g.bus != nil {
		g.bus.Publish(ctx, &storage.Event{
			ID:    fmt.Sprintf("cost-%s-%s-%d", runID, step, time.Now().UnixNano()),
			RunID: runID,
			Scope: scope,
			Type:  eventbus.SubjectCostRecorded,
			Data: map[string]any{
				"step":      step,
				"cost_usd":  usage.TotalCostUSD,
			},
		})
	}

	// Check budget
	status, err := g.CheckBudget(ctx, scope)
	if err != nil {
		return nil // don't fail on budget check errors
	}

	if status.AlertTriggered && g.bus != nil {
		g.bus.Publish(ctx, &storage.Event{
			ID:    fmt.Sprintf("budget-alert-%d", time.Now().UnixNano()),
			Scope: scope,
			Type:  eventbus.SubjectBudgetAlert,
			Message: fmt.Sprintf("Budget alert: daily spend $%.4f, monthly spend $%.4f", status.DailySpend, status.MonthlySpend),
			Data: map[string]any{
				"daily_spend":   status.DailySpend,
				"monthly_spend": status.MonthlySpend,
			},
		})
	}

	return nil
}

// CheckBudget returns current spend vs limits and emits an alert if threshold is crossed.
func (g *Governor) CheckBudget(ctx context.Context, scope storage.Scope) (*BudgetStatus, error) {
	budget, err := g.costStore.GetBudget(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	status := &BudgetStatus{
		Scope:  scope,
		Budget: budget,
	}

	if budget == nil {
		return status, nil
	}

	now := time.Now().UTC()

	// Daily spend
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dailyRecords, err := g.costStore.GetUsage(ctx, storage.UsageFilter{
		Scope: &scope,
		Since: dayStart,
		Until: now,
	})
	if err == nil {
		for _, r := range dailyRecords {
			status.DailySpend += r.TotalCostUSD
		}
	}

	// Monthly spend
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthlyRecords, err := g.costStore.GetUsage(ctx, storage.UsageFilter{
		Scope: &scope,
		Since: monthStart,
		Until: now,
	})
	if err == nil {
		for _, r := range monthlyRecords {
			status.MonthlySpend += r.TotalCostUSD
		}
	}

	// Check limits
	if budget.DailyLimit > 0 && status.DailySpend >= budget.DailyLimit {
		status.AtLimit = true
	}
	if budget.MonthlyLimit > 0 && status.MonthlySpend >= budget.MonthlyLimit {
		status.AtLimit = true
	}

	// Check alert threshold
	if budget.AlertAt > 0 {
		if budget.DailyLimit > 0 && status.DailySpend >= budget.DailyLimit*budget.AlertAt {
			status.AlertTriggered = true
		}
		if budget.MonthlyLimit > 0 && status.MonthlySpend >= budget.MonthlyLimit*budget.AlertAt {
			status.AlertTriggered = true
		}
	}

	return status, nil
}
