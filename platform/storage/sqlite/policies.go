package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

type policyStore struct {
	db *sql.DB
}

func (s *policyStore) Get(ctx context.Context, scope storage.Scope) (*storage.Policy, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, team_id, repo_id, max_iterations, max_cost_per_run,
			max_files_changed, allowed_models, blocked_patterns,
			require_tests, require_review, updated_at
		FROM policies WHERE org_id = ? AND team_id = ? AND repo_id = ?`,
		scope.OrgID, scope.TeamID, scope.RepoID,
	)

	return scanPolicy(row)
}

func (s *policyStore) Set(ctx context.Context, policy *storage.Policy) error {
	policy.UpdatedAt = time.Now().UTC()

	modelsJSON, err := json.Marshal(policy.AllowedModels)
	if err != nil {
		return fmt.Errorf("marshal allowed_models: %w", err)
	}
	patternsJSON, err := json.Marshal(policy.BlockedPatterns)
	if err != nil {
		return fmt.Errorf("marshal blocked_patterns: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO policies (id, org_id, team_id, repo_id, max_iterations, max_cost_per_run,
			max_files_changed, allowed_models, blocked_patterns,
			require_tests, require_review, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(org_id, team_id, repo_id)
		DO UPDATE SET max_iterations=excluded.max_iterations,
			max_cost_per_run=excluded.max_cost_per_run,
			max_files_changed=excluded.max_files_changed,
			allowed_models=excluded.allowed_models,
			blocked_patterns=excluded.blocked_patterns,
			require_tests=excluded.require_tests,
			require_review=excluded.require_review,
			updated_at=excluded.updated_at`,
		policy.ID, policy.Scope.OrgID, policy.Scope.TeamID, policy.Scope.RepoID,
		policy.MaxIterations, policy.MaxCostPerRun,
		policy.MaxFilesChanged, string(modelsJSON), string(patternsJSON),
		boolToInt(policy.RequireTests), boolToInt(policy.RequireReview),
		policy.UpdatedAt,
	)
	return err
}

func (s *policyStore) Delete(ctx context.Context, scope storage.Scope) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM policies WHERE org_id = ? AND team_id = ? AND repo_id = ?`,
		scope.OrgID, scope.TeamID, scope.RepoID,
	)
	return err
}

// GetEffectivePolicy merges policies from org -> team -> repo level.
func (s *policyStore) GetEffectivePolicy(ctx context.Context, scope storage.Scope) (*storage.Policy, error) {
	// Collect policies at each level: org, team, repo
	var policies []*storage.Policy

	// Org level
	if scope.OrgID != "" {
		orgPolicy, err := s.Get(ctx, storage.Scope{OrgID: scope.OrgID})
		if err == nil && orgPolicy != nil {
			policies = append(policies, orgPolicy)
		}
	}

	// Team level
	if scope.OrgID != "" && scope.TeamID != "" {
		teamPolicy, err := s.Get(ctx, storage.Scope{OrgID: scope.OrgID, TeamID: scope.TeamID})
		if err == nil && teamPolicy != nil {
			policies = append(policies, teamPolicy)
		}
	}

	// Repo level (full scope)
	if scope.OrgID != "" && scope.TeamID != "" && scope.RepoID != "" {
		repoPolicy, err := s.Get(ctx, scope)
		if err == nil && repoPolicy != nil {
			policies = append(policies, repoPolicy)
		}
	}

	if len(policies) == 0 {
		return nil, nil
	}

	// Merge: more specific levels override less specific
	effective := &storage.Policy{Scope: scope}
	for _, p := range policies {
		if p.MaxIterations > 0 {
			if effective.MaxIterations == 0 || p.MaxIterations < effective.MaxIterations {
				effective.MaxIterations = p.MaxIterations
			}
		}
		if p.MaxCostPerRun > 0 {
			if effective.MaxCostPerRun == 0 || p.MaxCostPerRun < effective.MaxCostPerRun {
				effective.MaxCostPerRun = p.MaxCostPerRun
			}
		}
		if p.MaxFilesChanged > 0 {
			if effective.MaxFilesChanged == 0 || p.MaxFilesChanged < effective.MaxFilesChanged {
				effective.MaxFilesChanged = p.MaxFilesChanged
			}
		}
		// Allowed models: intersect (if both set)
		if len(p.AllowedModels) > 0 {
			if len(effective.AllowedModels) == 0 {
				effective.AllowedModels = p.AllowedModels
			} else {
				effective.AllowedModels = intersectStrings(effective.AllowedModels, p.AllowedModels)
			}
		}
		// Blocked patterns: union
		if len(p.BlockedPatterns) > 0 {
			effective.BlockedPatterns = unionStrings(effective.BlockedPatterns, p.BlockedPatterns)
		}
		// Booleans: OR (if org requires, team can't disable)
		if p.RequireTests {
			effective.RequireTests = true
		}
		if p.RequireReview {
			effective.RequireReview = true
		}
	}

	return effective, nil
}

func scanPolicy(row *sql.Row) (*storage.Policy, error) {
	var p storage.Policy
	var modelsJSON, patternsJSON string
	var requireTests, requireReview int

	err := row.Scan(
		&p.ID, &p.Scope.OrgID, &p.Scope.TeamID, &p.Scope.RepoID,
		&p.MaxIterations, &p.MaxCostPerRun, &p.MaxFilesChanged,
		&modelsJSON, &patternsJSON,
		&requireTests, &requireReview, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	p.RequireTests = requireTests != 0
	p.RequireReview = requireReview != 0
	_ = json.Unmarshal([]byte(modelsJSON), &p.AllowedModels)
	_ = json.Unmarshal([]byte(patternsJSON), &p.BlockedPatterns)

	return &p, nil
}

func intersectStrings(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}
	var result []string
	for _, s := range b {
		if set[s] {
			result = append(result, s)
		}
	}
	return result
}

func unionStrings(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[s] = true
	}
	for _, s := range b {
		set[s] = true
	}
	var result []string
	for s := range set {
		result = append(result, s)
	}
	return result
}
