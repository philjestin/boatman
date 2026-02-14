import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SearchBar } from './SearchBar';

describe('SearchBar', () => {
  const mockOnChange = vi.fn();
  const mockOnSearch = vi.fn();
  const mockOnClear = vi.fn();
  const mockOnToggleFilters = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render search input', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      expect(input).toBeInTheDocument();
    });

    it('should display current value', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      expect(input).toHaveValue('test query');
    });

    it('should auto-focus when autoFocus is true', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
          autoFocus={true}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      expect(input).toHaveFocus();
    });
  });

  describe('Input changes', () => {
    it('should call onChange when typing', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.change(input, { target: { value: 'new query' } });

      expect(mockOnChange).toHaveBeenCalledWith('new query');
    });
  });

  describe('Search functionality', () => {
    it('should call onSearch when Enter key is pressed', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockOnSearch).toHaveBeenCalledTimes(1);
    });

    it('should not call onSearch for other keys', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.keyDown(input, { key: 'a' });

      expect(mockOnSearch).not.toHaveBeenCalled();
    });
  });

  describe('Clear functionality', () => {
    it('should call onClear when Escape key is pressed', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const input = screen.getByPlaceholderText('Search sessions...');
      fireEvent.keyDown(input, { key: 'Escape' });

      expect(mockOnClear).toHaveBeenCalledTimes(1);
    });

    it('should show clear button when value is not empty', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const clearButton = screen.getByTitle('Clear search (Esc)');
      expect(clearButton).toBeInTheDocument();
    });

    it('should not show clear button when value is empty', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const clearButton = screen.queryByTitle('Clear search (Esc)');
      expect(clearButton).not.toBeInTheDocument();
    });

    it('should call onClear when clear button is clicked', () => {
      render(
        <SearchBar
          value="test query"
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const clearButton = screen.getByTitle('Clear search (Esc)');
      fireEvent.click(clearButton);

      expect(mockOnClear).toHaveBeenCalledTimes(1);
    });
  });

  describe('Filter toggle', () => {
    it('should render filter toggle button', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      expect(filterButton).toBeInTheDocument();
    });

    it('should call onToggleFilters when filter button is clicked', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      fireEvent.click(filterButton);

      expect(mockOnToggleFilters).toHaveBeenCalledTimes(1);
    });

    it('should highlight filter button when showFilters is true', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={true}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      expect(filterButton).toHaveClass('text-blue-400');
    });

    it('should not highlight filter button when showFilters is false', () => {
      render(
        <SearchBar
          value=""
          onChange={mockOnChange}
          onSearch={mockOnSearch}
          onClear={mockOnClear}
          onToggleFilters={mockOnToggleFilters}
          showFilters={false}
        />
      );

      const filterButton = screen.getByTitle('Toggle filters');
      expect(filterButton).not.toHaveClass('text-blue-400');
      expect(filterButton).toHaveClass('text-slate-400');
    });
  });
});
