package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"boatman/agent"
	bmintegration "boatman/boatmanmode"

	"github.com/philjestin/boatman-ecosystem/shared/agentruntime/routines"
)

const routineRemediationConfidenceThreshold = 0.7

var githubPRURLPattern = regexp.MustCompile(`https://github\.com/[^\s)]+/pull/[0-9]+`)

// RoutineRemediationResult summarizes one automatic remediation child run.
type RoutineRemediationResult struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	PRURL      string `json:"prUrl,omitempty"`
	BranchName string `json:"branchName,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

type routineRemediationPlan struct {
	Candidates []routineRemediationCandidate `json:"candidates"`
}

type routineRemediationCandidate struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	ShouldRemediate   bool     `json:"should_remediate"`
	Confidence        float64  `json:"confidence"`
	Summary           string   `json:"summary"`
	TelemetryEvidence []string `json:"telemetry_evidence"`
	SuspectedFiles    []string `json:"suspected_files"`
	Validation        []string `json:"validation"`
	ExpectedImpact    string   `json:"expected_impact"`
	Risk              string   `json:"risk"`
	Prompt            string   `json:"prompt"`
}

func (a *App) maybeRunRoutineRemediations(ctx context.Context, built *builtRoutineRun, session *agent.Session, report string) ([]RoutineRemediationResult, string) {
	if !routineAutoRemediationEnabled(built.routine) {
		return nil, report
	}
	maxRemediations := routineMaxRemediations(built.values)
	if maxRemediations <= 0 {
		return nil, appendRoutineRemediationSummary(report, nil, "Automatic remediation disabled because max_remediations is 0.")
	}

	candidates, err := routineRemediationCandidatesFromReport(report)
	if err != nil {
		message := "Automatic remediation skipped: " + err.Error()
		session.AddBoatmanMessage("system", message)
		return nil, appendRoutineRemediationSummary(report, nil, message)
	}
	selected := selectRoutineRemediationCandidates(candidates, maxRemediations)
	if len(selected) == 0 {
		message := "Automatic remediation found no high-confidence actionable candidates."
		session.AddBoatmanMessage("system", message)
		return nil, appendRoutineRemediationSummary(report, nil, message)
	}
	session.SetStatus(agent.SessionStatusRunning)
	defer session.SetStatus(agent.SessionStatusIdle)

	prefs := a.config.GetPreferences()
	integration, err := bmintegration.NewIntegration("", prefs.APIKey, built.request.WorkDir)
	if err != nil {
		result := RoutineRemediationResult{
			Title:   "Automatic remediation setup",
			Status:  "failed",
			Error:   err.Error(),
			Message: "BoatmanMode could not start.",
		}
		session.AddBoatmanMessage("system", "Automatic remediation failed to start: "+err.Error())
		return []RoutineRemediationResult{result}, appendRoutineRemediationSummary(report, []RoutineRemediationResult{result}, "")
	}

	results := make([]RoutineRemediationResult, 0, len(selected))
	for i, candidate := range selected {
		result := a.runRoutineRemediationCandidate(ctx, built, session, integration, candidate, i+1, len(selected))
		results = append(results, result)
	}
	return results, appendRoutineRemediationSummary(report, results, "")
}

func (a *App) runRoutineRemediationCandidate(ctx context.Context, built *builtRoutineRun, session *agent.Session, integration *bmintegration.Integration, candidate routineRemediationCandidate, index, total int) RoutineRemediationResult {
	id := routineRemediationID(candidate, index)
	title := routineRemediationTitle(candidate, index)
	branchName := routineRemediationBranchName(built.request.RunID, candidate, index)
	taskID := "routine-remediation-" + id

	result := RoutineRemediationResult{
		ID:         id,
		Title:      title,
		Status:     "running",
		BranchName: branchName,
	}
	session.AddOrUpdateTask(taskID, "Remediate: "+title, candidate.Summary, "in_progress")
	session.AddBoatmanMessage("system", fmt.Sprintf("Starting remediation %d/%d: %s", index, total, title))

	outputChan := make(chan string, 100)
	var output strings.Builder
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range outputChan {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			output.WriteString(line)
			output.WriteString("\n")
			if strings.Contains(line, "PR") || strings.Contains(line, "Worktree") || strings.Contains(line, "Review") {
				session.AddBoatmanMessage("system", line)
			}
		}
	}()

	var streamBuilder strings.Builder
	var streamMsgID string
	onMessage := func(role, content string) {
		if role == "provider.raw" {
			session.ProcessExternalStreamLine(content, &streamBuilder, &streamMsgID)
			return
		}
		session.AddBoatmanMessage(role, content)
	}
	onEvent := func(event bmintegration.BoatmanEvent) {
		_ = a.HandleBoatmanModeEvent(session.ID, event.Type, boatmanEventMap(event))
	}

	_, err := integration.StreamExecutionWithOptions(ctx, session.ID, routineRemediationPrompt(built, candidate), "prompt", bmintegration.StreamExecutionOptions{
		KeepDraft:         true,
		ReviewSkill:       "peer-review",
		ExtraReviewSkills: []string{"lydia-code-review"},
		Models:            built.models,
		Title:             title,
		BranchName:        branchName,
		SuppressWailsEmit: true,
		OnEvent:           onEvent,
	}, outputChan, onMessage)
	close(outputChan)
	<-done

	result.PRURL = firstPRURL(output.String())
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.Message = "Remediation failed before a draft PR completed."
		session.AddOrUpdateTask(taskID, "Remediate: "+title, candidate.Summary, "failed")
		session.AddBoatmanMessage("system", fmt.Sprintf("Remediation failed: %s", err.Error()))
		return result
	}

	result.Status = "completed"
	if result.PRURL != "" {
		result.Message = "Draft PR created."
		session.AddBoatmanMessage("system", fmt.Sprintf("Remediation completed with draft PR: %s", result.PRURL))
	} else {
		result.Message = "Remediation completed; no PR URL was captured from output."
		session.AddBoatmanMessage("system", result.Message)
	}
	session.AddOrUpdateTask(taskID, "Remediate: "+title, candidate.Summary, "completed")
	return result
}

func routineAutoRemediationEnabled(routine routines.Routine) bool {
	value := strings.TrimSpace(routine.Metadata["automation"])
	return value == "remediation-loop"
}

func routineMaxRemediations(values map[string]string) int {
	value := strings.TrimSpace(values["max_remediations"])
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0
	}
	if n > 10 {
		return 10
	}
	return n
}

func routineRemediationCandidatesFromReport(report string) ([]routineRemediationCandidate, error) {
	body, ok := routineTaggedJSONBlock(report, "boatman_remediation_candidates")
	if !ok {
		return nil, fmt.Errorf("report did not include boatman_remediation_candidates JSON")
	}
	var plan routineRemediationPlan
	if err := json.Unmarshal([]byte(body), &plan); err != nil {
		return nil, fmt.Errorf("invalid remediation candidates JSON: %w", err)
	}
	return plan.Candidates, nil
}

func routineTaggedJSONBlock(text, marker string) (string, bool) {
	markerIndex := strings.LastIndex(text, marker)
	if markerIndex < 0 {
		return "", false
	}
	remaining := text[markerIndex+len(marker):]
	for _, fence := range []string{"```", "~~~"} {
		if body, ok := taggedJSONBlockWithFence(remaining, fence); ok {
			return body, true
		}
	}
	return "", false
}

func taggedJSONBlockWithFence(text, fence string) (string, bool) {
	remaining := text
	for {
		start := strings.Index(remaining, fence)
		if start < 0 {
			return "", false
		}
		remaining = remaining[start+len(fence):]
		lineEnd := strings.Index(remaining, "\n")
		if lineEnd < 0 {
			return "", false
		}
		lang := strings.TrimSpace(remaining[:lineEnd])
		remaining = remaining[lineEnd+1:]
		end := strings.Index(remaining, fence)
		if end < 0 {
			return "", false
		}
		body := strings.TrimSpace(remaining[:end])
		if lang == "" || strings.EqualFold(lang, "json") {
			return body, true
		}
		remaining = remaining[end+len(fence):]
	}
}

func selectRoutineRemediationCandidates(candidates []routineRemediationCandidate, limit int) []routineRemediationCandidate {
	eligible := make([]routineRemediationCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.ShouldRemediate || candidate.Confidence < routineRemediationConfidenceThreshold {
			continue
		}
		if strings.TrimSpace(candidate.Prompt) == "" && strings.TrimSpace(candidate.Summary) == "" {
			continue
		}
		eligible = append(eligible, candidate)
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		return eligible[i].Confidence > eligible[j].Confidence
	})
	if limit > len(eligible) {
		limit = len(eligible)
	}
	return eligible[:limit]
}

func routineRemediationPrompt(built *builtRoutineRun, candidate routineRemediationCandidate) string {
	evidence := strings.Join(candidate.TelemetryEvidence, "\n- ")
	if evidence != "" {
		evidence = "- " + evidence
	}
	files := strings.Join(candidate.SuspectedFiles, "\n- ")
	if files != "" {
		files = "- " + files
	}
	validation := strings.Join(candidate.Validation, "\n- ")
	if validation != "" {
		validation = "- " + validation
	}
	prompt := candidate.Prompt
	if strings.TrimSpace(prompt) == "" {
		prompt = candidate.Summary
	}
	return strings.TrimSpace(fmt.Sprintf(`# %s

This task was generated automatically by Boatman's routine %s (%s).

## Goal
%s

## Telemetry Evidence
%s

## Suspected Area
%s

## Expected Impact
%s

## Validation
%s

## Risk
%s

## Instructions
Use Boatman's normal work pipeline: plan the smallest safe remediation, work in the isolated worktree Boatman creates from latest main, implement the fix, run targeted validation, run peer-review and lydia-code-review, address actionable feedback, rerun relevant checks, and leave the PR as a draft for human review.
`, routineRemediationTitle(candidate, 1), built.routine.ID, built.request.RunID, prompt, emptySection(evidence), emptySection(files), emptySection(candidate.ExpectedImpact), emptySection(validation), emptySection(candidate.Risk)))
}

func appendRoutineRemediationSummary(report string, results []RoutineRemediationResult, note string) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(report))
	sb.WriteString("\n\n## Boatman Automatic Remediation\n\n")
	if note != "" {
		sb.WriteString(note)
		sb.WriteString("\n")
		return strings.TrimSpace(sb.String())
	}
	if len(results) == 0 {
		sb.WriteString("No remediation jobs were started.\n")
		return strings.TrimSpace(sb.String())
	}
	for _, result := range results {
		sb.WriteString("- ")
		sb.WriteString(result.Title)
		sb.WriteString(": ")
		sb.WriteString(result.Status)
		if result.PRURL != "" {
			sb.WriteString(" (")
			sb.WriteString(result.PRURL)
			sb.WriteString(")")
		}
		if result.Error != "" {
			sb.WriteString(" - ")
			sb.WriteString(result.Error)
		} else if result.Message != "" {
			sb.WriteString(" - ")
			sb.WriteString(result.Message)
		}
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}

func routineRemediationID(candidate routineRemediationCandidate, index int) string {
	if id := safeSlug(candidate.ID, 40); id != "" {
		return id
	}
	return fmt.Sprintf("candidate-%d", index)
}

func routineRemediationTitle(candidate routineRemediationCandidate, index int) string {
	if title := strings.TrimSpace(candidate.Title); title != "" {
		return title
	}
	return fmt.Sprintf("Routine remediation %d", index)
}

func routineRemediationBranchName(runID string, candidate routineRemediationCandidate, index int) string {
	slug := safeSlug(candidate.ID, 36)
	if slug == "" {
		slug = safeSlug(candidate.Title, 36)
	}
	if slug == "" {
		slug = fmt.Sprintf("candidate-%d", index)
	}
	run := safeSlug(runID, 24)
	if run == "" {
		run = time.Now().UTC().Format("20060102")
	}
	return "routine/" + run + "-" + slug
}

func safeSlug(value string, limit int) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastHyphen := false
	for _, r := range value {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if limit > 0 && len(out) > limit {
		out = strings.Trim(out[:limit], "-")
	}
	return out
}

func firstPRURL(output string) string {
	match := githubPRURLPattern.FindString(output)
	return strings.TrimRight(match, ".,")
}

func boatmanEventMap(event bmintegration.BoatmanEvent) map[string]interface{} {
	raw, err := json.Marshal(event)
	if err != nil {
		return map[string]interface{}{"type": event.Type}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]interface{}{"type": event.Type}
	}
	return out
}

func emptySection(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}
