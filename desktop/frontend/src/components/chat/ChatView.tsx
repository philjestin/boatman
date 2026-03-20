import { useRef, useEffect, useCallback, useState } from 'react';
import { MessageBubble } from './MessageBubble';
import { InputArea } from './InputArea';
import { AgentLogsPanel } from './AgentLogsPanel';
import { ChatHeader } from './ChatHeader';
import { Loader2, StopCircle, ArrowDown } from 'lucide-react';
import type { Message, SessionStatus } from '../../types';

interface ChatViewProps {
  messages: Message[];
  status: SessionStatus;
  onSendMessage: (content: string) => void;
  onStop?: () => void;
  isLoading?: boolean;
  hasMoreMessages?: boolean;
  onLoadMore?: () => void;
  isLoadingMore?: boolean;
  model?: string;
  reasoningEffort?: string;
  onModelChange?: (model: string) => void;
  onReasoningEffortChange?: (effort: string) => void;
  projectPath?: string;
  mode?: string;
}

export function ChatView({
  messages,
  status,
  onSendMessage,
  onStop,
  isLoading = false,
  hasMoreMessages = false,
  onLoadMore,
  isLoadingMore = false,
  model,
  reasoningEffort,
  onModelChange,
  onReasoningEffortChange,
  projectPath,
  mode,
}: ChatViewProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);
  const [showScrollButton, setShowScrollButton] = useState(false);

  const checkIfNearBottom = useCallback(() => {
    const el = scrollContainerRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 150;
    isNearBottomRef.current = nearBottom;
    setShowScrollButton(!nearBottom);
  }, []);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    if (isNearBottomRef.current) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  // Debug: log messages
  useEffect(() => {
    console.log('[ChatView] Messages updated:', {
      total: messages.length,
      roles: messages.map(m => ({ id: m.id, role: m.role, hasContent: !!m.content, contentLen: m.content?.length || 0 }))
    });
  }, [messages]);

  const getStatusMessage = () => {
    switch (status) {
      case 'running':
        return 'Claude is thinking...';
      case 'waiting':
        return 'Waiting for approval...';
      case 'error':
        return 'An error occurred';
      case 'stopped':
        return 'Session stopped';
      default:
        return null;
    }
  };

  const statusMessage = getStatusMessage();
  const isInputDisabled = status === 'running' || isLoading;

  return (
    <div className="flex flex-col h-full">
      {/* Claude Chat Header */}
      {projectPath && (
        <ChatHeader
          projectPath={projectPath}
          model={model}
          reasoningEffort={reasoningEffort}
          mode={mode}
          messageCount={messages.length}
        />
      )}

      {/* Messages */}
      <div className="flex-1 overflow-y-auto relative" ref={scrollContainerRef} onScroll={checkIfNearBottom}>
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <h3 className="text-lg font-medium text-slate-200 mb-2">
                Start a conversation
              </h3>
              <p className="text-slate-400 text-sm max-w-md mb-3">
                Ask Claude to help you with your code. You can ask questions, request changes,
                or get explanations about your project.
              </p>
              <p className="text-slate-500 text-xs max-w-md">
                Type <code className="px-1 py-0.5 bg-slate-800 rounded text-slate-400">/help</code> for commands
                {' '}&bull;{' '}
                <code className="px-1 py-0.5 bg-slate-800 rounded text-slate-400">/clear</code> to reset
                {' '}&bull;{' '}
                <code className="px-1 py-0.5 bg-slate-800 rounded text-slate-400">/model</code> to switch models
              </p>
            </div>
          </div>
        ) : (
          <div className="py-4">
            {/* Load More Button */}
            {hasMoreMessages && onLoadMore && (
              <div className="flex justify-center py-4">
                <button
                  onClick={onLoadMore}
                  disabled={isLoadingMore}
                  className="px-4 py-2 text-sm font-medium text-blue-400 bg-slate-800 rounded-lg hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isLoadingMore ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Loading earlier messages...
                    </span>
                  ) : (
                    'Load Earlier Messages'
                  )}
                </button>
              </div>
            )}

            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}

        {/* Scroll to bottom button */}
        {showScrollButton && messages.length > 0 && (
          <button
            onClick={scrollToBottom}
            className="sticky bottom-4 left-1/2 -translate-x-1/2 z-10 flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-slate-700 text-slate-200 rounded-full shadow-lg hover:bg-slate-600 transition-colors border border-slate-600"
          >
            <ArrowDown className="w-3.5 h-3.5" />
            New messages
          </button>
        )}
      </div>

      {/* Status indicator */}
      {statusMessage && (
        <div className="flex items-center justify-center gap-2 py-3 text-sm text-blue-400 bg-slate-800/50">
          {status === 'running' && <Loader2 className="w-4 h-4 animate-spin text-blue-400" />}
          <span>{statusMessage}</span>
          {status === 'running' && onStop && (
            <button
              onClick={onStop}
              className="ml-2 p-1 text-red-400 hover:text-red-300 hover:bg-slate-700 rounded transition-colors"
              aria-label="Stop session"
            >
              <StopCircle className="w-4 h-4" />
            </button>
          )}
        </div>
      )}

      {/* Agent Logs Panel */}
      <AgentLogsPanel messages={messages} isActive={status === 'running'} />

      {/* Input */}
      <InputArea
        onSend={onSendMessage}
        onStop={onStop}
        status={status}
        disabled={isInputDisabled}
        placeholder={
          status === 'waiting'
            ? 'Waiting for approval...'
            : status === 'running'
            ? 'Claude is thinking...'
            : 'Message Claude... (type /help for commands)'
        }
        model={model}
        reasoningEffort={reasoningEffort}
        onModelChange={onModelChange}
        onReasoningEffortChange={onReasoningEffortChange}
      />
    </div>
  );
}
