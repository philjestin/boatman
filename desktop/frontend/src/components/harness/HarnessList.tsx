import { RefreshCw, FolderOpen, FileCode, Package } from 'lucide-react';
import type { HarnessInfo } from '../../types';

interface HarnessListProps {
  harnesses: HarnessInfo[];
  onSelect: (harness: HarnessInfo) => void;
  onRefresh: () => void;
  isLoading: boolean;
}

export function HarnessList({ harnesses, onSelect, onRefresh, isLoading }: HarnessListProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-slate-300">Harness Projects</h3>
        <button
          onClick={onRefresh}
          disabled={isLoading}
          className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-slate-400 hover:text-slate-200 hover:bg-slate-700 rounded-md transition-colors disabled:opacity-50"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {harnesses.length === 0 ? (
        <div className="text-center py-8 text-slate-500">
          <FolderOpen className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p className="text-sm">No harness projects found</p>
          <p className="text-xs mt-1">Generate one using the Build tab</p>
        </div>
      ) : (
        <div className="space-y-2">
          {harnesses.map((harness) => (
            <button
              key={harness.path}
              onClick={() => onSelect(harness)}
              className="w-full text-left p-3 bg-slate-800/50 border border-slate-700 rounded-lg hover:border-slate-600 hover:bg-slate-800 transition-colors"
            >
              <div className="flex items-center gap-2 mb-1">
                <FolderOpen className="w-4 h-4 text-blue-400 flex-shrink-0" />
                <span className="text-sm font-medium text-slate-200">{harness.name}</span>
              </div>
              <p className="text-xs font-mono text-slate-500 ml-6 truncate">{harness.path}</p>
              <div className="flex items-center gap-3 mt-2 ml-6">
                <span className={`flex items-center gap-1 text-xs ${harness.hasGoMod ? 'text-green-400' : 'text-slate-600'}`}>
                  <Package className="w-3 h-3" />
                  go.mod
                </span>
                <span className={`flex items-center gap-1 text-xs ${harness.hasMain ? 'text-green-400' : 'text-slate-600'}`}>
                  <FileCode className="w-3 h-3" />
                  main.go
                </span>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
