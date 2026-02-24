package cost

import (
	"context"
	"time"

	harnesscost "github.com/philjestin/boatman-ecosystem/harness/cost"
	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// CostHooks creates runner.Hooks that record usage to the governor.
// The harness cost.Tracker continues to work for per-run aggregation;
// the hooks send each step's usage to the governor for org-level tracking.
func CostHooks(governor *Governor, tracker *harnesscost.Tracker, runID string, scope storage.Scope) runner.Hooks {
	ctx := context.Background()

	return runner.Hooks{
		OnStepEnd: func(step string, _ time.Duration, _ error) {
			if tracker == nil {
				return
			}
			usage := tracker.Total()
			governor.RecordStep(ctx, runID, step, usage, scope)
		},
	}
}
