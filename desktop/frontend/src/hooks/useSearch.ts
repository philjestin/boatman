import { useState, useCallback, useEffect } from 'react';
import { SearchSessions, GetAllTags } from '../../wailsjs/go/main/App';
import type { SearchFilters } from '../components/search/FilterPanel';
import type { SearchResult } from '../components/search/SearchResults';

export function useSearch() {
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [availableProjects, setAvailableProjects] = useState<string[]>([]);

  // Load available tags
  const loadTags = useCallback(async () => {
    try {
      const tags = await GetAllTags();
      setAvailableTags(tags || []);
    } catch (error) {
      console.error('Failed to load tags:', error);
      setAvailableTags([]);
    }
  }, []);

  // Load tags when search opens
  useEffect(() => {
    if (isSearchOpen) {
      loadTags();
    }
  }, [isSearchOpen, loadTags]);

  // Perform search
  const performSearch = useCallback(
    async (query: string, filters: SearchFilters): Promise<SearchResult[]> => {
      try {
        const results = await SearchSessions({
          query,
          tags: filters.tags,
          projectPath: filters.projectPath,
          isFavorite: filters.isFavorite,
          fromDate: filters.fromDate,
          toDate: filters.toDate,
        });

        return results.map((r: any) => ({
          sessionId: r.sessionId,
          projectPath: r.projectPath,
          createdAt: r.createdAt,
          updatedAt: r.updatedAt,
          tags: r.tags || [],
          isFavorite: r.isFavorite || false,
          messageCount: r.messageCount || 0,
          score: r.score || 0,
          matchReasons: r.matchReasons || [],
          status: r.status || 'idle',
        }));
      } catch (error) {
        console.error('Search failed:', error);
        return [];
      }
    },
    []
  );

  // Set up available projects from sessions
  const setProjects = useCallback((projects: string[]) => {
    setAvailableProjects(projects);
  }, []);

  // Open search modal
  const openSearch = useCallback(() => {
    setIsSearchOpen(true);
  }, []);

  // Close search modal
  const closeSearch = useCallback(() => {
    setIsSearchOpen(false);
  }, []);

  // Keyboard shortcut for global search (Cmd+Shift+F)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'f') {
        e.preventDefault();
        openSearch();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [openSearch]);

  return {
    isSearchOpen,
    openSearch,
    closeSearch,
    performSearch,
    availableTags,
    availableProjects,
    setProjects,
  };
}
