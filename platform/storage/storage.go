// Package storage defines the persistence contract for the Boatman platform.
// All services depend on these interfaces; backends (SQLite, PostgreSQL, etc.)
// implement them.
package storage

import (
	"context"
	"time"
)

// Store is the top-level storage interface grouping all sub-stores.
type Store interface {
	Runs() RunStore
	Memory() MemoryStore
	Costs() CostStore
	Policies() PolicyStore
	Events() EventStore
	Migrate(ctx context.Context) error
	Close() error
}

// Scope identifies the organizational context for any stored entity.
// The hierarchy is Org -> Team -> Repo. Empty fields mean "all" at that level.
type Scope struct {
	OrgID  string `json:"org_id"`
	TeamID string `json:"team_id"`
	RepoID string `json:"repo_id"`
}

// --- Run types ---

// RunStatus represents the lifecycle state of an agent run.
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusPassed    RunStatus = "passed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCanceled  RunStatus = "canceled"
	RunStatusError     RunStatus = "error"
)

// Run records a single agent run.
type Run struct {
	ID           string    `json:"id"`
	Scope        Scope     `json:"scope"`
	UserID       string    `json:"user_id"`
	Status       RunStatus `json:"status"`
	Prompt       string    `json:"prompt"`
	TotalCostUSD float64   `json:"total_cost_usd"`
	Iterations   int       `json:"iterations"`
	FilesChanged []string  `json:"files_changed"`
	Duration     time.Duration `json:"duration"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RunFilter selects runs by various criteria.
type RunFilter struct {
	Scope    *Scope
	UserID   string
	Status   RunStatus
	Since    time.Time
	Until    time.Time
	Limit    int
	Offset   int
}

// RunStore persists agent run records.
type RunStore interface {
	Create(ctx context.Context, run *Run) error
	Get(ctx context.Context, id string) (*Run, error)
	Update(ctx context.Context, run *Run) error
	List(ctx context.Context, filter RunFilter) ([]*Run, error)
}

// --- Memory types ---

// Pattern represents a learned code pattern with organizational scope.
type Pattern struct {
	ID          string    `json:"id"`
	Scope       Scope     `json:"scope"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Example     string    `json:"example,omitempty"`
	FileMatcher string    `json:"file_matcher,omitempty"`
	Weight      float64   `json:"weight"`
	UsageCount  int       `json:"usage_count"`
	SuccessRate float64   `json:"success_rate"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Preferences stores learned organizational preferences.
type Preferences struct {
	ID                     string            `json:"id"`
	Scope                  Scope             `json:"scope"`
	PreferredTestFramework string            `json:"preferred_test_framework"`
	NamingConventions      map[string]string `json:"naming_conventions"`
	FileOrganization       map[string]string `json:"file_organization"`
	CodeStyle              map[string]string `json:"code_style"`
	CommitMessageFormat    string            `json:"commit_message_format"`
	ReviewerThresholds     map[string]int    `json:"reviewer_thresholds"`
	UpdatedAt              time.Time         `json:"updated_at"`
}

// CommonIssue represents a frequently encountered issue with scope.
type CommonIssue struct {
	ID          string    `json:"id"`
	Scope       Scope     `json:"scope"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Solution    string    `json:"solution"`
	Frequency   int       `json:"frequency"`
	AutoFix     bool      `json:"auto_fix"`
	FileMatcher string    `json:"file_matcher,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemoryStore persists patterns, preferences, and common issues.
type MemoryStore interface {
	// Patterns
	CreatePattern(ctx context.Context, p *Pattern) error
	UpdatePattern(ctx context.Context, p *Pattern) error
	ListPatterns(ctx context.Context, scope Scope) ([]*Pattern, error)
	DeletePattern(ctx context.Context, id string) error

	// Preferences
	GetPreferences(ctx context.Context, scope Scope) (*Preferences, error)
	SetPreferences(ctx context.Context, prefs *Preferences) error

	// Common issues
	CreateIssue(ctx context.Context, issue *CommonIssue) error
	UpdateIssue(ctx context.Context, issue *CommonIssue) error
	ListIssues(ctx context.Context, scope Scope) ([]*CommonIssue, error)
}

// --- Cost types ---

// UsageRecord records token usage for a single step within a run.
type UsageRecord struct {
	ID               string    `json:"id"`
	RunID            string    `json:"run_id"`
	Scope            Scope     `json:"scope"`
	Step             string    `json:"step"`
	InputTokens      int       `json:"input_tokens"`
	OutputTokens     int       `json:"output_tokens"`
	CacheReadTokens  int       `json:"cache_read_tokens"`
	CacheWriteTokens int       `json:"cache_write_tokens"`
	TotalCostUSD     float64   `json:"total_cost_usd"`
	RecordedAt       time.Time `json:"recorded_at"`
}

// UsageFilter selects usage records by criteria.
type UsageFilter struct {
	Scope *Scope
	RunID string
	Since time.Time
	Until time.Time
}

// TimeGroup specifies how to group usage summaries.
type TimeGroup string

const (
	TimeGroupHour  TimeGroup = "hour"
	TimeGroupDay   TimeGroup = "day"
	TimeGroupWeek  TimeGroup = "week"
	TimeGroupMonth TimeGroup = "month"
)

// UsageSummary aggregates usage over a time period.
type UsageSummary struct {
	Period       time.Time `json:"period"`
	TotalRuns    int       `json:"total_runs"`
	TotalCostUSD float64  `json:"total_cost_usd"`
	InputTokens  int      `json:"input_tokens"`
	OutputTokens int      `json:"output_tokens"`
}

// Budget defines spending limits for a scope.
type Budget struct {
	ID           string    `json:"id"`
	Scope        Scope     `json:"scope"`
	MonthlyLimit float64   `json:"monthly_limit"`
	DailyLimit   float64   `json:"daily_limit"`
	PerRunLimit  float64   `json:"per_run_limit"`
	AlertAt      float64   `json:"alert_at"` // percentage (0-1) at which to alert
	UpdatedAt    time.Time `json:"updated_at"`
}

// CostStore persists usage records and budgets.
type CostStore interface {
	RecordUsage(ctx context.Context, record *UsageRecord) error
	GetUsage(ctx context.Context, filter UsageFilter) ([]*UsageRecord, error)
	GetUsageSummary(ctx context.Context, scope Scope, group TimeGroup, since, until time.Time) ([]*UsageSummary, error)
	GetBudget(ctx context.Context, scope Scope) (*Budget, error)
	SetBudget(ctx context.Context, budget *Budget) error
}

// --- Policy types ---

// Policy defines enforcement rules for a scope.
type Policy struct {
	ID               string   `json:"id"`
	Scope            Scope    `json:"scope"`
	MaxIterations    int      `json:"max_iterations,omitempty"`
	MaxCostPerRun    float64  `json:"max_cost_per_run,omitempty"`
	MaxFilesChanged  int      `json:"max_files_changed,omitempty"`
	AllowedModels    []string `json:"allowed_models,omitempty"`
	BlockedPatterns  []string `json:"blocked_patterns,omitempty"`
	RequireTests     bool     `json:"require_tests"`
	RequireReview    bool     `json:"require_review"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PolicyStore persists and retrieves policies.
type PolicyStore interface {
	Get(ctx context.Context, scope Scope) (*Policy, error)
	Set(ctx context.Context, policy *Policy) error
	Delete(ctx context.Context, scope Scope) error
	GetEffectivePolicy(ctx context.Context, scope Scope) (*Policy, error)
}

// --- Event types ---

// Event represents something that happened during a run.
type Event struct {
	ID        string         `json:"id"`
	RunID     string         `json:"run_id,omitempty"`
	Scope     Scope          `json:"scope"`
	Type      string         `json:"type"`
	Name      string         `json:"name,omitempty"`
	Message   string         `json:"message,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
}

// EventFilter selects events by criteria.
type EventFilter struct {
	Scope *Scope
	RunID string
	Types []string
	Since time.Time
	Until time.Time
	Limit int
}

// EventStore persists and queries events.
type EventStore interface {
	Publish(ctx context.Context, event *Event) error
	Query(ctx context.Context, filter EventFilter) ([]*Event, error)
}
