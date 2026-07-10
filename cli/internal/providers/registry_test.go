package providers

import (
	"context"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
)

type fakeProvider struct {
	name string
	caps agentruntime.Capabilities
}

func (p fakeProvider) Name() string {
	return p.name
}

func (p fakeProvider) Capabilities(context.Context) (agentruntime.Capabilities, error) {
	return p.caps, nil
}

func (p fakeProvider) StartRun(context.Context, agentruntime.RunRequest) (agentruntime.EventStream, error) {
	return nil, nil
}

func (p fakeProvider) ResumeRun(context.Context, string, agentruntime.RunInput) (agentruntime.EventStream, error) {
	return nil, nil
}

func (p fakeProvider) CancelRun(context.Context, string) error {
	return nil
}

func TestDefaultRegistryIncludesClaudeCLI(t *testing.T) {
	registry := NewDefaultRegistry()

	if _, ok := registry.Get("claude-cli"); !ok {
		t.Fatal("default registry should include claude-cli")
	}
	if _, ok := registry.Get("openai-responses"); !ok {
		t.Fatal("default registry should include openai-responses")
	}
}

func TestFromConfigUsesRoleProfileRouting(t *testing.T) {
	cfg := &config.Config{
		Runtime: config.RuntimeConfig{
			DefaultProvider: "missing",
			RoleProviders: map[string]string{
				"planner": "claude-cli",
			},
		},
	}

	provider, err := FromConfig(cfg, agentruntime.RolePlanner, "planner")
	if err != nil {
		t.Fatalf("FromConfig error: %v", err)
	}
	if provider.Name() != "claude-cli" {
		t.Fatalf("Provider = %q, want claude-cli", provider.Name())
	}
}

func TestFromConfigReturnsErrorForUnknownProvider(t *testing.T) {
	cfg := &config.Config{
		Runtime: config.RuntimeConfig{
			DefaultProvider: "missing-provider",
		},
	}

	if _, err := FromConfig(cfg, agentruntime.RoleExecutor, "executor"); err == nil {
		t.Fatal("FromConfig should fail for an unregistered provider")
	}
}

func TestRegistryMustGetMissingProvider(t *testing.T) {
	registry := NewRegistry()

	if _, err := registry.MustGet("missing"); err == nil {
		t.Fatal("MustGet should fail for missing provider")
	}
}

func TestRegistryNamesSorted(t *testing.T) {
	registry := NewRegistry()
	registry.Register(fakeProvider{name: "z-provider"})
	registry.Register(fakeProvider{name: "a-provider"})

	names := registry.Names()
	if len(names) != 2 || names[0] != "a-provider" || names[1] != "z-provider" {
		t.Fatalf("names = %#v, want sorted provider names", names)
	}
}

func TestFindByCapability(t *testing.T) {
	registry := NewRegistry()
	registry.Register(fakeProvider{
		name: "tools",
		caps: agentruntime.Capabilities{Provider: "tools", SupportsToolCalls: true},
	})
	registry.Register(fakeProvider{
		name: "plain",
		caps: agentruntime.Capabilities{Provider: "plain", SupportsToolCalls: false},
	})

	matches, err := registry.FindByCapability(context.Background(), func(caps agentruntime.Capabilities) bool {
		return caps.SupportsToolCalls
	})
	if err != nil {
		t.Fatalf("FindByCapability error: %v", err)
	}
	if len(matches) != 1 || matches[0].Name() != "tools" {
		t.Fatalf("matches = %#v, want only tools provider", matches)
	}
}
