package agent

import (
	"fmt"
	"strings"
	"time"
)

// SearchFilter represents search criteria for sessions
type SearchFilter struct {
	Query       string    // Full-text search query
	Tags        []string  // Filter by tags
	ProjectPath string    // Filter by project
	IsFavorite  *bool     // Filter favorites (nil = no filter)
	FromDate    time.Time // Filter by date range (start)
	ToDate      time.Time // Filter by date range (end)
}

// SearchResult represents a search result
type SearchResult struct {
	Session     *Session
	Score       int      // Relevance score
	MatchReason []string // Why this session matched
}

// SessionLoader is a function type for loading sessions (for testing)
type SessionLoader func() ([]*Session, error)

// defaultSessionLoader is the default implementation
var defaultSessionLoader SessionLoader = LoadAllSessions

// SearchSessions searches all sessions based on filter criteria
func SearchSessions(filter SearchFilter) ([]*SearchResult, error) {
	return searchSessionsWithLoader(filter, defaultSessionLoader)
}

// searchSessionsWithLoader allows dependency injection for testing
func searchSessionsWithLoader(filter SearchFilter, loader SessionLoader) ([]*SearchResult, error) {
	sessions, err := loader()
	if err != nil {
		return nil, err
	}

	var results []*SearchResult

	for _, session := range sessions {
		if matchesFilter(session, filter) {
			score, reasons := scoreSession(session, filter)
			results = append(results, &SearchResult{
				Session:     session,
				Score:       score,
				MatchReason: reasons,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

// matchesFilter checks if a session matches the filter criteria
func matchesFilter(session *Session, filter SearchFilter) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()

	// Filter by project path
	if filter.ProjectPath != "" && session.ProjectPath != filter.ProjectPath {
		return false
	}

	// Filter by favorite status
	if filter.IsFavorite != nil && session.IsFavorite != *filter.IsFavorite {
		return false
	}

	// Filter by tags
	if len(filter.Tags) > 0 {
		hasAllTags := true
		for _, filterTag := range filter.Tags {
			found := false
			for _, sessionTag := range session.Tags {
				if strings.EqualFold(sessionTag, filterTag) {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if !hasAllTags {
			return false
		}
	}

	// Filter by date range
	if !filter.FromDate.IsZero() && session.UpdatedAt.Before(filter.FromDate) {
		return false
	}
	if !filter.ToDate.IsZero() && session.UpdatedAt.After(filter.ToDate) {
		return false
	}

	// If there's a query, check if it matches
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)

		// Search in messages
		for _, msg := range session.Messages {
			if strings.Contains(strings.ToLower(msg.Content), query) {
				return true
			}
		}

		// Search in project path
		if strings.Contains(strings.ToLower(session.ProjectPath), query) {
			return true
		}

		// Search in tags
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				return true
			}
		}

		// No match found
		return false
	}

	return true
}

// scoreSession calculates relevance score and match reasons
func scoreSession(session *Session, filter SearchFilter) (int, []string) {
	session.mu.RLock()
	defer session.mu.RUnlock()

	score := 0
	var reasons []string

	// Base score from recency
	daysSinceUpdate := time.Since(session.UpdatedAt).Hours() / 24
	if daysSinceUpdate < 1 {
		score += 50
	} else if daysSinceUpdate < 7 {
		score += 30
	} else if daysSinceUpdate < 30 {
		score += 10
	}

	// Boost for favorites
	if session.IsFavorite {
		score += 20
		reasons = append(reasons, "Favorite")
	}

	// Boost for tag matches
	if len(filter.Tags) > 0 {
		score += len(filter.Tags) * 10
		for _, tag := range filter.Tags {
			reasons = append(reasons, "Tag: "+tag)
		}
	}

	// Query relevance
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		matchCount := 0

		// Count message matches
		for _, msg := range session.Messages {
			if strings.Contains(strings.ToLower(msg.Content), query) {
				matchCount++
			}
		}

		if matchCount > 0 {
			score += matchCount * 5
			if matchCount == 1 {
				reasons = append(reasons, "1 message match")
			} else {
				reasons = append(reasons, fmt.Sprintf("%d message matches", matchCount))
			}
		}

		// Project path match
		if strings.Contains(strings.ToLower(session.ProjectPath), query) {
			score += 15
			reasons = append(reasons, "Project path match")
		}

		// Tag match
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				score += 10
				reasons = append(reasons, "Tag match: "+tag)
				break
			}
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Matches all filters")
	}

	return score, reasons
}

// GetAllTags returns all unique tags across all sessions
func GetAllTags() ([]string, error) {
	return getAllTagsWithLoader(defaultSessionLoader)
}

// getAllTagsWithLoader allows dependency injection for testing
func getAllTagsWithLoader(loader SessionLoader) ([]string, error) {
	sessions, err := loader()
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]bool)
	for _, session := range sessions {
		session.mu.RLock()
		for _, tag := range session.Tags {
			tagSet[tag] = true
		}
		session.mu.RUnlock()
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	// Sort alphabetically
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[j] < tags[i] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	return tags, nil
}
