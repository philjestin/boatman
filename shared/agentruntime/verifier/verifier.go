// Package verifier defines independent verification contracts for agent output.
package verifier

import (
	"context"
	"regexp"
	"strings"
	"time"
)

// Status is the independent verifier verdict status.
type Status string

const (
	StatusPass  Status = "pass"
	StatusFail  Status = "fail"
	StatusError Status = "error"
)

// Severity describes finding severity.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Request is the verifier input for a completed or proposed agent change.
type Request struct {
	RunID        string            `json:"runId,omitempty"`
	WorkDir      string            `json:"workDir,omitempty"`
	Diff         string            `json:"diff,omitempty"`
	ChangedFiles []string          `json:"changedFiles,omitempty"`
	Checks       []Check           `json:"checks,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Check describes an external check expected to run around a verifier stage.
// Implementations can either execute it or report that it was delegated.
type Check struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name"`
	Command  string   `json:"command,omitempty"`
	Args     []string `json:"args,omitempty"`
	Required bool     `json:"required,omitempty"`
}

// Finding is one verifier issue.
type Finding struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Path     string   `json:"path,omitempty"`
	Line     int      `json:"line,omitempty"`
}

// Artifact is durable verifier evidence.
type Artifact struct {
	Kind    string `json:"kind"`
	Path    string `json:"path,omitempty"`
	URL     string `json:"url,omitempty"`
	Message string `json:"message,omitempty"`
}

// Verdict is the verifier output.
type Verdict struct {
	Status      Status     `json:"status"`
	Summary     string     `json:"summary,omitempty"`
	Findings    []Finding  `json:"findings,omitempty"`
	Artifacts   []Artifact `json:"artifacts,omitempty"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt time.Time  `json:"completedAt"`
}

// Verifier independently evaluates agent output.
type Verifier interface {
	Verify(ctx context.Context, req Request) (Verdict, error)
}

// PolicyVerifier performs local policy checks over diffs and file paths. It
// does not execute code; it provides the first independent verifier stage.
type PolicyVerifier struct {
	secretPatterns []*regexp.Regexp
}

// NewPolicyVerifier creates a verifier with Boatman's default policy checks.
func NewPolicyVerifier() *PolicyVerifier {
	return &PolicyVerifier{
		secretPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)AKIA[0-9A-Z]{12,}`),
			regexp.MustCompile(`(?i)(openai|anthropic|github|slack|datadog)[_-]?api[_-]?key\s*[:=]\s*['"]?[A-Za-z0-9_\-]{12,}`),
			regexp.MustCompile(`(?i)-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`),
			regexp.MustCompile(`ghp_[A-Za-z0-9_]{20,}`),
		},
	}
}

// Verify checks a diff and changed files for high-risk patterns.
func (v *PolicyVerifier) Verify(ctx context.Context, req Request) (Verdict, error) {
	started := time.Now().UTC()
	if err := ctx.Err(); err != nil {
		return Verdict{Status: StatusError, StartedAt: started, CompletedAt: time.Now().UTC()}, err
	}
	var findings []Finding
	for _, pattern := range v.secretPatterns {
		if pattern.MatchString(req.Diff) {
			findings = append(findings, Finding{
				Severity: SeverityCritical,
				Code:     "secret-like-diff",
				Message:  "diff appears to contain credential or private key material",
			})
			break
		}
	}
	lowerDiff := strings.ToLower(req.Diff)
	if strings.Contains(lowerDiff, "drop table") || strings.Contains(lowerDiff, "drop database") {
		findings = append(findings, Finding{
			Severity: SeverityCritical,
			Code:     "destructive-database-change",
			Message:  "diff contains destructive database operation text",
		})
	}
	for _, path := range req.ChangedFiles {
		lowerPath := strings.ToLower(path)
		switch {
		case containsAny(lowerPath, "db/migrate", "migration", "schema.sql"):
			findings = append(findings, Finding{
				Severity: SeverityWarning,
				Code:     "migration-changed",
				Message:  "database migration or schema file changed",
				Path:     path,
			})
		case containsAny(lowerPath, "auth", "oauth", "session", "password", "permission", "billing", "payment", "stripe"):
			findings = append(findings, Finding{
				Severity: SeverityWarning,
				Code:     "sensitive-domain-changed",
				Message:  "auth, permission, payment, or billing path changed",
				Path:     path,
			})
		}
	}
	status := StatusPass
	for _, finding := range findings {
		if finding.Severity == SeverityError || finding.Severity == SeverityCritical {
			status = StatusFail
			break
		}
	}
	summary := "independent policy checks passed"
	if status == StatusFail {
		summary = "independent policy checks found blocking issues"
	} else if len(findings) > 0 {
		summary = "independent policy checks passed with warnings"
	}
	return Verdict{
		Status:      status,
		Summary:     summary,
		Findings:    findings,
		StartedAt:   started,
		CompletedAt: time.Now().UTC(),
	}, nil
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
