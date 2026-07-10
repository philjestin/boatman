package providers

import (
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestDefaultRegistryIncludesDesktopProviders(t *testing.T) {
	registry := NewDefaultRegistry()

	for _, name := range []string{"claude-cli", "openai-responses"} {
		if _, ok := registry.Get(name); !ok {
			t.Fatalf("default registry should include %s", name)
		}
	}
}

func TestForRequestUsesDefaultProvider(t *testing.T) {
	provider, err := NewDefaultRegistry().ForRequest(agentruntime.RunRequest{})
	if err != nil {
		t.Fatalf("ForRequest error: %v", err)
	}
	if provider.Name() != DefaultProvider {
		t.Fatalf("Provider = %q, want %s", provider.Name(), DefaultProvider)
	}
}

func TestForRequestRejectsUnknownProvider(t *testing.T) {
	_, err := NewDefaultRegistry().ForRequest(agentruntime.RunRequest{Provider: "missing"})
	if err == nil {
		t.Fatal("ForRequest should fail for unknown provider")
	}
}
