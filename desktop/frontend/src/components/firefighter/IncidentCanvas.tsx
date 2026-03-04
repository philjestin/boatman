import { useState } from 'react';
import { CheckCircle, ChevronDown, ChevronRight } from 'lucide-react';
import { IncidentCard } from './IncidentCard';
import type { Incident } from '../../types';

interface IncidentCanvasProps {
  incidents: Incident[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onInvestigate: (id: string) => void;
}

export function IncidentCanvas({ incidents, selectedId, onSelect, onInvestigate }: IncidentCanvasProps) {
  const [showResolved, setShowResolved] = useState(false);

  const active = incidents.filter(i => !['resolved', 'failed'].includes(i.status));
  const resolved = incidents.filter(i => ['resolved', 'failed'].includes(i.status));

  if (incidents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 px-4 border border-dashed border-slate-700 rounded-lg">
        <CheckCircle className="w-12 h-12 text-green-500 mb-3" />
        <h3 className="text-lg font-medium text-slate-100 mb-1">All Clear</h3>
        <p className="text-sm text-slate-400">No incidents detected</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {/* Header counts */}
      <div className="text-xs text-slate-400 px-1 mb-3">
        {active.length} Active{resolved.length > 0 ? ` · ${resolved.length} Resolved` : ''}
      </div>

      {/* Active incidents */}
      {active.map(incident => (
        <IncidentCard
          key={incident.id}
          incident={incident}
          isSelected={selectedId === incident.id}
          onSelect={onSelect}
          onInvestigate={onInvestigate}
        />
      ))}

      {/* Resolved section (collapsible) */}
      {resolved.length > 0 && (
        <div className="pt-2">
          <button
            onClick={() => setShowResolved(!showResolved)}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200 transition-colors mb-2"
          >
            {showResolved ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
            Resolved ({resolved.length})
          </button>
          {showResolved && (
            <div className="space-y-2">
              {resolved.map(incident => (
                <IncidentCard
                  key={incident.id}
                  incident={incident}
                  isSelected={selectedId === incident.id}
                  onSelect={onSelect}
                  onInvestigate={onInvestigate}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
