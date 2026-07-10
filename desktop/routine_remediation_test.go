package main

import (
	"strings"
	"testing"
)

func TestRoutineRemediationCandidatesFromReport(t *testing.T) {
	report := `# Report

boatman_remediation_candidates
` + "```json" + `
{
  "candidates": [
    {
      "id": "slow-job-details",
      "title": "Fix slow job details resolver",
      "should_remediate": true,
      "confidence": 0.91,
      "summary": "N+1 query in resolver",
      "telemetry_evidence": ["p95 2400ms"],
      "suspected_files": ["packs/jobs"],
      "validation": ["go test ./..."],
      "expected_impact": "lower p95",
      "risk": "resolver fanout",
      "prompt": "Fix the resolver"
    }
  ]
}
` + "```"

	candidates, err := routineRemediationCandidatesFromReport(report)
	if err != nil {
		t.Fatalf("routineRemediationCandidatesFromReport error: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ID != "slow-job-details" || candidates[0].Confidence != 0.91 {
		t.Fatalf("candidates = %#v", candidates)
	}
}

func TestSelectRoutineRemediationCandidatesFiltersAndLimits(t *testing.T) {
	candidates := []routineRemediationCandidate{
		{ID: "low", ShouldRemediate: true, Confidence: 0.2, Prompt: "fix low"},
		{ID: "manual", ShouldRemediate: false, Confidence: 0.99, Prompt: "fix manual"},
		{ID: "second", ShouldRemediate: true, Confidence: 0.8, Prompt: "fix second"},
		{ID: "first", ShouldRemediate: true, Confidence: 0.95, Summary: "fix first"},
	}

	got := selectRoutineRemediationCandidates(candidates, 1)
	if len(got) != 1 || got[0].ID != "first" {
		t.Fatalf("selected = %#v, want highest-confidence actionable candidate", got)
	}
}

func TestAppendRoutineRemediationSummary(t *testing.T) {
	report := appendRoutineRemediationSummary("# Report", []RoutineRemediationResult{
		{Title: "Fix resolver", Status: "completed", PRURL: "https://github.com/acme/repo/pull/1", Message: "Draft PR created."},
	}, "")

	for _, want := range []string{"## Boatman Automatic Remediation", "Fix resolver", "https://github.com/acme/repo/pull/1"} {
		if !strings.Contains(report, want) {
			t.Fatalf("summary should contain %q:\n%s", want, report)
		}
	}
}
