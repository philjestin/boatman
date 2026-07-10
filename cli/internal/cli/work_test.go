package cli

import (
	"testing"

	"github.com/philjestin/boatmanmode/internal/config"
	"github.com/spf13/cobra"
)

func TestApplyWorkModelFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("plan-model", "", "")
	cmd.Flags().String("implementation-model", "", "")
	cmd.Flags().String("skill-model", "", "")
	if err := cmd.Flags().Set("plan-model", "opus"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("implementation-model", "sonnet"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("skill-model", "haiku"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	applyWorkModelFlags(cmd, cfg)

	if cfg.Claude.Models.Planner != "opus" {
		t.Fatalf("Planner = %q, want opus", cfg.Claude.Models.Planner)
	}
	if cfg.Claude.Models.Executor != "sonnet" || cfg.Claude.Models.Refactor != "sonnet" {
		t.Fatalf("implementation models = executor %q refactor %q, want sonnet", cfg.Claude.Models.Executor, cfg.Claude.Models.Refactor)
	}
	if cfg.Claude.Models.Reviewer != "haiku" {
		t.Fatalf("Reviewer = %q, want haiku", cfg.Claude.Models.Reviewer)
	}
}
