// Package memory provides hierarchical shared memory that merges org->team->repo patterns.
package memory

import (
	"context"
	"fmt"
	"sort"

	harnessmemory "github.com/philjestin/boatman-ecosystem/harness/memory"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// Service provides hierarchical memory management across organizational scopes.
type Service struct {
	store storage.MemoryStore
}

// NewService creates a new memory service.
func NewService(store storage.MemoryStore) *Service {
	return &Service{store: store}
}

// GetMergedPatterns returns patterns from all scope levels,
// with repo overriding team overriding org.
func (s *Service) GetMergedPatterns(ctx context.Context, scope storage.Scope) ([]*storage.Pattern, error) {
	var all []*storage.Pattern

	// Org-level patterns
	if scope.OrgID != "" {
		orgPatterns, err := s.store.ListPatterns(ctx, storage.Scope{OrgID: scope.OrgID})
		if err != nil {
			return nil, fmt.Errorf("list org patterns: %w", err)
		}
		all = append(all, orgPatterns...)
	}

	// Team-level patterns
	if scope.OrgID != "" && scope.TeamID != "" {
		teamPatterns, err := s.store.ListPatterns(ctx, storage.Scope{OrgID: scope.OrgID, TeamID: scope.TeamID})
		if err != nil {
			return nil, fmt.Errorf("list team patterns: %w", err)
		}
		all = append(all, teamPatterns...)
	}

	// Repo-level patterns
	if scope.OrgID != "" && scope.TeamID != "" && scope.RepoID != "" {
		repoPatterns, err := s.store.ListPatterns(ctx, scope)
		if err != nil {
			return nil, fmt.Errorf("list repo patterns: %w", err)
		}
		all = append(all, repoPatterns...)
	}

	// Deduplicate by ID, preferring more specific scope
	seen := make(map[string]*storage.Pattern)
	for _, p := range all {
		seen[p.ID] = p
	}

	var merged []*storage.Pattern
	for _, p := range seen {
		merged = append(merged, p)
	}

	// Sort by weight descending
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Weight > merged[j].Weight
	})

	return merged, nil
}

// LearnFromRun extracts patterns from a completed run and stores them at the repo scope level.
func (s *Service) LearnFromRun(ctx context.Context, run *storage.Run, reviewScore int) error {
	if reviewScore < 70 {
		return nil // only learn from reasonably successful runs
	}

	for _, file := range run.FilesChanged {
		pattern := &storage.Pattern{
			ID:          fmt.Sprintf("run-%s-%s", run.ID, file),
			Scope:       run.Scope,
			Type:        "success",
			Description: fmt.Sprintf("Pattern from run %s on file %s (score: %d)", run.ID, file, reviewScore),
			FileMatcher: file,
			Weight:      float64(reviewScore) / 100.0,
			UsageCount:  1,
			SuccessRate: float64(reviewScore) / 100.0,
		}

		if err := s.store.CreatePattern(ctx, pattern); err != nil {
			return fmt.Errorf("create pattern for %s: %w", file, err)
		}
	}

	return nil
}

// ToHarnessMemory builds a harness/memory.Memory from platform data
// so the existing runner pipeline works without modification.
func (s *Service) ToHarnessMemory(ctx context.Context, scope storage.Scope) (*harnessmemory.Memory, error) {
	patterns, err := s.GetMergedPatterns(ctx, scope)
	if err != nil {
		return nil, err
	}

	prefs, err := s.store.GetPreferences(ctx, scope)
	if err != nil {
		return nil, err
	}

	issues, err := s.store.ListIssues(ctx, scope)
	if err != nil {
		return nil, err
	}

	// Convert to harness types
	mem := &harnessmemory.Memory{
		ProjectID:    fmt.Sprintf("%s/%s/%s", scope.OrgID, scope.TeamID, scope.RepoID),
		FilePatterns: make(map[string][]string),
	}

	for _, p := range patterns {
		mem.Patterns = append(mem.Patterns, harnessmemory.Pattern{
			ID:          p.ID,
			Type:        p.Type,
			Description: p.Description,
			Example:     p.Example,
			FileMatcher: p.FileMatcher,
			Weight:      p.Weight,
			UsageCount:  p.UsageCount,
			SuccessRate: p.SuccessRate,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	for _, i := range issues {
		mem.CommonIssues = append(mem.CommonIssues, harnessmemory.CommonIssue{
			ID:          i.ID,
			Type:        i.Type,
			Description: i.Description,
			Solution:    i.Solution,
			Frequency:   i.Frequency,
			AutoFix:     i.AutoFix,
			FileMatcher: i.FileMatcher,
			CreatedAt:   i.CreatedAt,
		})
	}

	if prefs != nil {
		mem.Preferences = harnessmemory.Preferences{
			PreferredTestFramework: prefs.PreferredTestFramework,
			NamingConventions:      prefs.NamingConventions,
			FileOrganization:       prefs.FileOrganization,
			CodeStyle:              prefs.CodeStyle,
			CommitMessageFormat:    prefs.CommitMessageFormat,
			ReviewerThresholds:     prefs.ReviewerThresholds,
		}
	}

	return mem, nil
}
