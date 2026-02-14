import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FilterPanel, type SearchFilters } from './FilterPanel';

describe('FilterPanel', () => {
  const mockOnChange = vi.fn();
  const defaultFilters: SearchFilters = {
    tags: [],
    projectPath: '',
    isFavorite: undefined,
    fromDate: '',
    toDate: '',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render filter panel with title', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Filters')).toBeInTheDocument();
    });

    it('should render all filter sections', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Tags')).toBeInTheDocument();
      expect(screen.getByText('Project')).toBeInTheDocument();
      expect(screen.getByText('Favorites only')).toBeInTheDocument();
      expect(screen.getByText('Date Range')).toBeInTheDocument();
    });
  });

  describe('Tag filter', () => {
    it('should display available tags', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={['frontend', 'backend', 'bug-fix']}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('frontend')).toBeInTheDocument();
      expect(screen.getByText('backend')).toBeInTheDocument();
      expect(screen.getByText('bug-fix')).toBeInTheDocument();
    });

    it('should limit available tags to 8', () => {
      const manyTags = Array.from({ length: 15 }, (_, i) => `tag${i}`);

      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={manyTags}
          availableProjects={[]}
        />
      );

      // First 8 should be visible
      expect(screen.getByText('tag0')).toBeInTheDocument();
      expect(screen.getByText('tag7')).toBeInTheDocument();
      // 9th tag should not be visible
      expect(screen.queryByText('tag8')).not.toBeInTheDocument();
    });

    it('should add tag when clicked', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={['frontend']}
          availableProjects={[]}
        />
      );

      const tagButton = screen.getByText('frontend');
      fireEvent.click(tagButton);

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        tags: ['frontend'],
      });
    });

    it('should show selected tags separately', () => {
      const filtersWithTags = {
        ...defaultFilters,
        tags: ['frontend', 'backend'],
      };

      render(
        <FilterPanel
          filters={filtersWithTags}
          onChange={mockOnChange}
          availableTags={['frontend', 'backend', 'api']}
          availableProjects={[]}
        />
      );

      // Selected tags should appear with an X button
      const selectedTags = screen.getAllByText('frontend');
      expect(selectedTags.length).toBeGreaterThan(0);
    });

    it('should remove tag when X is clicked', () => {
      const filtersWithTags = {
        ...defaultFilters,
        tags: ['frontend', 'backend'],
      };

      const { container } = render(
        <FilterPanel
          filters={filtersWithTags}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      // Find the X button for frontend tag
      const xButtons = container.querySelectorAll('.lucide-x');
      if (xButtons[0]) {
        fireEvent.click(xButtons[0].parentElement!);
      }

      expect(mockOnChange).toHaveBeenCalledWith({
        ...filtersWithTags,
        tags: ['backend'],
      });
    });

    it('should show custom tag input when button clicked', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const addButton = screen.getByText('+ Add custom tag');
      fireEvent.click(addButton);

      expect(screen.getByPlaceholderText('Enter tag name...')).toBeInTheDocument();
    });

    it('should add custom tag on Enter key', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const addButton = screen.getByText('+ Add custom tag');
      fireEvent.click(addButton);

      const input = screen.getByPlaceholderText('Enter tag name...');
      fireEvent.change(input, { target: { value: 'custom-tag' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        tags: ['custom-tag'],
      });
    });

    it('should cancel custom tag input on Escape key', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const addButton = screen.getByText('+ Add custom tag');
      fireEvent.click(addButton);

      const input = screen.getByPlaceholderText('Enter tag name...');
      fireEvent.change(input, { target: { value: 'custom-tag' } });
      fireEvent.keyDown(input, { key: 'Escape' });

      expect(screen.queryByPlaceholderText('Enter tag name...')).not.toBeInTheDocument();
      expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('should not add duplicate tags', () => {
      const filtersWithTags = {
        ...defaultFilters,
        tags: ['frontend'],
      };

      render(
        <FilterPanel
          filters={filtersWithTags}
          onChange={mockOnChange}
          availableTags={['frontend']}
          availableProjects={[]}
        />
      );

      // Frontend tag should not be in available tags list since it's selected
      const availableFrontendButtons = screen.queryAllByText('frontend')
        .filter(el => el.tagName === 'BUTTON' && !el.querySelector('.lucide-x'));

      // Should only show selected tag, not in available list
      expect(availableFrontendButtons.length).toBe(0);
    });
  });

  describe('Project filter', () => {
    it('should render project dropdown', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={['/path/to/project1', '/path/to/project2']}
        />
      );

      const select = screen.getByRole('combobox');
      expect(select).toBeInTheDocument();
    });

    it('should show all projects option', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={['/path/to/project1']}
        />
      );

      expect(screen.getByText('All projects')).toBeInTheDocument();
    });

    it('should display project names from paths', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={['/path/to/project1', '/path/to/project2']}
        />
      );

      expect(screen.getByText('project1')).toBeInTheDocument();
      expect(screen.getByText('project2')).toBeInTheDocument();
    });

    it('should update project filter when changed', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={['/path/to/project1']}
        />
      );

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: '/path/to/project1' } });

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        projectPath: '/path/to/project1',
      });
    });
  });

  describe('Favorite filter', () => {
    it('should render favorites checkbox', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).toBeInTheDocument();
      expect(screen.getByText('Favorites only')).toBeInTheDocument();
    });

    it('should toggle favorite filter when clicked', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      fireEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        isFavorite: true,
      });
    });

    it('should uncheck when already checked', () => {
      const filtersWithFavorite = {
        ...defaultFilters,
        isFavorite: true,
      };

      render(
        <FilterPanel
          filters={filtersWithFavorite}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).toBeChecked();

      fireEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith({
        ...filtersWithFavorite,
        isFavorite: undefined,
      });
    });
  });

  describe('Date range filter', () => {
    it('should render from and to date inputs', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('From')).toBeInTheDocument();
      expect(screen.getByText('To')).toBeInTheDocument();
    });

    it('should update fromDate when changed', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const dateInputs = screen.getAllByDisplayValue('');
      const fromDateInput = dateInputs[0]; // First date input is "From"

      fireEvent.change(fromDateInput, { target: { value: '2024-01-01' } });

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        fromDate: '2024-01-01',
      });
    });

    it('should update toDate when changed', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const dateInputs = screen.getAllByDisplayValue('');
      const toDateInput = dateInputs[1]; // Second date input is "To"

      fireEvent.change(toDateInput, { target: { value: '2024-12-31' } });

      expect(mockOnChange).toHaveBeenCalledWith({
        ...defaultFilters,
        toDate: '2024-12-31',
      });
    });

    it('should display existing date values', () => {
      const filtersWithDates = {
        ...defaultFilters,
        fromDate: '2024-01-01',
        toDate: '2024-12-31',
      };

      render(
        <FilterPanel
          filters={filtersWithDates}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByDisplayValue('2024-01-01')).toBeInTheDocument();
      expect(screen.getByDisplayValue('2024-12-31')).toBeInTheDocument();
    });
  });

  describe('Clear filters', () => {
    it('should show clear all button when filters are active', () => {
      const filtersWithData = {
        tags: ['frontend'],
        projectPath: '/test',
        isFavorite: true,
        fromDate: '2024-01-01',
        toDate: '2024-12-31',
      };

      render(
        <FilterPanel
          filters={filtersWithData}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.getByText('Clear all')).toBeInTheDocument();
    });

    it('should not show clear all button when no filters active', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      expect(screen.queryByText('Clear all')).not.toBeInTheDocument();
    });

    it('should clear all filters when clicked', () => {
      const filtersWithData = {
        tags: ['frontend'],
        projectPath: '/test',
        isFavorite: true,
        fromDate: '2024-01-01',
        toDate: '2024-12-31',
      };

      render(
        <FilterPanel
          filters={filtersWithData}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      const clearButton = screen.getByText('Clear all');
      fireEvent.click(clearButton);

      expect(mockOnChange).toHaveBeenCalledWith({
        tags: [],
        projectPath: '',
        isFavorite: undefined,
        fromDate: '',
        toDate: '',
      });
    });
  });

  describe('Edge cases', () => {
    it('should handle empty available tags array', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      // Should show add custom tag button
      expect(screen.getByText('+ Add custom tag')).toBeInTheDocument();
    });

    it('should handle empty available projects array', () => {
      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[]}
        />
      );

      // Should show "All projects" option
      expect(screen.getByText('All projects')).toBeInTheDocument();
    });

    it('should handle very long project paths', () => {
      const longPath = '/very/long/path/to/project/with/many/nested/folders/projectname';

      render(
        <FilterPanel
          filters={defaultFilters}
          onChange={mockOnChange}
          availableTags={[]}
          availableProjects={[longPath]}
        />
      );

      // Should extract last part
      expect(screen.getByText('projectname')).toBeInTheDocument();
    });
  });
});
