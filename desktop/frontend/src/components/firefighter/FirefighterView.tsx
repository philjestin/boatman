import { useState, useMemo, useCallback } from 'react';
import { ChatView } from '../chat/ChatView';
import { IncidentCanvas } from './IncidentCanvas';
import { FirefighterMonitor } from './FirefighterMonitor';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import { parseIncidentsFromMessages } from '../../utils/parseIncidents';
import type { Message, SessionStatus } from '../../types';

interface FirefighterViewProps {
  sessionId: string;
  messages: Message[];
  status: SessionStatus;
  onSendMessage: (content: string) => void;
  onStop?: () => void;
  isLoading?: boolean;
  hasMoreMessages?: boolean;
  onLoadMore?: () => void;
  isLoadingMore?: boolean;
  monitoringActive?: boolean;
  onToggleMonitoring?: (active: boolean) => void;
  model?: string;
  reasoningEffort?: string;
  onModelChange?: (model: string) => void;
  onReasoningEffortChange?: (effort: string) => void;
}

export function FirefighterView({
  sessionId,
  messages,
  status,
  onSendMessage,
  onStop,
  isLoading,
  hasMoreMessages,
  onLoadMore,
  isLoadingMore,
  monitoringActive,
  onToggleMonitoring,
  model,
  reasoningEffort,
  onModelChange,
  onReasoningEffortChange,
}: FirefighterViewProps) {
  const [showSidebar, setShowSidebar] = useState(true);
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(null);

  const incidents = useMemo(() => parseIncidentsFromMessages(messages), [messages]);

  const selectedIncident = useMemo(
    () => selectedIncidentId ? incidents.find(i => i.id === selectedIncidentId) ?? null : null,
    [incidents, selectedIncidentId],
  );

  // When an incident is selected, show only its related messages
  const displayMessages = useMemo(() => {
    if (!selectedIncident) return messages;
    const idSet = new Set(selectedIncident.messageIds);
    return messages.filter(m => idSet.has(m.id));
  }, [messages, selectedIncident]);

  const handleSelectIncident = useCallback((id: string) => {
    setSelectedIncidentId(prev => prev === id ? null : id);
  }, []);

  const handleInvestigate = useCallback((id: string) => {
    const incident = incidents.find(i => i.id === id);
    if (incident) {
      const prompt = `Investigate this incident: ${incident.title}${incident.linearId ? ` (${incident.linearId})` : ''}${incident.url ? ` - ${incident.url}` : ''}`;
      onSendMessage(prompt);
    }
  }, [incidents, onSendMessage]);

  return (
    <div className="flex flex-col h-full">
      {/* Monitoring Status Bar */}
      {monitoringActive !== undefined && onToggleMonitoring && (
        <div className="flex-shrink-0 border-b border-slate-700">
          <FirefighterMonitor
            sessionId={sessionId}
            isActive={monitoringActive}
            onToggle={onToggleMonitoring}
          />
        </div>
      )}

      {/* Main Content Area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Incident Sidebar */}
        {showSidebar && (
          <div className="w-96 border-r border-slate-700 flex flex-col bg-slate-900">
            <div className="flex-shrink-0 px-4 py-3 border-b border-slate-700 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-100">Incidents</h2>
              <button
                onClick={() => setShowSidebar(false)}
                className="p-1 text-slate-400 hover:text-slate-200 transition-colors"
                title="Hide incidents"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              <IncidentCanvas
                incidents={incidents}
                selectedId={selectedIncidentId}
                onSelect={handleSelectIncident}
                onInvestigate={handleInvestigate}
              />
            </div>
          </div>
        )}

        {/* Chat Area */}
        <div className="flex-1 flex flex-col relative">
          {!showSidebar && (
            <button
              onClick={() => setShowSidebar(true)}
              className="absolute top-4 left-4 z-10 p-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-400 hover:text-slate-200 hover:border-slate-600 transition-colors"
              title="Show incidents"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          )}

          {/* Incident context bar */}
          {selectedIncident && (
            <div className="flex-shrink-0 flex items-center gap-3 px-4 py-2 bg-slate-800 border-b border-slate-700">
              <span className="text-xs text-slate-400">Viewing:</span>
              <span className="text-sm font-medium text-slate-100 truncate">{selectedIncident.title}</span>
              {selectedIncident.linearId && (
                <span className="text-xs text-slate-500">{selectedIncident.linearId}</span>
              )}
              <button
                onClick={() => setSelectedIncidentId(null)}
                className="ml-auto flex items-center gap-1 px-2 py-1 text-xs text-slate-400 hover:text-slate-200 bg-slate-700 rounded transition-colors"
                title="Show all messages"
              >
                <X className="w-3 h-3" />
                Show all
              </button>
            </div>
          )}

          <ChatView
            messages={displayMessages}
            status={status}
            onSendMessage={onSendMessage}
            onStop={onStop}
            isLoading={isLoading}
            hasMoreMessages={selectedIncidentId ? false : hasMoreMessages}
            onLoadMore={onLoadMore}
            isLoadingMore={isLoadingMore}
            model={model}
            reasoningEffort={reasoningEffort}
            onModelChange={onModelChange}
            onReasoningEffortChange={onReasoningEffortChange}
          />
        </div>
      </div>
    </div>
  );
}
