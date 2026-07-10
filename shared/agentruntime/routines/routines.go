// Package routines defines repeatable provider-neutral Boatman runs.
package routines

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	agentruntime "github.com/philjestin/boatman-ecosystem/shared/agentruntime"
)

const DatadogGQLSlowQueriesID = "datadog-gql-slow-queries"

// ParameterType describes the expected value shape for a routine parameter.
type ParameterType string

const (
	ParameterString   ParameterType = "string"
	ParameterInteger  ParameterType = "integer"
	ParameterDuration ParameterType = "duration"
)

// Parameter is one user-supplied routine input.
type Parameter struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Description string        `json:"description,omitempty"`
	Default     string        `json:"default,omitempty"`
	Required    bool          `json:"required,omitempty"`
}

// Output describes where and how the routine should produce durable output.
type Output struct {
	Format      string `json:"format"`
	DefaultPath string `json:"defaultPath,omitempty"`
}

// Routine is a saved, repeatable run definition.
type Routine struct {
	ID               string            `json:"id"`
	Extends          string            `json:"extends,omitempty"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Schedule         string            `json:"schedule,omitempty"`
	WorkflowTemplate string            `json:"workflowTemplate,omitempty"`
	Role             agentruntime.Role `json:"role"`
	Profile          string            `json:"profile"`
	Integrations     []string          `json:"integrations,omitempty"`
	Parameters       []Parameter       `json:"parameters,omitempty"`
	Defaults         map[string]string `json:"defaults,omitempty"`
	Output           Output            `json:"output"`
	Instructions     string            `json:"instructions,omitempty"`
	PromptTemplate   string            `json:"promptTemplate,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Library stores routines by ID.
type Library struct {
	routines map[string]Routine
}

// DefaultLibrary returns Boatman's built-in routines.
func DefaultLibrary() Library {
	return NewLibrary(DefaultRoutines())
}

// NewLibrary builds a routine library.
func NewLibrary(items []Routine) Library {
	values := make(map[string]Routine, len(items))
	for _, item := range items {
		item = Normalize(item)
		values[item.ID] = cloneRoutine(item)
	}
	return Library{routines: values}
}

// Get returns a routine by ID.
func (l Library) Get(id string) (Routine, bool) {
	item, ok := l.routines[strings.TrimSpace(id)]
	if !ok {
		return Routine{}, false
	}
	return cloneRoutine(item), true
}

// List returns routines sorted by ID.
func (l Library) List() []Routine {
	items := make([]Routine, 0, len(l.routines))
	for _, item := range l.routines {
		items = append(items, cloneRoutine(item))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

// DefaultRoutines returns built-in repeatable runs.
func DefaultRoutines() []Routine {
	return []Routine{
		{
			ID:               DatadogGQLSlowQueriesID,
			Name:             "Datadog GraphQL Slow Queries",
			Description:      "Investigate the slowest GraphQL operations for a graph area using Datadog MCP evidence.",
			Schedule:         "0 8 * * *",
			WorkflowTemplate: "research",
			Role:             agentruntime.RoleRoutine,
			Profile:          "routine.datadog-gql-slow-queries",
			Integrations:     []string{"datadog"},
			Parameters: []Parameter{
				{Name: "graph_area", Type: ParameterString, Description: "Graph or product area to inspect, such as employer, candidate, search, or jobs.", Required: true},
				{Name: "top_n", Type: ParameterInteger, Description: "Number of slow operations to inspect.", Default: "20"},
				{Name: "lookback", Type: ParameterDuration, Description: "Datadog lookback window.", Default: "24h"},
				{Name: "environment", Type: ParameterString, Description: "Datadog environment tag.", Default: "prod"},
				{Name: "service", Type: ParameterString, Description: "Optional service tag to restrict the query."},
			},
			Output: Output{
				Format:      "markdown",
				DefaultPath: ".boatman/routines/datadog-gql-slow-queries",
			},
			Instructions: strings.TrimSpace(`
You are Boatman running a repeatable performance investigation routine.

Use the Datadog MCP integration as the source of truth for telemetry. Do not make code changes, create tickets, or write to external systems. Read only observability data and produce a Markdown report that another engineer can act on.

Be explicit about confidence. If telemetry is missing, stale, sampled, or ambiguous, say so and list the Datadog query or span evidence you attempted to inspect.
`),
			PromptTemplate: strings.TrimSpace(`
Investigate the top {{top_n}} slowest GraphQL operations for graph area "{{graph_area}}" over the last {{lookback}} in environment "{{environment}}".

{{service_filter}}

Use Datadog MCP to inspect APM traces, GraphQL operation names, span/resource names, latency percentiles, request volume, error rate, downstream DB/cache/external spans, and recent deploy or regression signals. Prefer p95/p99 plus volume over one-off maximum latency. When possible, include Datadog links or exact query dimensions.

Return a Markdown report with these sections:

1. Executive summary
2. Scope and Datadog query assumptions
3. Top slow GraphQL operations table
   - operation/resource
   - p95/p99 or comparable latency
   - volume
   - error rate
   - dominant slow spans
   - first-seen or regression signal
4. Clustered likely causes
5. Suggested fixes
   - code or schema area to inspect
   - expected impact
   - validation query/check
6. Follow-up questions and missing telemetry

Do not invent numbers. If Datadog does not expose a metric, mark it as unavailable and explain the closest evidence you found.
`),
			Metadata: map[string]string{
				"category": "observability",
				"cadence":  "daily",
			},
		},
	}
}

// ProjectLibrary returns built-in routines plus routines defined under workDir.
// Project routines override built-ins by ID and may extend any routine loaded
// before them.
func ProjectLibrary(workDir string) (Library, error) {
	library := DefaultLibrary()
	projectRoutines, err := LoadProjectRoutines(workDir, library)
	if err != nil {
		return Library{}, err
	}
	for _, routine := range projectRoutines {
		library.routines[routine.ID] = cloneRoutine(Normalize(routine))
	}
	return library, nil
}

// LoadProjectRoutines loads routine definitions from .boatman/routines.json
// and .boatman/routines/*.json under workDir.
func LoadProjectRoutines(workDir string, base Library) ([]Routine, error) {
	paths, err := projectRoutinePaths(workDir)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}
	library := base
	out := []Routine{}
	for _, path := range paths {
		items, err := readRoutineFile(path)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			resolved, err := resolveRoutine(item, library)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			if err := Validate(resolved); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			library.routines[resolved.ID] = cloneRoutine(resolved)
			out = append(out, cloneRoutine(resolved))
		}
	}
	return out, nil
}

// Normalize applies routine defaults that make project routine files concise.
func Normalize(routine Routine) Routine {
	routine.ID = strings.TrimSpace(routine.ID)
	routine.Extends = strings.TrimSpace(routine.Extends)
	if routine.Role == "" {
		routine.Role = agentruntime.RoleRoutine
	}
	if strings.TrimSpace(routine.Profile) == "" && routine.ID != "" {
		routine.Profile = "routine." + routine.ID
	}
	if strings.TrimSpace(routine.Output.Format) == "" {
		routine.Output.Format = "markdown"
	}
	if strings.TrimSpace(routine.Output.DefaultPath) == "" && routine.ID != "" {
		routine.Output.DefaultPath = filepath.Join(".boatman", "routines", routine.ID)
	}
	return routine
}

// Values resolves user inputs against defaults and validates required fields.
func Values(routine Routine, overrides map[string]string) (map[string]string, error) {
	routine = Normalize(routine)
	values := make(map[string]string, len(routine.Parameters))
	for _, param := range routine.Parameters {
		value := strings.TrimSpace(param.Default)
		if routine.Defaults != nil {
			if defaultValue, ok := routine.Defaults[param.Name]; ok {
				value = strings.TrimSpace(defaultValue)
			}
		}
		if overrides != nil {
			if override, ok := overrides[param.Name]; ok {
				value = strings.TrimSpace(override)
			}
		}
		if param.Required && value == "" {
			return nil, fmt.Errorf("routine %q requires parameter %q", routine.ID, param.Name)
		}
		if value != "" {
			if err := validateParameter(param, value); err != nil {
				return nil, err
			}
			values[param.Name] = value
		}
	}
	for key, value := range overrides {
		if _, ok := values[key]; !ok && strings.TrimSpace(value) != "" {
			values[key] = strings.TrimSpace(value)
		}
	}
	return values, nil
}

// BuildOptions customizes the runtime request generated for a routine run.
type BuildOptions struct {
	RunID      string
	WorkDir    string
	Provider   string
	Model      string
	MCPServers []agentruntime.MCPServerRef
	Metadata   map[string]string
}

// BuildRequest turns a routine definition and values into a provider-neutral
// runtime request.
func BuildRequest(routine Routine, values map[string]string, opts BuildOptions) (agentruntime.RunRequest, error) {
	routine = Normalize(routine)
	if err := Validate(routine); err != nil {
		return agentruntime.RunRequest{}, err
	}
	if strings.TrimSpace(opts.RunID) == "" {
		opts.RunID = NewRunID(routine.ID, time.Now())
	}
	prompt, err := RenderPrompt(routine, values)
	if err != nil {
		return agentruntime.RunRequest{}, err
	}
	metadata := cloneStringMap(routine.Metadata)
	for key, value := range opts.Metadata {
		metadata[key] = value
	}
	metadata["routineId"] = routine.ID
	metadata["phaseId"] = "routine"
	return agentruntime.RunRequest{
		RunID:          opts.RunID,
		Role:           routine.Role,
		Profile:        routine.Profile,
		Provider:       opts.Provider,
		Model:          opts.Model,
		WorkDir:        opts.WorkDir,
		Instructions:   routine.Instructions,
		Messages:       []agentruntime.Message{{Role: "user", Content: prompt}},
		MCPServers:     cloneMCPRefs(opts.MCPServers),
		ApprovalPolicy: agentruntime.ApprovalSuggest,
		Reasoning:      &agentruntime.ReasoningOptions{Effort: "high"},
		Metadata:       metadata,
	}, nil
}

// RenderPrompt fills a routine prompt with validated values.
func RenderPrompt(routine Routine, values map[string]string) (string, error) {
	routine = Normalize(routine)
	resolved, err := Values(routine, values)
	if err != nil {
		return "", err
	}
	prompt := routine.PromptTemplate
	for key, value := range resolved {
		prompt = strings.ReplaceAll(prompt, "{{"+key+"}}", value)
	}
	serviceFilter := "No service filter was provided; infer the relevant GraphQL service from Datadog tags and state the assumption."
	if service := strings.TrimSpace(resolved["service"]); service != "" {
		serviceFilter = fmt.Sprintf("Restrict the investigation to Datadog service %q when that tag is available.", service)
	}
	prompt = strings.ReplaceAll(prompt, "{{service_filter}}", serviceFilter)
	return strings.TrimSpace(prompt), nil
}

// Validate checks routine shape.
func Validate(routine Routine) error {
	routine = Normalize(routine)
	if strings.TrimSpace(routine.ID) == "" {
		return fmt.Errorf("routine ID is required")
	}
	if strings.TrimSpace(routine.Name) == "" {
		return fmt.Errorf("routine %q name is required", routine.ID)
	}
	if routine.Role == "" {
		return fmt.Errorf("routine %q role is required", routine.ID)
	}
	if strings.TrimSpace(routine.Profile) == "" {
		return fmt.Errorf("routine %q profile is required", routine.ID)
	}
	if strings.TrimSpace(routine.PromptTemplate) == "" {
		return fmt.Errorf("routine %q prompt template is required", routine.ID)
	}
	seen := map[string]bool{}
	parameters := map[string]Parameter{}
	for _, param := range routine.Parameters {
		name := strings.TrimSpace(param.Name)
		if name == "" {
			return fmt.Errorf("routine %q has a parameter with no name", routine.ID)
		}
		if seen[name] {
			return fmt.Errorf("routine %q has duplicate parameter %q", routine.ID, name)
		}
		seen[name] = true
		parameters[name] = param
		if param.Type == "" {
			return fmt.Errorf("routine %q parameter %q type is required", routine.ID, name)
		}
		if param.Default != "" {
			if err := validateParameter(param, param.Default); err != nil {
				return err
			}
		}
	}
	for name, value := range routine.Defaults {
		param, ok := parameters[name]
		if !ok {
			return fmt.Errorf("routine %q has a default for unknown parameter %q", routine.ID, name)
		}
		if strings.TrimSpace(value) != "" {
			if err := validateParameter(param, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewRunID returns a stable-ish run ID prefix for a routine execution.
func NewRunID(routineID string, now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	return fmt.Sprintf("%s-%s", strings.TrimSpace(routineID), now.UTC().Format("20060102T150405Z"))
}

func validateParameter(param Parameter, value string) error {
	switch param.Type {
	case ParameterString:
		return nil
	case ParameterInteger:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("parameter %q must be an integer", param.Name)
		}
	case ParameterDuration:
		if _, err := parseDuration(value); err != nil {
			return fmt.Errorf("parameter %q must be a duration like 24h or 7d: %w", param.Name, err)
		}
	default:
		return fmt.Errorf("parameter %q has unknown type %q", param.Name, param.Type)
	}
	return nil
}

func projectRoutinePaths(workDir string) ([]string, error) {
	if strings.TrimSpace(workDir) == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, err
	}
	root := filepath.Join(workDir, ".boatman")
	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	seen := map[string]bool{}
	var paths []string
	add := func(path string) {
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	file := filepath.Join(root, "routines.json")
	if info, err := os.Stat(file); err == nil && !info.IsDir() {
		add(file)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	matches, err := filepath.Glob(filepath.Join(root, "routines", "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	for _, path := range matches {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			add(path)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return paths, nil
}

func readRoutineFile(path string) ([]Routine, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Routines []Routine `json:"routines"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Routines != nil {
		return wrapper.Routines, nil
	}
	var list []Routine
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var item Routine
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("failed to parse routine JSON: %w", err)
	}
	return []Routine{item}, nil
}

func resolveRoutine(item Routine, library Library) (Routine, error) {
	item = Normalize(item)
	if item.Extends == "" {
		return item, nil
	}
	base, ok := library.Get(item.Extends)
	if !ok {
		return Routine{}, fmt.Errorf("routine %q extends unknown routine %q", item.ID, item.Extends)
	}
	return mergeRoutine(base, item), nil
}

func mergeRoutine(base Routine, override Routine) Routine {
	base = Normalize(base)
	override = Normalize(override)
	out := cloneRoutine(base)
	out.Extends = override.Extends
	if override.ID != "" {
		out.ID = override.ID
	}
	if override.Name != "" {
		out.Name = override.Name
	}
	if override.Description != "" {
		out.Description = override.Description
	}
	if override.Schedule != "" {
		out.Schedule = override.Schedule
	}
	if override.WorkflowTemplate != "" {
		out.WorkflowTemplate = override.WorkflowTemplate
	}
	if override.Role != "" {
		out.Role = override.Role
	}
	if override.Profile != "" {
		out.Profile = override.Profile
	}
	if override.Integrations != nil {
		out.Integrations = append([]string(nil), override.Integrations...)
	}
	if override.Parameters != nil {
		out.Parameters = mergeParameters(out.Parameters, override.Parameters)
	}
	out.Defaults = mergeStringMaps(out.Defaults, override.Defaults)
	if override.Output.Format != "" {
		out.Output.Format = override.Output.Format
	}
	if override.Output.DefaultPath != "" {
		out.Output.DefaultPath = override.Output.DefaultPath
	}
	if override.Instructions != "" {
		out.Instructions = override.Instructions
	}
	if override.PromptTemplate != "" {
		out.PromptTemplate = override.PromptTemplate
	}
	out.Metadata = mergeStringMaps(out.Metadata, override.Metadata)
	return Normalize(out)
}

func mergeParameters(base []Parameter, overrides []Parameter) []Parameter {
	out := append([]Parameter(nil), base...)
	indexByName := map[string]int{}
	for i, param := range out {
		indexByName[param.Name] = i
	}
	for _, override := range overrides {
		if idx, ok := indexByName[override.Name]; ok {
			out[idx] = mergeParameter(out[idx], override)
			continue
		}
		indexByName[override.Name] = len(out)
		out = append(out, override)
	}
	return out
}

func mergeParameter(base Parameter, override Parameter) Parameter {
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Type != "" {
		base.Type = override.Type
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.Default != "" {
		base.Default = override.Default
	}
	if override.Required {
		base.Required = true
	}
	return base
}

func parseDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func cloneRoutine(routine Routine) Routine {
	routine.Integrations = append([]string(nil), routine.Integrations...)
	routine.Parameters = append([]Parameter(nil), routine.Parameters...)
	routine.Defaults = cloneStringMap(routine.Defaults)
	routine.Metadata = cloneStringMap(routine.Metadata)
	return routine
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func mergeStringMaps(base, override map[string]string) map[string]string {
	out := cloneStringMap(base)
	if out == nil && len(override) > 0 {
		out = map[string]string{}
	}
	for key, value := range override {
		out[key] = value
	}
	return out
}

func cloneMCPRefs(refs []agentruntime.MCPServerRef) []agentruntime.MCPServerRef {
	out := make([]agentruntime.MCPServerRef, len(refs))
	for i, ref := range refs {
		out[i] = ref
		out[i].Args = append([]string(nil), ref.Args...)
		out[i].Env = cloneStringMap(ref.Env)
	}
	return out
}
