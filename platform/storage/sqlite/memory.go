package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type memoryStore struct {
	db *sql.DB
}

// --- Patterns ---

func (s *memoryStore) CreatePattern(ctx context.Context, p *storage.Pattern) error {
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO patterns (id, org_id, team_id, repo_id, type, description, example,
			file_matcher, weight, usage_count, success_rate, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Scope.OrgID, p.Scope.TeamID, p.Scope.RepoID,
		p.Type, p.Description, p.Example,
		p.FileMatcher, p.Weight, p.UsageCount, p.SuccessRate,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (s *memoryStore) UpdatePattern(ctx context.Context, p *storage.Pattern) error {
	p.UpdatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx, `
		UPDATE patterns SET org_id=?, team_id=?, repo_id=?, type=?, description=?,
			example=?, file_matcher=?, weight=?, usage_count=?, success_rate=?, updated_at=?
		WHERE id=?`,
		p.Scope.OrgID, p.Scope.TeamID, p.Scope.RepoID,
		p.Type, p.Description, p.Example,
		p.FileMatcher, p.Weight, p.UsageCount, p.SuccessRate,
		p.UpdatedAt, p.ID,
	)
	return err
}

func (s *memoryStore) ListPatterns(ctx context.Context, scope storage.Scope) ([]*storage.Pattern, error) {
	query := `SELECT id, org_id, team_id, repo_id, type, description, example,
		file_matcher, weight, usage_count, success_rate, created_at, updated_at
		FROM patterns WHERE org_id = ? AND team_id = ? AND repo_id = ?
		ORDER BY weight DESC`

	rows, err := s.db.QueryContext(ctx, query, scope.OrgID, scope.TeamID, scope.RepoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []*storage.Pattern
	for rows.Next() {
		var p storage.Pattern
		if err := rows.Scan(
			&p.ID, &p.Scope.OrgID, &p.Scope.TeamID, &p.Scope.RepoID,
			&p.Type, &p.Description, &p.Example,
			&p.FileMatcher, &p.Weight, &p.UsageCount, &p.SuccessRate,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		patterns = append(patterns, &p)
	}
	return patterns, rows.Err()
}

func (s *memoryStore) DeletePattern(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM patterns WHERE id = ?", id)
	return err
}

// --- Preferences ---

func (s *memoryStore) GetPreferences(ctx context.Context, scope storage.Scope) (*storage.Preferences, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, team_id, repo_id, data, updated_at
		FROM preferences WHERE org_id = ? AND team_id = ? AND repo_id = ?`,
		scope.OrgID, scope.TeamID, scope.RepoID,
	)

	var prefs storage.Preferences
	var dataJSON string
	err := row.Scan(&prefs.ID, &prefs.Scope.OrgID, &prefs.Scope.TeamID, &prefs.Scope.RepoID,
		&dataJSON, &prefs.UpdatedAt)
	if err == sql.ErrNoRows {
		return &storage.Preferences{
			Scope:              scope,
			NamingConventions:  make(map[string]string),
			FileOrganization:   make(map[string]string),
			CodeStyle:          make(map[string]string),
			ReviewerThresholds: make(map[string]int),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(dataJSON), &prefs); err != nil {
		return nil, fmt.Errorf("unmarshal preferences: %w", err)
	}
	prefs.Scope = scope
	return &prefs, nil
}

func (s *memoryStore) SetPreferences(ctx context.Context, prefs *storage.Preferences) error {
	prefs.UpdatedAt = time.Now().UTC()

	dataJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO preferences (id, org_id, team_id, repo_id, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(org_id, team_id, repo_id)
		DO UPDATE SET data=excluded.data, updated_at=excluded.updated_at`,
		prefs.ID, prefs.Scope.OrgID, prefs.Scope.TeamID, prefs.Scope.RepoID,
		string(dataJSON), prefs.UpdatedAt,
	)
	return err
}

// --- Common Issues ---

func (s *memoryStore) CreateIssue(ctx context.Context, issue *storage.CommonIssue) error {
	now := time.Now().UTC()
	if issue.CreatedAt.IsZero() {
		issue.CreatedAt = now
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO common_issues (id, org_id, team_id, repo_id, type, description,
			solution, frequency, auto_fix, file_matcher, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		issue.ID, issue.Scope.OrgID, issue.Scope.TeamID, issue.Scope.RepoID,
		issue.Type, issue.Description, issue.Solution,
		issue.Frequency, boolToInt(issue.AutoFix), issue.FileMatcher,
		issue.CreatedAt,
	)
	return err
}

func (s *memoryStore) UpdateIssue(ctx context.Context, issue *storage.CommonIssue) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE common_issues SET org_id=?, team_id=?, repo_id=?, type=?, description=?,
			solution=?, frequency=?, auto_fix=?, file_matcher=?
		WHERE id=?`,
		issue.Scope.OrgID, issue.Scope.TeamID, issue.Scope.RepoID,
		issue.Type, issue.Description, issue.Solution,
		issue.Frequency, boolToInt(issue.AutoFix), issue.FileMatcher,
		issue.ID,
	)
	return err
}

func (s *memoryStore) ListIssues(ctx context.Context, scope storage.Scope) ([]*storage.CommonIssue, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, org_id, team_id, repo_id, type, description, solution,
			frequency, auto_fix, file_matcher, created_at
		FROM common_issues WHERE org_id = ? AND team_id = ? AND repo_id = ?
		ORDER BY frequency DESC`,
		scope.OrgID, scope.TeamID, scope.RepoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*storage.CommonIssue
	for rows.Next() {
		var issue storage.CommonIssue
		var autoFix int
		if err := rows.Scan(
			&issue.ID, &issue.Scope.OrgID, &issue.Scope.TeamID, &issue.Scope.RepoID,
			&issue.Type, &issue.Description, &issue.Solution,
			&issue.Frequency, &autoFix, &issue.FileMatcher,
			&issue.CreatedAt,
		); err != nil {
			return nil, err
		}
		issue.AutoFix = autoFix != 0
		issues = append(issues, &issue)
	}
	return issues, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
