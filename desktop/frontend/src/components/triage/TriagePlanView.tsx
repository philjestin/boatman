import { CheckCircle2, XCircle, AlertTriangle, FileCode, ChevronRight } from 'lucide-react';
import type { PlanResult, PlanStats, TriageNormalizedTicket } from '../../types';

interface TriagePlanViewProps {
  plans: PlanResult[];
  planStats: PlanStats;
  tickets: TriageNormalizedTicket[];
  onPlanClick: (plan: PlanResult) => void;
}

export function TriagePlanView({ plans, planStats, tickets, onPlanClick }: TriagePlanViewProps) {
  const ticketTitles = new Map(tickets.map((t) => [t.ticketId, t.title]));

  return (
    <div className="p-4 space-y-4">
      {/* Summary bar */}
      <div className="flex items-center gap-4 p-3 bg-slate-800/50 rounded-lg border border-slate-700">
        <div className="flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-green-400" />
          <span className="text-sm text-slate-300">
            <span className="font-medium text-green-400">{planStats.passed}</span> passed
          </span>
        </div>
        <div className="flex items-center gap-2">
          <XCircle className="w-4 h-4 text-red-400" />
          <span className="text-sm text-slate-300">
            <span className="font-medium text-red-400">{planStats.failed}</span> failed
          </span>
        </div>
        <div className="text-sm text-slate-500">
          {planStats.total} total
        </div>
        {planStats.totalTokensUsed > 0 && (
          <div className="ml-auto text-xs text-slate-500">
            {(planStats.totalTokensUsed / 1000).toFixed(1)}K tokens | ${planStats.totalCostUsd.toFixed(4)}
          </div>
        )}
      </div>

      {/* Plan rows */}
      <div className="space-y-1">
        {plans.map((pr) => {
          const title = ticketTitles.get(pr.ticketId) || pr.ticketId;
          const truncTitle = title.length > 60 ? title.slice(0, 57) + '...' : title;
          const passed = pr.validation?.passed ?? false;
          const hasError = !!pr.error;
          const fileCount = pr.plan?.candidateFiles?.length ?? 0;

          return (
            <button
              key={pr.ticketId}
              onClick={() => onPlanClick(pr)}
              className="w-full flex items-center gap-3 px-3 py-2.5 bg-slate-800/30 hover:bg-slate-700/50 rounded-md transition-colors text-left group"
            >
              {/* Status icon */}
              {hasError ? (
                <AlertTriangle className="w-4 h-4 text-red-400 flex-shrink-0" />
              ) : passed ? (
                <CheckCircle2 className="w-4 h-4 text-green-400 flex-shrink-0" />
              ) : (
                <XCircle className="w-4 h-4 text-yellow-400 flex-shrink-0" />
              )}

              {/* Ticket ID */}
              <span className="text-xs font-mono text-amber-400 w-20 flex-shrink-0">{pr.ticketId}</span>

              {/* Title */}
              <span className="text-sm text-slate-300 flex-1 truncate">{truncTitle}</span>

              {/* File count */}
              {fileCount > 0 && (
                <span className="flex items-center gap-1 text-xs text-slate-500">
                  <FileCode className="w-3 h-3" />
                  {fileCount}
                </span>
              )}

              {/* Approach preview */}
              {pr.plan?.approach && (
                <span className="text-xs text-slate-500 max-w-[200px] truncate hidden lg:inline">
                  {pr.plan.approach}
                </span>
              )}

              <ChevronRight className="w-4 h-4 text-slate-600 group-hover:text-slate-400 flex-shrink-0" />
            </button>
          );
        })}
      </div>
    </div>
  );
}
