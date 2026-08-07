// Package frontendqueue plans safe parallel execution for frontend Linear tickets.
package frontendqueue

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// TicketType describes the dominant implementation shape of a frontend ticket.
type TicketType string

const (
	TicketCopyOnly            TicketType = "copy-only"
	TicketStyleToken          TicketType = "style-token"
	TicketComponentRestyle    TicketType = "component-restyle"
	TicketPageRestyle         TicketType = "page-restyle"
	TicketInteractionBehavior TicketType = "interaction-behavior"
	TicketAmbiguousDesign     TicketType = "ambiguous-design"
)

// AutomationPolicy describes how far Boatman should take a ticket unattended.
type AutomationPolicy string

const (
	PolicyAutoPR   AutomationPolicy = "auto-pr"
	PolicyDraftPR  AutomationPolicy = "draft-pr"
	PolicyPlanOnly AutomationPolicy = "plan-only"
	PolicyBlocked  AutomationPolicy = "blocked"
)

// ValidationKind classifies one validation gate for a planned worker.
type ValidationKind string

const (
	ValidationStatic     ValidationKind = "static"
	ValidationUnit       ValidationKind = "unit"
	ValidationVisual     ValidationKind = "visual"
	ValidationDOM        ValidationKind = "dom"
	ValidationReview     ValidationKind = "review"
	ValidationA11y       ValidationKind = "accessibility"
	ValidationAcceptance ValidationKind = "acceptance"
)

// Ticket is a normalized frontend ticket from Linear or another issue tracker.
type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	BranchName  string    `json:"branchName,omitempty"`
	Status      string    `json:"status,omitempty"`
	StatusType  string    `json:"statusType,omitempty"`
	Team        string    `json:"team,omitempty"`
	Project     string    `json:"project,omitempty"`
	ParentID    string    `json:"parentId,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Estimate    *float64  `json:"estimate,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// PlanOptions control ticket classification and batching.
type PlanOptions struct {
	ProjectPath       string   `json:"projectPath,omitempty"`
	LinearProject     string   `json:"linearProject,omitempty"`
	MaxParallel       int      `json:"maxParallel,omitempty"`
	AllowedPolicies   []string `json:"allowedPolicies,omitempty"`
	IncludePlanOnly   bool     `json:"includePlanOnly,omitempty"`
	DefaultBaseBranch string   `json:"defaultBaseBranch,omitempty"`
}

// TargetHint is an inferred code, route, Figma, or component target.
type TargetHint struct {
	Kind       string  `json:"kind"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

// ValidationStep is one deterministic or evaluator-backed quality gate.
type ValidationStep struct {
	Kind        ValidationKind `json:"kind"`
	Command     string         `json:"command,omitempty"`
	Description string         `json:"description"`
	Required    bool           `json:"required"`
}

// TicketPlan is the orchestration plan for one ticket.
type TicketPlan struct {
	Ticket          Ticket           `json:"ticket"`
	Type            TicketType       `json:"type"`
	Policy          AutomationPolicy `json:"policy"`
	Confidence      float64          `json:"confidence"`
	RiskScore       int              `json:"riskScore"`
	TargetKey       string           `json:"targetKey"`
	TargetHints     []TargetHint     `json:"targetHints,omitempty"`
	FigmaRefs       []string         `json:"figmaRefs,omitempty"`
	Validation      []ValidationStep `json:"validation,omitempty"`
	Reasons         []string         `json:"reasons,omitempty"`
	Blockers        []string         `json:"blockers,omitempty"`
	WorkerPrompt    string           `json:"workerPrompt"`
	PassKCandidates int              `json:"passKCandidates,omitempty"`
}

// Batch is a set of tickets Boatman can attempt concurrently.
type Batch struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	TicketIDs   []string `json:"ticketIds"`
	TargetKeys  []string `json:"targetKeys"`
	Parallelism int      `json:"parallelism"`
	Rationale   string   `json:"rationale"`
}

// Stats summarizes a frontend queue plan.
type Stats struct {
	TotalTickets   int            `json:"totalTickets"`
	AutoPRCount    int            `json:"autoPrCount"`
	DraftPRCount   int            `json:"draftPrCount"`
	PlanOnlyCount  int            `json:"planOnlyCount"`
	BlockedCount   int            `json:"blockedCount"`
	BatchCount     int            `json:"batchCount"`
	MaxParallel    int            `json:"maxParallel"`
	ByType         map[string]int `json:"byType"`
	ByTarget       map[string]int `json:"byTarget"`
	EligibleCount  int            `json:"eligibleCount"`
	PassKCandidate int            `json:"passKCandidateCount"`
}

// Plan is the full deterministic queue plan.
type Plan struct {
	GeneratedAt time.Time    `json:"generatedAt"`
	Options     PlanOptions  `json:"options"`
	Stats       Stats        `json:"stats"`
	Tickets     []TicketPlan `json:"tickets"`
	Batches     []Batch      `json:"batches"`
}

var (
	figmaURLPattern = regexp.MustCompile(`https://www\.figma\.com/[^\s)>]+`)
	filePattern     = regexp.MustCompile(`(?:^|[\s` + "`" + `(\[])([A-Za-z0-9_./-]+/[A-Za-z0-9_./-]+\.(?:tsx|ts|jsx|js|css|scss|mdx|json))(?:$|[\s` + "`" + `)\],.;:])`)
)

// PlanTickets classifies frontend tickets and creates a conflict-aware batch plan.
func PlanTickets(tickets []Ticket, options PlanOptions) Plan {
	options = normalizeOptions(options)
	plans := make([]TicketPlan, 0, len(tickets))
	for _, ticket := range tickets {
		plans = append(plans, ClassifyTicket(ticket, options))
	}
	sort.SliceStable(plans, func(i, j int) bool {
		if policyRank(plans[i].Policy) != policyRank(plans[j].Policy) {
			return policyRank(plans[i].Policy) < policyRank(plans[j].Policy)
		}
		if plans[i].RiskScore != plans[j].RiskScore {
			return plans[i].RiskScore < plans[j].RiskScore
		}
		return strings.Compare(plans[i].Ticket.ID, plans[j].Ticket.ID) < 0
	})

	batches := buildBatches(plans, options)
	return Plan{
		GeneratedAt: time.Now().UTC(),
		Options:     options,
		Stats:       buildStats(plans, batches, options.MaxParallel),
		Tickets:     plans,
		Batches:     batches,
	}
}

// ClassifyTicket classifies one frontend ticket with deterministic rules.
func ClassifyTicket(ticket Ticket, options PlanOptions) TicketPlan {
	options = normalizeOptions(options)
	ticket = normalizeTicket(ticket)
	text := ticketText(ticket)
	figmaRefs := extractFigmaRefs(ticket.Description)
	fileRefs := extractFileRefs(ticket.Description)
	hints := inferTargetHints(ticket, text, fileRefs, figmaRefs)
	ticketType, typeReasons := inferTicketType(ticket, text, hints)
	riskScore, riskReasons := inferRiskScore(ticket, text, ticketType, len(fileRefs) > 0)
	policy, policyReasons, blockers := inferPolicy(ticket, text, ticketType, riskScore)
	confidence := inferConfidence(ticket, text, ticketType, policy, riskScore, len(fileRefs) > 0, len(figmaRefs) > 0)
	targetKey := inferTargetKey(ticket, hints)
	validation := validationSteps(ticketType, policy, len(figmaRefs) > 0)
	passK := 0
	if policy == PolicyAutoPR && (ticketType == TicketCopyOnly || ticketType == TicketStyleToken) {
		passK = 3
	}

	reasons := append([]string{}, typeReasons...)
	reasons = append(reasons, riskReasons...)
	reasons = append(reasons, policyReasons...)
	return TicketPlan{
		Ticket:          ticket,
		Type:            ticketType,
		Policy:          policy,
		Confidence:      confidence,
		RiskScore:       riskScore,
		TargetKey:       targetKey,
		TargetHints:     hints,
		FigmaRefs:       figmaRefs,
		Validation:      validation,
		Reasons:         uniqueStrings(reasons),
		Blockers:        blockers,
		WorkerPrompt:    workerPrompt(ticket, ticketType, policy, targetKey, hints, figmaRefs, validation, options),
		PassKCandidates: passK,
	}
}

func normalizeOptions(options PlanOptions) PlanOptions {
	if options.MaxParallel <= 0 {
		options.MaxParallel = 3
	}
	if strings.TrimSpace(options.DefaultBaseBranch) == "" {
		options.DefaultBaseBranch = "main"
	}
	return options
}

func normalizeTicket(ticket Ticket) Ticket {
	ticket.ID = strings.TrimSpace(ticket.ID)
	ticket.Title = strings.TrimSpace(ticket.Title)
	ticket.Description = strings.TrimSpace(ticket.Description)
	ticket.Status = strings.TrimSpace(ticket.Status)
	ticket.StatusType = strings.TrimSpace(ticket.StatusType)
	ticket.Team = strings.TrimSpace(ticket.Team)
	ticket.Project = strings.TrimSpace(ticket.Project)
	ticket.ParentID = strings.TrimSpace(ticket.ParentID)
	ticket.BranchName = strings.TrimSpace(ticket.BranchName)
	ticket.Labels = uniqueStrings(ticket.Labels)
	if ticket.ID == "" {
		ticket.ID = stableID(ticket.Title)
	}
	return ticket
}

func ticketText(ticket Ticket) string {
	parts := []string{ticket.ID, ticket.Title, ticket.Description, ticket.Team, ticket.Project, ticket.ParentID}
	parts = append(parts, ticket.Labels...)
	return strings.ToLower(strings.Join(parts, " "))
}

func inferTicketType(ticket Ticket, text string, hints []TargetHint) (TicketType, []string) {
	ambiguous := hasAny(text, "reconsider", "is that intended", "fully investigate", "unexpected behavior", "consider if", "modal or full", "full page", "paradigm")
	if ambiguous {
		return TicketAmbiguousDesign, []string{"Ticket asks for design/product judgment before implementation."}
	}
	if hasAny(text, "interaction", "takes user", "behavior", "persistent", "dismiss", "stack errors", "expand/collapse", "navigation") {
		return TicketInteractionBehavior, []string{"Ticket changes behavior or interaction semantics."}
	}
	if hasAny(text, "copy", "cta", "sentence-case", "sentence case", "rename", "label", "metadata", "text") &&
		!hasAny(text, "fully update", "all styling", "all filters") {
		return TicketCopyOnly, []string{"Ticket primarily changes copy or labels."}
	}
	if hasAny(text, "icon") && !hasAny(text, "all ", "fully ", "standardize", "across") {
		return TicketStyleToken, []string{"Ticket is a concrete icon-level visual change."}
	}
	if hasAny(text, "margin", "margins", "spacing", "font-size", "font size", "15pt", "14px", "12px", "color", "border", "border-less", "width") {
		if hasComponentHint(hints) {
			return TicketComponentRestyle, []string{"Ticket is a concrete visual restyle against a known component."}
		}
		return TicketStyleToken, []string{"Ticket is a concrete token, spacing, icon, border, or typography change."}
	}
	if hasAny(text, "rosetta", "filters", "menu", "table v2", "pagination", "empty state", "section divider", "badge", "sidesheet", "side sheet", "modal") {
		if hasComponentHint(hints) {
			return TicketComponentRestyle, []string{"Ticket maps to a known reusable component or pattern."}
		}
		return TicketStyleToken, []string{"Ticket is a Rosetta or design-token migration."}
	}
	if hasAny(text, "page", "view", "match design", "figma") {
		return TicketPageRestyle, []string{"Ticket targets a page or view with design reference."}
	}
	return TicketAmbiguousDesign, []string{"Ticket does not expose enough implementation shape for unattended work."}
}

func inferRiskScore(ticket Ticket, text string, ticketType TicketType, hasFileRefs bool) (int, []string) {
	score := 2
	reasons := []string{}
	switch ticketType {
	case TicketCopyOnly:
		score = 1
	case TicketStyleToken:
		score = 2
	case TicketComponentRestyle:
		score = 3
	case TicketPageRestyle:
		score = 4
	case TicketInteractionBehavior:
		score = 5
	case TicketAmbiguousDesign:
		score = 7
	}
	if hasFileRefs {
		score--
		reasons = append(reasons, "Ticket mentions concrete source files.")
	}
	if hasAny(text, "all ", "fully ", "standardize", "across", "nine pages", "shared component") {
		score += 2
		reasons = append(reasons, "Scope appears broad or shared.")
	}
	if hasAny(text, "required", "acceptance criteria", "- [ ]", "## scope") {
		score--
		reasons = append(reasons, "Ticket includes acceptance or scope details.")
	}
	if ticket.Estimate != nil && *ticket.Estimate <= 1 {
		score--
		reasons = append(reasons, "Estimate is small enough for automation.")
	}
	if score < 1 {
		score = 1
	}
	if score > 10 {
		score = 10
	}
	return score, reasons
}

func inferPolicy(ticket Ticket, text string, ticketType TicketType, riskScore int) (AutomationPolicy, []string, []string) {
	status := strings.ToLower(strings.TrimSpace(ticket.StatusType))
	if status == "completed" || status == "canceled" || status == "cancelled" {
		return PolicyBlocked, []string{"Ticket status is not runnable."}, []string{"Ticket is completed or canceled."}
	}
	if strings.Contains(text, "is that intended") {
		return PolicyPlanOnly, []string{"Ticket asks an explicit product question."}, []string{"Needs product confirmation."}
	}
	if ticketType == TicketAmbiguousDesign {
		return PolicyPlanOnly, []string{"Ticket needs a design decision before implementation."}, nil
	}
	if ticketType == TicketInteractionBehavior {
		return PolicyDraftPR, []string{"Interaction changes should be reviewed before ready-for-review."}, nil
	}
	if riskScore <= 2 && (ticketType == TicketCopyOnly || ticketType == TicketStyleToken) {
		return PolicyAutoPR, []string{"Low-risk frontend change is eligible for unattended draft PR creation."}, nil
	}
	if riskScore <= 5 {
		return PolicyDraftPR, []string{"Ticket is automatable but should stay draft until review artifacts are inspected."}, nil
	}
	return PolicyPlanOnly, []string{"Risk is too high for automatic implementation."}, nil
}

func inferConfidence(ticket Ticket, text string, ticketType TicketType, policy AutomationPolicy, riskScore int, hasFileRefs bool, hasFigma bool) float64 {
	confidence := 0.55
	switch ticketType {
	case TicketCopyOnly, TicketStyleToken:
		confidence += 0.18
	case TicketComponentRestyle:
		confidence += 0.12
	case TicketPageRestyle:
		confidence += 0.04
	case TicketInteractionBehavior:
		confidence -= 0.03
	case TicketAmbiguousDesign:
		confidence -= 0.18
	}
	if hasFileRefs {
		confidence += 0.1
	}
	if hasFigma {
		confidence += 0.06
	}
	if hasAny(text, "acceptance criteria", "## scope", "- [ ]") {
		confidence += 0.08
	}
	if policy == PolicyBlocked || policy == PolicyPlanOnly {
		confidence -= 0.08
	}
	confidence -= float64(max(0, riskScore-4)) * 0.04
	if confidence < 0.1 {
		confidence = 0.1
	}
	if confidence > 0.95 {
		confidence = 0.95
	}
	return round2(confidence)
}

func inferTargetHints(ticket Ticket, text string, fileRefs, figmaRefs []string) []TargetHint {
	hints := []TargetHint{}
	for _, file := range fileRefs {
		hints = append(hints, TargetHint{Kind: "file", Value: file, Confidence: 0.92})
	}
	componentKeywords := []struct {
		Needles []string
		Value   string
	}{
		{[]string{"table v2", "tablev2", "pagination", "empty state", "section divider", "grouped-row", "group row"}, "component:TableV2"},
		{[]string{"toast", "toast.error", "error toast"}, "component:Toast"},
		{[]string{"filter bar", "filters", "smart pills"}, "component:Filters"},
		{[]string{"menu", "columns menu"}, "component:Menu"},
		{[]string{"modal"}, "component:Modal"},
		{[]string{"side sheet", "sidesheet"}, "component:SideSheet"},
		{[]string{"json editor", "json text"}, "component:JSONEditor"},
		{[]string{"badge", "badges"}, "component:Badge"},
	}
	for _, item := range componentKeywords {
		if hasAny(text, item.Needles...) {
			hints = append(hints, TargetHint{Kind: "component", Value: item.Value, Confidence: 0.78})
		}
	}
	for _, figma := range figmaRefs {
		hints = append(hints, TargetHint{Kind: "figma", Value: figma, Confidence: 0.72})
	}
	if ticket.ParentID != "" {
		hints = append(hints, TargetHint{Kind: "parent", Value: ticket.ParentID, Confidence: 0.65})
	}
	for _, label := range ticket.Labels {
		if strings.TrimSpace(label) != "" {
			hints = append(hints, TargetHint{Kind: "label", Value: label, Confidence: 0.45})
		}
	}
	return dedupeHints(hints)
}

func inferTargetKey(ticket Ticket, hints []TargetHint) string {
	for _, hint := range hints {
		if hint.Kind == "file" || hint.Kind == "component" {
			return strings.ToLower(strings.TrimSpace(hint.Value))
		}
	}
	if ticket.ParentID != "" {
		return "parent:" + strings.ToLower(ticket.ParentID)
	}
	for _, hint := range hints {
		if hint.Kind == "label" {
			return "label:" + strings.ToLower(hint.Value)
		}
	}
	if ticket.Team != "" {
		return "team:" + strings.ToLower(ticket.Team)
	}
	return "unknown"
}

func validationSteps(ticketType TicketType, policy AutomationPolicy, hasFigma bool) []ValidationStep {
	if policy == PolicyBlocked {
		return nil
	}
	steps := []ValidationStep{
		{Kind: ValidationStatic, Command: "npm run build", Description: "Typecheck and build the frontend package.", Required: true},
	}
	switch ticketType {
	case TicketCopyOnly:
		steps = append(steps,
			ValidationDOMStep("Assert the updated copy/label/CTA appears in the targeted UI state."),
			ValidationVisualStep("Capture before/after screenshots for the affected route or story."),
		)
	case TicketStyleToken, TicketComponentRestyle, TicketPageRestyle:
		steps = append(steps,
			ValidationVisualStep("Capture before/after screenshots at desktop and narrow viewports."),
			ValidationStep{Kind: ValidationA11y, Description: "Run an accessibility smoke check for labels, contrast, and focus after the visual change.", Required: false},
		)
	case TicketInteractionBehavior:
		steps = append(steps,
			ValidationDOMStep("Exercise the changed interaction with Playwright and assert the visible state transition."),
			ValidationVisualStep("Capture screenshots before and after the interaction."),
		)
	default:
		steps = append(steps, ValidationAcceptanceStep("Produce a plan and screenshots, but stop before implementation until ambiguity is resolved."))
	}
	if hasFigma {
		steps = append(steps, ValidationAcceptanceStep("Compare the final screenshot against the linked Figma reference or Linear attachment."))
	}
	steps = append(steps, ValidationStep{Kind: ValidationReview, Description: "Run peer-review and lydia-code-review, then address actionable feedback.", Required: true})
	return steps
}

// ValidationDOMStep creates a DOM validation step.
func ValidationDOMStep(description string) ValidationStep {
	return ValidationStep{Kind: ValidationDOM, Command: "npx playwright test", Description: description, Required: true}
}

// ValidationVisualStep creates a visual validation step.
func ValidationVisualStep(description string) ValidationStep {
	return ValidationStep{Kind: ValidationVisual, Command: "npx playwright test --update-snapshots=false", Description: description, Required: true}
}

// ValidationAcceptanceStep creates an acceptance validation step.
func ValidationAcceptanceStep(description string) ValidationStep {
	return ValidationStep{Kind: ValidationAcceptance, Description: description, Required: true}
}

func buildBatches(plans []TicketPlan, options PlanOptions) []Batch {
	eligible := []TicketPlan{}
	for _, plan := range plans {
		if plan.Policy == PolicyAutoPR || plan.Policy == PolicyDraftPR || (options.IncludePlanOnly && plan.Policy == PolicyPlanOnly) {
			eligible = append(eligible, plan)
		}
	}
	queuesByTarget := map[string][]TicketPlan{}
	targets := []string{}
	for _, plan := range eligible {
		key := plan.TargetKey
		if key == "" {
			key = "unknown"
		}
		if _, ok := queuesByTarget[key]; !ok {
			targets = append(targets, key)
		}
		queuesByTarget[key] = append(queuesByTarget[key], plan)
	}
	sort.Strings(targets)
	for _, key := range targets {
		sort.SliceStable(queuesByTarget[key], func(i, j int) bool {
			if queuesByTarget[key][i].RiskScore != queuesByTarget[key][j].RiskScore {
				return queuesByTarget[key][i].RiskScore < queuesByTarget[key][j].RiskScore
			}
			return queuesByTarget[key][i].Ticket.ID < queuesByTarget[key][j].Ticket.ID
		})
	}

	batches := []Batch{}
	for {
		ticketIDs := []string{}
		targetKeys := []string{}
		for _, key := range targets {
			queue := queuesByTarget[key]
			if len(queue) == 0 {
				continue
			}
			plan := queue[0]
			queuesByTarget[key] = queue[1:]
			ticketIDs = append(ticketIDs, plan.Ticket.ID)
			targetKeys = append(targetKeys, key)
			if len(ticketIDs) >= options.MaxParallel {
				break
			}
		}
		if len(ticketIDs) == 0 {
			break
		}
		batchID := fmt.Sprintf("batch-%02d", len(batches)+1)
		batches = append(batches, Batch{
			ID:          batchID,
			Name:        fmt.Sprintf("Batch %d", len(batches)+1),
			TicketIDs:   ticketIDs,
			TargetKeys:  targetKeys,
			Parallelism: len(ticketIDs),
			Rationale:   "At most one ticket per inferred target key is included, so these workers can run in separate worktrees with low merge-conflict risk.",
		})
	}
	return batches
}

func buildStats(plans []TicketPlan, batches []Batch, maxParallel int) Stats {
	stats := Stats{
		TotalTickets: len(plans),
		BatchCount:   len(batches),
		MaxParallel:  maxParallel,
		ByType:       map[string]int{},
		ByTarget:     map[string]int{},
	}
	for _, plan := range plans {
		stats.ByType[string(plan.Type)]++
		stats.ByTarget[plan.TargetKey]++
		if plan.PassKCandidates > 0 {
			stats.PassKCandidate++
		}
		switch plan.Policy {
		case PolicyAutoPR:
			stats.AutoPRCount++
			stats.EligibleCount++
		case PolicyDraftPR:
			stats.DraftPRCount++
			stats.EligibleCount++
		case PolicyPlanOnly:
			stats.PlanOnlyCount++
		case PolicyBlocked:
			stats.BlockedCount++
		}
	}
	return stats
}

func workerPrompt(ticket Ticket, ticketType TicketType, policy AutomationPolicy, targetKey string, hints []TargetHint, figmaRefs []string, validation []ValidationStep, options PlanOptions) string {
	lines := []string{
		fmt.Sprintf("Implement frontend ticket %s: %s", ticket.ID, ticket.Title),
		"",
		fmt.Sprintf("Automation policy: %s", policy),
		fmt.Sprintf("Ticket type: %s", ticketType),
		fmt.Sprintf("Inferred target key: %s", targetKey),
		fmt.Sprintf("Base branch: %s", options.DefaultBaseBranch),
		"",
		"Linear context:",
		strings.TrimSpace(ticket.Description),
	}
	if ticket.URL != "" {
		lines = append(lines, "", "Linear URL: "+ticket.URL)
	}
	if len(figmaRefs) > 0 {
		lines = append(lines, "", "Figma references:")
		for _, ref := range figmaRefs {
			lines = append(lines, "- "+ref)
		}
	}
	if len(hints) > 0 {
		lines = append(lines, "", "Target hints:")
		for _, hint := range hints {
			lines = append(lines, fmt.Sprintf("- %s: %s", hint.Kind, hint.Value))
		}
	}
	lines = append(lines,
		"",
		"Execution instructions:",
		"- Start from the latest base branch in a fresh git worktree.",
		"- Make the smallest code change that satisfies the ticket.",
		"- Prefer existing design-system components and tokens over bespoke CSS.",
		"- Capture before/after screenshots for every targeted route or story.",
		"- If the ticket needs product/design judgment, stop with a plan and do not invent the decision.",
		"- Open a draft PR with Linear/Figma links, screenshots, validation output, and residual risk.",
		"",
		"Validation gates:",
	)
	for _, step := range validation {
		line := "- " + step.Description
		if step.Command != "" {
			line += " (`" + step.Command + "`)"
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func extractFigmaRefs(text string) []string {
	return uniqueStrings(figmaURLPattern.FindAllString(text, -1))
}

func extractFileRefs(text string) []string {
	matches := filePattern.FindAllStringSubmatch(text, -1)
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			refs = append(refs, strings.TrimSpace(match[1]))
		}
	}
	return uniqueStrings(refs)
}

func hasComponentHint(hints []TargetHint) bool {
	for _, hint := range hints {
		if hint.Kind == "component" || hint.Kind == "file" {
			return true
		}
	}
	return false
}

func hasAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func policyRank(policy AutomationPolicy) int {
	switch policy {
	case PolicyAutoPR:
		return 0
	case PolicyDraftPR:
		return 1
	case PolicyPlanOnly:
		return 2
	case PolicyBlocked:
		return 3
	default:
		return 4
	}
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func dedupeHints(hints []TargetHint) []TargetHint {
	seen := map[string]bool{}
	out := []TargetHint{}
	for _, hint := range hints {
		if strings.TrimSpace(hint.Value) == "" {
			continue
		}
		key := strings.ToLower(hint.Kind + ":" + hint.Value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, hint)
	}
	return out
}

func stableID(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	title = strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-").Replace(title)
	title = strings.Trim(title, "-")
	if title == "" {
		return "ticket"
	}
	return title
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
