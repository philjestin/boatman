// Package bridge provides adapters between the event bus and the harness/runner hooks system.
package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/philjestin/boatman-ecosystem/harness/runner"
	"github.com/philjestin/boatman-ecosystem/platform/eventbus"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// HooksAdapter creates runner.Hooks that publish events to the Bus.
func HooksAdapter(bus *eventbus.Bus, runID string, scope storage.Scope) runner.Hooks {
	ctx := context.Background()

	return runner.Hooks{
		OnStepStart: func(stepName string) {
			bus.Publish(ctx, &storage.Event{
				ID:    fmt.Sprintf("%s-%s-start-%d", runID, stepName, time.Now().UnixNano()),
				RunID: runID,
				Scope: scope,
				Type:  eventbus.SubjectStepPrefix + stepName,
				Name:  stepName,
				Data:  map[string]any{"phase": "start"},
			})
		},
		OnStepEnd: func(stepName string, duration time.Duration, err error) {
			data := map[string]any{
				"phase":    "end",
				"duration": duration.String(),
			}
			if err != nil {
				data["error"] = err.Error()
			}
			bus.Publish(ctx, &storage.Event{
				ID:    fmt.Sprintf("%s-%s-end-%d", runID, stepName, time.Now().UnixNano()),
				RunID: runID,
				Scope: scope,
				Type:  eventbus.SubjectStepPrefix + stepName,
				Name:  stepName,
				Data:  data,
			})
		},
		OnIterationComplete: func(_ context.Context, iteration int, passed bool) {
			bus.Publish(ctx, &storage.Event{
				ID:    fmt.Sprintf("%s-iteration-%d-%d", runID, iteration, time.Now().UnixNano()),
				RunID: runID,
				Scope: scope,
				Type:  "iteration.complete",
				Data: map[string]any{
					"iteration": iteration,
					"passed":    passed,
				},
			})
		},
	}
}

// ObserverAdapter creates a runner.Observer that publishes events to the Bus.
func ObserverAdapter(bus *eventbus.Bus, runID string, scope storage.Scope) runner.Observer {
	return &busObserver{
		bus:   bus,
		runID: runID,
		scope: scope,
	}
}

type busObserver struct {
	bus   *eventbus.Bus
	runID string
	scope storage.Scope
}

func (o *busObserver) OnRunStart(ctx context.Context, req *runner.Request) {
	o.bus.Publish(ctx, &storage.Event{
		ID:      fmt.Sprintf("%s-run-start-%d", o.runID, time.Now().UnixNano()),
		RunID:   o.runID,
		Scope:   o.scope,
		Type:    eventbus.SubjectRunStarted,
		Name:    "run_start",
		Message: req.Description,
	})
}

func (o *busObserver) OnRunComplete(ctx context.Context, result *runner.Result) {
	data := map[string]any{
		"status":     result.Status.String(),
		"iterations": result.Iterations,
		"duration":   result.Duration.String(),
	}
	if result.Error != nil {
		data["error"] = result.Error.Error()
	}

	o.bus.Publish(ctx, &storage.Event{
		ID:    fmt.Sprintf("%s-run-complete-%d", o.runID, time.Now().UnixNano()),
		RunID: o.runID,
		Scope: o.scope,
		Type:  eventbus.SubjectRunCompleted,
		Name:  "run_complete",
		Data:  data,
	})
}

func (o *busObserver) OnStepStart(ctx context.Context, step string) {
	o.bus.Publish(ctx, &storage.Event{
		ID:    fmt.Sprintf("%s-%s-obs-start-%d", o.runID, step, time.Now().UnixNano()),
		RunID: o.runID,
		Scope: o.scope,
		Type:  eventbus.SubjectStepPrefix + step,
		Name:  step,
		Data:  map[string]any{"phase": "start"},
	})
}

func (o *busObserver) OnStepComplete(ctx context.Context, step string, duration time.Duration, err error) {
	data := map[string]any{
		"phase":    "complete",
		"duration": duration.String(),
	}
	if err != nil {
		data["error"] = err.Error()
	}

	o.bus.Publish(ctx, &storage.Event{
		ID:    fmt.Sprintf("%s-%s-obs-complete-%d", o.runID, step, time.Now().UnixNano()),
		RunID: o.runID,
		Scope: o.scope,
		Type:  eventbus.SubjectStepPrefix + step,
		Name:  step,
		Data:  data,
	})
}
