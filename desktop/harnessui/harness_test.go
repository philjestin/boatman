package harnessui

import (
	"strings"
	"testing"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

func TestHarnessRunEnvAddsModelProfile(t *testing.T) {
	env := harnessRunEnv([]string{"BASE=1"}, RunRequest{
		Model: "dropdown-model",
		Models: agentruntime.ModelProfile{
			Plan:   "plan-model",
			Skills: "skill-model",
		},
	})
	joined := "\n" + strings.Join(env, "\n") + "\n"
	for _, want := range []string{
		"\nHARNESS_MODEL=dropdown-model\n",
		"\nHARNESS_PLAN_MODEL=plan-model\n",
		"\nHARNESS_IMPLEMENTATION_MODEL=dropdown-model\n",
		"\nHARNESS_SKILLS_MODEL=skill-model\n",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("env missing %q in %#v", strings.TrimSpace(want), env)
		}
	}
}
