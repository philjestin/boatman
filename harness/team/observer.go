package team

import (
	"context"
	"time"
)

// TeamObserver provides structured observation of team lifecycle events.
type TeamObserver interface {
	OnTeamStart(ctx context.Context, teamName string, task *Task)
	OnTeamComplete(ctx context.Context, teamName string, result *Result, err error)
	OnAgentStart(ctx context.Context, teamName, agentName string, task *Task)
	OnAgentComplete(ctx context.Context, teamName, agentName string, result *Result, duration time.Duration, err error)
	OnRouteDecision(ctx context.Context, teamName string, selected []Selection)
}

// NopTeamObserver is a no-op TeamObserver that ignores all events.
type NopTeamObserver struct{}

func (NopTeamObserver) OnTeamStart(_ context.Context, _ string, _ *Task)      {}
func (NopTeamObserver) OnTeamComplete(_ context.Context, _ string, _ *Result, _ error) {
}
func (NopTeamObserver) OnAgentStart(_ context.Context, _, _ string, _ *Task) {}
func (NopTeamObserver) OnAgentComplete(_ context.Context, _, _ string, _ *Result, _ time.Duration, _ error) {
}
func (NopTeamObserver) OnRouteDecision(_ context.Context, _ string, _ []Selection) {}

// MultiTeamObserver fans out events to multiple observers.
type MultiTeamObserver struct {
	Observers []TeamObserver
}

// NewMultiTeamObserver creates a MultiTeamObserver from the given observers.
func NewMultiTeamObserver(observers ...TeamObserver) *MultiTeamObserver {
	return &MultiTeamObserver{Observers: observers}
}

func (m *MultiTeamObserver) OnTeamStart(ctx context.Context, teamName string, task *Task) {
	for _, o := range m.Observers {
		o.OnTeamStart(ctx, teamName, task)
	}
}

func (m *MultiTeamObserver) OnTeamComplete(ctx context.Context, teamName string, result *Result, err error) {
	for _, o := range m.Observers {
		o.OnTeamComplete(ctx, teamName, result, err)
	}
}

func (m *MultiTeamObserver) OnAgentStart(ctx context.Context, teamName, agentName string, task *Task) {
	for _, o := range m.Observers {
		o.OnAgentStart(ctx, teamName, agentName, task)
	}
}

func (m *MultiTeamObserver) OnAgentComplete(ctx context.Context, teamName, agentName string, result *Result, duration time.Duration, err error) {
	for _, o := range m.Observers {
		o.OnAgentComplete(ctx, teamName, agentName, result, duration, err)
	}
}

func (m *MultiTeamObserver) OnRouteDecision(ctx context.Context, teamName string, selected []Selection) {
	for _, o := range m.Observers {
		o.OnRouteDecision(ctx, teamName, selected)
	}
}
