package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
	"github.com/philjestin/boatmanmode/internal/config"
	runtimeproviders "github.com/philjestin/boatmanmode/internal/providers"
)

func TestWriteProvidersText(t *testing.T) {
	var buf bytes.Buffer
	err := writeProviders(&buf, []providerInfo{
		{
			Name: "openai-responses",
			Capabilities: agentruntime.Capabilities{
				Provider:              "openai-responses",
				SupportsToolCalls:     true,
				SupportsMCP:           true,
				SupportsApprovals:     true,
				SupportsStructuredOut: true,
				SupportsUsage:         true,
				Experimental:          []string{"responses-api", "remote-mcp"},
			},
		},
	}, false)
	if err != nil {
		t.Fatalf("writeProviders error: %v", err)
	}
	output := buf.String()
	for _, want := range []string{"PROVIDER", "openai-responses", "yes", "responses-api,remote-mcp"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q should contain %q", output, want)
		}
	}
}

func TestWriteProvidersJSON(t *testing.T) {
	var buf bytes.Buffer
	err := writeProviders(&buf, []providerInfo{
		{
			Name: "claude-cli",
			Capabilities: agentruntime.Capabilities{
				Provider:          "claude-cli",
				SupportsStreaming: true,
			},
		},
	}, true)
	if err != nil {
		t.Fatalf("writeProviders error: %v", err)
	}

	var decoded []providerInfo
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("json output did not decode: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Name != "claude-cli" || !decoded[0].Capabilities.SupportsStreaming {
		t.Fatalf("decoded = %#v, want claude-cli streaming provider", decoded)
	}
}

func TestMissingConfiguredProviders(t *testing.T) {
	cfg := &config.Config{
		Runtime: config.RuntimeConfig{
			DefaultProvider: "claude-cli",
			RoleProviders: map[string]string{
				"executor": "missing-role-provider",
			},
			ProfileProviders: map[string]string{
				"triage-scorer": "missing-profile-provider",
			},
		},
	}

	missing := missingConfiguredProviders(cfg, runtimeproviders.NewDefaultRegistry())
	if len(missing) != 2 ||
		missing[0] != "role.executor=missing-role-provider" ||
		missing[1] != "profile.triage-scorer=missing-profile-provider" {
		t.Fatalf("missing = %#v, want missing role/profile routes", missing)
	}
}
