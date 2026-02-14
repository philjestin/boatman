import { Check, X, AlertTriangle, FileEdit, Terminal } from 'lucide-react';

interface ApprovalBarProps {
  visible: boolean;
  onApprove: () => void;
  onReject: () => void;
  actionType?: 'edit' | 'bash' | 'other';
  actionDescription?: string;
  filePath?: string;
}

function getActionIcon(type: ApprovalBarProps['actionType']) {
  switch (type) {
    case 'edit':
      return <FileEdit className="w-5 h-5" />;
    case 'bash':
      return <Terminal className="w-5 h-5" />;
    default:
      return <AlertTriangle className="w-5 h-5" />;
  }
}

function getActionLabel(type: ApprovalBarProps['actionType']) {
  switch (type) {
    case 'edit':
      return 'File Edit';
    case 'bash':
      return 'Command';
    default:
      return 'Action';
  }
}

export function ApprovalBar({
  visible,
  onApprove,
  onReject,
  actionType = 'other',
  actionDescription,
  filePath,
}: ApprovalBarProps) {
  if (!visible) return null;

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 animate-in slide-in-from-bottom">
      <div className="bg-slate-800 border-t border-slate-600 shadow-lg">
        <div className="max-w-4xl mx-auto px-4 py-3">
          <div className="flex items-center justify-between gap-4">
            {/* Action info */}
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-yellow-500/20 text-yellow-500">
                {getActionIcon(actionType)}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-slate-100">
                    {getActionLabel(actionType)} requires approval
                  </span>
                  {filePath && (
                    <code className="text-xs px-1.5 py-0.5 bg-slate-700 rounded text-slate-300">
                      {filePath}
                    </code>
                  )}
                </div>
                {actionDescription && (
                  <p className="text-xs text-slate-400 mt-0.5">{actionDescription}</p>
                )}
              </div>
            </div>

            {/* Action buttons */}
            <div className="flex items-center gap-2">
              <button
                onClick={onReject}
                className="flex items-center gap-2 px-4 py-2 text-sm text-red-500 hover:bg-red-500/10 rounded-lg transition-colors"
              >
                <X className="w-4 h-4" />
                Reject
              </button>
              <button
                onClick={onApprove}
                className="flex items-center gap-2 px-4 py-2 text-sm bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
              >
                <Check className="w-4 h-4" />
                Approve
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Inline approval component for chat messages
interface InlineApprovalProps {
  onApprove: () => void;
  onReject: () => void;
  description?: string;
}

export function InlineApproval({ onApprove, onReject, description }: InlineApprovalProps) {
  return (
    <div className="mt-3 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
      <div className="flex items-center gap-2 text-yellow-500 mb-2">
        <AlertTriangle className="w-4 h-4" />
        <span className="text-sm font-medium">Approval Required</span>
      </div>
      {description && <p className="text-xs text-slate-300 mb-3">{description}</p>}
      <div className="flex items-center gap-2">
        <button
          onClick={onReject}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-red-500 hover:bg-red-500/10 rounded transition-colors"
        >
          <X className="w-3 h-3" />
          Reject
        </button>
        <button
          onClick={onApprove}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
        >
          <Check className="w-3 h-3" />
          Approve
        </button>
      </div>
    </div>
  );
}
