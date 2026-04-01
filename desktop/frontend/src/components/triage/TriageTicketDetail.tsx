import { X, PlayCircle } from 'lucide-react';
import type { TriageClassification, TriageNormalizedTicket } from '../../types';
import { TriageCategoryBadge } from './TriageCategoryBadge';

interface TriageTicketDetailProps {
  classification: TriageClassification;
  ticket?: TriageNormalizedTicket;
  onClose: () => void;
  onExecute?: (ticketID: string) => void;
}

const rubricLabels: Record<string, { label: string; positive: boolean }> = {
  clarity: { label: 'Clarity', positive: true },
  codeLocality: { label: 'Code Locality', positive: true },
  patternMatch: { label: 'Pattern Match', positive: true },
  validationStrength: { label: 'Validation Strength', positive: true },
  dependencyRisk: { label: 'Dependency Risk', positive: false },
  productAmbiguity: { label: 'Product Ambiguity', positive: false },
  blastRadius: { label: 'Blast Radius', positive: false },
};

function ScoreBar({ score, positive }: { score: number; positive: boolean }) {
  const percentage = (score / 5) * 100;
  const barColor = positive
    ? score >= 4 ? 'bg-green-500' : score >= 3 ? 'bg-blue-500' : 'bg-yellow-500'
    : score <= 1 ? 'bg-green-500' : score <= 2 ? 'bg-blue-500' : 'bg-red-500';

  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-2 bg-slate-700 rounded-full overflow-hidden">
        <div className={`h-full ${barColor} rounded-full`} style={{ width: `${percentage}%` }} />
      </div>
      <span className="text-xs text-slate-400 w-4 text-right">{score}</span>
    </div>
  );
}

export function TriageTicketDetail({ classification, ticket, onClose, onExecute }: TriageTicketDetailProps) {
  const canExecute = classification.category === 'AI_DEFINITE' || classification.category === 'AI_LIKELY';

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 border border-slate-700 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-700">
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-lg font-semibold text-slate-100">{classification.ticketId}</h2>
              <TriageCategoryBadge category={classification.category} />
            </div>
            {ticket && <p className="text-sm text-slate-400 mt-1">{ticket.title}</p>}
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-700 rounded-md">
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Rubric Scores */}
          <div>
            <h3 className="text-sm font-medium text-slate-200 mb-3">Rubric Scores</h3>
            <div className="space-y-2">
              {Object.entries(rubricLabels).map(([key, { label, positive }]) => (
                <div key={key} className="flex items-center gap-3">
                  <span className="text-xs text-slate-400 w-32">{label}</span>
                  <div className="flex-1">
                    <ScoreBar
                      score={classification.rubric[key as keyof typeof classification.rubric]}
                      positive={positive}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Gate Results */}
          {classification.gateResults && classification.gateResults.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-slate-200 mb-2">Gate Results</h3>
              <div className="space-y-1">
                {classification.gateResults.map((gate) => (
                  <div key={gate.gate} className="flex items-center gap-2 text-sm">
                    <span className={gate.passed ? 'text-green-400' : 'text-red-400'}>
                      {gate.passed ? 'PASS' : 'FAIL'}
                    </span>
                    <span className="text-slate-400">{gate.gate}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Hard Stops */}
          {classification.hardStops && classification.hardStops.length > 0 && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
              <h3 className="text-sm font-medium text-red-300 mb-1">Hard Stops</h3>
              <ul className="text-sm text-red-400 space-y-0.5">
                {classification.hardStops.map((stop) => (
                  <li key={stop}>{stop}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Reasons */}
          {classification.reasons && classification.reasons.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-slate-200 mb-2">Reasons</h3>
              <ul className="text-sm text-slate-400 space-y-1">
                {classification.reasons.map((reason, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <span className="text-slate-500 mt-0.5">-</span>
                    <span>{reason}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Uncertain Axes */}
          {classification.uncertainAxes && classification.uncertainAxes.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-slate-200 mb-2">Uncertain Dimensions</h3>
              <div className="flex flex-wrap gap-1">
                {classification.uncertainAxes.map((axis) => (
                  <span key={axis} className="px-2 py-0.5 text-xs bg-slate-700 text-slate-300 rounded">
                    {axis}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        {onExecute && (
          <div className="flex items-center justify-end gap-3 p-6 border-t border-slate-700">
            <button
              onClick={() => onExecute(classification.ticketId)}
              className={`flex items-center gap-2 px-4 py-2 text-sm text-white rounded-md transition-colors ${
                canExecute
                  ? 'bg-purple-500 hover:bg-purple-600'
                  : 'bg-slate-600 hover:bg-slate-500'
              }`}
            >
              <PlayCircle className="w-4 h-4" />
              Execute with Boatman Mode
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
