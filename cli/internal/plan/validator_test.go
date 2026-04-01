package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/philjestin/boatmanmode/internal/triage"
)

func TestValidatePlan_AllGatesPass(t *testing.T) {
	dir := t.TempDir()
	// Create candidate files.
	for _, f := range []string{"src/foo.ts", "src/bar.ts"} {
		full := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte("x"), 0o644)
	}

	plan := &TicketPlan{
		CandidateFiles: []string{"src/foo.ts", "src/bar.ts"},
		StopConditions: []string{"If tests fail, stop"},
		Validation:     []string{"yarn test src/foo.test.ts"},
	}

	ctx := &triage.ContextDoc{
		RepoAreas: []string{"src/"},
	}

	v := ValidatePlan(plan, dir, ctx)

	if !v.Passed {
		t.Errorf("expected all gates to pass, got Passed=false")
		for _, g := range v.GateResults {
			t.Logf("  gate %s: passed=%v reason=%s", g.Gate, g.Passed, g.Reason)
		}
	}
	if len(v.ValidatedFiles) != 2 {
		t.Errorf("expected 2 validated files, got %d", len(v.ValidatedFiles))
	}
	if len(v.MissingFiles) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(v.MissingFiles))
	}
}

func TestValidatePlan_NoCandidateFiles(t *testing.T) {
	plan := &TicketPlan{
		CandidateFiles: nil,
		StopConditions: []string{"stop if blocked"},
		Validation:     []string{"yarn test"},
	}

	v := ValidatePlan(plan, t.TempDir(), nil)

	if v.Passed {
		t.Error("expected validation to fail with no candidate files")
	}

	found := false
	for _, g := range v.GateResults {
		if g.Gate == "files_exist" && !g.Passed {
			found = true
		}
	}
	if !found {
		t.Error("expected files_exist gate to fail")
	}
}

func TestValidatePlan_AllFilesMissing(t *testing.T) {
	plan := &TicketPlan{
		CandidateFiles: []string{"nonexistent/a.ts", "nonexistent/b.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	v := ValidatePlan(plan, t.TempDir(), nil)

	if v.Passed {
		t.Error("expected validation to fail when all files are missing")
	}
	if len(v.MissingFiles) != 2 {
		t.Errorf("expected 2 missing files, got %d", len(v.MissingFiles))
	}
}

func TestValidatePlan_SomeFilesMissing(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "src/real.ts")
	os.MkdirAll(filepath.Dir(existing), 0o755)
	os.WriteFile(existing, []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"src/real.ts", "src/missing.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	v := ValidatePlan(plan, dir, nil)

	// Should still pass (partial is OK).
	filesGate := findGate(v.GateResults, "files_exist")
	if !filesGate.Passed {
		t.Error("expected files_exist gate to pass with partial files")
	}
	if len(v.ValidatedFiles) != 1 {
		t.Errorf("expected 1 validated file, got %d", len(v.ValidatedFiles))
	}
	if len(v.MissingFiles) != 1 {
		t.Errorf("expected 1 missing file, got %d", len(v.MissingFiles))
	}
}

func TestValidatePlan_OutOfScopeFiles(t *testing.T) {
	dir := t.TempDir()
	// Create files in two areas.
	for _, f := range []string{"src/a.ts", "lib/b.ts", "lib/c.ts"} {
		full := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte("x"), 0o644)
	}

	plan := &TicketPlan{
		CandidateFiles: []string{"src/a.ts", "lib/b.ts", "lib/c.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	ctx := &triage.ContextDoc{
		RepoAreas: []string{"src/"},
	}

	v := ValidatePlan(plan, dir, ctx)

	// Majority of files (2/3) are outside repo areas — gate should fail.
	scopeGate := findGate(v.GateResults, "within_repo_areas")
	if scopeGate.Passed {
		t.Error("expected within_repo_areas gate to fail when majority of files are out of scope")
	}
	if len(v.OutOfScopeFiles) != 2 {
		t.Errorf("expected 2 out-of-scope files, got %d", len(v.OutOfScopeFiles))
	}
}

func TestValidatePlan_OutOfScopeWithinTolerance(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"src/a.ts", "src/b.ts", "lib/c.ts"} {
		full := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, []byte("x"), 0o644)
	}

	plan := &TicketPlan{
		CandidateFiles: []string{"src/a.ts", "src/b.ts", "lib/c.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	ctx := &triage.ContextDoc{
		RepoAreas: []string{"src/"},
	}

	v := ValidatePlan(plan, dir, ctx)

	// Only 1/3 files outside scope — within tolerance.
	scopeGate := findGate(v.GateResults, "within_repo_areas")
	if !scopeGate.Passed {
		t.Error("expected within_repo_areas gate to pass (1/3 files outside is within tolerance)")
	}
}

func TestValidatePlan_NoRepoAreas(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	ctx := &triage.ContextDoc{
		RepoAreas: nil,
	}

	v := ValidatePlan(plan, dir, ctx)

	scopeGate := findGate(v.GateResults, "within_repo_areas")
	if !scopeGate.Passed {
		t.Error("expected within_repo_areas gate to pass with no repo areas defined")
	}
}

func TestValidatePlan_NoContextDoc(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test"},
	}

	v := ValidatePlan(plan, dir, nil)

	scopeGate := findGate(v.GateResults, "within_repo_areas")
	if !scopeGate.Passed {
		t.Error("expected within_repo_areas gate to pass with nil context doc")
	}
}

func TestValidatePlan_NoStopConditions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: nil,
		Validation:     []string{"yarn test"},
	}

	v := ValidatePlan(plan, dir, nil)

	if v.Passed {
		t.Error("expected validation to fail with no stop conditions")
	}

	gate := findGate(v.GateResults, "stop_conditions")
	if gate.Passed {
		t.Error("expected stop_conditions gate to fail")
	}
}

func TestValidatePlan_NoValidationCommands(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: []string{"stop"},
		Validation:     nil,
	}

	v := ValidatePlan(plan, dir, nil)

	if v.Passed {
		t.Error("expected validation to fail with no validation commands")
	}

	gate := findGate(v.GateResults, "validation_commands")
	if gate.Passed {
		t.Error("expected validation_commands gate to fail")
	}
}

func TestValidatePlan_UnknownValidationCommands(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"my-custom-runner test", "another-thing"},
	}

	v := ValidatePlan(plan, dir, nil)

	gate := findGate(v.GateResults, "validation_commands")
	if gate.Passed {
		t.Error("expected validation_commands gate to fail for unknown runners")
	}
}

func TestValidatePlan_KnownValidationRunners(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"yarn test", "yarn test src/foo.test.ts"},
		{"yarn check-types", "yarn check-types"},
		{"yarn lint", "yarn lint --fix"},
		{"bundle exec rspec", "bundle exec rspec spec/foo_spec.rb"},
		{"bundle exec rubocop", "bundle exec rubocop app/"},
		{"npm test", "npm test"},
		{"npm run", "npm run lint"},
		{"npx", "npx tsc --noEmit"},
		{"make", "make test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

			plan := &TicketPlan{
				CandidateFiles: []string{"a.ts"},
				StopConditions: []string{"stop"},
				Validation:     []string{tt.cmd},
			}

			v := ValidatePlan(plan, dir, nil)
			gate := findGate(v.GateResults, "validation_commands")
			if !gate.Passed {
				t.Errorf("expected validation_commands gate to pass for %q", tt.cmd)
			}
		})
	}
}

func TestValidatePlan_MixedValidationCommands(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.ts"), []byte("x"), 0o644)

	plan := &TicketPlan{
		CandidateFiles: []string{"a.ts"},
		StopConditions: []string{"stop"},
		Validation:     []string{"yarn test foo", "unknown-tool bar"},
	}

	v := ValidatePlan(plan, dir, nil)

	gate := findGate(v.GateResults, "validation_commands")
	if !gate.Passed {
		t.Error("expected validation_commands gate to pass with at least one known runner")
	}
}

func findGate(gates []PlanGateResult, name string) PlanGateResult {
	for _, g := range gates {
		if g.Gate == name {
			return g
		}
	}
	return PlanGateResult{}
}
