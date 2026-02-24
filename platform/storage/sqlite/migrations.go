package sqlite

import (
	"context"
	"database/sql"
)

// migrations contains ordered DDL statements. Each is idempotent (IF NOT EXISTS).
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS runs (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		user_id TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		prompt TEXT NOT NULL DEFAULT '',
		total_cost_usd REAL NOT NULL DEFAULT 0,
		iterations INTEGER NOT NULL DEFAULT 0,
		files_changed TEXT NOT NULL DEFAULT '[]',
		duration_ns INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_runs_scope ON runs(org_id, team_id, repo_id)`,
	`CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status)`,
	`CREATE INDEX IF NOT EXISTS idx_runs_created ON runs(created_at)`,

	`CREATE TABLE IF NOT EXISTS patterns (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		example TEXT NOT NULL DEFAULT '',
		file_matcher TEXT NOT NULL DEFAULT '',
		weight REAL NOT NULL DEFAULT 0,
		usage_count INTEGER NOT NULL DEFAULT 0,
		success_rate REAL NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_patterns_scope ON patterns(org_id, team_id, repo_id)`,

	`CREATE TABLE IF NOT EXISTS preferences (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		data TEXT NOT NULL DEFAULT '{}',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE UNIQUE INDEX IF NOT EXISTS idx_preferences_scope ON preferences(org_id, team_id, repo_id)`,

	`CREATE TABLE IF NOT EXISTS common_issues (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		solution TEXT NOT NULL DEFAULT '',
		frequency INTEGER NOT NULL DEFAULT 0,
		auto_fix INTEGER NOT NULL DEFAULT 0,
		file_matcher TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_issues_scope ON common_issues(org_id, team_id, repo_id)`,

	`CREATE TABLE IF NOT EXISTS usage_records (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL DEFAULT '',
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		step TEXT NOT NULL DEFAULT '',
		input_tokens INTEGER NOT NULL DEFAULT 0,
		output_tokens INTEGER NOT NULL DEFAULT 0,
		cache_read_tokens INTEGER NOT NULL DEFAULT 0,
		cache_write_tokens INTEGER NOT NULL DEFAULT 0,
		total_cost_usd REAL NOT NULL DEFAULT 0,
		recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_usage_run ON usage_records(run_id)`,
	`CREATE INDEX IF NOT EXISTS idx_usage_scope ON usage_records(org_id, team_id, repo_id)`,
	`CREATE INDEX IF NOT EXISTS idx_usage_recorded ON usage_records(recorded_at)`,

	`CREATE TABLE IF NOT EXISTS budgets (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		monthly_limit REAL NOT NULL DEFAULT 0,
		daily_limit REAL NOT NULL DEFAULT 0,
		per_run_limit REAL NOT NULL DEFAULT 0,
		alert_at REAL NOT NULL DEFAULT 0.8,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE UNIQUE INDEX IF NOT EXISTS idx_budgets_scope ON budgets(org_id, team_id, repo_id)`,

	`CREATE TABLE IF NOT EXISTS policies (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		max_iterations INTEGER NOT NULL DEFAULT 0,
		max_cost_per_run REAL NOT NULL DEFAULT 0,
		max_files_changed INTEGER NOT NULL DEFAULT 0,
		allowed_models TEXT NOT NULL DEFAULT '[]',
		blocked_patterns TEXT NOT NULL DEFAULT '[]',
		require_tests INTEGER NOT NULL DEFAULT 0,
		require_review INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE UNIQUE INDEX IF NOT EXISTS idx_policies_scope ON policies(org_id, team_id, repo_id)`,

	`CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		run_id TEXT NOT NULL DEFAULT '',
		org_id TEXT NOT NULL DEFAULT '',
		team_id TEXT NOT NULL DEFAULT '',
		repo_id TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL DEFAULT '',
		message TEXT NOT NULL DEFAULT '',
		data TEXT NOT NULL DEFAULT '{}',
		version INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE INDEX IF NOT EXISTS idx_events_run ON events(run_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_scope ON events(org_id, team_id, repo_id)`,
	`CREATE INDEX IF NOT EXISTS idx_events_type ON events(type)`,
	`CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at)`,
}

// runMigrations executes all DDL statements within a transaction.
func runMigrations(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, ddl := range migrations {
		if _, err := tx.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}

	return tx.Commit()
}
