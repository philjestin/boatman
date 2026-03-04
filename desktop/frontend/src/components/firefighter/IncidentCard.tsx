import { Flame, AlertCircle, Monitor, MessageSquare, PlayCircle, Clock } from 'lucide-react';
import type { Incident } from '../../types';

interface IncidentCardProps {
  incident: Incident;
  isSelected: boolean;
  onSelect: (id: string) => void;
  onInvestigate: (id: string) => void;
}

const severityConfig: Record<Incident['severity'], { border: string; text: string; bg: string; label: string }> = {
  urgent: { border: 'border-l-red-500', text: 'text-red-500', bg: 'bg-red-500/10', label: 'Urgent' },
  high: { border: 'border-l-orange-500', text: 'text-orange-500', bg: 'bg-orange-500/10', label: 'High' },
  medium: { border: 'border-l-yellow-500', text: 'text-yellow-500', bg: 'bg-yellow-500/10', label: 'Medium' },
  low: { border: 'border-l-blue-500', text: 'text-blue-500', bg: 'bg-blue-500/10', label: 'Low' },
};

const statusConfig: Record<Incident['status'], { text: string; bg: string; label: string }> = {
  new: { text: 'text-red-400', bg: 'bg-red-500/10', label: 'New' },
  investigating: { text: 'text-yellow-400', bg: 'bg-yellow-500/10', label: 'Investigating' },
  fixing: { text: 'text-orange-400', bg: 'bg-orange-500/10', label: 'Fixing' },
  testing: { text: 'text-blue-400', bg: 'bg-blue-500/10', label: 'Testing' },
  resolved: { text: 'text-green-400', bg: 'bg-green-500/10', label: 'Resolved' },
  failed: { text: 'text-red-400', bg: 'bg-red-500/10', label: 'Failed' },
};

function SourceIcon({ source }: { source: Incident['source'] }) {
  switch (source) {
    case 'linear': return <Flame className="w-3.5 h-3.5" />;
    case 'bugsnag': return <AlertCircle className="w-3.5 h-3.5" />;
    case 'datadog': return <Monitor className="w-3.5 h-3.5" />;
    case 'slack': return <MessageSquare className="w-3.5 h-3.5" />;
  }
}

function timeAgo(timestamp: string): string {
  const seconds = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function IncidentCard({ incident, isSelected, onSelect, onInvestigate }: IncidentCardProps) {
  const sev = severityConfig[incident.severity];
  const stat = statusConfig[incident.status];
  const canInvestigate = incident.status === 'new';

  return (
    <div
      onClick={() => onSelect(incident.id)}
      className={`p-3 border-l-4 ${sev.border} bg-slate-800 border border-slate-700 rounded-r-lg cursor-pointer transition-colors ${
        isSelected ? 'ring-1 ring-blue-500 border-slate-600' : 'hover:border-slate-600'
      }`}
    >
      {/* Header: severity + source + time */}
      <div className="flex items-center gap-2 mb-1.5">
        <span className={`flex items-center gap-1 text-xs font-medium ${sev.text}`}>
          <SourceIcon source={incident.source} />
          {sev.label}
        </span>
        <span className={`px-1.5 py-0.5 text-xs rounded ${stat.bg} ${stat.text}`}>
          {stat.label}
        </span>
        <span className="ml-auto flex items-center gap-1 text-xs text-slate-500">
          <Clock className="w-3 h-3" />
          {timeAgo(incident.lastUpdated)}
        </span>
      </div>

      {/* Title */}
      <h4 className="text-sm font-medium text-slate-100 mb-1 line-clamp-2">
        {incident.title}
      </h4>

      {/* Linear ID + message count */}
      <div className="flex items-center gap-2 text-xs text-slate-500">
        {incident.linearId && <span>{incident.linearId}</span>}
        <span>{incident.messageIds.length} message{incident.messageIds.length !== 1 ? 's' : ''}</span>
      </div>

      {/* Investigate button */}
      {canInvestigate && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onInvestigate(incident.id);
          }}
          className="mt-2 flex items-center gap-1.5 px-3 py-1.5 text-xs bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors"
        >
          <PlayCircle className="w-3 h-3" />
          Investigate
        </button>
      )}
    </div>
  );
}
