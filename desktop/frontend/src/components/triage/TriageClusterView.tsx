import { PlayCircle, FileCode, AlertTriangle, ClipboardCheck } from 'lucide-react';
import type { TriageCluster, TriageContextDoc } from '../../types';

interface TriageClusterViewProps {
  clusters: TriageCluster[];
  contextDocs: TriageContextDoc[];
  onTicketClick?: (ticketId: string) => void;
  onExecuteTicket?: (ticketId: string) => void;
}

function isFilePath(s: string): boolean {
  return /^[a-zA-Z0-9_/.]/.test(s) && !s.includes('//') && (s.includes('/') || s.includes('.'));
}

export function TriageClusterView({ clusters, contextDocs, onTicketClick, onExecuteTicket }: TriageClusterViewProps) {
  const docMap = new Map(contextDocs.map((d) => [d.clusterId, d]));

  if (clusters.length === 0) {
    return (
      <div className="p-4 text-sm text-slate-400">No clusters formed.</div>
    );
  }

  return (
    <div className="p-4 space-y-4">
      {clusters.map((cluster) => {
        const doc = docMap.get(cluster.clusterId);
        const tickets = cluster.tickets || [];
        const repoAreas = (doc?.repoAreas || cluster.repoAreas || []).filter(isFilePath);
        const risks = (doc?.risks || []);
        const validationPlan = (doc?.validationPlan || []);

        return (
          <div key={cluster.clusterId} className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-3">
                <h3 className="text-sm font-medium text-slate-200">{cluster.clusterId}</h3>
                <span className="text-xs text-slate-500">
                  {tickets.length} ticket{tickets.length !== 1 ? 's' : ''}
                </span>
              </div>
              {onExecuteTicket && tickets.length > 0 && (
                <button
                  onClick={() => tickets.forEach((id) => onExecuteTicket(id))}
                  className="flex items-center gap-1.5 px-3 py-1 text-xs bg-purple-500/20 text-purple-300 border border-purple-500/30 rounded-md hover:bg-purple-500/30 transition-colors"
                >
                  <PlayCircle className="w-3 h-3" />
                  Execute All
                </button>
              )}
            </div>

            {/* Rationale */}
            <p className="text-sm text-slate-400 mb-3">{cluster.rationale}</p>

            {/* Ticket IDs */}
            <div className="flex flex-wrap gap-1.5 mb-3">
              {tickets.map((id) => (
                <button
                  key={id}
                  onClick={() => onTicketClick?.(id)}
                  className="px-2 py-0.5 text-xs bg-slate-700 text-slate-300 rounded font-mono hover:bg-slate-600 transition-colors cursor-pointer"
                >
                  {id}
                </button>
              ))}
            </div>

            {/* Repo Areas (file paths only, deduplicated) */}
            {repoAreas.length > 0 && (
              <div className="mb-3">
                <div className="flex items-center gap-1.5 mb-1.5">
                  <FileCode className="w-3 h-3 text-slate-500" />
                  <span className="text-xs text-slate-500">Repo Areas</span>
                </div>
                <div className="flex flex-wrap gap-1">
                  {repoAreas.map((area) => (
                    <span key={area} className="px-1.5 py-0.5 text-xs bg-blue-500/10 text-blue-300 border border-blue-500/20 rounded font-mono">
                      {area}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Context Doc details */}
            {doc && (
              <div className="mt-3 pt-3 border-t border-slate-700 space-y-2">
                {validationPlan.length > 0 && (
                  <div className="flex items-start gap-1.5">
                    <ClipboardCheck className="w-3 h-3 text-slate-500 mt-0.5 flex-shrink-0" />
                    <div>
                      <span className="text-xs text-slate-500">Validation Plan</span>
                      <ul className="text-xs text-slate-400 mt-0.5 space-y-0.5">
                        {validationPlan.map((step, i) => (
                          <li key={i}>{step}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                )}
                {risks.length > 0 && (
                  <div className="flex items-start gap-1.5">
                    <AlertTriangle className="w-3 h-3 text-yellow-500 mt-0.5 flex-shrink-0" />
                    <div>
                      <span className="text-xs text-slate-500">Risks</span>
                      <ul className="text-xs text-yellow-400 mt-0.5 space-y-0.5">
                        {risks.map((risk, i) => (
                          <li key={i}>{risk}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                )}
                {doc.costCeiling && (
                  <div className="text-xs text-slate-500">
                    Cost ceiling: {(doc.costCeiling.maxTokensPerTicket / 1000).toFixed(0)}K tokens, {doc.costCeiling.maxAgentMinutesPerTicket}min per ticket
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
