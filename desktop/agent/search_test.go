package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper function to create a test session
func createTestSession(id, projectPath string, tags []string, isFavorite bool, messages []string) *Session {
	session := NewSession(id, projectPath)
	session.Tags = tags
	session.IsFavorite = isFavorite

	// Add messages
	for _, content := range messages {
		msg := Message{
			ID:        "msg-" + content,
			Role:      "user",
			Content:   content,
			Timestamp: time.Now(),
		}
		session.Messages = append(session.Messages, msg)
	}

	return session
}

// Helper to create temp test directory
func createTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "boatman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tempDir
}

// Helper to cleanup test directory
func cleanupTestDir(t *testing.T, dir string) {
	os.RemoveAll(dir)
}

// Test basic search functionality
func TestSearchSessions_BasicQuery(t *testing.T) {
	// Create test sessions in memory
	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{"bug"}, false, []string{"Fix authentication bug"}),
		createTestSession("2", "/project/bar", []string{"feature"}, true, []string{"Add new feature"}),
		createTestSession("3", "/project/baz", []string{"refactor"}, false, []string{"Refactor database layer"}),
	}

	// Create test loader
	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	tests := []struct {
		name          string
		filter        SearchFilter
		expectedCount int
		expectedFirst string // Expected first result ID
	}{
		{
			name: "Search for 'bug'",
			filter: SearchFilter{
				Query: "bug",
			},
			expectedCount: 1,
			expectedFirst: "1",
		},
		{
			name: "Search for 'feature'",
			filter: SearchFilter{
				Query: "feature",
			},
			expectedCount: 1,
			expectedFirst: "2",
		},
		{
			name: "Search for 'database'",
			filter: SearchFilter{
				Query: "database",
			},
			expectedCount: 1,
			expectedFirst: "3",
		},
		{
			name: "Search for 'project' (in path)",
			filter: SearchFilter{
				Query: "project",
			},
			expectedCount: 3, // All sessions have /project in path
		},
		{
			name: "Search with no matches",
			filter: SearchFilter{
				Query: "nonexistent",
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := searchSessionsWithLoader(tt.filter, loader)
			if err != nil {
				t.Fatalf("SearchSessions failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			if tt.expectedCount > 0 && tt.expectedFirst != "" {
				if results[0].Session.ID != tt.expectedFirst {
					t.Errorf("Expected first result ID %s, got %s", tt.expectedFirst, results[0].Session.ID)
				}
			}
		})
	}
}

// Test tag filtering
func TestSearchSessions_TagFilter(t *testing.T) {
	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{"bug", "critical"}, false, []string{"Critical bug"}),
		createTestSession("2", "/project/bar", []string{"bug", "minor"}, false, []string{"Minor bug"}),
		createTestSession("3", "/project/baz", []string{"feature"}, false, []string{"New feature"}),
	}

	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	tests := []struct {
		name          string
		tags          []string
		expectedCount int
	}{
		{
			name:          "Filter by 'bug' tag",
			tags:          []string{"bug"},
			expectedCount: 2,
		},
		{
			name:          "Filter by 'critical' tag",
			tags:          []string{"critical"},
			expectedCount: 1,
		},
		{
			name:          "Filter by 'bug' and 'critical' tags",
			tags:          []string{"bug", "critical"},
			expectedCount: 1,
		},
		{
			name:          "Filter by non-existent tag",
			tags:          []string{"nonexistent"},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := SearchFilter{
				Tags: tt.tags,
			}

			results, err := searchSessionsWithLoader(filter, loader)
			if err != nil {
				t.Fatalf("SearchSessions failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

// Test favorite filtering
func TestSearchSessions_FavoriteFilter(t *testing.T) {
	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{}, true, []string{"Favorite 1"}),
		createTestSession("2", "/project/bar", []string{}, false, []string{"Not favorite"}),
		createTestSession("3", "/project/baz", []string{}, true, []string{"Favorite 2"}),
	}

	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	t.Run("Filter favorites only", func(t *testing.T) {
		isFavorite := true
		filter := SearchFilter{
			IsFavorite: &isFavorite,
		}

		results, err := searchSessionsWithLoader(filter, loader)
		if err != nil {
			t.Fatalf("SearchSessions failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 favorite results, got %d", len(results))
		}
	})

	t.Run("Filter non-favorites only", func(t *testing.T) {
		isFavorite := false
		filter := SearchFilter{
			IsFavorite: &isFavorite,
		}

		results, err := searchSessionsWithLoader(filter, loader)
		if err != nil {
			t.Fatalf("SearchSessions failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 non-favorite result, got %d", len(results))
		}
	})
}

// Test date range filtering
func TestSearchSessions_DateRangeFilter(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)
	lastMonth := now.AddDate(0, -1, 0)

	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{}, false, []string{"Today"}),
		createTestSession("2", "/project/bar", []string{}, false, []string{"Last week"}),
		createTestSession("3", "/project/baz", []string{}, false, []string{"Last month"}),
	}
	sessions[0].UpdatedAt = now
	sessions[1].UpdatedAt = lastWeek
	sessions[2].UpdatedAt = lastMonth

	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	t.Run("Filter from yesterday", func(t *testing.T) {
		filter := SearchFilter{
			FromDate: yesterday,
		}

		results, err := searchSessionsWithLoader(filter, loader)
		if err != nil {
			t.Fatalf("SearchSessions failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result from yesterday, got %d", len(results))
		}
	})

	t.Run("Filter from last month", func(t *testing.T) {
		filter := SearchFilter{
			FromDate: lastMonth.AddDate(0, 0, -1),
		}

		results, err := searchSessionsWithLoader(filter, loader)
		if err != nil {
			t.Fatalf("SearchSessions failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results from last month, got %d", len(results))
		}
	})
}

// Test project path filtering
func TestSearchSessions_ProjectPathFilter(t *testing.T) {
	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{}, false, []string{"Session 1"}),
		createTestSession("2", "/project/bar", []string{}, false, []string{"Session 2"}),
		createTestSession("3", "/project/foo", []string{}, false, []string{"Session 3"}),
	}

	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	t.Run("Filter by project path /project/foo", func(t *testing.T) {
		filter := SearchFilter{
			ProjectPath: "/project/foo",
		}

		results, err := searchSessionsWithLoader(filter, loader)
		if err != nil {
			t.Fatalf("SearchSessions failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results for /project/foo, got %d", len(results))
		}
	})
}

// Test GetAllTags
func TestGetAllTags(t *testing.T) {
	sessions := []*Session{
		createTestSession("1", "/project/foo", []string{"bug", "critical"}, false, []string{"Session 1"}),
		createTestSession("2", "/project/bar", []string{"bug", "minor"}, false, []string{"Session 2"}),
		createTestSession("3", "/project/baz", []string{"feature", "ui"}, false, []string{"Session 3"}),
	}

	loader := func() ([]*Session, error) {
		return sessions, nil
	}

	tags, err := getAllTagsWithLoader(loader)
	if err != nil {
		t.Fatalf("GetAllTags failed: %v", err)
	}

	expectedTags := map[string]bool{
		"bug":      true,
		"critical": true,
		"minor":    true,
		"feature":  true,
		"ui":       true,
	}

	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d unique tags, got %d", len(expectedTags), len(tags))
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

// Test session tag management
func TestSession_TagManagement(t *testing.T) {
	session := NewSession("test-session", "/project/test")

	t.Run("Add tag", func(t *testing.T) {
		session.AddTag("bug")
		tags := session.GetTags()
		if len(tags) != 1 || tags[0] != "bug" {
			t.Errorf("Expected tag 'bug', got %v", tags)
		}
	})

	t.Run("Add duplicate tag (case insensitive)", func(t *testing.T) {
		session.AddTag("BUG")
		tags := session.GetTags()
		if len(tags) != 1 {
			t.Errorf("Expected 1 tag (duplicate should be ignored), got %d", len(tags))
		}
	})

	t.Run("Add another tag", func(t *testing.T) {
		session.AddTag("feature")
		tags := session.GetTags()
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}
	})

	t.Run("Remove tag", func(t *testing.T) {
		session.RemoveTag("bug")
		tags := session.GetTags()
		if len(tags) != 1 || tags[0] != "feature" {
			t.Errorf("Expected only 'feature' tag, got %v", tags)
		}
	})

	t.Run("Remove non-existent tag", func(t *testing.T) {
		session.RemoveTag("nonexistent")
		tags := session.GetTags()
		if len(tags) != 1 {
			t.Errorf("Expected 1 tag (no change), got %d", len(tags))
		}
	})
}

// Test session favorite management
func TestSession_FavoriteManagement(t *testing.T) {
	session := NewSession("test-session", "/project/test")

	t.Run("Initial state is not favorite", func(t *testing.T) {
		if session.IsFavorite {
			t.Error("New session should not be favorite")
		}
	})

	t.Run("Set as favorite", func(t *testing.T) {
		session.SetFavorite(true)
		if !session.IsFavorite {
			t.Error("Session should be favorite after SetFavorite(true)")
		}
	})

	t.Run("Unset favorite", func(t *testing.T) {
		session.SetFavorite(false)
		if session.IsFavorite {
			t.Error("Session should not be favorite after SetFavorite(false)")
		}
	})
}

// Test persistence of tags and favorites
func TestSession_PersistenceTagsAndFavorites(t *testing.T) {
	tempDir := createTestDir(t)
	defer cleanupTestDir(t, tempDir)

	// Override sessions directory for testing
	originalGetter := defaultSessionsDirGetter
	defer func() { defaultSessionsDirGetter = originalGetter }()

	defaultSessionsDirGetter = func() (string, error) {
		sessionsDir := filepath.Join(tempDir, "sessions")
		if err := os.MkdirAll(sessionsDir, 0755); err != nil {
			return "", err
		}
		return sessionsDir, nil
	}

	// Create and save a session with tags and favorite
	session := NewSession("test-session", "/project/test")
	session.AddTag("bug")
	session.AddTag("critical")
	session.SetFavorite(true)

	err := SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load the session back
	loadedSession, err := LoadSession("test-session")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify tags
	tags := loadedSession.GetTags()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	expectedTags := map[string]bool{"bug": true, "critical": true}
	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}

	// Verify favorite
	if !loadedSession.IsFavorite {
		t.Error("Loaded session should be favorite")
	}
}
