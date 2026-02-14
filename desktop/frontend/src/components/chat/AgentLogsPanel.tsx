import { useState, useEffect, useRef, useMemo } from 'react';
import { Terminal, ChevronDown, ChevronUp, Trash2, GitBranch } from 'lucide-react';
import type { Message, AgentInfo } from '../../types';

interface AgentLogsPanelProps {
  messages: Message[];
  isActive: boolean;
}

interface LogEntry {
  timestamp: Date;
  type: 'system' | 'tool_use' | 'tool_result' | 'assistant' | 'user' | 'event';
  content: string;
  agentId: string;
  metadata?: any;
}

interface AgentData {
  info: AgentInfo;
  logs: LogEntry[];
}

export function AgentLogsPanel({ messages, isActive }: AgentLogsPanelProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [activeAgentId, setActiveAgentId] = useState<string>('main');
  const [showHierarchy, setShowHierarchy] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);

  // Memoize agent map construction to avoid rebuilding on every render
  const agents = useMemo(() => {
    console.log('[AgentLogsPanel] Processing messages:', messages.length);

    // Group messages by agent
    const agentMap = new Map<string, AgentData>();

    messages.forEach(msg => {
      const agentId = msg.metadata?.agent?.agentId || 'main';
      const agentInfo = msg.metadata?.agent || {
        agentId: 'main',
        agentType: 'main',
      };

      // Initialize agent if not exists
      if (!agentMap.has(agentId)) {
        agentMap.set(agentId, {
          info: agentInfo,
          logs: [],
        });
      }

      // Add log entry
      const logEntry: LogEntry = {
        timestamp: new Date(msg.timestamp),
        type: msg.role === 'system' ? 'system' :
              msg.metadata?.toolUse ? 'tool_use' :
              msg.metadata?.toolResult ? 'tool_result' :
              msg.role as any,
        content: msg.content,
        agentId,
        metadata: msg.metadata,
      };

      agentMap.get(agentId)!.logs.push(logEntry);
    });

    console.log('[AgentLogsPanel] Agents found:', Array.from(agentMap.keys()));
    return agentMap;
  }, [messages]);

  // Memoize agents with activity indicator separately
  const agentsWithActivity = useMemo(() => {
    if (!isActive || agents.size === 0) return agents;

    const agentMapCopy = new Map(agents);
    agentMapCopy.forEach((data, agentId) => {
      agentMapCopy.set(agentId, {
        ...data,
        logs: [...data.logs, {
          timestamp: new Date(),
          type: 'event' as const,
          content: '⚡ Working...',
          agentId,
        }],
      });
    });

    return agentMapCopy;
  }, [agents, isActive]);

  // Auto-select first agent if current selection doesn't exist
  useEffect(() => {
    if (!agentsWithActivity.has(activeAgentId) && agentsWithActivity.size > 0) {
      setActiveAgentId(agentsWithActivity.keys().next().value);
    }
  }, [agentsWithActivity, activeAgentId]);

  useEffect(() => {
    // Auto-scroll to bottom
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [agentsWithActivity, activeAgentId]);

  const getAgentColor = (agentType: string) => {
    switch (agentType) {
      case 'main': return 'text-blue-400 border-blue-500';
      case 'task': return 'text-purple-400 border-purple-500';
      case 'Explore': return 'text-green-400 border-green-500';
      case 'Plan': return 'text-amber-400 border-amber-500';
      case 'general-purpose': return 'text-cyan-400 border-cyan-500';
      default: return 'text-slate-400 border-slate-500';
    }
  };

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

  const renderAgentHierarchy = () => {
    const mainAgent = agentsWithActivity.get('main');
    if (!mainAgent) return null;

    const subAgents = Array.from(agentsWithActivity.entries())
      .filter(([id]) => id !== 'main')
      .map(([id, data]) => ({ id, ...data }));

    return (
      <div className="p-3 bg-slate-900 border-b border-slate-700 text-xs font-mono">
        <div className="flex items-center gap-2 text-blue-400">
          <GitBranch className="w-3 h-3" />
          <span className="font-bold">Main Agent</span>
          <span className="text-slate-500">({mainAgent.logs.length} events)</span>
        </div>
        {subAgents.length > 0 && (
          <div className="ml-4 mt-2 space-y-1">
            {subAgents.map(agent => {
              const color = getAgentColor(agent.info.agentType);
              return (
                <div key={agent.id} className="flex items-center gap-2">
                  <span className="text-slate-600">└─</span>
                  <span className={color.split(' ')[0]}>
                    {agent.info.agentType || 'Agent'}
                  </span>
                  {agent.info.description && (
                    <span className="text-slate-500">
                      ({agent.info.description.substring(0, 40)}...)
                    </span>
                  )}
                  <span className="text-slate-600">
                    ({agent.logs.length} events)
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </div>
    );
  };

  const currentAgent = agentsWithActivity.get(activeAgentId);
  const currentLogs = currentAgent?.logs || [];

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
            Agent Activity
          </span>
          {isActive && (
            <span className="flex items-center gap-1 text-xs text-cyan-400">
              <span className="inline-block w-2 h-2 bg-cyan-400 rounded-full animate-pulse" />
              Active
            </span>
          )}
          <span className="text-xs text-slate-500">
            ({agentsWithActivity.size} agent{agentsWithActivity.size !== 1 ? 's' : ''})
          </span>
        </div>
        <div className="flex items-center gap-2">
          {agentsWithActivity.size > 1 && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                setShowHierarchy(!showHierarchy);
              }}
              className={`p-1 transition-colors ${
                showHierarchy ? 'text-blue-400' : 'text-slate-500 hover:text-slate-300'
              }`}
              title="Show hierarchy"
            >
              <GitBranch className="w-3 h-3" />
            </button>
          )}
          {/* Clear button removed since logs are now derived from messages */}
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-slate-400" />
          ) : (
            <ChevronUp className="w-4 h-4 text-slate-400" />
          )}
        </div>
      </button>

      {/* Hierarchy view */}
      {isExpanded && showHierarchy && renderAgentHierarchy()}

      {/* Agent tabs */}
      {isExpanded && agentsWithActivity.size > 1 && (
        <div className="flex gap-1 px-2 py-2 bg-slate-900 border-b border-slate-700 overflow-x-auto">
          {Array.from(agentsWithActivity.entries()).map(([agentId, data]) => {
            const color = getAgentColor(data.info.agentType);
            const isActive = agentId === activeAgentId;
            return (
              <button
                key={agentId}
                onClick={() => setActiveAgentId(agentId)}
                className={`px-3 py-1 text-xs rounded transition-colors border ${
                  isActive
                    ? `${color} bg-opacity-10`
                    : 'text-slate-500 border-slate-700 hover:border-slate-600'
                }`}
              >
                {data.info.agentType === 'main' ? '⭐ Main' : `🤖 ${data.info.agentType || 'Agent'}`}
                <span className="ml-1 text-slate-600">({data.logs.length})</span>
              </button>
            );
          })}
        </div>
      )}

      {/* Logs content */}
      {isExpanded && (
        <div className="h-64 overflow-y-auto bg-slate-950 p-3 font-mono text-xs">
          {currentLogs.length === 0 ? (
            <div className="flex items-center justify-center h-full text-slate-600">
              No activity yet. Send a message to see logs.
            </div>
          ) : (
            <div className="space-y-1">
              {currentLogs.map((log, index) => (
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
