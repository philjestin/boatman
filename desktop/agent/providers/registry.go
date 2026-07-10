// Package providers contains desktop runtime provider registration.
package providers

import (
	"fmt"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/providers/openairesponses"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"

	"boatman/agent/providers/claudecli"
)

const DefaultProvider = "claude-cli"

// Registry stores desktop provider adapters by name.
type Registry struct {
	providers map[string]agentruntime.Provider
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]agentruntime.Provider)}
}

// NewDefaultRegistry creates the provider set available to desktop sessions.
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registry.Register(claudecli.New())
	registry.Register(openairesponses.New(openairesponses.WithToolBroker(toolbroker.NewLocal())))
	return registry
}

// Register adds or replaces a provider adapter.
func (r *Registry) Register(provider agentruntime.Provider) {
	if provider == nil {
		return
	}
	r.providers[provider.Name()] = provider
}

// Get returns a provider by exact name.
func (r *Registry) Get(name string) (agentruntime.Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

// ForRequest returns the provider requested by the run, falling back to the
// desktop default.
func (r *Registry) ForRequest(req agentruntime.RunRequest) (agentruntime.Provider, error) {
	name := req.Provider
	if name == "" {
		name = DefaultProvider
	}
	provider, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("desktop provider %q is not registered", name)
	}
	return provider, nil
}
