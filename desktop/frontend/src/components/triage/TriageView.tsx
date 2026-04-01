import { useState, useEffect, useRef } from 'react';
import { Loader2, CheckCircle2, Download, Cpu, Layers, FolderTree, FileDown, ClipboardList } from 'lucide-react';
import type { TriageResult, TriageClassification, TriageEvent, PlanResult, PlanStats } from '../../types';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { ExportTriagePDF } from '../../../wailsjs/go/main/App';
import { TriageStatsBar } from './TriageStatsBar';
import { TriageResultsTable } from './TriageResultsTable';
import { TriageTicketDetail } from './TriageTicketDetail';
import { TriageClusterView } from './TriageClusterView';
import { TriagePlanView } from './TriagePlanView';
import { TriagePlanDetail } from './TriagePlanDetail';

interface TriageViewProps {
  sessionId: string;
  status: string;
  getTriageResult: (sessionId: string) => Promise<Record<string, any> | null>;
  onExecuteTicket?: (ticketID: string) => void;
}

type TriageTab = 'results' | 'clusters' | 'plans';

interface ProgressEntry {
  id: string;
  type: string;
  message: string;
  timestamp: number;
  ticketID?: string;
  scored?: number;
  total?: number;
  status?: string;
  error?: string;
}

type PipelineStage = 'starting' | 'fetching' | 'scoring' | 'classifying' | 'clustering' | 'planning' | 'complete';

const stageLabels: Record<PipelineStage, string> = {
  starting: 'Initializing pipeline...',
  fetching: 'Fetching tickets from Linear',
  scoring: 'Scoring tickets with Claude',
  classifying: 'Classifying tickets',
  clustering: 'Clustering related tickets',
  planning: 'Generating execution plans',
  complete: 'Pipeline complete',
};

const stageIcons: Record<PipelineStage, React.ReactNode> = {
  starting: <Loader2 className="w-4 h-4 animate-spin" />,
  fetching: <Download className="w-4 h-4" />,
  scoring: <Cpu className="w-4 h-4" />,
  classifying: <Layers className="w-4 h-4" />,
  clustering: <FolderTree className="w-4 h-4" />,
  planning: <ClipboardList className="w-4 h-4" />,
  complete: <CheckCircle2 className="w-4 h-4" />,
};

export function TriageView({ sessionId, status, getTriageResult, onExecuteTicket }: TriageViewProps) {
  const [triageResult, setTriageResult] = useState<TriageResult | null>(null);
  const [selectedClassification, setSelectedClassification] = useState<TriageClassification | null>(null);
  const [selectedPlan, setSelectedPlan] = useState<PlanResult | null>(null);
  const [activeTab, setActiveTab] = useState<TriageTab>('results');
  const [stage, setStage] = useState<PipelineStage>('starting');
  const [progress, setProgress] = useState<ProgressEntry[]>([]);
  const [scoredCount, setScoredCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const feedRef = useRef<HTMLDivElement>(null);

  // Reset all state when session changes
  useEffect(() => {
    setTriageResult(null);
    setSelectedClassification(null);
    setSelectedPlan(null);
    setActiveTab('results');
    setStage('starting');
    setProgress([]);
    setScoredCount(0);
    setTotalCount(0);
  }, [sessionId]);

  // Subscribe to triage events for live progress
  useEffect(() => {
    const handler = (data: { sessionId: string; event: TriageEvent }) => {
      if (data.sessionId !== sessionId) return;

      const evt = data.event;
      const entry: ProgressEntry = {
        id: `${evt.type}-${Date.now()}-${Math.random()}`,
        type: evt.type,
        message: evt.message || evt.type,
        timestamp: Date.now(),
        ticketID: evt.id,
        status: evt.status,
        error: (evt.data as any)?.error,
      };

      switch (evt.type) {
        case 'triage_started':
          setStage('fetching');
          break;
        case 'triage_fetch_complete':
          setTotalCount((evt.data as any)?.ticketCount || 0);
          setStage('scoring');
          break;
        case 'triage_scoring_started':
          setTotalCount((evt.data as any)?.ticketCount || 0);
          setStage('scoring');
          break;
        case 'triage_ticket_scoring':
          entry.scored = Number((evt.data as any)?.index ?? 0);
          entry.total = Number((evt.data as any)?.total ?? 0);
          break;
        case 'triage_ticket_scored': {
          const tot = Number((evt.data as any)?.total ?? 0);
          // Use functional updater to count completions correctly
          // (index is array position, not sequential with concurrency)
          setScoredCount((prev) => prev + 1);
          setTotalCount(tot);
          break;
        }
        case 'triage_scoring_complete':
          setStage('classifying');
          break;
        case 'triage_classifying':
          setStage('classifying');
          break;
        case 'triage_clustering':
          setStage('clustering');
          break;
        case 'plan_started':
          setStage('planning');
          break;
        case 'plan_ticket_planned':
          break;
        case 'plan_complete':
          setStage('complete');
          break;
        case 'triage_complete':
          // If plans are being generated, clustering → planning will happen next.
          // Otherwise this is the final stage.
          setStage('complete');
          break;
      }

      setProgress((prev) => [...prev, entry]);
    };

    EventsOn('triage:event', handler);
    return () => { EventsOff('triage:event'); };
  }, [sessionId]);

  // Auto-scroll the progress feed
  useEffect(() => {
    if (feedRef.current) {
      feedRef.current.scrollTop = feedRef.current.scrollHeight;
    }
  }, [progress]);

  // Fetch triage result when status changes to stopped (completed)
  useEffect(() => {
    if (status === 'stopped' || status === 'idle') {
      const loadResult = async () => {
        const config = await getTriageResult(sessionId);
        if (config?.triageResult) {
          const raw = config.triageResult as Record<string, any>;
          // Normalize null arrays from Go nil slices
          const result: TriageResult = {
            tickets: raw.tickets || [],
            classifications: raw.classifications || [],
            clusters: raw.clusters || [],
            contextDocs: raw.contextDocs || [],
            stats: raw.stats || {
              totalTickets: 0,
              aiDefiniteCount: 0,
              aiLikelyCount: 0,
              humanReviewCount: 0,
              humanOnlyCount: 0,
              clusterCount: 0,
              totalTokensUsed: 0,
              totalCostUsd: 0,
            },
          };
          // Load plan data if available
          if (config.plans) {
            result.plans = config.plans as PlanResult[];
          }
          if (config.planStats) {
            result.planStats = config.planStats as PlanStats;
          }
          setTriageResult(result);
        }
      };
      loadResult();
    }
  }, [sessionId, status, getTriageResult]);

  // Progress phase
  if (!triageResult) {
    const progressPct = totalCount > 0 ? Math.round((scoredCount / totalCount) * 100) : 0;

    return (
      <div className="flex-1 flex flex-col overflow-hidden">
        {status === 'running' ? (
          <div className="flex-1 flex flex-col p-6 space-y-6">
            {/* Stage indicator */}
            <div className="flex items-center justify-center gap-3">
              <Loader2 className="w-6 h-6 text-amber-500 animate-spin" />
              <span className="text-lg text-slate-200 font-medium">Triage in progress</span>
            </div>

            {/* Pipeline stages */}
            <div className="flex items-center justify-center gap-1 text-xs">
              {(['fetching', 'scoring', 'classifying', 'clustering', 'planning'] as PipelineStage[]).map((s, i) => {
                const allStages = ['fetching', 'scoring', 'classifying', 'clustering', 'planning'];
                const isActive = s === stage;
                const isDone =
                  (allStages.indexOf(stage) > allStages.indexOf(s)) ||
                  stage === 'complete';

                return (
                  <div key={s} className="flex items-center gap-1">
                    {i > 0 && (
                      <div className={`w-8 h-px ${isDone ? 'bg-amber-500' : 'bg-slate-700'}`} />
                    )}
                    <div
                      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full transition-colors ${
                        isActive
                          ? 'bg-amber-500/20 text-amber-400 ring-1 ring-amber-500/40'
                          : isDone
                            ? 'bg-green-500/10 text-green-400'
                            : 'bg-slate-800 text-slate-500'
                      }`}
                    >
                      {isActive ? (
                        <Loader2 className="w-3 h-3 animate-spin" />
                      ) : isDone ? (
                        <CheckCircle2 className="w-3 h-3" />
                      ) : (
                        stageIcons[s]
                      )}
                      <span>{s.charAt(0).toUpperCase() + s.slice(1)}</span>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Progress bar (during scoring) */}
            {stage === 'scoring' && totalCount > 0 && (
              <div className="max-w-md mx-auto w-full space-y-2">
                <div className="flex justify-between text-sm text-slate-400">
                  <span>Scoring tickets</span>
                  <span>{scoredCount} / {totalCount} ({progressPct}%)</span>
                </div>
                <div className="w-full h-2 bg-slate-700 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-amber-500 rounded-full transition-all duration-300"
                    style={{ width: `${progressPct}%` }}
                  />
                </div>
              </div>
            )}

            {/* Live event feed */}
            <div className="flex-1 min-h-0">
              <div
                ref={feedRef}
                className="h-full max-h-[400px] overflow-y-auto bg-slate-900/50 rounded-lg border border-slate-700 p-3 font-mono text-xs space-y-1"
              >
                {progress.map((entry) => (
                  <div key={entry.id} className="flex items-start gap-2">
                    <span className="text-slate-600 flex-shrink-0">
                      {new Date(entry.timestamp).toLocaleTimeString()}
                    </span>
                    <span
                      className={
                        entry.status === 'error' || entry.error
                          ? 'text-red-400'
                          : entry.type === 'triage_ticket_scored'
                            ? 'text-green-400'
                            : entry.type === 'triage_ticket_scoring'
                              ? 'text-amber-400'
                              : entry.type.includes('complete') || entry.type.includes('classifying') || entry.type.includes('clustering')
                                ? 'text-blue-400'
                                : 'text-slate-300'
                      }
                    >
                      {entry.message}
                    </span>
                  </div>
                ))}
                {progress.length === 0 && (
                  <div className="text-slate-500">Waiting for events...</div>
                )}
              </div>
            </div>
          </div>
        ) : status === 'error' ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center space-y-3">
              <p className="text-red-400 font-medium">Triage failed</p>
              <p className="text-sm text-slate-400">Check the session messages for details</p>
              {progress.length > 0 && (
                <div className="mt-4 max-w-lg mx-auto">
                  <div
                    ref={feedRef}
                    className="max-h-[200px] overflow-y-auto bg-slate-900/50 rounded-lg border border-slate-700 p-3 font-mono text-xs space-y-1 text-left"
                  >
                    {progress.slice(-10).map((entry) => (
                      <div key={entry.id} className="flex items-start gap-2">
                        <span className="text-slate-600 flex-shrink-0">
                          {new Date(entry.timestamp).toLocaleTimeString()}
                        </span>
                        <span className="text-slate-300">{entry.message}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-slate-400">No triage results available</p>
          </div>
        )}
      </div>
    );
  }

  // Results phase
  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Stats */}
      <TriageStatsBar stats={triageResult.stats} />

      {/* Sub-tabs */}
      <div className="flex items-center border-b border-slate-700 px-4">
        <button
          onClick={() => setActiveTab('results')}
          className={`px-4 py-2 text-sm border-b-2 transition-colors ${
            activeTab === 'results'
              ? 'border-amber-500 text-slate-100'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          Results ({triageResult.classifications.length})
        </button>
        <button
          onClick={() => setActiveTab('clusters')}
          className={`px-4 py-2 text-sm border-b-2 transition-colors ${
            activeTab === 'clusters'
              ? 'border-amber-500 text-slate-100'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          Clusters ({triageResult.clusters.length})
        </button>
        {triageResult.plans && triageResult.plans.length > 0 && (
          <button
            onClick={() => setActiveTab('plans')}
            className={`px-4 py-2 text-sm border-b-2 transition-colors ${
              activeTab === 'plans'
                ? 'border-amber-500 text-slate-100'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            Plans ({triageResult.plans.length})
          </button>
        )}

        {/* Cost info + Export */}
        <div className="ml-auto flex items-center gap-3 text-xs text-slate-500">
          {triageResult.stats.totalTokensUsed > 0 && (
            <span>
              {(triageResult.stats.totalTokensUsed / 1000).toFixed(1)}K tokens | ${triageResult.stats.totalCostUsd.toFixed(4)}
            </span>
          )}
          <button
            onClick={async () => {
              try {
                const path = await ExportTriagePDF(sessionId);
                if (path) {
                  console.log('PDF exported to:', path);
                }
              } catch (err) {
                console.error('PDF export failed:', err);
              }
            }}
            className="flex items-center gap-1.5 px-3 py-1 text-xs bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded-md hover:bg-amber-500/30 transition-colors"
          >
            <FileDown className="w-3 h-3" />
            Export PDF
          </button>
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 min-h-0 overflow-y-auto">
        {activeTab === 'results' && (
          <TriageResultsTable
            classifications={triageResult.classifications}
            tickets={triageResult.tickets}
            clusters={triageResult.clusters}
            onTicketClick={setSelectedClassification}
          />
        )}
        {activeTab === 'clusters' && (
          <TriageClusterView
            clusters={triageResult.clusters}
            contextDocs={triageResult.contextDocs}
            onTicketClick={(ticketId) => {
              const c = triageResult.classifications.find((cl) => cl.ticketId === ticketId);
              if (c) setSelectedClassification(c);
            }}
            onExecuteTicket={onExecuteTicket}
          />
        )}
        {activeTab === 'plans' && triageResult.plans && triageResult.planStats && (
          <TriagePlanView
            plans={triageResult.plans}
            planStats={triageResult.planStats}
            tickets={triageResult.tickets}
            onPlanClick={setSelectedPlan}
          />
        )}
      </div>

      {/* Ticket Detail Modal */}
      {selectedClassification && (
        <TriageTicketDetail
          classification={selectedClassification}
          ticket={triageResult.tickets.find((t) => t.ticketId === selectedClassification.ticketId)}
          onClose={() => setSelectedClassification(null)}
          onExecute={onExecuteTicket}
        />
      )}

      {/* Plan Detail Modal */}
      {selectedPlan && (
        <TriagePlanDetail
          planResult={selectedPlan}
          ticketTitle={triageResult.tickets.find((t) => t.ticketId === selectedPlan.ticketId)?.title}
          onClose={() => setSelectedPlan(null)}
        />
      )}
    </div>
  );
}
