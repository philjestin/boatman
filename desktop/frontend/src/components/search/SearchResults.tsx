import { Star, Folder, MessageSquare, Calendar } from 'lucide-react';

export interface SearchResult {
  sessionId: string;
  projectPath: string;
  createdAt: string;
  updatedAt: string;
  tags: string[];
  isFavorite: boolean;
  messageCount: number;
  score: number;
  matchReasons: string[];
  status: string;
}

interface SearchResultsProps {
  results: SearchResult[];
  onSelectSession: (sessionId: string) => void;
  isLoading?: boolean;
}

export function SearchResults({
  results,
  onSelectSession,
  isLoading = false,
}: SearchResultsProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-sm text-slate-400">Searching...</p>
        </div>
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <MessageSquare className="w-12 h-12 text-slate-600 mx-auto mb-3" />
          <h3 className="text-sm font-medium text-slate-300 mb-1">No sessions found</h3>
          <p className="text-xs text-slate-500">
            Try adjusting your search or filters
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="divide-y divide-slate-700">
      {results.map((result) => (
        <SearchResultItem
          key={result.sessionId}
          result={result}
          onSelect={onSelectSession}
        />
      ))}
    </div>
  );
}

interface SearchResultItemProps {
  result: SearchResult;
  onSelect: (sessionId: string) => void;
}

function SearchResultItem({ result, onSelect }: SearchResultItemProps) {
  const projectName = result.projectPath.split('/').pop() || result.projectPath;
  const updatedDate = new Date(result.updatedAt);
  const relativeTime = getRelativeTime(updatedDate);

  return (
    <button
      onClick={() => onSelect(result.sessionId)}
      className="w-full px-4 py-3 text-left hover:bg-slate-800 transition-colors"
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2 flex-1 min-w-0">
          {/* Favorite Star */}
          {result.isFavorite && (
            <Star className="w-4 h-4 text-yellow-500 fill-yellow-500 flex-shrink-0" />
          )}

          {/* Project Name */}
          <div className="flex items-center gap-2 text-sm text-slate-200 truncate">
            <Folder className="w-4 h-4 text-slate-400 flex-shrink-0" />
            <span className="font-medium truncate">{projectName}</span>
          </div>
        </div>

        {/* Score Badge */}
        {result.score > 0 && (
          <span className="px-2 py-0.5 text-xs bg-blue-500/20 text-blue-400 rounded border border-blue-500/30 flex-shrink-0">
            {result.score}
          </span>
        )}
      </div>

      {/* Tags */}
      {result.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {result.tags.map((tag) => (
            <span
              key={tag}
              className="px-1.5 py-0.5 text-xs bg-slate-700 text-slate-300 rounded"
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Match Reasons */}
      {result.matchReasons.length > 0 && (
        <div className="mb-2">
          <p className="text-xs text-slate-400">
            {result.matchReasons.slice(0, 2).join(' • ')}
          </p>
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center gap-4 text-xs text-slate-500">
        <div className="flex items-center gap-1">
          <MessageSquare className="w-3 h-3" />
          <span>{result.messageCount} messages</span>
        </div>

        <div className="flex items-center gap-1">
          <Calendar className="w-3 h-3" />
          <span>{relativeTime}</span>
        </div>

        <div className={`px-1.5 py-0.5 rounded text-xs ${getStatusColor(result.status)}`}>
          {result.status}
        </div>
      </div>
    </button>
  );
}

function getRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)}mo ago`;
  return `${Math.floor(diffDays / 365)}y ago`;
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'running':
      return 'bg-green-500/20 text-green-400 border border-green-500/30';
    case 'idle':
      return 'bg-blue-500/20 text-blue-400 border border-blue-500/30';
    case 'error':
      return 'bg-red-500/20 text-red-400 border border-red-500/30';
    case 'stopped':
      return 'bg-slate-600/20 text-slate-400 border border-slate-600/30';
    default:
      return 'bg-slate-600/20 text-slate-400 border border-slate-600/30';
  }
}
