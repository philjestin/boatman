package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateData is the data passed to every template.
type TemplateData struct {
	// ModulePath is the Go module path for the generated project.
	ModulePath string

	// PackageName is the Go package name (last segment of module path).
	PackageName string

	// HarnessImport is the import path for the harness module.
	HarnessImport string

	// Provider is the selected LLM provider.
	Provider LLMProvider

	// ProviderMeta contains provider-specific hints and names.
	ProviderMeta ProviderMeta

	// ProjectLang is the target project language.
	ProjectLang ProjectLanguage

	// LangTestCommand is the suggested test command for the language.
	LangTestCommand string

	// IncludePlanner indicates whether to generate Planner code.
	IncludePlanner bool

	// IncludeTester indicates whether to generate Tester code.
	IncludeTester bool

	// IncludeCostTracking indicates whether to wire cost tracking.
	IncludeCostTracking bool

	// MaxIterations is the default max review iterations.
	MaxIterations int

	// ReviewCriteria is an optional reviewer focus description.
	ReviewCriteria string
}

// templateEntry maps a template file to its output filename.
type templateEntry struct {
	// tmplName is the filename inside templates/ (e.g. "main.go.tmpl").
	tmplName string
	// outName is the output filename (e.g. "main.go").
	outName string
	// optional means this template is only rendered when the condition is true.
	optional bool
	// condition returns true if this optional template should be rendered.
	condition func(data *TemplateData) bool
}

// templateRegistry defines which templates to render and their output names.
var templateRegistry = []templateEntry{
	{tmplName: "go.mod.tmpl", outName: "go.mod"},
	{tmplName: "main.go.tmpl", outName: "main.go"},
	{tmplName: "developer.go.tmpl", outName: "developer.go"},
	{tmplName: "reviewer.go.tmpl", outName: "reviewer.go"},
	{tmplName: "config.go.tmpl", outName: "config.go"},
	{
		tmplName:  "planner.go.tmpl",
		outName:   "planner.go",
		optional:  true,
		condition: func(d *TemplateData) bool { return d.IncludePlanner },
	},
	{
		tmplName:  "tester.go.tmpl",
		outName:   "tester.go",
		optional:  true,
		condition: func(d *TemplateData) bool { return d.IncludeTester },
	},
}

// newTemplateData builds template data from a scaffold config.
func newTemplateData(cfg *ScaffoldConfig) *TemplateData {
	parts := strings.Split(cfg.ProjectName, "/")
	pkgName := parts[len(parts)-1]

	return &TemplateData{
		ModulePath:          cfg.ProjectName,
		PackageName:         pkgName,
		HarnessImport:       "github.com/philjestin/boatman-ecosystem/harness",
		Provider:            cfg.Provider,
		ProviderMeta:        getProviderMeta(cfg.Provider),
		ProjectLang:         cfg.ProjectLang,
		LangTestCommand:     LangTestCommand(cfg.ProjectLang),
		IncludePlanner:      cfg.IncludePlanner,
		IncludeTester:       cfg.IncludeTester,
		IncludeCostTracking: cfg.IncludeCostTracking,
		MaxIterations:       cfg.MaxIterations,
		ReviewCriteria:      cfg.ReviewCriteria,
	}
}

// renderTemplate renders a single template by name and returns the output.
func renderTemplate(name string, data *TemplateData) ([]byte, error) {
	tmplPath := path.Join("templates", name)

	content, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}
