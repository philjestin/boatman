package integrations

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

// State describes an integration's broker-visible lifecycle.
type State string

const (
	StateUnknown            State = "unknown"
	StateDisabled           State = "disabled"
	StateNeedsConfiguration State = "needs_configuration"
	StateReady              State = "ready"
	StateConnected          State = "connected"
	StateDegraded           State = "degraded"
	StateFailed             State = "failed"
)

// ResolveOptions supplies session- or user-specific integration configuration.
type ResolveOptions struct {
	Enabled bool              `json:"enabled"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
}

// Status is the provider-neutral health view for one integration.
type Status struct {
	Name        string            `json:"name"`
	State       State             `json:"state"`
	Message     string            `json:"message,omitempty"`
	MissingEnv  []string          `json:"missingEnv,omitempty"`
	LastChecked time.Time         `json:"lastChecked"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Broker resolves integration descriptors and emits provider-neutral health
// state. It is intentionally descriptor-only for now; long-lived connection
// lifecycles can sit behind this contract later.
type Broker struct {
	Catalog Catalog
	Now     func() time.Time
}

// NewBroker creates a descriptor-backed integration broker.
func NewBroker(catalog Catalog) *Broker {
	return &Broker{Catalog: catalog}
}

// ResolveMCP returns an MCP ref plus the broker-visible integration status.
func (b *Broker) ResolveMCP(ctx context.Context, name string, opts ResolveOptions) (agentruntime.MCPServerRef, Status, error) {
	if err := ctx.Err(); err != nil {
		return agentruntime.MCPServerRef{}, Status{}, err
	}
	status, err := b.Status(ctx, name, opts)
	if err != nil {
		return agentruntime.MCPServerRef{}, status, err
	}
	if status.State != StateReady && status.State != StateConnected {
		return agentruntime.MCPServerRef{}, status, nil
	}

	item, _ := b.Catalog.Lookup(name)
	ref, ok := item.MCPRef()
	if !ok {
		return agentruntime.MCPServerRef{}, status, fmt.Errorf("integration %q does not expose MCP", normalizeName(name))
	}
	if opts.URL != "" {
		ref.URL = strings.TrimSpace(opts.URL)
		ref.Command = ""
		ref.Args = nil
	}
	ref.Env = mergeEnv(ref.Env, opts.Env)
	return ref, status, nil
}

// Status returns the descriptor-level health for one integration.
func (b *Broker) Status(ctx context.Context, name string, opts ResolveOptions) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	name = normalizeName(name)
	now := b.now()
	status := Status{Name: name, State: StateUnknown, LastChecked: now}
	if name == "" {
		status.State = StateFailed
		status.Message = "integration name is required"
		return status, fmt.Errorf("%s", status.Message)
	}
	item, ok := b.Catalog.Lookup(name)
	if !ok {
		status.State = StateFailed
		status.Message = fmt.Sprintf("unknown integration %q", name)
		return status, fmt.Errorf("%s", status.Message)
	}
	if !opts.Enabled {
		status.State = StateDisabled
		status.Message = "integration is disabled"
		return status, nil
	}

	missing := missingEnv(item, opts.Env)
	if len(missing) > 0 {
		status.State = StateNeedsConfiguration
		status.Message = "integration is missing required configuration"
		status.MissingEnv = missing
		return status, nil
	}
	status.State = StateReady
	status.Message = "integration descriptor is ready"
	return status, nil
}

// Event converts integration health into the normalized runtime event stream.
func (b *Broker) Event(status Status) agentruntime.Event {
	event := agentruntime.NewEvent(agentruntime.EventIntegrationState)
	event.Name = status.Name
	event.Status = runtimeStatus(status.State)
	event.Message = status.Message
	event.Data = map[string]any{
		"state":       status.State,
		"missing_env": status.MissingEnv,
	}
	if status.Metadata != nil {
		event.Data["metadata"] = status.Metadata
	}
	if !status.LastChecked.IsZero() {
		event.Timestamp = status.LastChecked
	}
	return event
}

func (b *Broker) now() time.Time {
	if b != nil && b.Now != nil {
		return b.Now().UTC()
	}
	return time.Now().UTC()
}

func missingEnv(item Integration, overrides map[string]string) []string {
	if item.MCP == nil {
		return nil
	}
	var missing []string
	for _, key := range item.MCP.RequiredEnv() {
		if strings.TrimSpace(overrides[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func mergeEnv(base, overrides map[string]string) map[string]string {
	out := cloneStringMap(base)
	if out == nil {
		out = make(map[string]string)
	}
	for key, value := range overrides {
		if strings.TrimSpace(value) != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func runtimeStatus(state State) agentruntime.Status {
	switch state {
	case StateDisabled:
		return agentruntime.StatusSkipped
	case StateNeedsConfiguration:
		return agentruntime.StatusWaiting
	case StateReady, StateConnected:
		return agentruntime.StatusSucceeded
	case StateDegraded:
		return agentruntime.StatusRunning
	case StateFailed, StateUnknown:
		return agentruntime.StatusFailed
	default:
		return agentruntime.StatusCompleted
	}
}

// RequiredEnv returns sorted environment keys that must be configured because
// the descriptor does not provide a default value.
func (m MCPDescriptor) RequiredEnv() []string {
	keys := make([]string, 0, len(m.Env))
	for key, value := range m.Env {
		if strings.TrimSpace(value) == "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
