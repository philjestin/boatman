import { useState, useEffect } from 'react';
import { Flame, AlertCircle, Clock, CheckCircle, PlayCircle, ExternalLink } from 'lucide-react';
import { InvestigateLinearTicket } from '../../../wailsjs/go/main/App';

interface LinearTicket {
  id: string;
  identifier: string; // e.g., "EMP-123"
  title: string;
  description: string;
  priority: number; // 0=None, 1=Urgent, 2=High, 3=Medium, 4=Low
  status: string;
  labels: string[];
  url: string;
  createdAt: string;
  updatedAt: string;
}

interface FirefighterTicketListProps {
  sessionId: string;
  onInvestigate?: (ticketId: string) => void;
}

export function FirefighterTicketList({ sessionId, onInvestigate }: FirefighterTicketListProps) {
  const [tickets, setTickets] = useState<LinearTicket[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [investigating, setInvestigating] = useState<string | null>(null);

  useEffect(() => {
    // In a real implementation, this would call a backend method
    // that uses the Linear MCP tool to fetch tickets
    // For now, this is a placeholder
    const fetchTickets = async () => {
      try {
        setLoading(true);
        setError(null);

        // TODO: Implement backend method GetLinearTriageTickets
        // const tickets = await GetLinearTriageTickets();

        // Mock data for now
        const mockTickets: LinearTicket[] = [
          {
            id: '1',
            identifier: 'EMP-456',
            title: 'Production error: NullPointerException in payment service',
            description: 'Bugsnag error: https://app.bugsnag.com/...',
            priority: 1, // Urgent
            status: 'Triage',
            labels: ['firefighter', 'production', 'payment'],
            url: 'https://linear.app/...',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
          {
            id: '2',
            identifier: 'EMP-457',
            title: 'Datadog alert: High error rate in auth service',
            description: 'Datadog monitor: https://app.datadoghq.com/...',
            priority: 2, // High
            status: 'Triage',
            labels: ['firefighter', 'auth', 'monitoring'],
            url: 'https://linear.app/...',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        ];

        setTickets(mockTickets);
      } catch (err: any) {
        console.error('Failed to fetch tickets:', err);
        setError(err.toString());
      } finally {
        setLoading(false);
      }
    };

    fetchTickets();

    // Refresh every 30 seconds
    const interval = setInterval(fetchTickets, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleInvestigate = async (ticket: LinearTicket) => {
    try {
      setInvestigating(ticket.id);
      await InvestigateLinearTicket(sessionId, ticket.id);
      onInvestigate?.(ticket.id);
    } catch (err: any) {
      console.error('Failed to investigate ticket:', err);
      alert(`Failed to investigate: ${err}`);
    } finally {
      setInvestigating(null);
    }
  };

  const getPriorityColor = (priority: number) => {
    switch (priority) {
      case 1: return 'text-red-500';
      case 2: return 'text-orange-500';
      case 3: return 'text-yellow-500';
      case 4: return 'text-blue-500';
      default: return 'text-slate-500';
    }
  };

  const getPriorityLabel = (priority: number) => {
    switch (priority) {
      case 1: return 'Urgent';
      case 2: return 'High';
      case 3: return 'Medium';
      case 4: return 'Low';
      default: return 'None';
    }
  };

  const getPriorityIcon = (priority: number) => {
    switch (priority) {
      case 1: return <Flame className="w-4 h-4" />;
      case 2: return <AlertCircle className="w-4 h-4" />;
      default: return <Clock className="w-4 h-4" />;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full"></div>
        <span className="ml-3 text-sm text-slate-400">Loading triage queue...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
        <p className="text-sm text-red-400">Failed to load tickets: {error}</p>
      </div>
    );
  }

  if (tickets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 px-4 border border-dashed border-slate-700 rounded-lg">
        <CheckCircle className="w-12 h-12 text-green-500 mb-3" />
        <h3 className="text-lg font-medium text-slate-100 mb-1">All Clear!</h3>
        <p className="text-sm text-slate-400">No tickets in triage queue</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-slate-100">
          Triage Queue ({tickets.length})
        </h3>
        <button
          onClick={() => window.location.reload()}
          className="text-xs text-slate-400 hover:text-slate-200 transition-colors"
        >
          Refresh
        </button>
      </div>

      {tickets.map((ticket) => (
        <div
          key={ticket.id}
          className="p-4 bg-slate-800 border border-slate-700 rounded-lg hover:border-slate-600 transition-colors"
        >
          <div className="flex items-start justify-between gap-3">
            <div className="flex-1 min-w-0">
              {/* Header */}
              <div className="flex items-center gap-2 mb-2">
                <span className={`flex items-center gap-1 text-xs font-medium ${getPriorityColor(ticket.priority)}`}>
                  {getPriorityIcon(ticket.priority)}
                  {getPriorityLabel(ticket.priority)}
                </span>
                <span className="text-xs text-slate-500">•</span>
                <span className="text-xs text-slate-400">{ticket.identifier}</span>
                <span className="text-xs text-slate-500">•</span>
                <span className="px-2 py-0.5 text-xs bg-slate-700 text-slate-300 rounded">
                  {ticket.status}
                </span>
              </div>

              {/* Title */}
              <h4 className="text-sm font-medium text-slate-100 mb-2">
                {ticket.title}
              </h4>

              {/* Description Preview */}
              <p className="text-xs text-slate-400 line-clamp-2 mb-3">
                {ticket.description}
              </p>

              {/* Labels */}
              {ticket.labels.length > 0 && (
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {ticket.labels.map((label) => (
                    <span
                      key={label}
                      className="px-2 py-0.5 text-xs bg-blue-500/10 text-blue-400 rounded-full"
                    >
                      {label}
                    </span>
                  ))}
                </div>
              )}

              {/* Actions */}
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleInvestigate(ticket)}
                  disabled={investigating === ticket.id}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {investigating === ticket.id ? (
                    <>
                      <div className="w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      Investigating...
                    </>
                  ) : (
                    <>
                      <PlayCircle className="w-3 h-3" />
                      Investigate
                    </>
                  )}
                </button>

                <a
                  href={ticket.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-slate-300 hover:text-slate-100 transition-colors"
                >
                  <ExternalLink className="w-3 h-3" />
                  Open in Linear
                </a>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
