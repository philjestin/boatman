import { useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import {
  AlertTriangle,
  CheckCircle2,
  Clipboard,
  ExternalLink,
  FileCode2,
  GitPullRequest,
  Layers,
  Loader2,
  Play,
  Search,
  ShieldCheck,
  Workflow,
} from 'lucide-react';
import { PlanFrontendTicketQueue, RunFrontendTicketQueue } from '../../../wailsjs/go/main/App';
import { main } from '../../../wailsjs/go/models';
import { MODEL_OPTIONS } from '../../types';
import type {
  FrontendAutomationPolicy,
  FrontendQueueBatch,
  FrontendTicketPlan,
  FrontendTicketQueueResult,
  FrontendTicketQueueRunResult,
  FrontendTicketType,
  ModelProfile,
} from '../../types';

interface FrontendQueueViewProps {
  projectPath?: string;
  onOpenSettings?: () => void;
}

type RunState = 'idle' | 'loading' | 'error' | 'complete';

const DEFAULT_STATES = 'backlog,todo,unstarted';
const INPUT_CLASS = 'w-full px-3 py-2 text-sm bg-slate-950 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500';

export function FrontendQueueView({ projectPath, onOpenSettings }: FrontendQueueViewProps) {
  const [linearProject, setLinearProject] = useState('');
  const [query, setQuery] = useState('');
  const [states, setStates] = useState(DEFAULT_STATES);
  const [limit, setLimit] = useState(50);
  const [maxParallel, setMaxParallel] = useState(3);
  const [includeCanceled, setIncludeCanceled] = useState(false);
  const [models, setModels] = useState<ModelProfile>({});
  const [result, setResult] = useState<FrontendTicketQueueResult | null>(null);
  const [runResult, setRunResult] = useState<FrontendTicketQueueRunResult | null>(null);
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);
  const [runState, setRunState] = useState<RunState>('idle');
  const [workerRunning, setWorkerRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [runError, setRunError] = useState<string | null>(null);
  const [copiedTicketId, setCopiedTicketId] = useState<string | null>(null);

  const selectedTicket = useMemo(() => {
    if (!result) return null;
    return result.plan.tickets.find((ticket) => ticket.ticket.id === selectedTicketId)
      || result.plan.tickets[0]
      || null;
  }, [result, selectedTicketId]);

  const ticketById = useMemo(() => {
    const byId = new Map<string, FrontendTicketPlan>();
    result?.plan.tickets.forEach((ticket) => byId.set(ticket.ticket.id, ticket));
    return byId;
  }, [result]);

  const handlePlan = async () => {
    if (!linearProject.trim()) {
      setError('Linear project is required.');
      return;
    }
    setRunState('loading');
    setError(null);
    setRunError(null);
    setResult(null);
    setRunResult(null);
    setSelectedTicketId(null);
    try {
      const request = main.FrontendTicketQueueRequest.createFrom({
        projectPath: projectPath || '',
        linearProject: linearProject.trim(),
        query: query.trim(),
        limit,
        states: states.split(',').map((state) => state.trim()).filter(Boolean),
        includeCanceled,
        maxParallel,
      });
      const response = (await PlanFrontendTicketQueue(request)) as unknown as FrontendTicketQueueResult;
      setResult(response);
      setSelectedTicketId(response.plan.tickets[0]?.ticket.id || null);
      setRunState('complete');
    } catch (err) {
      setError(errorMessage(err, 'Failed to plan frontend queue'));
      setRunState('error');
    }
  };

  const handleRunTickets = async (ticketIds: string[], batchId?: string) => {
    if (!result) return;
    const resolvedProjectPath = projectPath || result.plan.options?.projectPath || '';
    if (!resolvedProjectPath) {
      setRunError('Open a project before starting frontend workers.');
      return;
    }
    setWorkerRunning(true);
    setRunError(null);
    try {
      const request = main.FrontendTicketQueueRunRequest.createFrom({
        projectPath: resolvedProjectPath,
        linearProject: result.linearProject || linearProject.trim(),
        plan: result.plan,
        batchId: batchId || '',
        ticketIds,
        models,
      });
      const response = (await RunFrontendTicketQueue(request)) as unknown as FrontendTicketQueueRunResult;
      setRunResult(response);
    } catch (err) {
      setRunError(errorMessage(err, 'Failed to start frontend workers'));
    } finally {
      setWorkerRunning(false);
    }
  };

  const handleModelChange = (key: keyof ModelProfile, value: string) => {
    setModels((current) => ({
      ...current,
      [key]: value || undefined,
    }));
  };

  const handleCopyPrompt = async (ticket: FrontendTicketPlan) => {
    await navigator.clipboard.writeText(ticket.workerPrompt);
    setCopiedTicketId(ticket.ticket.id);
    window.setTimeout(() => setCopiedTicketId(null), 1400);
  };

  return (
    <div className="flex flex-col h-full bg-slate-900">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-2 min-w-0">
          <Workflow className="w-5 h-5 text-emerald-400 flex-shrink-0" />
          <h2 className="text-sm font-semibold text-slate-200">Frontend Queue</h2>
          {result && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs bg-slate-700 rounded text-slate-400">
              <Layers className="w-3 h-3" />
              {result.plan.stats.batchCount} batches
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {result?.plan.batches[0] && (
            <button
              onClick={() => handleRunTickets(result.plan.batches[0].ticketIds, result.plan.batches[0].id)}
              disabled={workerRunning}
              className="inline-flex items-center gap-2 px-3 py-1.5 text-xs border border-slate-700 text-slate-200 rounded-md hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {workerRunning ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <GitPullRequest className="w-3.5 h-3.5" />}
              Run First Batch
            </button>
          )}
          <button
            onClick={handlePlan}
            disabled={runState === 'loading'}
            className="inline-flex items-center gap-2 px-3 py-1.5 text-xs bg-emerald-600 text-white rounded-md hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {runState === 'loading' ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
            Plan Queue
          </button>
        </div>
      </div>

      {error && (
        <div className="flex items-center justify-between gap-3 px-4 py-2 bg-red-950/50 border-b border-red-900/60 text-xs text-red-200">
          <span className="inline-flex items-center gap-2">
            <AlertTriangle className="w-3.5 h-3.5 text-red-400 flex-shrink-0" />
            {error}
          </span>
          {error.toLowerCase().includes('linear api key') && onOpenSettings && (
            <button
              onClick={onOpenSettings}
              className="px-2 py-1 text-xs text-red-100 border border-red-800 rounded hover:bg-red-900/40"
            >
              Settings
            </button>
          )}
        </div>
      )}

      {runError && (
        <div className="flex items-center gap-2 px-4 py-2 bg-amber-950/50 border-b border-amber-900/60 text-xs text-amber-200">
          <AlertTriangle className="w-3.5 h-3.5 text-amber-400 flex-shrink-0" />
          {runError}
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-[360px_1fr] flex-1 min-h-0">
        <div className="border-b xl:border-b-0 xl:border-r border-slate-800 overflow-y-auto">
          <div className="p-4 space-y-4">
            <div className="space-y-3">
              <label className="block">
                <span className="block text-xs font-medium text-slate-400 mb-1">Linear Project</span>
                <input
                  value={linearProject}
                  onChange={(event) => setLinearProject(event.target.value)}
                  placeholder="project slug, name, id, or URL"
                  className={INPUT_CLASS}
                />
              </label>
              <label className="block">
                <span className="block text-xs font-medium text-slate-400 mb-1">Filter</span>
                <div className="relative">
                  <Search className="absolute left-3 top-2.5 w-4 h-4 text-slate-500" />
                  <input
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder="label, area, title"
                    className={`${INPUT_CLASS} pl-9`}
                  />
                </div>
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block">
                  <span className="block text-xs font-medium text-slate-400 mb-1">Limit</span>
                  <input
                    type="number"
                    value={limit}
                    min={1}
                    max={250}
                    onChange={(event) => setLimit(parseInt(event.target.value, 10) || 50)}
                    className={INPUT_CLASS}
                  />
                </label>
                <label className="block">
                  <span className="block text-xs font-medium text-slate-400 mb-1">Workers</span>
                  <input
                    type="number"
                    value={maxParallel}
                    min={1}
                    max={8}
                    onChange={(event) => setMaxParallel(parseInt(event.target.value, 10) || 3)}
                    className={INPUT_CLASS}
                  />
                </label>
              </div>
              <label className="block">
                <span className="block text-xs font-medium text-slate-400 mb-1">States</span>
                <input
                  value={states}
                  onChange={(event) => setStates(event.target.value)}
                  className={INPUT_CLASS}
                />
              </label>
              <label className="flex items-center gap-2 text-sm text-slate-300">
                <input
                  type="checkbox"
                  checked={includeCanceled}
                  onChange={(event) => setIncludeCanceled(event.target.checked)}
                  className="rounded border-slate-600 bg-slate-950 text-emerald-500 focus:ring-emerald-500"
                />
                Include closed tickets
              </label>
            </div>

            <div className="space-y-2">
              <div className="text-xs font-medium uppercase tracking-wide text-slate-500">Model Routing</div>
              <div className="grid grid-cols-1 gap-2">
                <ModelSelect label="Plan" value={models.plan || ''} onChange={(value) => handleModelChange('plan', value)} />
                <ModelSelect label="Implementation" value={models.implementation || ''} onChange={(value) => handleModelChange('implementation', value)} />
                <ModelSelect label="Skills" value={models.skills || ''} onChange={(value) => handleModelChange('skills', value)} />
              </div>
            </div>

            {result && <StatsPanel result={result} />}
            {runResult && <RunResultPanel result={runResult} />}
          </div>
        </div>

        <div className="min-h-0 overflow-hidden">
          {!result ? (
            <EmptyState loading={runState === 'loading'} />
          ) : (
            <div className="grid grid-cols-1 2xl:grid-cols-[1fr_420px] h-full min-h-0">
              <div className="min-h-0 overflow-y-auto p-4 space-y-4">
                <BatchList
                  batches={result.plan.batches}
                  ticketById={ticketById}
                  onSelectTicket={setSelectedTicketId}
                  onRunBatch={(batch) => handleRunTickets(batch.ticketIds, batch.id)}
                  runDisabled={workerRunning}
                />
                <TicketTable
                  tickets={result.plan.tickets}
                  selectedTicketId={selectedTicket?.ticket.id || null}
                  onSelectTicket={setSelectedTicketId}
                />
              </div>
              <div className="border-t 2xl:border-t-0 2xl:border-l border-slate-800 min-h-0 overflow-y-auto">
                {selectedTicket && (
                  <TicketDetail
                    ticket={selectedTicket}
                    copied={copiedTicketId === selectedTicket.ticket.id}
                    onCopyPrompt={() => handleCopyPrompt(selectedTicket)}
                    onRunTicket={() => handleRunTickets([selectedTicket.ticket.id])}
                    runDisabled={workerRunning || !frontendQueuePolicyRunnable(selectedTicket.policy)}
                  />
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function EmptyState({ loading }: { loading: boolean }) {
  return (
    <div className="h-full flex items-center justify-center">
      <div className="text-center text-slate-500">
        {loading ? (
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-3 text-emerald-400" />
        ) : (
          <Workflow className="w-8 h-8 mx-auto mb-3 text-slate-600" />
        )}
        <p className="text-sm">{loading ? 'Planning queue...' : 'No frontend queue planned'}</p>
      </div>
    </div>
  );
}

function StatsPanel({ result }: { result: FrontendTicketQueueResult }) {
  const stats = result.plan.stats;
  return (
    <div className="grid grid-cols-2 gap-2">
      <Stat label="Tickets" value={stats.totalTickets} />
      <Stat label="Eligible" value={stats.eligibleCount} />
      <Stat label="Auto PR" value={stats.autoPrCount} />
      <Stat label="Draft PR" value={stats.draftPrCount} />
      <Stat label="Plan Only" value={stats.planOnlyCount} />
      <Stat label="Blocked" value={stats.blockedCount} />
    </div>
  );
}

function RunResultPanel({ result }: { result: FrontendTicketQueueRunResult }) {
  return (
    <div className="rounded-md border border-slate-800 bg-slate-950 p-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="text-xs font-medium uppercase tracking-wide text-slate-500">Last Run</div>
          <div className="mt-1 text-sm font-mono text-slate-300 truncate">{result.runId}</div>
        </div>
        <div className="text-right text-xs text-slate-500">
          <div>{result.startedCount} started</div>
          <div>{result.skippedCount} skipped</div>
        </div>
      </div>
      <div className="mt-3 space-y-1.5">
        {result.sessions.map((session) => (
          <div key={`${session.ticketId}:${session.sessionId || session.status}`} className="flex items-center justify-between gap-2 text-xs">
            <span className="font-mono text-slate-400">{session.ticketId}</span>
            <span className={session.status === 'started' ? 'text-emerald-300' : 'text-amber-300'}>{session.status}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border border-slate-800 bg-slate-950 px-3 py-2">
      <div className="text-lg font-semibold text-slate-100">{value}</div>
      <div className="text-[11px] uppercase tracking-wide text-slate-500">{label}</div>
    </div>
  );
}

function BatchList({
  batches,
  ticketById,
  onSelectTicket,
  onRunBatch,
  runDisabled,
}: {
  batches: FrontendQueueBatch[];
  ticketById: Map<string, FrontendTicketPlan>;
  onSelectTicket: (ticketId: string) => void;
  onRunBatch: (batch: FrontendQueueBatch) => void;
  runDisabled: boolean;
}) {
  if (batches.length === 0) {
    return (
      <div className="rounded-md border border-slate-800 bg-slate-950 p-4 text-sm text-slate-500">
        No runnable batches.
      </div>
    );
  }
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-slate-500">
        <Layers className="w-3.5 h-3.5" />
        Parallel Batches
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
        {batches.map((batch) => (
          <div key={batch.id} className="rounded-md border border-slate-800 bg-slate-950 p-3">
            <div className="flex items-center justify-between gap-3">
              <h3 className="text-sm font-medium text-slate-200">{batch.name}</h3>
              <div className="flex items-center gap-2">
                <span className="text-xs text-slate-500">{batch.parallelism} workers</span>
                <button
                  onClick={() => onRunBatch(batch)}
                  disabled={runDisabled}
                  className="inline-flex items-center gap-1 px-2 py-1 text-xs rounded border border-emerald-700 text-emerald-300 hover:bg-emerald-950/50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <GitPullRequest className="w-3 h-3" />
                  Run
                </button>
              </div>
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              {batch.ticketIds.map((ticketId) => {
                const ticket = ticketById.get(ticketId);
                if (!ticket) return null;
                return (
                  <button
                    key={ticketId}
                    onClick={() => onSelectTicket(ticketId)}
                    className="inline-flex items-center gap-1.5 px-2 py-1 text-xs rounded border border-slate-700 text-slate-300 hover:bg-slate-800"
                    title={ticket.ticket.title}
                  >
                    <PolicyDot policy={ticket.policy} />
                    {ticketId}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function TicketTable({
  tickets,
  selectedTicketId,
  onSelectTicket,
}: {
  tickets: FrontendTicketPlan[];
  selectedTicketId: string | null;
  onSelectTicket: (ticketId: string) => void;
}) {
  return (
    <div className="rounded-md border border-slate-800 overflow-hidden">
      <div className="grid grid-cols-[110px_1fr_150px_100px_90px] gap-3 px-3 py-2 bg-slate-800 text-xs font-medium uppercase tracking-wide text-slate-500">
        <span>Ticket</span>
        <span>Title</span>
        <span>Type</span>
        <span>Policy</span>
        <span>Risk</span>
      </div>
      <div className="divide-y divide-slate-800 bg-slate-950">
        {tickets.map((ticket) => (
          <button
            key={ticket.ticket.id}
            onClick={() => onSelectTicket(ticket.ticket.id)}
            className={`grid grid-cols-[110px_1fr_150px_100px_90px] gap-3 w-full px-3 py-2 text-left text-sm transition-colors ${
              selectedTicketId === ticket.ticket.id ? 'bg-slate-800/80' : 'hover:bg-slate-900'
            }`}
          >
            <span className="font-mono text-xs text-slate-400 truncate">{ticket.ticket.id}</span>
            <span className="text-slate-200 truncate">{ticket.ticket.title}</span>
            <span className="text-slate-400 truncate">{typeLabel(ticket.type)}</span>
            <PolicyBadge policy={ticket.policy} />
            <span className="text-slate-400">{ticket.riskScore}/10</span>
          </button>
        ))}
      </div>
    </div>
  );
}

function TicketDetail({
  ticket,
  copied,
  onCopyPrompt,
  onRunTicket,
  runDisabled,
}: {
  ticket: FrontendTicketPlan;
  copied: boolean;
  onCopyPrompt: () => void;
  onRunTicket: () => void;
  runDisabled: boolean;
}) {
  return (
    <div className="p-4 space-y-4">
      <div>
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-xs font-mono text-slate-500">{ticket.ticket.id}</p>
            <h3 className="mt-1 text-base font-semibold text-slate-100">{ticket.ticket.title}</h3>
          </div>
          {ticket.ticket.url && (
            <a
              href={ticket.ticket.url}
              target="_blank"
              rel="noreferrer"
              className="p-1.5 text-slate-400 hover:text-slate-100 hover:bg-slate-800 rounded"
              aria-label="Open Linear ticket"
            >
              <ExternalLink className="w-4 h-4" />
            </a>
          )}
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          <PolicyBadge policy={ticket.policy} />
          <span className="px-2 py-1 text-xs rounded border border-slate-700 text-slate-300">{typeLabel(ticket.type)}</span>
          <span className="px-2 py-1 text-xs rounded border border-slate-700 text-slate-300">{Math.round(ticket.confidence * 100)}% confidence</span>
          {ticket.passKCandidates ? (
            <span className="px-2 py-1 text-xs rounded border border-emerald-700 text-emerald-300">Pass@{ticket.passKCandidates}</span>
          ) : null}
        </div>
      </div>

      <DetailSection icon={<FileCode2 className="w-4 h-4" />} title="Target">
        <p className="text-sm font-mono text-slate-300 break-all">{ticket.targetKey}</p>
        <div className="mt-2 flex flex-wrap gap-1.5">
          {(ticket.targetHints || []).slice(0, 8).map((hint) => (
            <span key={`${hint.kind}:${hint.value}`} className="px-1.5 py-0.5 text-[11px] rounded bg-slate-800 text-slate-400">
              {hint.kind}: {hint.value}
            </span>
          ))}
        </div>
      </DetailSection>

      <DetailSection icon={<ShieldCheck className="w-4 h-4" />} title="Validation">
        <div className="space-y-2">
          {(ticket.validation || []).map((step) => (
            <div key={`${step.kind}:${step.description}`} className="text-sm text-slate-300">
              <div className="flex items-start gap-2">
                <CheckCircle2 className={`w-3.5 h-3.5 mt-0.5 ${step.required ? 'text-emerald-400' : 'text-slate-500'}`} />
                <span>{step.description}</span>
              </div>
              {step.command && <p className="ml-5 mt-0.5 text-xs font-mono text-slate-500">{step.command}</p>}
            </div>
          ))}
        </div>
      </DetailSection>

      {(ticket.reasons || []).length > 0 && (
        <DetailSection icon={<GitPullRequest className="w-4 h-4" />} title="Decision">
          <ul className="space-y-1 text-sm text-slate-300">
            {(ticket.reasons || []).map((reason) => (
              <li key={reason}>{reason}</li>
            ))}
          </ul>
        </DetailSection>
      )}

      {(ticket.blockers || []).length > 0 && (
        <DetailSection icon={<AlertTriangle className="w-4 h-4" />} title="Blockers">
          <ul className="space-y-1 text-sm text-amber-200">
            {(ticket.blockers || []).map((blocker) => (
              <li key={blocker}>{blocker}</li>
            ))}
          </ul>
        </DetailSection>
      )}

      <div className="rounded-md border border-slate-800 bg-slate-950 overflow-hidden">
        <div className="flex items-center justify-between gap-3 px-3 py-2 border-b border-slate-800">
          <span className="text-xs font-medium uppercase tracking-wide text-slate-500">Worker Prompt</span>
          <div className="flex items-center gap-1">
            <button
              onClick={onRunTicket}
              disabled={runDisabled}
              className="inline-flex items-center gap-1.5 px-2 py-1 text-xs text-emerald-300 hover:text-emerald-100 hover:bg-emerald-950/50 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <GitPullRequest className="w-3.5 h-3.5" />
              Run
            </button>
            <button
              onClick={onCopyPrompt}
              className="inline-flex items-center gap-1.5 px-2 py-1 text-xs text-slate-300 hover:text-slate-100 hover:bg-slate-800 rounded"
            >
              <Clipboard className="w-3.5 h-3.5" />
              {copied ? 'Copied' : 'Copy'}
            </button>
          </div>
        </div>
        <pre className="max-h-72 overflow-auto p-3 text-xs text-slate-300 whitespace-pre-wrap">{ticket.workerPrompt}</pre>
      </div>
    </div>
  );
}

function DetailSection({ icon, title, children }: { icon: ReactNode; title: string; children: ReactNode }) {
  return (
    <section className="rounded-md border border-slate-800 bg-slate-950 p-3">
      <div className="flex items-center gap-2 mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
        {icon}
        {title}
      </div>
      {children}
    </section>
  );
}

function PolicyBadge({ policy }: { policy: FrontendAutomationPolicy }) {
  return (
    <span className={`inline-flex items-center justify-center px-2 py-0.5 text-xs rounded border ${policyClass(policy)}`}>
      {policyLabel(policy)}
    </span>
  );
}

function PolicyDot({ policy }: { policy: FrontendAutomationPolicy }) {
  return <span className={`w-2 h-2 rounded-full ${policyDotClass(policy)}`} />;
}

function ModelSelect({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  return (
    <label className="grid grid-cols-[110px_1fr] items-center gap-2 text-xs text-slate-400">
      <span>{label}</span>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="w-full px-2 py-1.5 text-xs bg-slate-950 border border-slate-700 rounded-md text-slate-100 focus:outline-none focus:border-emerald-500"
      >
        <option value="">Default</option>
        {MODEL_OPTIONS.map((option) => (
          <option key={option.value} value={option.value}>{option.label}</option>
        ))}
      </select>
    </label>
  );
}

function frontendQueuePolicyRunnable(policy: FrontendAutomationPolicy): boolean {
  return policy === 'auto-pr' || policy === 'draft-pr';
}

function policyClass(policy: FrontendAutomationPolicy): string {
  switch (policy) {
    case 'auto-pr':
      return 'border-emerald-700 bg-emerald-950/60 text-emerald-300';
    case 'draft-pr':
      return 'border-blue-700 bg-blue-950/60 text-blue-300';
    case 'plan-only':
      return 'border-amber-700 bg-amber-950/60 text-amber-300';
    case 'blocked':
      return 'border-red-800 bg-red-950/60 text-red-300';
    default:
      return 'border-slate-700 bg-slate-900 text-slate-300';
  }
}

function policyDotClass(policy: FrontendAutomationPolicy): string {
  switch (policy) {
    case 'auto-pr':
      return 'bg-emerald-400';
    case 'draft-pr':
      return 'bg-blue-400';
    case 'plan-only':
      return 'bg-amber-400';
    case 'blocked':
      return 'bg-red-400';
    default:
      return 'bg-slate-500';
  }
}

function policyLabel(policy: FrontendAutomationPolicy): string {
  switch (policy) {
    case 'auto-pr':
      return 'Auto PR';
    case 'draft-pr':
      return 'Draft PR';
    case 'plan-only':
      return 'Plan Only';
    case 'blocked':
      return 'Blocked';
    default:
      return policy;
  }
}

function typeLabel(type: FrontendTicketType): string {
  return type
    .split('-')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message;
  if (typeof err === 'string' && err) return err;
  return fallback;
}
