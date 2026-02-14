import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SearchModal } from './SearchModal';
import type { SearchResult } from './SearchResults';

describe('SearchModal', () => {
  const mockOnClose = vi.fn();
  const mockOnSelectSession = vi.fn();
  const mockOnSearch = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('Modal visibility', () => {
    it('should not render when isOpen is false', () => {
      render(
        <SearchModal
          isOpen={false}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.queryByText('Search Sessions')).not.toBeInTheDocument();
    });

    it('should render when isOpen is true', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Search Sessions')).toBeInTheDocument();
    });
  });

  describe('Modal closing', () => {
    it('should call onClose when X button is clicked', () => {
      const { container } = render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const closeButton = container.querySelector('.lucide-x')?.parentElement;
      if (closeButton) {
        fireEvent.click(closeButton);
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when backdrop is clicked', () => {
      const { container } = render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const backdrop = container.querySelector('.bg-black\\/60');
      if (backdrop) {
        fireEvent.click(backdrop);
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when Escape key is pressed', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      fireEvent.keyDown(window, { key: 'Escape' });

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Search bar integration', () => {
    it('should render search bar', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByPlaceholderText('Search sessions...')).toBeInTheDocument();
    });

    it('should update search query when typing', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test query' } });

      expect(input).toHaveValue('test query');
    });

    it('should debounce search by 300ms', async () => {
      mockOnSearch.mockResolvedValue([]);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');

      // Type query
      fireEvent.change(input, { target: { value: 'test' } });

      // Search should not be called immediately
      expect(mockOnSearch).not.toHaveBeenCalled();

      // Wait for debounce and search to complete
      await waitFor(() => {
        expect(mockOnSearch).toHaveBeenCalledWith('test', expect.any(Object));
      }, { timeout: 1000 });
    });

    it('should cancel previous search when query changes', async () => {
      mockOnSearch.mockResolvedValue([]);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');

      // Type first query
      fireEvent.change(input, { target: { value: 'test' } });

      // Type second query quickly before debounce completes
      await new Promise(resolve => setTimeout(resolve, 100));
      fireEvent.change(input, { target: { value: 'test2' } });

      // Wait for debounce and verify only latest query was searched
      await waitFor(() => {
        expect(mockOnSearch).toHaveBeenCalledTimes(1);
        expect(mockOnSearch).toHaveBeenCalledWith('test2', expect.any(Object));
      }, { timeout: 1000 });
    });
  });

  describe('Filter panel integration', () => {
    it('should not show filters by default', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.queryByText('Filters')).not.toBeInTheDocument();
    });

    it('should toggle filter panel when filter button clicked', async () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.getByText('Filters')).toBeInTheDocument();
      });

      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.queryByText('Filters')).not.toBeInTheDocument();
      });
    });

    it('should pass available tags to filter panel', async () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={['frontend', 'backend']}
          availableProjects={[]}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.getByText('frontend')).toBeInTheDocument();
        expect(screen.getByText('backend')).toBeInTheDocument();
      });
    });

    it('should pass available projects to filter panel', async () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={['/path/to/project1', '/path/to/project2']}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.getByText('project1')).toBeInTheDocument();
        expect(screen.getByText('project2')).toBeInTheDocument();
      });
    });
  });

  describe('Search results integration', () => {
    it('should show empty state initially', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Search your sessions')).toBeInTheDocument();
    });

    it('should display search results', async () => {
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

      mockOnSearch.mockResolvedValue(mockResults);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        expect(screen.getByText('project1')).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should show loading state during search', async () => {
      mockOnSearch.mockImplementation(() => new Promise(resolve => setTimeout(() => resolve([]), 500)));

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        expect(screen.getByText('Searching...')).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should show result count in footer', async () => {
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
        {
          sessionId: 'session-2',
          projectPath: '/path/to/project2',
          createdAt: '2024-01-15T10:00:00Z',
          updatedAt: '2024-01-15T15:30:00Z',
          tags: [],
          isFavorite: false,
          messageCount: 5,
          score: 0,
          matchReasons: [],
          status: 'idle',
        },
      ];

      mockOnSearch.mockResolvedValue(mockResults);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        expect(screen.getByText('Found 2 sessions')).toBeInTheDocument();
      }, { timeout: 1000 });
    });

    it('should show singular form for one result', async () => {
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

      mockOnSearch.mockResolvedValue(mockResults);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        expect(screen.getByText('Found 1 session')).toBeInTheDocument();
      }, { timeout: 1000 });
    });
  });

  describe('Session selection', () => {
    it('should select session and close modal when result clicked', async () => {
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

      mockOnSearch.mockResolvedValue(mockResults);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        expect(screen.getByText('project1')).toBeInTheDocument();
      }, { timeout: 1000 });

      const resultButton = screen.getByText('project1').closest('button');
      if (resultButton) {
        fireEvent.click(resultButton);
      }

      expect(mockOnSelectSession).toHaveBeenCalledWith('session-1');
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Clear functionality', () => {
    it('should clear search when clear button clicked', async () => {
      mockOnSearch.mockResolvedValue([]);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      await waitFor(() => {
        const clearButton = screen.getByTitle('Clear search (Esc)');
        expect(clearButton).toBeInTheDocument();
      }, { timeout: 1000 });

      const clearButton = screen.getByTitle('Clear search (Esc)');
      fireEvent.click(clearButton);

      expect(input).toHaveValue('');
    });

    it('should clear filters and results', async () => {
      mockOnSearch.mockResolvedValue([]);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={['frontend']}
          availableProjects={[]}
        />
      );

      // Open filters
      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.getByText('Filters')).toBeInTheDocument();
      });

      // Add a tag filter
      const tagButton = screen.getByText('frontend');
      fireEvent.click(tagButton);

      // Add search query
      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'test' } });

      // Wait for search to complete
      await new Promise(resolve => setTimeout(resolve, 400));

      // Click clear
      const clearButton = await screen.findByTitle('Clear search (Esc)');
      fireEvent.click(clearButton);

      expect(input).toHaveValue('');
    });
  });

  describe('Keyboard shortcuts hint', () => {
    it('should show keyboard shortcuts in empty state', () => {
      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Enter')).toBeInTheDocument();
      expect(screen.getByText('to search')).toBeInTheDocument();
      expect(screen.getByText('Esc')).toBeInTheDocument();
      expect(screen.getByText('to close')).toBeInTheDocument();
    });
  });

  describe('Filter-triggered search', () => {
    it('should search when filters change', async () => {
      mockOnSearch.mockResolvedValue([]);

      render(
        <SearchModal
          isOpen={true}
          onClose={mockOnClose}
          onSelectSession={mockOnSelectSession}
          onSearch={mockOnSearch}
          availableTags={['frontend']}
          availableProjects={[]}
        />
      );

      // Open filters
      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      await waitFor(() => {
        expect(screen.getByText('Filters')).toBeInTheDocument();
      });

      // Add a tag filter
      const tagButton = screen.getByText('frontend');
      fireEvent.click(tagButton);

      await waitFor(() => {
        expect(mockOnSearch).toHaveBeenCalledWith('', expect.objectContaining({
          tags: ['frontend'],
        }));
      }, { timeout: 1000 });
    });
  });
});
