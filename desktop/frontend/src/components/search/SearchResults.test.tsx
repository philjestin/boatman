import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SearchResults, type SearchResult } from './SearchResults';

describe('SearchResults', () => {
  const mockOnSelectSession = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading state', () => {
    it('should show loading indicator when isLoading is true', () => {
      render(
        <SearchResults
          results={[]}
          onSelectSession={mockOnSelectSession}
          isLoading={true}
        />
      );

      expect(screen.getByText('Searching...')).toBeInTheDocument();
    });

    it('should show spinner animation when loading', () => {
      const { container } = render(
        <SearchResults
          results={[]}
          onSelectSession={mockOnSelectSession}
          isLoading={true}
        />
      );

      const spinner = container.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('Empty state', () => {
    it('should show empty state when no results', () => {
      render(
        <SearchResults
          results={[]}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('No sessions found')).toBeInTheDocument();
      expect(screen.getByText('Try adjusting your search or filters')).toBeInTheDocument();
    });

    it('should not show empty state when loading', () => {
      render(
        <SearchResults
          results={[]}
          onSelectSession={mockOnSelectSession}
          isLoading={true}
        />
      );

      expect(screen.queryByText('No sessions found')).not.toBeInTheDocument();
    });
  });

  describe('Results display', () => {
    const mockResults: SearchResult[] = [
      {
        sessionId: 'session-1',
        projectPath: '/path/to/project1',
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-15T15:30:00Z',
        tags: ['frontend', 'bug-fix'],
        isFavorite: true,
        messageCount: 42,
        score: 95,
        matchReasons: ['Matched query in message content', 'Tagged with frontend'],
        status: 'idle',
      },
      {
        sessionId: 'session-2',
        projectPath: '/path/to/project2',
        createdAt: '2024-01-10T08:00:00Z',
        updatedAt: '2024-01-14T12:00:00Z',
        tags: [],
        isFavorite: false,
        messageCount: 15,
        score: 0,
        matchReasons: [],
        status: 'running',
      },
    ];

    it('should render all search results', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('project1')).toBeInTheDocument();
      expect(screen.getByText('project2')).toBeInTheDocument();
    });

    it('should display project names from paths', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Should extract last part of path
      expect(screen.getByText('project1')).toBeInTheDocument();
      expect(screen.getByText('project2')).toBeInTheDocument();
    });

    it('should show favorite star for favorite sessions', () => {
      const { container } = render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      const stars = container.querySelectorAll('.text-yellow-500');
      expect(stars.length).toBeGreaterThan(0);
    });

    it('should display tags when present', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('frontend')).toBeInTheDocument();
      expect(screen.getByText('bug-fix')).toBeInTheDocument();
    });

    it('should not render tags section when no tags', () => {
      const resultsWithoutTags: SearchResult[] = [{
        ...mockResults[1],
        tags: [],
      }];

      const { container } = render(
        <SearchResults
          results={resultsWithoutTags}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Tags should not be visible for session-2
      const tagContainers = container.querySelectorAll('.flex-wrap');
      expect(tagContainers.length).toBe(0);
    });

    it('should display score badge when score > 0', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('95')).toBeInTheDocument();
    });

    it('should not display score badge when score is 0', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Only one score badge should be visible (95)
      const { container } = render(
        <SearchResults
          results={[mockResults[1]]}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      const scoreBadges = container.querySelectorAll('.text-blue-400');
      expect(scoreBadges.length).toBe(0);
    });

    it('should display message count', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('42 messages')).toBeInTheDocument();
      expect(screen.getByText('15 messages')).toBeInTheDocument();
    });

    it('should display match reasons when present', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText(/Matched query in message content/)).toBeInTheDocument();
    });

    it('should limit match reasons to 2', () => {
      const resultWithManyReasons: SearchResult[] = [{
        ...mockResults[0],
        matchReasons: ['Reason 1', 'Reason 2', 'Reason 3', 'Reason 4'],
      }];

      render(
        <SearchResults
          results={resultWithManyReasons}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Should show first 2 reasons joined with •
      expect(screen.getByText('Reason 1 • Reason 2')).toBeInTheDocument();
    });

    it('should display session status', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      expect(screen.getByText('idle')).toBeInTheDocument();
      expect(screen.getByText('running')).toBeInTheDocument();
    });

    it('should apply correct status colors', () => {
      const { container } = render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      const runningStatus = screen.getByText('running');
      expect(runningStatus).toHaveClass('text-green-400');
    });
  });

  describe('User interactions', () => {
    const mockResults: SearchResult[] = [
      {
        sessionId: 'session-1',
        projectPath: '/path/to/project1',
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-15T15:30:00Z',
        tags: [],
        isFavorite: false,
        messageCount: 10,
        score: 0,
        matchReasons: [],
        status: 'idle',
      },
    ];

    it('should call onSelectSession when result is clicked', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      const resultButton = screen.getByText('project1').closest('button');
      if (resultButton) {
        fireEvent.click(resultButton);
      }

      expect(mockOnSelectSession).toHaveBeenCalledWith('session-1');
    });

    it('should highlight result on hover', () => {
      render(
        <SearchResults
          results={mockResults}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      const resultButton = screen.getByText('project1').closest('button');
      expect(resultButton).toHaveClass('hover:bg-slate-800');
    });
  });

  describe('Relative time display', () => {
    it('should display relative time for recent updates', () => {
      const now = new Date();
      const recentResult: SearchResult[] = [{
        sessionId: 'session-1',
        projectPath: '/path/to/project',
        createdAt: now.toISOString(),
        updatedAt: new Date(now.getTime() - 5 * 60000).toISOString(), // 5 minutes ago
        tags: [],
        isFavorite: false,
        messageCount: 5,
        score: 0,
        matchReasons: [],
        status: 'idle',
      }];

      render(
        <SearchResults
          results={recentResult}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Should show relative time like "5m ago"
      expect(screen.getByText(/ago/)).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('should handle empty project path', () => {
      const resultWithEmptyPath: SearchResult[] = [{
        sessionId: 'session-1',
        projectPath: '',
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-15T15:30:00Z',
        tags: [],
        isFavorite: false,
        messageCount: 0,
        score: 0,
        matchReasons: [],
        status: 'idle',
      }];

      render(
        <SearchResults
          results={resultWithEmptyPath}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Should not crash and render the session
      expect(screen.getByText('0 messages')).toBeInTheDocument();
    });

    it('should handle very long tag lists', () => {
      const resultWithManyTags: SearchResult[] = [{
        sessionId: 'session-1',
        projectPath: '/project',
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-15T15:30:00Z',
        tags: ['tag1', 'tag2', 'tag3', 'tag4', 'tag5', 'tag6'],
        isFavorite: false,
        messageCount: 0,
        score: 0,
        matchReasons: [],
        status: 'idle',
      }];

      render(
        <SearchResults
          results={resultWithManyTags}
          onSelectSession={mockOnSelectSession}
          isLoading={false}
        />
      );

      // Should render all tags
      expect(screen.getByText('tag1')).toBeInTheDocument();
      expect(screen.getByText('tag6')).toBeInTheDocument();
    });
  });
});
