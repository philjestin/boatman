// Package config handles configuration management for boatman.
package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const DefaultRuntimeProvider = "claude-cli"

// Config holds all configuration for the boatman agent.
type Config struct {
	// Linear API
	LinearKey string

	// Workflow settings
	MaxIterations int
	BaseBranch    string
	AutoPR        bool
	ReviewSkill   string
	// ExtraReviewSkills are additional Claude reviewer skills run after the
	// primary review skill; all blocking feedback is merged into the refactor loop.
	ExtraReviewSkills []string
	// KeepDraftPR leaves the final pull request in draft state after successful
	// validation/review. The draft checkpoint is still created either way.
	KeepDraftPR bool

	// Review pass criteria
	Review ReviewConfig

	// Coordinator settings
	Coordinator CoordinatorConfig

	// Retry settings
	Retry RetryConfig

	// Claude settings
	Claude ClaudeConfig

	// Runtime settings
	Runtime RuntimeConfig

	// Token budgets
	TokenBudget TokenBudgetConfig

	// Brain settings
	Brain BrainConfig

	// Triage settings
	Triage TriageConfig

	// Debug enables verbose logging
	Debug bool

	// EnableTools enables Claude CLI tool capabilities for agents
	EnableTools bool
}

// RuntimeConfig holds provider routing settings.
type RuntimeConfig struct {
	// DefaultProvider is the provider used when no role or profile override matches.
	DefaultProvider string

	// RoleProviders maps runtime roles such as planner, executor, reviewer, or scorer
	// to provider IDs.
	RoleProviders map[string]string

	// ProfileProviders maps narrower workflow profiles such as triage-planner or
	// refactor to provider IDs. Profile matches take precedence over role matches.
	ProfileProviders map[string]string
}

// ProviderFor returns the configured provider for a workflow role/profile.
func (r RuntimeConfig) ProviderFor(role, profile string) string {
	if provider := lookupProvider(r.ProfileProviders, profile); provider != "" {
		return provider
	}
	if provider := lookupProvider(r.RoleProviders, role); provider != "" {
		return provider
	}
	if provider := strings.TrimSpace(r.DefaultProvider); provider != "" {
		return provider
	}
	return DefaultRuntimeProvider
}

// BrainConfig holds brain domain knowledge settings.
type BrainConfig struct {
	// Enabled controls whether brain injection is active.
	Enabled bool

	// Directories is an optional list of additional brain directories.
	Directories []string

	// MaxBrains is the maximum number of brains to inject per task.
	MaxBrains int

	// TokenBudget is the token budget for brain content in prompts.
	TokenBudget int
}

// ReviewConfig holds review pass criteria settings.
type ReviewConfig struct {
	// MaxCriticalIssues is the maximum number of critical issues allowed to pass.
	MaxCriticalIssues int

	// MaxMajorIssues is the maximum number of major issues allowed to pass.
	MaxMajorIssues int

	// MinVerificationConfidence is the minimum confidence percentage (0-100) for diff verification.
	MinVerificationConfidence int

	// StrictParsing enables strict keyword matching in natural language review parsing.
	StrictParsing bool
}

// CoordinatorConfig holds coordinator-specific settings.
type CoordinatorConfig struct {
	// MessageBufferSize is the size of the main message channel buffer.
	MessageBufferSize int

	// SubscriberBufferSize is the size of per-subscriber channel buffers.
	SubscriberBufferSize int
}

// RetryConfig holds retry behavior settings.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int

	// InitialDelay is the initial delay before first retry.
	InitialDelay time.Duration

	// MaxDelay caps the maximum delay between retries.
	MaxDelay time.Duration
}

// ClaudeConfig holds Claude CLI settings.
type ClaudeConfig struct {
	// Command is the claude command to use.
	Command string

	// UseTmux enables tmux for large prompts.
	UseTmux bool

	// LargePromptThreshold is the character count above which to use tmux.
	LargePromptThreshold int

	// Timeout for Claude operations (0 = no timeout).
	Timeout time.Duration

	// Model configuration per agent type
	Models ModelConfig

	// EnablePromptCaching enables prompt caching for cost reduction.
	// Note: Requires Claude CLI version that supports --cache-system-prompt flag.
	// Set to true only if your CLI version supports it.
	EnablePromptCaching bool

	// Effort sets the reasoning effort level for all agents ("low", "medium", "high").
	// Empty = CLI default. Use "high" with Opus 4.6 for maximum reasoning.
	Effort string
}

// ModelConfig holds model selection per agent type.
// Leave empty to use the Claude CLI's default model.
// Model names vary by provider (Anthropic API vs Vertex AI vs AWS Bedrock).
type ModelConfig struct {
	// Planner model for planning phase (empty = CLI default)
	Planner string

	// Executor model for code generation (empty = CLI default)
	Executor string

	// Reviewer model for code review (empty = CLI default)
	Reviewer string

	// Refactor model for fixing issues (empty = CLI default)
	Refactor string

	// Preflight model for validation (empty = CLI default)
	Preflight string

	// TestRunner model for test output parsing (empty = CLI default)
	TestRunner string

	// Scorer model for triage rubric scoring (empty = CLI default)
	Scorer string
}

// TriageConfig holds triage pipeline settings.
type TriageConfig struct {
	// StalenessHours is the TTL for normalized ticket records (default: 168 = 7 days).
	StalenessHours int

	// DefaultTeams are the default team keys for batch fetch when --teams is not specified.
	DefaultTeams []string

	// OutputDir is where decision logs and context docs are stored.
	OutputDir string

	// MaxConcurrency is the maximum number of concurrent Claude scoring calls.
	MaxConcurrency int

	// PostComments controls whether to post rubric breakdown to Linear by default.
	PostComments bool
}

// TokenBudgetConfig holds context token budget settings.
type TokenBudgetConfig struct {
	// Context is the token budget for context in prompts.
	Context int

	// Plan is the token budget for planning information.
	Plan int

	// Review is the token budget for review feedback.
	Review int
}

// Load reads configuration from viper and environment variables.
func Load() (*Config, error) {
	cfg := loadValues()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadRuntime reads provider/runtime configuration without requiring external
// service credentials. Use it for local inspection commands that do not call
// Linear or model APIs.
func LoadRuntime() (*Config, error) {
	cfg := loadValues()
	cfg.ensureRuntimeDefaults()
	return cfg, nil
}

func loadValues() *Config {
	return &Config{
		LinearKey:     getEnvOrViper("LINEAR_API_KEY", "linear_key"),
		MaxIterations: getIntOrDefault("max_iterations", 5), // Increased from 3 to 5
		BaseBranch:    getStringOrDefault("base_branch", "main"),
		AutoPR:        viper.GetBool("auto_pr"),
		ReviewSkill:   getStringOrDefault("review_skill", "peer-review"),
		ExtraReviewSkills: sanitizeStringSlice(
			append(viper.GetStringSlice("extra_review_skills"), splitCSV(os.Getenv("BOATMAN_EXTRA_REVIEW_SKILLS"))...),
		),
		KeepDraftPR: getBoolOrDefault("keep_draft_pr", false),
		Debug:       os.Getenv("BOATMAN_DEBUG") == "1",
		EnableTools: getBoolOrDefault("enable_tools", true),

		Review: ReviewConfig{
			MaxCriticalIssues:         getIntOrDefault("review.max_critical_issues", 1),          // Allow 1 critical (was 0)
			MaxMajorIssues:            getIntOrDefault("review.max_major_issues", 3),             // Allow 3 major (was 2)
			MinVerificationConfidence: getIntOrDefault("review.min_verification_confidence", 50), // 50% confidence threshold
			StrictParsing:             getBoolOrDefault("review.strict_parsing", false),          // Relaxed by default
		},

		Coordinator: CoordinatorConfig{
			MessageBufferSize:    getIntOrDefault("coordinator.message_buffer_size", 1000),
			SubscriberBufferSize: getIntOrDefault("coordinator.subscriber_buffer_size", 100),
		},

		Retry: RetryConfig{
			MaxAttempts:  getIntOrDefault("retry.max_attempts", 3),
			InitialDelay: getDurationOrDefault("retry.initial_delay", 500*time.Millisecond),
			MaxDelay:     getDurationOrDefault("retry.max_delay", 30*time.Second),
		},

		Claude: ClaudeConfig{
			Command:              getStringOrDefault("claude.command", "claude"),
			UseTmux:              viper.GetBool("claude.use_tmux"),
			LargePromptThreshold: getIntOrDefault("claude.large_prompt_threshold", 100000),
			Timeout:              getDurationOrDefault("claude.timeout", 0),
			EnablePromptCaching:  getBoolOrDefault("claude.enable_prompt_caching", false),
			Effort:               getStringOrDefault("claude.effort", ""),
			Models: ModelConfig{
				Planner:    getStringOrDefault("claude.models.planner", ""),     // Empty = use CLI default
				Executor:   getStringOrDefault("claude.models.executor", ""),    // Empty = use CLI default
				Reviewer:   getStringOrDefault("claude.models.reviewer", ""),    // Empty = use CLI default
				Refactor:   getStringOrDefault("claude.models.refactor", ""),    // Empty = use CLI default
				Preflight:  getStringOrDefault("claude.models.preflight", ""),   // Empty = use CLI default
				TestRunner: getStringOrDefault("claude.models.test_runner", ""), // Empty = use CLI default
				Scorer:     getStringOrDefault("claude.models.scorer", ""),      // Empty = use CLI default
			},
		},

		Runtime: RuntimeConfig{
			DefaultProvider:  getEnvOrViper("BOATMAN_PROVIDER", "runtime.default_provider"),
			RoleProviders:    normalizeProviderMap(viper.GetStringMapString("runtime.role_providers")),
			ProfileProviders: normalizeProviderMap(viper.GetStringMapString("runtime.profile_providers")),
		},

		TokenBudget: TokenBudgetConfig{
			Context: getIntOrDefault("token_budget.context", 8000),
			Plan:    getIntOrDefault("token_budget.plan", 2000),
			Review:  getIntOrDefault("token_budget.review", 4000),
		},

		Brain: BrainConfig{
			Enabled:     getBoolOrDefault("brain.enabled", true),
			MaxBrains:   getIntOrDefault("brain.max_brains", 3),
			TokenBudget: getIntOrDefault("brain.token_budget", 2000),
		},

		Triage: TriageConfig{
			StalenessHours: getIntOrDefault("triage.staleness_hours", 168),
			DefaultTeams:   viper.GetStringSlice("triage.default_teams"),
			OutputDir:      getStringOrDefault("triage.output_dir", ".boatman-triage"),
			MaxConcurrency: getIntOrDefault("triage.max_concurrency", 3),
			PostComments:   getBoolOrDefault("triage.post_comments", false),
		},
	}
}

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	if c.LinearKey == "" {
		return errors.New("linear API key is required (set LINEAR_API_KEY or --linear-key)")
	}
	c.ensureRuntimeDefaults()
	return nil
}

func (c *Config) ensureRuntimeDefaults() {
	if strings.TrimSpace(c.Runtime.DefaultProvider) == "" {
		c.Runtime.DefaultProvider = DefaultRuntimeProvider
	}
}

func lookupProvider(providers map[string]string, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if provider := strings.TrimSpace(providers[key]); provider != "" {
		return provider
	}
	return ""
}

func normalizeProviderMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sanitizeStringSlice(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}

// getEnvOrViper returns the value from environment variable or viper config.
func getEnvOrViper(envKey, viperKey string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return viper.GetString(viperKey)
}

// getIntOrDefault returns viper int value or default if not set.
func getIntOrDefault(key string, defaultVal int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	return defaultVal
}

// getStringOrDefault returns viper string value or default if not set.
func getStringOrDefault(key string, defaultVal string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultVal
}

// getDurationOrDefault returns viper duration value or default if not set.
func getDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	if viper.IsSet(key) {
		return viper.GetDuration(key)
	}
	return defaultVal
}

// getBoolOrDefault returns viper bool value or default if not set.
func getBoolOrDefault(key string, defaultVal bool) bool {
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}
	return defaultVal
}
