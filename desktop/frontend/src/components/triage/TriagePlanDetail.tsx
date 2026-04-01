import { X, CheckCircle2, XCircle, FileCode, AlertTriangle, Terminal, Undo2, OctagonX, HelpCircle } from 'lucide-react';
import type { PlanResult } from '../../types';

interface TriagePlanDetailProps {
  planResult: PlanResult;
  ticketTitle?: string;
  onClose: () => void;
}

function GateBadge({ passed }: { passed: boolean }) {
  return passed ? (
    <span className="flex items-center gap-1 text-xs text-green-400">
      <CheckCircle2 className="w-3 h-3" /> Passed
    </span>
  ) : (
    <span className="flex items-center gap-1 text-xs text-red-400">
      <XCircle className="w-3 h-3" /> Failed
    </span>
  );
}

function FileStatus({ file, validated, missing, outOfScope }: {
  file: string;
  validated: string[];
  missing: string[];
  outOfScope: string[];
}) {
  const isValidated = validated.includes(file);
  const isMissing = missing.includes(file);
  const isOutOfScope = outOfScope.includes(file);

  return (
    <div className="flex items-center gap-2 py-0.5">
      {isMissing ? (
        <XCircle className="w-3 h-3 text-red-400 flex-shrink-0" />
      ) : isOutOfScope ? (
        <AlertTriangle className="w-3 h-3 text-yellow-400 flex-shrink-0" />
      ) : isValidated ? (
        <CheckCircle2 className="w-3 h-3 text-green-400 flex-shrink-0" />
      ) : (
        <FileCode className="w-3 h-3 text-slate-500 flex-shrink-0" />
      )}
      <span className={`font-mono text-xs ${isMissing ? 'text-red-400' : isOutOfScope ? 'text-yellow-400' : 'text-slate-300'}`}>
        {file}
      </span>
    </div>
  );
}

export function TriagePlanDetail({ planResult, ticketTitle, onClose }: TriagePlanDetailProps) {
  const { plan, validation, error } = planResult;
  const validatedFiles = validation?.validatedFiles || [];
  const missingFiles = validation?.missingFiles || [];
  const outOfScopeFiles = validation?.outOfScopeFiles || [];

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-lg shadow-xl max-w-3xl w-full mx-4 border border-slate-700 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-700">
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-lg font-semibold text-slate-100">{planResult.ticketId}</h2>
              {validation && (
                <span className={`px-2 py-0.5 text-xs rounded-full ${
                  validation.passed
                    ? 'bg-green-500/20 text-green-400 border border-green-500/30'
                    : 'bg-red-500/20 text-red-400 border border-red-500/30'
                }`}>
                  {validation.passed ? 'Gates Passed' : 'Gates Failed'}
                </span>
              )}
            </div>
            {ticketTitle && (
              <p className="text-sm text-slate-400 mt-1">{ticketTitle}</p>
            )}
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-700 rounded-md">
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Error state */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-500/10 border border-red-500/30 rounded-md">
            <p className="text-sm text-red-400">{error}</p>
          </div>
        )}

        {plan && (
          <div className="p-6 space-y-6">
            {/* Approach */}
            <div>
              <h3 className="text-sm font-medium text-slate-300 mb-2">Approach</h3>
              <p className="text-sm text-slate-400 leading-relaxed whitespace-pre-wrap">{plan.approach}</p>
            </div>

            {/* Candidate files */}
            {plan.candidateFiles?.length > 0 && (
              <div>
                <h3 className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-2">
                  <FileCode className="w-4 h-4" />
                  Candidate Files ({plan.candidateFiles.length})
                </h3>
                <div className="bg-slate-900/50 rounded-md p-3 space-y-0.5">
                  {plan.candidateFiles.map((f) => (
                    <FileStatus
                      key={f}
                      file={f}
                      validated={validatedFiles}
                      missing={missingFiles}
                      outOfScope={outOfScopeFiles}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* New / Deleted files */}
            {plan.newFiles?.length > 0 && (
              <div>
                <h3 className="text-sm font-medium text-green-400 mb-2">New Files ({plan.newFiles.length})</h3>
                <div className="bg-slate-900/50 rounded-md p-3 font-mono text-xs text-green-300 space-y-0.5">
                  {plan.newFiles.map((f) => <div key={f}>+ {f}</div>)}
                </div>
              </div>
            )}
            {plan.deletedFiles?.length > 0 && (
              <div>
                <h3 className="text-sm font-medium text-red-400 mb-2">Deleted Files ({plan.deletedFiles.length})</h3>
                <div className="bg-slate-900/50 rounded-md p-3 font-mono text-xs text-red-300 space-y-0.5">
                  {plan.deletedFiles.map((f) => <div key={f}>- {f}</div>)}
                </div>
              </div>
            )}

            {/* Validation commands */}
            {plan.validation?.length > 0 && (
              <div>
                <h3 className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-2">
                  <Terminal className="w-4 h-4" />
                  Validation Commands
                </h3>
                <div className="bg-slate-900/50 rounded-md p-3 font-mono text-xs text-slate-300 space-y-1">
                  {plan.validation.map((cmd, i) => (
                    <div key={i} className="flex items-start gap-2">
                      <span className="text-slate-600">$</span>
                      <span>{cmd}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Stop conditions */}
            {plan.stopConditions?.length > 0 && (
              <div>
                <h3 className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-2">
                  <OctagonX className="w-4 h-4" />
                  Stop Conditions
                </h3>
                <ul className="space-y-1">
                  {plan.stopConditions.map((sc, i) => (
                    <li key={i} className="flex items-start gap-2 text-sm text-slate-400">
                      <span className="text-red-400 mt-0.5">&#x2022;</span>
                      {sc}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Uncertainties */}
            {plan.uncertainties?.length > 0 && (
              <div>
                <h3 className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-2">
                  <HelpCircle className="w-4 h-4" />
                  Uncertainties
                </h3>
                <ul className="space-y-1">
                  {plan.uncertainties.map((u, i) => (
                    <li key={i} className="flex items-start gap-2 text-sm text-slate-400">
                      <span className="text-yellow-400 mt-0.5">?</span>
                      {u}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Rollback */}
            {plan.rollback && (
              <div>
                <h3 className="flex items-center gap-2 text-sm font-medium text-slate-300 mb-2">
                  <Undo2 className="w-4 h-4" />
                  Rollback
                </h3>
                <p className="text-sm text-slate-400">{plan.rollback}</p>
              </div>
            )}

            {/* Gate results */}
            {validation && (
              <div>
                <h3 className="text-sm font-medium text-slate-300 mb-2">Validation Gates</h3>
                <div className="space-y-2">
                  {validation.gateResults.map((g) => (
                    <div key={g.gate} className="flex items-center justify-between px-3 py-2 bg-slate-900/50 rounded-md">
                      <span className="text-sm text-slate-300">{g.gate.replace(/_/g, ' ')}</span>
                      <div className="flex items-center gap-3">
                        {g.reason && <span className="text-xs text-slate-500">{g.reason}</span>}
                        <GateBadge passed={g.passed} />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Usage */}
            {planResult.usage && (
              <div className="text-xs text-slate-500 pt-2 border-t border-slate-700">
                {planResult.usage.input_tokens + planResult.usage.output_tokens} tokens | ${planResult.usage.total_cost_usd.toFixed(4)}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
