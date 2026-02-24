package team

import (
	"context"
	"fmt"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/cost"
)

// TeamGuard is consulted before each agent execution. If AllowAgent returns
// an error, the agent is skipped.
type TeamGuard interface {
	AllowAgent(ctx context.Context, agentName string, task *Task, state *TeamGuardState) error
}

// TeamGuardState provides current team execution metrics.
type TeamGuardState struct {
	TeamName        string
	AgentsCompleted int
	AgentsRemaining int
	TotalUsage      cost.Usage
	ElapsedTime     time.Duration
}

// NopTeamGuard always allows every agent.
type NopTeamGuard struct{}

// AllowAgent always returns nil, allowing the agent to proceed.
func (NopTeamGuard) AllowAgent(_ context.Context, _ string, _ *Task, _ *TeamGuardState) error {
	return nil
}

// CostLimitGuard rejects agent execution when accumulated cost exceeds MaxCostUSD.
type CostLimitGuard struct {
	MaxCostUSD float64
}

// AllowAgent returns an error if the total cost so far exceeds the limit.
func (g *CostLimitGuard) AllowAgent(_ context.Context, agentName string, _ *Task, state *TeamGuardState) error {
	if state.TotalUsage.TotalCostUSD >= g.MaxCostUSD {
		return fmt.Errorf("cost limit exceeded: $%.4f >= $%.4f limit, rejecting agent %s",
			state.TotalUsage.TotalCostUSD, g.MaxCostUSD, agentName)
	}
	return nil
}
