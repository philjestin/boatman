import { useState } from 'react';
import { ChatView } from '../chat/ChatView';
import { FirefighterTicketList } from './FirefighterTicketList';
import { FirefighterMonitor } from './FirefighterMonitor';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import type { Message, SessionStatus } from '../../types';

interface FirefighterViewProps {
  sessionId: string;
  messages: Message[];
  status: SessionStatus;
  onSendMessage: (content: string) => void;
  isLoading?: boolean;
  hasMoreMessages?: boolean;
  onLoadMore?: () => void;
  isLoadingMore?: boolean;
  monitoringActive?: boolean;
  onToggleMonitoring?: () => void;
}

export function FirefighterView({
  sessionId,
  messages,
  status,
  onSendMessage,
  isLoading,
  hasMoreMessages,
  onLoadMore,
  isLoadingMore,
  monitoringActive,
  onToggleMonitoring,
}: FirefighterViewProps) {
  const [showTicketList, setShowTicketList] = useState(true);

  const handleInvestigateTicket = (ticketId: string) => {
    // The investigation has been triggered, chat will show the investigation
    console.log('Investigation started for ticket:', ticketId);
  };

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
        {/* Ticket List Sidebar */}
        {showTicketList && (
          <div className="w-96 border-r border-slate-700 flex flex-col bg-slate-900">
            <div className="flex-shrink-0 px-4 py-3 border-b border-slate-700 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-100">Linear Triage Queue</h2>
              <button
                onClick={() => setShowTicketList(false)}
                className="p-1 text-slate-400 hover:text-slate-200 transition-colors"
                title="Hide ticket list"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              <FirefighterTicketList
                sessionId={sessionId}
                onInvestigate={handleInvestigateTicket}
              />
            </div>
          </div>
        )}

        {/* Chat Area */}
        <div className="flex-1 flex flex-col relative">
          {!showTicketList && (
            <button
              onClick={() => setShowTicketList(true)}
              className="absolute top-4 left-4 z-10 p-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-400 hover:text-slate-200 hover:border-slate-600 transition-colors"
              title="Show ticket list"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          )}

          <ChatView
            messages={messages}
            status={status}
            onSendMessage={onSendMessage}
            isLoading={isLoading}
            hasMoreMessages={hasMoreMessages}
            onLoadMore={onLoadMore}
            isLoadingMore={isLoadingMore}
          />
        </div>
      </div>
    </div>
  );
}
