import { X, FileText, AlertCircle, GitBranch, Lightbulb } from 'lucide-react';
import type { Task } from '../../types';

interface TaskDetailModalProps {
  task: Task;
  onClose: () => void;
}

export function TaskDetailModal({ task, onClose }: TaskDetailModalProps) {
  const metadata = task.metadata || {};

  const hasDiff = metadata.diff && typeof metadata.diff === 'string';
  const hasFeedback = metadata.feedback && typeof metadata.feedback === 'string';
  const hasIssues = Array.isArray(metadata.issues) && metadata.issues.length > 0;
  const hasPlan = metadata.plan && typeof metadata.plan === 'string';
  const hasRefactorDiff = metadata.refactor_diff && typeof metadata.refactor_diff === 'string';

  const hasContent = hasDiff || hasFeedback || hasIssues || hasPlan || hasRefactorDiff;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-4xl max-h-[90vh] bg-slate-900 rounded-lg border border-slate-700 shadow-2xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700">
          <div className="flex-1">
            <h2 className="text-lg font-semibold text-slate-100">{task.subject}</h2>
            {task.description && (
              <p className="mt-1 text-sm text-slate-400">{task.description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-slate-800 transition-colors"
            aria-label="Close"
          >
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {!hasContent && (
            <div className="text-center py-12 text-slate-500">
              <FileText className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>No details available for this task yet.</p>
            </div>
          )}

          {hasPlan && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <Lightbulb className="w-4 h-4 text-yellow-500" />
                <span>Plan</span>
              </div>
              <pre className="bg-slate-800 rounded-lg p-4 text-xs text-slate-300 overflow-x-auto whitespace-pre-wrap font-mono">
                {metadata.plan}
              </pre>
            </div>
          )}

          {hasDiff && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <GitBranch className="w-4 h-4 text-blue-500" />
                <span>Execution Diff</span>
              </div>
              <pre className="bg-slate-800 rounded-lg p-4 text-xs text-slate-300 overflow-x-auto whitespace-pre-wrap font-mono">
                {metadata.diff}
              </pre>
            </div>
          )}

          {hasFeedback && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <FileText className="w-4 h-4 text-purple-500" />
                <span>Review Feedback</span>
              </div>
              <div className="bg-slate-800 rounded-lg p-4 text-sm text-slate-300 whitespace-pre-wrap">
                {metadata.feedback}
              </div>
            </div>
          )}

          {hasRefactorDiff && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <GitBranch className="w-4 h-4 text-green-500" />
                <span>Refactor Diff</span>
              </div>
              <pre className="bg-slate-800 rounded-lg p-4 text-xs text-slate-300 overflow-x-auto whitespace-pre-wrap font-mono">
                {metadata.refactor_diff}
              </pre>
            </div>
          )}

          {hasIssues && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <AlertCircle className="w-4 h-4 text-red-500" />
                <span>Issues Found</span>
              </div>
              <div className="space-y-2">
                {metadata.issues.map((issue: any, idx: number) => (
                  <div
                    key={idx}
                    className="bg-slate-800 rounded-lg p-3 text-sm text-slate-300"
                  >
                    {typeof issue === 'string' ? issue : JSON.stringify(issue, null, 2)}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-700 flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-100 rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
