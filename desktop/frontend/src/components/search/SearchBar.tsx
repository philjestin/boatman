import { useState, useRef, useEffect } from 'react';
import { Search, X, Filter } from 'lucide-react';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  onSearch: () => void;
  onClear: () => void;
  onToggleFilters?: () => void;
  showFilters?: boolean;
  placeholder?: string;
  autoFocus?: boolean;
}

export function SearchBar({
  value,
  onChange,
  onSearch,
  onClear,
  onToggleFilters,
  showFilters = false,
  placeholder = 'Search sessions...',
  autoFocus = false,
}: SearchBarProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isFocused, setIsFocused] = useState(false);

  useEffect(() => {
    if (autoFocus && inputRef.current) {
      inputRef.current.focus();
    }
  }, [autoFocus]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      onSearch();
    } else if (e.key === 'Escape') {
      if (value) {
        onClear();
      } else {
        inputRef.current?.blur();
      }
    }
  };

  return (
    <div
      className={`flex items-center gap-2 px-4 py-2 bg-slate-800 border rounded-lg transition-colors ${
        isFocused ? 'border-blue-500' : 'border-slate-700'
      }`}
    >
      {/* Search Icon */}
      <Search className="w-4 h-4 text-slate-400 flex-shrink-0" />

      {/* Input */}
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        placeholder={placeholder}
        className="flex-1 bg-transparent text-sm text-slate-100 placeholder-slate-500 outline-none"
      />

      {/* Clear Button */}
      {value && (
        <button
          onClick={onClear}
          className="p-1 text-slate-400 hover:text-slate-200 transition-colors"
          title="Clear search (Esc)"
        >
          <X className="w-4 h-4" />
        </button>
      )}

      {/* Filter Toggle */}
      {onToggleFilters && (
        <button
          onClick={onToggleFilters}
          className={`p-1 transition-colors ${
            showFilters
              ? 'text-blue-400 hover:text-blue-300'
              : 'text-slate-400 hover:text-slate-200'
          }`}
          title="Toggle filters"
        >
          <Filter className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}
