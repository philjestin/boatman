// Package workflows defines provider-neutral workflow templates for Boatman's
// central agent plane.
package workflows

import (
	"fmt"
	"sort"
	"strings"
)

// StageKind identifies a workflow stage by responsibility instead of by a
// specific CLI or provider implementation.
type StageKind string

const (
	StageIntake         StageKind = "intake"
	StageContext        StageKind = "context"
	StagePlanning       StageKind = "planning"
	StageImplementation StageKind = "implementation"
	StageValidation     StageKind = "validation"
	StagePullRequest    StageKind = "pull_request"
	StageSynthesis      StageKind = "synthesis"
)

// GatePolicy describes how a stage should be approved before it can run or
// publish side effects.
type GatePolicy string

const (
	GateNone    GatePolicy = "none"
	GateDynamic GatePolicy = "dynamic"
	GateHuman   GatePolicy = "human"
)

// Template is a durable workflow definition. It is intentionally data-first so
// CLI, desktop, and a future control-plane service can share the same library.
type Template struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Stages      []Stage `json:"stages"`
}

// Stage is one node in a workflow template. Next and OnFailure are explicit so
// loops and skips stay visible instead of being hidden in free-form code.
type Stage struct {
	ID           string     `json:"id"`
	Kind         StageKind  `json:"kind"`
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	Optional     bool       `json:"optional,omitempty"`
	Gate         GatePolicy `json:"gate"`
	Capabilities []string   `json:"capabilities,omitempty"`
	Preview      bool       `json:"preview,omitempty"`
	Next         []string   `json:"next,omitempty"`
	OnFailure    []string   `json:"onFailure,omitempty"`
}

// Library is an immutable-ish collection of templates keyed by ID.
type Library struct {
	templates map[string]Template
}

// DefaultLibrary returns Boatman's built-in workflow templates.
func DefaultLibrary() Library {
	return NewLibrary(DefaultTemplates())
}

// NewLibrary builds a template library from templates. Later duplicates replace
// earlier entries so callers can override built-ins intentionally.
func NewLibrary(templates []Template) Library {
	values := make(map[string]Template, len(templates))
	for _, template := range templates {
		values[template.ID] = cloneTemplate(template)
	}
	return Library{templates: values}
}

// Get returns a template by ID.
func (l Library) Get(id string) (Template, bool) {
	template, ok := l.templates[strings.TrimSpace(id)]
	if !ok {
		return Template{}, false
	}
	return cloneTemplate(template), true
}

// List returns templates sorted by ID.
func (l Library) List() []Template {
	templates := make([]Template, 0, len(l.templates))
	for _, template := range l.templates {
		templates = append(templates, cloneTemplate(template))
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].ID < templates[j].ID
	})
	return templates
}

// DefaultTemplates returns the core local templates used to model the central
// agent-plane workflows without requiring a service runtime.
func DefaultTemplates() []Template {
	return []Template{
		{
			ID:          "bugfix",
			Name:        "Bug Fix",
			Description: "Diagnose, patch, validate, and publish a focused defect fix.",
			Stages: []Stage{
				stage("intake", StageIntake, "Intake", "Normalize the ticket, reproduction steps, severity, and stop conditions.", GateNone, false, "context"),
				stage("context", StageContext, "Context", "Load repo, memory, linked issues, logs, and ownership hints.", GateNone, true, "planning"),
				stage("planning", StagePlanning, "Plan", "Choose the smallest fix plan and required checks.", GateDynamic, true, "implementation"),
				stage("implementation", StageImplementation, "Implement", "Apply the defect fix in an isolated worktree.", GateDynamic, false, "validation"),
				failureStage("validation", StageValidation, "Validate", "Run tests, static checks, and independent diff verification.", GateNone, false, []string{"pull-request"}, []string{"implementation"}),
				stage("pull-request", StagePullRequest, "Pull Request", "Create or update the reviewable PR artifact.", GateHuman, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Summarize outcome, evidence, residual risks, and follow-ups.", GateNone, true),
			},
		},
		{
			ID:          "code-review",
			Name:        "Code Review",
			Description: "Inspect a diff independently and produce actionable findings.",
			Stages: []Stage{
				stage("intake", StageIntake, "Intake", "Normalize the PR, author intent, and changed files.", GateNone, false, "context"),
				stage("context", StageContext, "Context", "Load ownership, memory, prior comments, and repo conventions.", GateNone, true, "validation"),
				stage("validation", StageValidation, "Review", "Run verifier checks and model review over the diff.", GateNone, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Publish findings, confidence, and test gaps.", GateNone, true),
			},
		},
		{
			ID:          "feature",
			Name:        "Feature",
			Description: "Plan, implement, validate, and PR a product feature with review gates.",
			Stages: []Stage{
				stage("intake", StageIntake, "Intake", "Normalize user request, acceptance criteria, and product constraints.", GateNone, false, "context"),
				stage("context", StageContext, "Context", "Load repo, memory, linked docs, prior work, and integration state.", GateNone, true, "planning"),
				stage("planning", StagePlanning, "Plan", "Generate an inspectable implementation plan and validation strategy.", GateHuman, true, "implementation"),
				stage("implementation", StageImplementation, "Implement", "Edit code and docs in an isolated worktree.", GateDynamic, false, "validation"),
				failureStage("validation", StageValidation, "Validate", "Run tests, docs checks, verifier checks, and review loops.", GateNone, false, []string{"pull-request"}, []string{"implementation"}),
				stage("pull-request", StagePullRequest, "Pull Request", "Create or update the draft PR and attach evidence.", GateHuman, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Summarize shipped behavior, checks, artifacts, and next risks.", GateNone, true),
			},
		},
		{
			ID:          "firefighter",
			Name:        "Firefighter",
			Description: "Investigate incidents with preview-first remediation and explicit gates.",
			Stages: []Stage{
				stage("intake", StageIntake, "Intake", "Normalize incident, customer impact, timelines, and severity.", GateNone, false, "context"),
				stage("context", StageContext, "Context", "Gather telemetry, deploy history, tickets, ownership, and runbooks.", GateNone, true, "planning"),
				stage("planning", StagePlanning, "Plan", "Choose investigation branches and safe remediation previews.", GateDynamic, true, "implementation"),
				stage("implementation", StageImplementation, "Remediate", "Prepare a fix, rollback, or mitigation without silent production writes.", GateHuman, true, "validation"),
				failureStage("validation", StageValidation, "Validate", "Verify blast radius, tests, dashboards, and rollback criteria.", GateDynamic, true, []string{"pull-request", "synthesis"}, []string{"planning"}),
				stage("pull-request", StagePullRequest, "Pull Request", "Publish the fix or mitigation artifact when appropriate.", GateHuman, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Write incident findings, confidence, and handoff notes.", GateNone, true),
			},
		},
		{
			ID:          "research",
			Name:        "Research",
			Description: "Answer a question with provenance and no code side effects.",
			Stages: []Stage{
				stage("intake", StageIntake, "Intake", "Normalize the question, constraints, and desired evidence.", GateNone, false, "context"),
				stage("context", StageContext, "Context", "Gather docs, memory, code references, and external facts.", GateNone, true, "planning"),
				stage("planning", StagePlanning, "Analysis", "Compare options and cite sources or repository evidence.", GateNone, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Return answer, assumptions, and remaining uncertainty.", GateNone, true),
			},
		},
		{
			ID:          "triage",
			Name:        "Triage",
			Description: "Score, classify, cluster, and optionally plan backlog work.",
			Stages: []Stage{
				stage("intake", StageIntake, "Fetch", "Fetch tickets and normalize source metadata.", GateNone, false, "context"),
				stage("context", StageContext, "Ingest", "Extract code, domain, dependency, and validation signals.", GateNone, false, "planning"),
				stage("planning", StagePlanning, "Score And Classify", "Score AI-readiness and classify deterministic gates.", GateNone, true, "validation"),
				stage("validation", StageValidation, "Cluster And Plan", "Cluster related tickets and validate generated plans.", GateNone, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Synthesis", "Publish backlog summary, plans, and stop conditions.", GateNone, true),
			},
		},
		{
			ID:          "frontend-ticket-orchestration",
			Name:        "Frontend Ticket Orchestration",
			Description: "Classify frontend tickets, plan conflict-aware parallel workers, validate visually, and publish draft PR evidence.",
			Stages: []Stage{
				stage("intake", StageIntake, "Linear Intake", "Fetch frontend tickets, linked Figma refs, labels, status, estimate, and parent relationships.", GateNone, false, "context"),
				stage("context", StageContext, "Frontend Context", "Map tickets to routes, stories, components, design tokens, and learned Boatman Brain hints.", GateNone, true, "planning"),
				stage("planning", StagePlanning, "Queue Planning", "Classify automation policy and build conflict-aware parallel batches.", GateDynamic, true, "implementation"),
				failureStage("implementation", StageImplementation, "Parallel Workers", "Run eligible tickets in isolated worktrees with browser sandboxes and Pass@K where useful.", GateDynamic, false, []string{"validation"}, []string{"planning"}),
				failureStage("validation", StageValidation, "Visual Validation", "Run static checks, DOM assertions, screenshots, visual diffs, accessibility smoke checks, and review skills.", GateNone, false, []string{"pull-request"}, []string{"implementation"}),
				stage("pull-request", StagePullRequest, "Draft PRs", "Create draft PRs with Linear/Figma links, screenshots, validation output, and residual risk.", GateDynamic, true, "synthesis"),
				terminalStage("synthesis", StageSynthesis, "Queue Summary", "Summarize shipped PRs, blocked tickets, confidence, and learned mappings.", GateNone, true),
			},
		},
	}
}

// Validate checks a workflow template for shape errors.
func Validate(template Template) error {
	if strings.TrimSpace(template.ID) == "" {
		return fmt.Errorf("template ID is required")
	}
	if strings.TrimSpace(template.Name) == "" {
		return fmt.Errorf("template %q name is required", template.ID)
	}
	if len(template.Stages) == 0 {
		return fmt.Errorf("template %q must have at least one stage", template.ID)
	}
	seen := make(map[string]Stage, len(template.Stages))
	for _, stage := range template.Stages {
		if strings.TrimSpace(stage.ID) == "" {
			return fmt.Errorf("template %q has a stage with no ID", template.ID)
		}
		if _, ok := seen[stage.ID]; ok {
			return fmt.Errorf("template %q has duplicate stage %q", template.ID, stage.ID)
		}
		if !validKind(stage.Kind) {
			return fmt.Errorf("template %q stage %q has unknown kind %q", template.ID, stage.ID, stage.Kind)
		}
		if strings.TrimSpace(stage.Name) == "" {
			return fmt.Errorf("template %q stage %q name is required", template.ID, stage.ID)
		}
		if !validGate(stage.Gate) {
			return fmt.Errorf("template %q stage %q has unknown gate %q", template.ID, stage.ID, stage.Gate)
		}
		seen[stage.ID] = stage
	}
	terminalCount := 0
	for _, stage := range template.Stages {
		if len(stage.Next) == 0 {
			terminalCount++
		}
		for _, next := range append(append([]string{}, stage.Next...), stage.OnFailure...) {
			if _, ok := seen[next]; !ok {
				return fmt.Errorf("template %q stage %q references unknown stage %q", template.ID, stage.ID, next)
			}
		}
	}
	if terminalCount == 0 {
		return fmt.Errorf("template %q must have a terminal stage", template.ID)
	}
	return nil
}

func stage(id string, kind StageKind, name, description string, gate GatePolicy, preview bool, next ...string) Stage {
	return Stage{
		ID:          id,
		Kind:        kind,
		Name:        name,
		Description: description,
		Gate:        gate,
		Preview:     preview,
		Next:        append([]string(nil), next...),
	}
}

func failureStage(id string, kind StageKind, name, description string, gate GatePolicy, preview bool, next []string, onFailure []string) Stage {
	value := stage(id, kind, name, description, gate, preview, next...)
	value.OnFailure = append([]string(nil), onFailure...)
	return value
}

func terminalStage(id string, kind StageKind, name, description string, gate GatePolicy, preview bool) Stage {
	return stage(id, kind, name, description, gate, preview)
}

func validKind(kind StageKind) bool {
	switch kind {
	case StageIntake, StageContext, StagePlanning, StageImplementation, StageValidation, StagePullRequest, StageSynthesis:
		return true
	default:
		return false
	}
}

func validGate(gate GatePolicy) bool {
	switch gate {
	case GateNone, GateDynamic, GateHuman:
		return true
	default:
		return false
	}
}

func cloneTemplate(template Template) Template {
	template.Stages = append([]Stage(nil), template.Stages...)
	for i := range template.Stages {
		template.Stages[i].Capabilities = append([]string(nil), template.Stages[i].Capabilities...)
		template.Stages[i].Next = append([]string(nil), template.Stages[i].Next...)
		template.Stages[i].OnFailure = append([]string(nil), template.Stages[i].OnFailure...)
	}
	return template
}
