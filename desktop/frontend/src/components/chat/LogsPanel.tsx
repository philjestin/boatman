import { useState, useEffect, useRef, useMemo } from 'react';
import { Terminal, ChevronDown, ChevronUp, Trash2 } from 'lucide-react';
import type { Message } from '../../types';

interface LogsPanelProps {
  messages: Message[];
  isActive: boolean;
}

interface LogEntry {
  timestamp: Date;
  type: 'system' | 'tool_use' | 'tool_result' | 'assistant' | 'user' | 'event';
  content: string;
  metadata?: any;
}

export function LogsPanel({ messages, isActive }: LogsPanelProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);

  // Memoize log conversion to avoid rebuilding on every render
  const logs = useMemo(() => {
    console.log('[LogsPanel] Messages updated:', messages.length, 'messages');

    // Convert messages to log entries
    const newLogs: LogEntry[] = messages.map(msg => ({
      timestamp: new Date(msg.timestamp),
      type: msg.role === 'system' ? 'system' :
            msg.metadata?.toolUse ? 'tool_use' :
            msg.metadata?.toolResult ? 'tool_result' :
            msg.role as any,
      content: msg.content,
      metadata: msg.metadata,
    }));

    console.log('[LogsPanel] Converted to logs:', newLogs.length, 'entries');
    return newLogs;
  }, [messages]);

  // Memoize activity indicator separately to avoid rebuilding logs
  const logsWithActivity = useMemo(() => {
    if (isActive && logs.length > 0) {
      return [...logs, {
        timestamp: new Date(),
        type: 'event' as const,
        content: '⚡ Claude is working...',
      }];
    }
    return logs;
  }, [logs, isActive]);

  useEffect(() => {
    // Auto-scroll to bottom
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logsWithActivity]);

  const getLogColor = (type: string) => {
    switch (type) {
      case 'user': return 'text-blue-400';
      case 'assistant': return 'text-purple-400';
      case 'tool_use': return 'text-amber-400';
      case 'tool_result': return 'text-green-400';
      case 'system': return 'text-slate-400';
      case 'event': return 'text-cyan-400';
      default: return 'text-slate-300';
    }
  };

  const getLogPrefix = (type: string) => {
    switch (type) {
      case 'user': return '[USER]';
      case 'assistant': return '[ASSISTANT]';
      case 'tool_use': return '[TOOL→]';
      case 'tool_result': return '[←RESULT]';
      case 'system': return '[SYSTEM]';
      case 'event': return '[EVENT]';
      default: return '[LOG]';
    }
  };

  return (
    <div className="border-t border-slate-700 bg-slate-950">
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between px-4 py-2 bg-slate-900 hover:bg-slate-800 transition-colors"
      >
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-slate-400" />
          <span className="text-sm font-medium text-slate-300">
            Activity Logs
          </span>
          {isActive && (
            <span className="flex items-center gap-1 text-xs text-cyan-400">
              <span className="inline-block w-2 h-2 bg-cyan-400 rounded-full animate-pulse" />
              Active
            </span>
          )}
          <span className="text-xs text-slate-500">({logsWithActivity.length} entries)</span>
        </div>
        <div className="flex items-center gap-2">
          {/* Clear button removed since logs are now derived from messages */}
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-slate-400" />
          ) : (
            <ChevronUp className="w-4 h-4 text-slate-400" />
          )}
        </div>
      </button>

      {/* Logs content */}
      {isExpanded && (
        <div className="h-64 overflow-y-auto bg-slate-950 p-3 font-mono text-xs">
          {logsWithActivity.length === 0 ? (
            <div className="flex items-center justify-center h-full text-slate-600">
              No activity yet. Send a message to see logs.
            </div>
          ) : (
            <div className="space-y-1">
              {logsWithActivity.map((log, index) => (
                <div key={index} className="flex gap-2">
                  <span className="text-slate-600 flex-shrink-0">
                    {log.timestamp.toLocaleTimeString()}
                  </span>
                  <span className={`flex-shrink-0 ${getLogColor(log.type)}`}>
                    {getLogPrefix(log.type)}
                  </span>
                  <span className="text-slate-300 break-all">
                    {log.content.length > 200
                      ? log.content.substring(0, 200) + '...'
                      : log.content}
                  </span>
                </div>
              ))}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
