package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/philjestin/boatmanmode/internal/triage"
)

// ValidatePlan checks a generated plan against the four ADR-004 gates.
func ValidatePlan(plan *TicketPlan, repoPath string, contextDoc *triage.ContextDoc) *PlanValidation {
	v := &PlanValidation{
		Passed: true,
	}

	// Gate 1: Candidate files exist
	v.GateResults = append(v.GateResults, checkFilesExist(plan, repoPath, v))

	// Gate 2: Candidate files within cluster repoAreas
	if contextDoc != nil {
		v.GateResults = append(v.GateResults, checkWithinRepoAreas(plan, contextDoc.RepoAreas, v))
	} else {
		v.GateResults = append(v.GateResults, PlanGateResult{
			Gate:   "within_repo_areas",
			Passed: true,
			Reason: "No cluster context — skipping scope check",
		})
	}

	// Gate 3: StopConditions non-empty
	v.GateResults = append(v.GateResults, checkStopConditions(plan))

	// Gate 4: Validation commands reference known runners
	v.GateResults = append(v.GateResults, checkValidationCommands(plan, repoPath))

	// Overall pass/fail
	for _, g := range v.GateResults {
		if !g.Passed {
			v.Passed = false
			break
		}
	}

	return v
}

// checkFilesExist verifies each candidateFile exists on disk.
func checkFilesExist(plan *TicketPlan, repoPath string, v *PlanValidation) PlanGateResult {
	if len(plan.CandidateFiles) == 0 {
		return PlanGateResult{
			Gate:   "files_exist",
			Passed: false,
			Reason: "No candidate files specified",
		}
	}

	for _, f := range plan.CandidateFiles {
		fullPath := filepath.Join(repoPath, f)
		if _, err := os.Stat(fullPath); err == nil {
			v.ValidatedFiles = append(v.ValidatedFiles, f)
		} else {
			v.MissingFiles = append(v.MissingFiles, f)
		}
	}

	if len(v.MissingFiles) == len(plan.CandidateFiles) {
		return PlanGateResult{
			Gate:   "files_exist",
			Passed: false,
			Reason: fmt.Sprintf("All %d candidate files are missing", len(plan.CandidateFiles)),
		}
	}

	if len(v.MissingFiles) > 0 {
		return PlanGateResult{
			Gate:   "files_exist",
			Passed: true,
			Reason: fmt.Sprintf("%d of %d files exist (%d missing)", len(v.ValidatedFiles), len(plan.CandidateFiles), len(v.MissingFiles)),
		}
	}

	return PlanGateResult{
		Gate:   "files_exist",
		Passed: true,
		Reason: fmt.Sprintf("All %d candidate files exist", len(plan.CandidateFiles)),
	}
}

// checkWithinRepoAreas verifies candidate files have paths within the cluster's repoAreas.
func checkWithinRepoAreas(plan *TicketPlan, repoAreas []string, v *PlanValidation) PlanGateResult {
	if len(repoAreas) == 0 {
		return PlanGateResult{
			Gate:   "within_repo_areas",
			Passed: true,
			Reason: "No repo areas defined — skipping scope check",
		}
	}

	for _, f := range plan.CandidateFiles {
		inScope := false
		for _, area := range repoAreas {
			if strings.HasPrefix(f, area) {
				inScope = true
				break
			}
		}
		if !inScope {
			v.OutOfScopeFiles = append(v.OutOfScopeFiles, f)
		}
	}

	if len(v.OutOfScopeFiles) > len(plan.CandidateFiles)/2 {
		return PlanGateResult{
			Gate:   "within_repo_areas",
			Passed: false,
			Reason: fmt.Sprintf("%d of %d files are outside cluster repo areas", len(v.OutOfScopeFiles), len(plan.CandidateFiles)),
		}
	}

	if len(v.OutOfScopeFiles) > 0 {
		return PlanGateResult{
			Gate:   "within_repo_areas",
			Passed: true,
			Reason: fmt.Sprintf("%d file(s) outside repo areas (within tolerance)", len(v.OutOfScopeFiles)),
		}
	}

	return PlanGateResult{
		Gate:   "within_repo_areas",
		Passed: true,
		Reason: "All candidate files within cluster repo areas",
	}
}

// checkStopConditions verifies the plan has at least one stop condition.
func checkStopConditions(plan *TicketPlan) PlanGateResult {
	if len(plan.StopConditions) == 0 {
		return PlanGateResult{
			Gate:   "stop_conditions",
			Passed: false,
			Reason: "No stop conditions defined — plan must include at least one",
		}
	}
	return PlanGateResult{
		Gate:   "stop_conditions",
		Passed: true,
		Reason: fmt.Sprintf("%d stop condition(s) defined", len(plan.StopConditions)),
	}
}

// knownRunners are command prefixes that we recognize as valid test/lint runners.
var knownRunners = []string{
	"yarn test",
	"yarn check-types",
	"yarn lint",
	"yarn test:affected",
	"bundle exec rspec",
	"bundle exec rubocop",
	"bin/packwerk",
	"make ",
	"npx ",
	"npm test",
	"npm run",
}

// checkValidationCommands verifies validation commands use known runners.
func checkValidationCommands(plan *TicketPlan, repoPath string) PlanGateResult {
	if len(plan.Validation) == 0 {
		return PlanGateResult{
			Gate:   "validation_commands",
			Passed: false,
			Reason: "No validation commands specified",
		}
	}

	unknownCount := 0
	for _, cmd := range plan.Validation {
		known := false
		for _, runner := range knownRunners {
			if strings.HasPrefix(cmd, runner) {
				known = true
				break
			}
		}
		if !known {
			unknownCount++
		}
	}

	if unknownCount == len(plan.Validation) {
		return PlanGateResult{
			Gate:   "validation_commands",
			Passed: false,
			Reason: fmt.Sprintf("None of the %d validation commands use a known test runner", len(plan.Validation)),
		}
	}

	return PlanGateResult{
		Gate:   "validation_commands",
		Passed: true,
		Reason: fmt.Sprintf("%d validation command(s) recognized", len(plan.Validation)-unknownCount),
	}
}
