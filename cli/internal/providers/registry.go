// Package providers contains runtime provider registration for Boatman CLI.
package providers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/providers/openairesponses"
	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/toolbroker"
	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/philjestin/boatmanmode/internal/providers/claudecli"
)

// Registry stores available provider adapters by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]agentruntime.Provider
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]agentruntime.Provider),
	}
}

// NewDefaultRegistry creates a registry with the adapters currently available
// in the CLI.
func NewDefaultRegistry() *Registry {
	return NewRegistryForConfig(nil)
}

// NewRegistryForConfig creates a registry with adapters configured from Boatman
// runtime settings.
func NewRegistryForConfig(cfg *config.Config) *Registry {
	registry := NewRegistry()
	var opts []claudecli.Option
	if cfg != nil && cfg.Claude.Command != "" {
		opts = append(opts, claudecli.WithCommand(cfg.Claude.Command))
	}
	registry.Register(claudecli.New(opts...))
	registry.Register(openairesponses.New(openairesponses.WithToolBroker(toolbroker.NewLocal())))
	return registry
}

// FromConfig returns the provider configured for a workflow role/profile.
func FromConfig(cfg *config.Config, role agentruntime.Role, profile string) (agentruntime.Provider, error) {
	registry := NewRegistryForConfig(cfg)
	providerName := config.DefaultRuntimeProvider
	if cfg != nil {
		providerName = cfg.Runtime.ProviderFor(string(role), profile)
	}
	return registry.MustGet(providerName)
}

// MustFromConfig returns the provider configured for a workflow role/profile or
// panics when configuration references an unavailable adapter.
func MustFromConfig(cfg *config.Config, role agentruntime.Role, profile string) agentruntime.Provider {
	provider, err := FromConfig(cfg, role, profile)
	if err != nil {
		panic(err)
	}
	return provider
}

// Register adds or replaces a provider adapter.
func (r *Registry) Register(provider agentruntime.Provider) {
	if provider == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// Get returns a provider by exact name.
func (r *Registry) Get(name string) (agentruntime.Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

// MustGet returns a provider by name or an error.
func (r *Registry) MustGet(name string) (agentruntime.Provider, error) {
	provider, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("provider %q is not registered", name)
	}
	return provider, nil
}

// Names returns registered provider names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FindByCapability returns providers whose capabilities satisfy predicate.
func (r *Registry) FindByCapability(ctx context.Context, predicate func(agentruntime.Capabilities) bool) ([]agentruntime.Provider, error) {
	r.mu.RLock()
	providers := make([]agentruntime.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	r.mu.RUnlock()

	var matches []agentruntime.Provider
	for _, provider := range providers {
		caps, err := provider.Capabilities(ctx)
		if err != nil {
			return nil, fmt.Errorf("provider %q capabilities: %w", provider.Name(), err)
		}
		if predicate(caps) {
			matches = append(matches, provider)
		}
	}
	return matches, nil
}
