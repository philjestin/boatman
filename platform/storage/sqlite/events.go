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

type eventStore struct {
	db *sql.DB
}

func (s *eventStore) Publish(ctx context.Context, event *storage.Event) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Version == 0 {
		event.Version = 1
	}

	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO events (id, run_id, org_id, team_id, repo_id, type, name,
			message, data, version, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.RunID, event.Scope.OrgID, event.Scope.TeamID, event.Scope.RepoID,
		event.Type, event.Name, event.Message,
		string(dataJSON), event.Version, event.CreatedAt,
	)
	return err
}

func (s *eventStore) Query(ctx context.Context, filter storage.EventFilter) ([]*storage.Event, error) {
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
	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		where = append(where, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}
	if !filter.Since.IsZero() {
		where = append(where, "created_at >= ?")
		args = append(args, filter.Since)
	}
	if !filter.Until.IsZero() {
		where = append(where, "created_at <= ?")
		args = append(args, filter.Until)
	}

	query := `SELECT id, run_id, org_id, team_id, repo_id, type, name,
		message, data, version, created_at FROM events`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*storage.Event
	for rows.Next() {
		var e storage.Event
		var dataJSON string
		if err := rows.Scan(
			&e.ID, &e.RunID, &e.Scope.OrgID, &e.Scope.TeamID, &e.Scope.RepoID,
			&e.Type, &e.Name, &e.Message,
			&dataJSON, &e.Version, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		if dataJSON != "" && dataJSON != "{}" {
			_ = json.Unmarshal([]byte(dataJSON), &e.Data)
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}
