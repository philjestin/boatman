package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type runStore struct {
	db *sql.DB
}

func (s *runStore) Create(ctx context.Context, run *storage.Run) error {
	filesJSON, err := json.Marshal(run.FilesChanged)
	if err != nil {
		return fmt.Errorf("marshal files: %w", err)
	}

	now := time.Now().UTC()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO runs (id, org_id, team_id, repo_id, user_id, status, prompt,
			total_cost_usd, iterations, files_changed, duration_ns, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.Scope.OrgID, run.Scope.TeamID, run.Scope.RepoID,
		run.UserID, string(run.Status), run.Prompt,
		run.TotalCostUSD, run.Iterations, string(filesJSON),
		int64(run.Duration), run.CreatedAt, run.UpdatedAt,
	)
	return err
}

func (s *runStore) Get(ctx context.Context, id string) (*storage.Run, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, team_id, repo_id, user_id, status, prompt,
			total_cost_usd, iterations, files_changed, duration_ns, created_at, updated_at
		FROM runs WHERE id = ?`, id)

	return scanRun(row)
}

func (s *runStore) Update(ctx context.Context, run *storage.Run) error {
	filesJSON, err := json.Marshal(run.FilesChanged)
	if err != nil {
		return fmt.Errorf("marshal files: %w", err)
	}

	run.UpdatedAt = time.Now().UTC()

	_, err = s.db.ExecContext(ctx, `
		UPDATE runs SET org_id=?, team_id=?, repo_id=?, user_id=?, status=?, prompt=?,
			total_cost_usd=?, iterations=?, files_changed=?, duration_ns=?, updated_at=?
		WHERE id=?`,
		run.Scope.OrgID, run.Scope.TeamID, run.Scope.RepoID,
		run.UserID, string(run.Status), run.Prompt,
		run.TotalCostUSD, run.Iterations, string(filesJSON),
		int64(run.Duration), run.UpdatedAt, run.ID,
	)
	return err
}

func (s *runStore) List(ctx context.Context, filter storage.RunFilter) ([]*storage.Run, error) {
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
	if filter.UserID != "" {
		where = append(where, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, string(filter.Status))
	}
	if !filter.Since.IsZero() {
		where = append(where, "created_at >= ?")
		args = append(args, filter.Since)
	}
	if !filter.Until.IsZero() {
		where = append(where, "created_at <= ?")
		args = append(args, filter.Until)
	}

	query := "SELECT id, org_id, team_id, repo_id, user_id, status, prompt, total_cost_usd, iterations, files_changed, duration_ns, created_at, updated_at FROM runs"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*storage.Run
	for rows.Next() {
		run, err := scanRunRows(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func scanRun(row *sql.Row) (*storage.Run, error) {
	var run storage.Run
	var filesJSON string
	var durationNS int64
	var status string

	err := row.Scan(
		&run.ID, &run.Scope.OrgID, &run.Scope.TeamID, &run.Scope.RepoID,
		&run.UserID, &status, &run.Prompt,
		&run.TotalCostUSD, &run.Iterations, &filesJSON,
		&durationNS, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	run.Status = storage.RunStatus(status)
	run.Duration = time.Duration(durationNS)
	if err := json.Unmarshal([]byte(filesJSON), &run.FilesChanged); err != nil {
		run.FilesChanged = nil
	}

	return &run, nil
}

func scanRunRows(rows *sql.Rows) (*storage.Run, error) {
	var run storage.Run
	var filesJSON string
	var durationNS int64
	var status string

	err := rows.Scan(
		&run.ID, &run.Scope.OrgID, &run.Scope.TeamID, &run.Scope.RepoID,
		&run.UserID, &status, &run.Prompt,
		&run.TotalCostUSD, &run.Iterations, &filesJSON,
		&durationNS, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	run.Status = storage.RunStatus(status)
	run.Duration = time.Duration(durationNS)
	if err := json.Unmarshal([]byte(filesJSON), &run.FilesChanged); err != nil {
		run.FilesChanged = nil
	}

	return &run, nil
}
