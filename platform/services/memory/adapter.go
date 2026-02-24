package memory

import (
	"context"

	harnessmemory "github.com/philjestin/boatman-ecosystem/harness/memory"
	"github.com/philjestin/boatman-ecosystem/platform/storage"
)

// PlatformMemoryStore implements harness/memory.MemoryProvider
// so the CLI can swap in platform-backed memory transparently.
type PlatformMemoryStore struct {
	service *Service
	scope   storage.Scope
}

// Compile-time check that PlatformMemoryStore implements MemoryProvider.
var _ harnessmemory.MemoryProvider = (*PlatformMemoryStore)(nil)

// NewPlatformMemoryStore creates a memory provider backed by the platform service.
func NewPlatformMemoryStore(service *Service, scope storage.Scope) *PlatformMemoryStore {
	return &PlatformMemoryStore{
		service: service,
		scope:   scope,
	}
}

// Get retrieves memory for the given project path. The projectPath is ignored
// since platform memory is scoped by org/team/repo.
func (s *PlatformMemoryStore) Get(_ string) (*harnessmemory.Memory, error) {
	return s.service.ToHarnessMemory(context.Background(), s.scope)
}

// Save persists memory back to the platform. This converts harness patterns
// and issues back to platform storage types.
func (s *PlatformMemoryStore) Save(mem *harnessmemory.Memory) error {
	ctx := context.Background()

	// Save patterns
	for _, p := range mem.Patterns {
		sp := &storage.Pattern{
			ID:          p.ID,
			Scope:       s.scope,
			Type:        p.Type,
			Description: p.Description,
			Example:     p.Example,
			FileMatcher: p.FileMatcher,
			Weight:      p.Weight,
			UsageCount:  p.UsageCount,
			SuccessRate: p.SuccessRate,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
		// Try create, fall back to update
		if err := s.service.store.CreatePattern(ctx, sp); err != nil {
			if err := s.service.store.UpdatePattern(ctx, sp); err != nil {
				return err
			}
		}
	}

	// Save preferences
	prefs := &storage.Preferences{
		ID:                     "prefs-" + s.scope.OrgID + "-" + s.scope.TeamID + "-" + s.scope.RepoID,
		Scope:                  s.scope,
		PreferredTestFramework: mem.Preferences.PreferredTestFramework,
		NamingConventions:      mem.Preferences.NamingConventions,
		FileOrganization:       mem.Preferences.FileOrganization,
		CodeStyle:              mem.Preferences.CodeStyle,
		CommitMessageFormat:    mem.Preferences.CommitMessageFormat,
		ReviewerThresholds:     mem.Preferences.ReviewerThresholds,
	}
	if err := s.service.store.SetPreferences(ctx, prefs); err != nil {
		return err
	}

	return nil
}
