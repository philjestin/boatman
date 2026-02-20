import { useRef, useEffect } from 'react';
import { X, FileText, AlertCircle, GitBranch, Lightbulb, CheckCircle, Loader2, Clock, Terminal, MessageSquare, Wrench } from 'lucide-react';
import type { Task, Message } from '../../types';

interface TaskDetailModalProps {
  task: Task;
  messages: Message[];
  onClose: () => void;
}

export function TaskDetailModal({ task, messages, onClose }: TaskDetailModalProps) {
  const metadata = task.metadata || {};
  const activityEndRef = useRef<HTMLDivElement>(null);

  const hasDiff = metadata.diff && typeof metadata.diff === 'string';
  const hasFeedback = metadata.feedback && typeof metadata.feedback === 'string';
  const hasIssues = Array.isArray(metadata.issues) && metadata.issues.length > 0;
  const hasPlan = metadata.plan && typeof metadata.plan === 'string';
  const hasRefactorDiff = metadata.refactor_diff && typeof metadata.refactor_diff === 'string';

  const hasMetadataContent = hasDiff || hasFeedback || hasIssues || hasPlan || hasRefactorDiff;

  // Filter messages by agent ID matching this task's ID
  const taskMessages = messages.filter(
    (m) => m.metadata?.agent?.agentId === task.id
  );

  // Get log entries from task metadata (backend fallback)
  const logEntries: string[] = Array.isArray(metadata.log) ? metadata.log : [];

  const hasActivity = taskMessages.length > 0 || logEntries.length > 0;
  const hasContent = hasMetadataContent || hasActivity;

  // Auto-scroll to bottom for in-progress tasks
  useEffect(() => {
    if (task.status === 'in_progress' && activityEndRef.current) {
      activityEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [task.status, taskMessages.length, logEntries.length]);

  const statusIcon = () => {
    switch (task.status) {
      case 'completed':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'in_progress':
        return <Loader2 className="w-5 h-5 text-blue-400 animate-spin" />;
      default:
        return <Clock className="w-5 h-5 text-slate-500" />;
    }
  };

  const statusLabel = () => {
    switch (task.status) {
      case 'completed':
        return 'Completed';
      case 'in_progress':
        return 'In Progress';
      default:
        return 'Pending';
    }
  };

  const renderActivityEntry = (msg: Message, idx: number) => {
    // Tool use messages
    if (msg.metadata?.toolUse) {
      return (
        <div key={msg.id || idx} className="flex items-start gap-2 py-1.5 text-sm">
          <Wrench className="w-3.5 h-3.5 text-blue-400 mt-0.5 shrink-0" />
          <span className="text-slate-300">{msg.content}</span>
        </div>
      );
    }

    // Tool result messages
    if (msg.metadata?.toolResult) {
      const isError = msg.metadata.toolResult.isError;
      return (
        <div key={msg.id || idx} className="flex items-start gap-2 py-1.5 text-sm">
          <Terminal className="w-3.5 h-3.5 text-slate-500 mt-0.5 shrink-0" />
          <span className={`${isError ? 'text-red-400' : 'text-slate-500'} truncate`}>
            {msg.content.length > 200 ? msg.content.slice(0, 200) + '...' : msg.content}
          </span>
        </div>
      );
    }

    // Cost info messages — skip in activity log
    if (msg.metadata?.costInfo) {
      return null;
    }

    // System messages
    if (msg.role === 'system') {
      return (
        <div key={msg.id || idx} className="flex items-start gap-2 py-1.5 text-sm">
          <MessageSquare className="w-3.5 h-3.5 text-yellow-500 mt-0.5 shrink-0" />
          <span className="text-slate-400">{msg.content}</span>
        </div>
      );
    }

    // Assistant messages
    if (msg.role === 'assistant') {
      const displayContent = msg.content.length > 300
        ? msg.content.slice(0, 300) + '...'
        : msg.content;
      return (
        <div key={msg.id || idx} className="py-1.5 text-sm">
          <div className="text-slate-300 whitespace-pre-wrap">{displayContent}</div>
        </div>
      );
    }

    return null;
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-4xl max-h-[90vh] bg-slate-900 rounded-lg border border-slate-700 shadow-2xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700">
          <div className="flex-1">
            <div className="flex items-center gap-2">
              {statusIcon()}
              <h2 className="text-lg font-semibold text-slate-100">{task.subject}</h2>
            </div>
            <div className="mt-1 flex items-center gap-2">
              <span className="text-xs px-2 py-0.5 rounded-full bg-slate-700 text-slate-300">
                {statusLabel()}
              </span>
              {task.description && (
                <p className="text-sm text-slate-400">{task.description}</p>
              )}
            </div>
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
          {!hasContent && task.status === 'in_progress' && (
            <div className="text-center py-12 text-slate-500">
              <Loader2 className="w-12 h-12 mx-auto mb-3 opacity-50 animate-spin" />
              <p>This task is currently running.</p>
              <p className="text-xs mt-1">Details will appear when activity begins.</p>
            </div>
          )}

          {!hasContent && task.status === 'pending' && (
            <div className="text-center py-12 text-slate-500">
              <Clock className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>This task is waiting to start.</p>
            </div>
          )}

          {!hasContent && task.status === 'completed' && (
            <div className="text-center py-12 text-slate-500">
              <CheckCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>This step completed successfully.</p>
            </div>
          )}

          {/* Metadata sections — shown above the activity log when present */}
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

          {/* Activity Log */}
          {hasActivity && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                <Terminal className="w-4 h-4 text-slate-400" />
                <span>Activity</span>
                <span className="text-xs text-slate-500">
                  ({taskMessages.length + logEntries.length} entries)
                </span>
              </div>
              <div className="bg-slate-800 rounded-lg p-3 max-h-96 overflow-y-auto divide-y divide-slate-700/50">
                {/* Backend fallback log entries (shown first as they're typically early progress) */}
                {logEntries.map((entry, idx) => (
                  <div key={`log-${idx}`} className="flex items-start gap-2 py-1.5 text-sm">
                    <MessageSquare className="w-3.5 h-3.5 text-yellow-500 mt-0.5 shrink-0" />
                    <span className="text-slate-400">{String(entry)}</span>
                  </div>
                ))}
                {/* Agent-attributed messages */}
                {taskMessages.map((msg, idx) => renderActivityEntry(msg, idx))}
                <div ref={activityEndRef} />
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
