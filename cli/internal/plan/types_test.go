package plan

import (
	"encoding/json"
	"testing"
)

func TestTicketPlan_JSON(t *testing.T) {
	plan := TicketPlan{
		TicketID:       "ENG-42",
		Approach:       "Update the button component",
		CandidateFiles: []string{"src/Button.tsx", "src/Button.test.tsx"},
		NewFiles:       []string{"src/NewHelper.ts"},
		DeletedFiles:   nil,
		Validation:     []string{"yarn test src/Button.test.tsx", "yarn check-types"},
		Rollback:       "Revert PR",
		StopConditions: []string{"If tests fail, stop"},
		Uncertainties:  []string{"Unclear if Button is used in admin"},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed TicketPlan
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.TicketID != "ENG-42" {
		t.Errorf("expected ticketId ENG-42, got %s", parsed.TicketID)
	}
	if parsed.Approach != "Update the button component" {
		t.Errorf("unexpected approach: %s", parsed.Approach)
	}
	if len(parsed.CandidateFiles) != 2 {
		t.Errorf("expected 2 candidate files, got %d", len(parsed.CandidateFiles))
	}
	if len(parsed.NewFiles) != 1 {
		t.Errorf("expected 1 new file, got %d", len(parsed.NewFiles))
	}
	if len(parsed.StopConditions) != 1 {
		t.Errorf("expected 1 stop condition, got %d", len(parsed.StopConditions))
	}
}

func TestPlanValidation_JSON(t *testing.T) {
	v := PlanValidation{
		Passed: true,
		GateResults: []PlanGateResult{
			{Gate: "files_exist", Passed: true, Reason: "All 3 files exist"},
			{Gate: "within_repo_areas", Passed: true, Reason: "All within scope"},
			{Gate: "stop_conditions", Passed: true, Reason: "2 stop conditions defined"},
			{Gate: "validation_commands", Passed: true, Reason: "3 commands recognized"},
		},
		ValidatedFiles:  []string{"src/a.ts", "src/b.ts"},
		MissingFiles:    nil,
		OutOfScopeFiles: nil,
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed PlanValidation
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !parsed.Passed {
		t.Error("expected Passed to be true")
	}
	if len(parsed.GateResults) != 4 {
		t.Errorf("expected 4 gate results, got %d", len(parsed.GateResults))
	}
	if len(parsed.ValidatedFiles) != 2 {
		t.Errorf("expected 2 validated files, got %d", len(parsed.ValidatedFiles))
	}
}

func TestPlanResult_JSON_WithError(t *testing.T) {
	result := PlanResult{
		TicketID: "ENG-99",
		Plan:     nil,
		Error:    "planner Claude call failed",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed PlanResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.TicketID != "ENG-99" {
		t.Errorf("expected ENG-99, got %s", parsed.TicketID)
	}
	if parsed.Plan != nil {
		t.Error("expected nil plan")
	}
	if parsed.Error != "planner Claude call failed" {
		t.Errorf("expected error message, got %s", parsed.Error)
	}
}

func TestPlanStats_JSON(t *testing.T) {
	stats := PlanStats{
		Total:           10,
		Passed:          7,
		Failed:          3,
		TotalTokensUsed: 50000,
		TotalCostUSD:    0.25,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed PlanStats
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Total != 10 {
		t.Errorf("expected total 10, got %d", parsed.Total)
	}
	if parsed.Passed != 7 {
		t.Errorf("expected passed 7, got %d", parsed.Passed)
	}
	if parsed.Failed != 3 {
		t.Errorf("expected failed 3, got %d", parsed.Failed)
	}
	if parsed.Passed+parsed.Failed != parsed.Total {
		t.Error("passed + failed should equal total")
	}
}

func TestPlanGateResult_OmitsEmptyReason(t *testing.T) {
	g := PlanGateResult{Gate: "files_exist", Passed: true}

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// reason should be omitted (omitempty).
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if _, ok := raw["reason"]; ok {
		t.Error("expected reason to be omitted when empty")
	}
}
