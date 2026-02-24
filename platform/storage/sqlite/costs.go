package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type costStore struct {
	db *sql.DB
}

func (s *costStore) RecordUsage(ctx context.Context, record *storage.UsageRecord) error {
	if record.RecordedAt.IsZero() {
		record.RecordedAt = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO usage_records (id, run_id, org_id, team_id, repo_id, step,
			input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
			total_cost_usd, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.RunID, record.Scope.OrgID, record.Scope.TeamID, record.Scope.RepoID,
		record.Step, record.InputTokens, record.OutputTokens,
		record.CacheReadTokens, record.CacheWriteTokens,
		record.TotalCostUSD, record.RecordedAt,
	)
	return err
}

func (s *costStore) GetUsage(ctx context.Context, filter storage.UsageFilter) ([]*storage.UsageRecord, error) {
	var where []string
	var args []any

	if filter.Scope != nil {
		if filter.Scope.OrgID != "" {
			where = append(where, "org_id = ?")
			args = append(args, filter.Scope.OrgID)
		}
		if filter.Scope.TeamID != "" {
			where = append(where, "team_id = ?")
			args = append(args, filter.Scope.TeamID)
		}
		if filter.Scope.RepoID != "" {
			where = append(where, "repo_id = ?")
			args = append(args, filter.Scope.RepoID)
		}
	}
	if filter.RunID != "" {
		where = append(where, "run_id = ?")
		args = append(args, filter.RunID)
	}
	if !filter.Since.IsZero() {
		where = append(where, "recorded_at >= ?")
		args = append(args, filter.Since)
	}
	if !filter.Until.IsZero() {
		where = append(where, "recorded_at <= ?")
		args = append(args, filter.Until)
	}

	query := `SELECT id, run_id, org_id, team_id, repo_id, step,
		input_tokens, output_tokens, cache_read_tokens, cache_write_tokens,
		total_cost_usd, recorded_at FROM usage_records`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY recorded_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*storage.UsageRecord
	for rows.Next() {
		var r storage.UsageRecord
		if err := rows.Scan(
			&r.ID, &r.RunID, &r.Scope.OrgID, &r.Scope.TeamID, &r.Scope.RepoID,
			&r.Step, &r.InputTokens, &r.OutputTokens,
			&r.CacheReadTokens, &r.CacheWriteTokens,
			&r.TotalCostUSD, &r.RecordedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}
	return records, rows.Err()
}

func (s *costStore) GetUsageSummary(ctx context.Context, scope storage.Scope, group storage.TimeGroup, since, until time.Time) ([]*storage.UsageSummary, error) {
	var groupExpr string
	switch group {
	case storage.TimeGroupHour:
		groupExpr = "strftime('%Y-%m-%d %H:00:00', recorded_at)"
	case storage.TimeGroupDay:
		groupExpr = "strftime('%Y-%m-%d', recorded_at)"
	case storage.TimeGroupWeek:
		groupExpr = "strftime('%Y-%W', recorded_at)"
	case storage.TimeGroupMonth:
		groupExpr = "strftime('%Y-%m', recorded_at)"
	default:
		groupExpr = "strftime('%Y-%m-%d', recorded_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s as period,
			COUNT(DISTINCT run_id) as total_runs,
			SUM(total_cost_usd) as total_cost,
			SUM(input_tokens) as input_tokens,
			SUM(output_tokens) as output_tokens
		FROM usage_records
		WHERE org_id = ? AND team_id = ? AND repo_id = ?
			AND recorded_at >= ? AND recorded_at <= ?
		GROUP BY period
		ORDER BY period ASC`,
		groupExpr,
	)

	rows, err := s.db.QueryContext(ctx, query,
		scope.OrgID, scope.TeamID, scope.RepoID, since, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*storage.UsageSummary
	for rows.Next() {
		var s storage.UsageSummary
		var period string
		if err := rows.Scan(&period, &s.TotalRuns, &s.TotalCostUSD, &s.InputTokens, &s.OutputTokens); err != nil {
			return nil, err
		}
		// Parse period string back to time
		s.Period, _ = time.Parse("2006-01-02", period)
		if s.Period.IsZero() {
			s.Period, _ = time.Parse("2006-01-02 15:04:05", period)
		}
		summaries = append(summaries, &s)
	}
	return summaries, rows.Err()
}

func (s *costStore) GetBudget(ctx context.Context, scope storage.Scope) (*storage.Budget, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, team_id, repo_id, monthly_limit, daily_limit,
			per_run_limit, alert_at, updated_at
		FROM budgets WHERE org_id = ? AND team_id = ? AND repo_id = ?`,
		scope.OrgID, scope.TeamID, scope.RepoID,
	)

	var b storage.Budget
	err := row.Scan(
		&b.ID, &b.Scope.OrgID, &b.Scope.TeamID, &b.Scope.RepoID,
		&b.MonthlyLimit, &b.DailyLimit, &b.PerRunLimit,
		&b.AlertAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *costStore) SetBudget(ctx context.Context, budget *storage.Budget) error {
	budget.UpdatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO budgets (id, org_id, team_id, repo_id, monthly_limit, daily_limit,
			per_run_limit, alert_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(org_id, team_id, repo_id)
		DO UPDATE SET monthly_limit=excluded.monthly_limit, daily_limit=excluded.daily_limit,
			per_run_limit=excluded.per_run_limit, alert_at=excluded.alert_at,
			updated_at=excluded.updated_at`,
		budget.ID, budget.Scope.OrgID, budget.Scope.TeamID, budget.Scope.RepoID,
		budget.MonthlyLimit, budget.DailyLimit, budget.PerRunLimit,
		budget.AlertAt, budget.UpdatedAt,
	)
	return err
}
