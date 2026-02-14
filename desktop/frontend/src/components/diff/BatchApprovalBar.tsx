import { Check, X, FileCheck } from 'lucide-react';

interface BatchApprovalBarProps {
  selectedCount: number;
  totalCount: number;
  onApproveSelected: () => void;
  onRejectSelected: () => void;
  onClearSelection: () => void;
}

export function BatchApprovalBar({
  selectedCount,
  totalCount,
  onApproveSelected,
  onRejectSelected,
  onClearSelection,
}: BatchApprovalBarProps) {
  if (selectedCount === 0) {
    return null;
  }

  return (
    <div className="fixed bottom-0 left-0 right-0 bg-slate-800 border-t border-slate-700 px-4 py-3 shadow-lg">
      <div className="flex items-center justify-between max-w-7xl mx-auto">
        <div className="flex items-center gap-3">
          <FileCheck className="w-5 h-5 text-blue-500" />
          <span className="text-sm text-slate-200">
            <span className="font-medium">{selectedCount}</span> of{' '}
            <span className="font-medium">{totalCount}</span> file
            {totalCount > 1 ? 's' : ''} selected
          </span>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={onClearSelection}
            className="px-3 py-1.5 text-sm text-slate-400 hover:text-slate-200 rounded transition-colors"
          >
            Clear
          </button>
          <button
            onClick={onRejectSelected}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-red-500 hover:bg-red-500/10 rounded transition-colors"
          >
            <X className="w-4 h-4" />
            Reject Selected
          </button>
          <button
            onClick={onApproveSelected}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
          >
            <Check className="w-4 h-4" />
            Approve Selected
          </button>
        </div>
      </div>
    </div>
  );
}
