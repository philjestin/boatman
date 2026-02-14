import { useState, useEffect, useCallback } from 'react';
import { X } from 'lucide-react';
import { SearchBar } from './SearchBar';
import { FilterPanel, type SearchFilters } from './FilterPanel';
import { SearchResults, type SearchResult } from './SearchResults';

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectSession: (sessionId: string) => void;
  onSearch: (query: string, filters: SearchFilters) => Promise<SearchResult[]>;
  availableTags: string[];
  availableProjects: string[];
}

export function SearchModal({
  isOpen,
  onClose,
  onSelectSession,
  onSearch,
  availableTags,
  availableProjects,
}: SearchModalProps) {
  const [query, setQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState<SearchFilters>({
    tags: [],
    projectPath: '',
    isFavorite: undefined,
    fromDate: '',
    toDate: '',
  });
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  // Perform search
  const handleSearch = useCallback(async () => {
    setIsLoading(true);
    setHasSearched(true);
    try {
      const searchResults = await onSearch(query, filters);
      setResults(searchResults);
    } catch (error) {
      console.error('Search failed:', error);
      setResults([]);
    } finally {
      setIsLoading(false);
    }
  }, [query, filters, onSearch]);

  // Auto-search when query or filters change (debounced)
  useEffect(() => {
    if (!isOpen) return;

    const timer = setTimeout(() => {
      if (query || hasActiveFilters(filters)) {
        handleSearch();
      } else {
        setResults([]);
        setHasSearched(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query, filters, isOpen, handleSearch]);

  // Handle session selection
  const handleSelectSession = (sessionId: string) => {
    onSelectSession(sessionId);
    onClose();
  };

  // Clear search
  const handleClear = () => {
    setQuery('');
    setFilters({
      tags: [],
      projectPath: '',
      isFavorite: undefined,
      fromDate: '',
      toDate: '',
    });
    setResults([]);
    setHasSearched(false);
  };

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Escape to close
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-20">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-slate-900 rounded-xl shadow-2xl w-full max-w-3xl max-h-[80vh] overflow-hidden border border-slate-700 flex flex-col">
        {/* Header */}
        <div className="flex-shrink-0 px-4 pt-4 pb-3 border-b border-slate-700">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold text-slate-100">Search Sessions</h2>
            <button
              onClick={onClose}
              className="p-1 text-slate-400 hover:text-slate-200 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <SearchBar
            value={query}
            onChange={setQuery}
            onSearch={handleSearch}
            onClear={handleClear}
            onToggleFilters={() => setShowFilters(!showFilters)}
            showFilters={showFilters}
            autoFocus
          />
        </div>

        {/* Filters */}
        {showFilters && (
          <div className="flex-shrink-0">
            <FilterPanel
              filters={filters}
              onChange={setFilters}
              availableTags={availableTags}
              availableProjects={availableProjects}
            />
          </div>
        )}

        {/* Results */}
        <div className="flex-1 overflow-y-auto">
          {hasSearched || query || hasActiveFilters(filters) ? (
            <SearchResults
              results={results}
              onSelectSession={handleSelectSession}
              isLoading={isLoading}
            />
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center px-4">
                <h3 className="text-sm font-medium text-slate-300 mb-2">
                  Search your sessions
                </h3>
                <p className="text-xs text-slate-500 max-w-md">
                  Search by message content, tags, project, or use filters to find
                  specific sessions
                </p>
                <div className="mt-4 space-y-2 text-xs text-slate-500">
                  <p>
                    <kbd className="px-2 py-1 bg-slate-800 rounded border border-slate-700">
                      Enter
                    </kbd>{' '}
                    to search
                  </p>
                  <p>
                    <kbd className="px-2 py-1 bg-slate-800 rounded border border-slate-700">
                      Esc
                    </kbd>{' '}
                    to close
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        {results.length > 0 && (
          <div className="flex-shrink-0 px-4 py-2 border-t border-slate-700 bg-slate-900">
            <p className="text-xs text-slate-500">
              Found {results.length} {results.length === 1 ? 'session' : 'sessions'}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

function hasActiveFilters(filters: SearchFilters): boolean {
  return (
    filters.tags.length > 0 ||
    filters.projectPath !== '' ||
    filters.isFavorite !== undefined ||
    filters.fromDate !== '' ||
    filters.toDate !== ''
  );
}
