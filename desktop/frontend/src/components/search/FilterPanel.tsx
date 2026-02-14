import { useState } from 'react';
import { Calendar, Folder, Star, Tag, X } from 'lucide-react';

export interface SearchFilters {
  tags: string[];
  projectPath: string;
  isFavorite?: boolean;
  fromDate: string;
  toDate: string;
}

interface FilterPanelProps {
  filters: SearchFilters;
  onChange: (filters: SearchFilters) => void;
  availableTags: string[];
  availableProjects: string[];
}

export function FilterPanel({
  filters,
  onChange,
  availableTags,
  availableProjects,
}: FilterPanelProps) {
  const [newTag, setNewTag] = useState('');
  const [showTagInput, setShowTagInput] = useState(false);

  const handleAddTag = (tag: string) => {
    if (tag && !filters.tags.includes(tag)) {
      onChange({
        ...filters,
        tags: [...filters.tags, tag],
      });
    }
    setNewTag('');
    setShowTagInput(false);
  };

  const handleRemoveTag = (tag: string) => {
    onChange({
      ...filters,
      tags: filters.tags.filter((t) => t !== tag),
    });
  };

  const handleFavoriteToggle = () => {
    onChange({
      ...filters,
      isFavorite: filters.isFavorite === true ? undefined : true,
    });
  };

  const handleClearFilters = () => {
    onChange({
      tags: [],
      projectPath: '',
      isFavorite: undefined,
      fromDate: '',
      toDate: '',
    });
  };

  const hasActiveFilters =
    filters.tags.length > 0 ||
    filters.projectPath ||
    filters.isFavorite !== undefined ||
    filters.fromDate ||
    filters.toDate;

  return (
    <div className="p-4 bg-slate-900 border-b border-slate-700 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-slate-300">Filters</h3>
        {hasActiveFilters && (
          <button
            onClick={handleClearFilters}
            className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
          >
            Clear all
          </button>
        )}
      </div>

      {/* Tags Filter */}
      <div>
        <label className="flex items-center gap-2 text-xs font-medium text-slate-400 mb-2">
          <Tag className="w-3 h-3" />
          Tags
        </label>

        {/* Selected Tags */}
        {filters.tags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2">
            {filters.tags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-blue-500/20 text-blue-400 rounded border border-blue-500/30"
              >
                {tag}
                <button
                  onClick={() => handleRemoveTag(tag)}
                  className="hover:text-blue-300"
                >
                  <X className="w-3 h-3" />
                </button>
              </span>
            ))}
          </div>
        )}

        {/* Tag Input */}
        {showTagInput ? (
          <div className="flex gap-2">
            <input
              type="text"
              value={newTag}
              onChange={(e) => setNewTag(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleAddTag(newTag);
                } else if (e.key === 'Escape') {
                  setShowTagInput(false);
                  setNewTag('');
                }
              }}
              placeholder="Enter tag name..."
              className="flex-1 px-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-100 placeholder-slate-500 outline-none focus:border-blue-500"
              autoFocus
            />
            <button
              onClick={() => handleAddTag(newTag)}
              className="px-3 py-1.5 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
            >
              Add
            </button>
          </div>
        ) : (
          <div className="space-y-1">
            {/* Available Tags */}
            {availableTags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {availableTags
                  .filter((tag) => !filters.tags.includes(tag))
                  .slice(0, 8)
                  .map((tag) => (
                    <button
                      key={tag}
                      onClick={() => handleAddTag(tag)}
                      className="px-2 py-1 text-xs bg-slate-800 text-slate-300 rounded border border-slate-700 hover:border-blue-500 hover:text-blue-400 transition-colors"
                    >
                      {tag}
                    </button>
                  ))}
              </div>
            )}

            <button
              onClick={() => setShowTagInput(true)}
              className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
            >
              + Add custom tag
            </button>
          </div>
        )}
      </div>

      {/* Project Filter */}
      <div>
        <label className="flex items-center gap-2 text-xs font-medium text-slate-400 mb-2">
          <Folder className="w-3 h-3" />
          Project
        </label>
        <select
          value={filters.projectPath}
          onChange={(e) => onChange({ ...filters, projectPath: e.target.value })}
          className="w-full px-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-100 outline-none focus:border-blue-500"
        >
          <option value="">All projects</option>
          {availableProjects.map((project) => (
            <option key={project} value={project}>
              {project.split('/').pop() || project}
            </option>
          ))}
        </select>
      </div>

      {/* Favorite Filter */}
      <div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={filters.isFavorite === true}
            onChange={handleFavoriteToggle}
            className="w-4 h-4 rounded"
          />
          <Star className="w-3 h-3 text-slate-400" />
          <span className="text-xs text-slate-300">Favorites only</span>
        </label>
      </div>

      {/* Date Range Filter */}
      <div>
        <label className="flex items-center gap-2 text-xs font-medium text-slate-400 mb-2">
          <Calendar className="w-3 h-3" />
          Date Range
        </label>
        <div className="grid grid-cols-2 gap-2">
          <div>
            <label className="block text-xs text-slate-500 mb-1">From</label>
            <input
              type="date"
              value={filters.fromDate}
              onChange={(e) => onChange({ ...filters, fromDate: e.target.value })}
              className="w-full px-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-100 outline-none focus:border-blue-500"
            />
          </div>
          <div>
            <label className="block text-xs text-slate-500 mb-1">To</label>
            <input
              type="date"
              value={filters.toDate}
              onChange={(e) => onChange({ ...filters, toDate: e.target.value })}
              className="w-full px-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-100 outline-none focus:border-blue-500"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
