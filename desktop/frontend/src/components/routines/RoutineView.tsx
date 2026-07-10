import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  DatabaseZap,
  FileText,
  KeyRound,
  Loader2,
  Play,
  RefreshCw,
  Settings,
  Sparkles,
} from 'lucide-react';
import {
  AuthenticateDatadogMCP,
  DryRunRoutine,
  ListProjectRoutines,
  ListRoutines,
  RunRoutine,
} from '../../../wailsjs/go/main/App';
import type {
  DesktopRoutine,
  DatadogMCPAuthResult,
  IntegrationStatus,
  RoutineDryRunResult,
  RoutineRunResult,
  RoutineRunRequest,
} from '../../types';

interface RoutineViewProps {
  projectPath?: string;
  onOpenSettings?: () => void;
  onRoutineSessionOpen?: (sessionId: string) => void | Promise<void>;
}

type RunState = 'idle' | 'checking' | 'running' | 'complete' | 'error';
type DatadogAuthPhase = 'idle' | 'opening' | 'checking';

const DEFAULT_ROUTINE_ID = 'datadog-gql-slow-queries';
const INPUT_CLASS = 'w-full px-3 py-2 text-sm bg-slate-950 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:border-teal-500';
const DATADOG_AUTH_POLL_ATTEMPTS = 40;
const DATADOG_AUTH_POLL_DELAY_MS = 3000;

export function RoutineView({ projectPath, onOpenSettings, onRoutineSessionOpen }: RoutineViewProps) {
  const [routines, setRoutines] = useState<DesktopRoutine[]>([]);
  const [selectedRoutineId, setSelectedRoutineId] = useState(DEFAULT_ROUTINE_ID);
  const [values, setValues] = useState<Record<string, string>>({});
  const [model, setModel] = useState('');
  const [dryRun, setDryRun] = useState<RoutineDryRunResult | null>(null);
  const [result, setResult] = useState<RoutineRunResult | null>(null);
  const [runState, setRunState] = useState<RunState>('idle');
  const [datadogAuthPhase, setDatadogAuthPhase] = useState<DatadogAuthPhase>('idle');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    setDryRun(null);
    setResult(null);
    setRunState('idle');
    setNotice(null);
    const list = projectPath ? ListProjectRoutines(projectPath) : ListRoutines();
    list
      .then((items) => {
        if (!mounted) return;
        const loaded = (items || []) as unknown as DesktopRoutine[];
        setRoutines(loaded);
        if (loaded.length > 0) {
          const next = loaded.find((routine) => routine.id === selectedRoutineId)
            || loaded.find((routine) => routine.id === DEFAULT_ROUTINE_ID)
            || loaded[0];
          setSelectedRoutineId(next.id);
        }
      })
      .catch((err) => {
        if (!mounted) return;
        setError(errorMessage(err, 'Failed to load routines'));
      });
    return () => {
      mounted = false;
    };
  }, [projectPath]);

  const selectedRoutine = useMemo(
    () => routines.find((routine) => routine.id === selectedRoutineId) || routines[0] || null,
    [routines, selectedRoutineId]
  );

  useEffect(() => {
    if (!selectedRoutine) {
      setValues({});
      return;
    }
    setValues(defaultValuesForRoutine(selectedRoutine));
  }, [selectedRoutine]);

  const selectedStatuses = result?.integrations || dryRun?.integrations || [];
  const hasMissingRequired = selectedRoutine
    ? (selectedRoutine.parameters || []).some((param) => param.required && !String(values[param.name] || '').trim())
    : true;
  const canRun = Boolean(projectPath && selectedRoutine && !hasMissingRequired && runState !== 'checking' && runState !== 'running');

  const buildRequest = (): RoutineRunRequest => ({
    routineId: selectedRoutine?.id || selectedRoutineId,
    projectPath: projectPath || '',
    model: model.trim(),
    values: routineValues(selectedRoutine, values),
  });

  const refreshRoutineCheck = async () => {
    const response = (await DryRunRoutine(buildRequest())) as unknown as RoutineDryRunResult;
    setDryRun(response);
    setResult(null);
    return response;
  };

  const handleDryRun = async () => {
    if (!selectedRoutine || !projectPath) {
      setError('Open a project before checking a routine.');
      return;
    }
    setRunState('checking');
    setError(null);
    setNotice(null);
    setResult(null);
    try {
      await refreshRoutineCheck();
      setRunState('idle');
    } catch (err) {
      setError(errorMessage(err, 'Routine check failed'));
      setRunState('error');
    }
  };

  const handleRun = async () => {
    if (!canRun) {
      setError(projectPath ? 'Required routine parameters are missing.' : 'Open a project before running a routine.');
      return;
    }
    setRunState('running');
    setError(null);
    setNotice(null);
    setResult(null);
    try {
      const response = await RunRoutine(buildRequest());
      setResult(response as unknown as RoutineRunResult);
      setDryRun(null);
      if (response?.sessionId) {
        await onRoutineSessionOpen?.(response.sessionId);
      }
      setRunState('complete');
    } catch (err) {
      setError(errorMessage(err, 'Routine run failed'));
      setRunState('error');
    }
  };

  const pollDatadogReadiness = async () => {
    let latest: RoutineDryRunResult | null = null;
    for (let attempt = 0; attempt < DATADOG_AUTH_POLL_ATTEMPTS; attempt += 1) {
      if (attempt > 0) {
        await sleep(DATADOG_AUTH_POLL_DELAY_MS);
      }
      latest = await refreshRoutineCheck();
      const datadog = integrationStatus(latest.integrations || [], 'datadog');
      if (isReadyIntegration(datadog)) {
        setNotice('Datadog MCP is authenticated and ready for routines.');
        return;
      }
    }

    const datadog = integrationStatus(latest?.integrations || [], 'datadog');
    setNotice(datadog?.message
      ? `Datadog MCP is still not ready: ${datadog.message}`
      : 'Datadog MCP is still not ready. Finish the browser auth flow, then click Check.');
  };

  const handleAuthenticateDatadog = async () => {
    setDatadogAuthPhase('opening');
    setError(null);
    setNotice(null);
    try {
      const response = (await AuthenticateDatadogMCP()) as unknown as DatadogMCPAuthResult;
      const message = response.message || 'Datadog MCP auth opened. Complete the Claude auth flow.';
      if (projectPath && selectedRoutine && !hasMissingRequired) {
        setNotice(`${message} Boatman will keep checking for readiness.`);
        setDatadogAuthPhase('checking');
        await pollDatadogReadiness();
      } else {
        setNotice(`${message} Click Check when the browser flow is complete.`);
      }
    } catch (err) {
      setError(errorMessage(err, 'Datadog MCP authentication failed'));
    } finally {
      setDatadogAuthPhase('idle');
    }
  };

  return (
    <div className="flex flex-col h-full bg-slate-900">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-2 min-w-0">
          <Sparkles className="w-5 h-5 text-teal-400 flex-shrink-0" />
          <h2 className="text-sm font-semibold text-slate-200">Routines</h2>
          {selectedRoutine?.schedule && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 text-xs bg-slate-700 rounded text-slate-400">
              <Clock className="w-3 h-3" />
              {selectedRoutine.schedule}
            </span>
          )}
        </div>
        <button
          onClick={handleDryRun}
          disabled={!projectPath || runState === 'checking' || runState === 'running'}
          className="inline-flex items-center gap-2 px-3 py-1.5 text-xs border border-slate-700 text-slate-300 rounded-md hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {runState === 'checking' ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <RefreshCw className="w-3.5 h-3.5" />}
          Check
        </button>
      </div>

      {error && (
        <div className="flex items-center gap-2 px-4 py-2 bg-red-950/50 border-b border-red-900/60 text-xs text-red-200">
          <AlertCircle className="w-3.5 h-3.5 text-red-400 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}
      {notice && (
        <div className="flex items-center gap-2 px-4 py-2 bg-teal-950/40 border-b border-teal-900/60 text-xs text-teal-100">
          <CheckCircle2 className="w-3.5 h-3.5 text-teal-300 flex-shrink-0" />
          <span>{notice}</span>
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-[420px_1fr] flex-1 min-h-0">
        <div className="border-b xl:border-b-0 xl:border-r border-slate-800 overflow-y-auto">
          <div className="p-4 space-y-4">
            <RoutinePicker
              routines={routines}
              selectedRoutineId={selectedRoutine?.id || selectedRoutineId}
              onSelect={(routineId) => {
                setSelectedRoutineId(routineId);
                setDryRun(null);
                setResult(null);
                setError(null);
                setNotice(null);
              }}
            />

            {selectedRoutine && <RoutineSummary routine={selectedRoutine} />}

            {selectedRoutine && (
              <RoutineParameters
                routine={selectedRoutine}
                values={values}
                onChange={(name, value) => setValues((current) => ({ ...current, [name]: value }))}
              />
            )}

            <Field label="Model">
              <input
                value={model}
                onChange={(event) => setModel(event.target.value)}
                placeholder="default"
                className={INPUT_CLASS}
              />
            </Field>

            {selectedRoutine && (
              <IntegrationsPanel
                routine={selectedRoutine}
                statuses={selectedStatuses}
                onOpenSettings={onOpenSettings}
                onAuthenticateDatadog={handleAuthenticateDatadog}
                datadogAuthPhase={datadogAuthPhase}
              />
            )}

            <button
              onClick={handleRun}
              disabled={!canRun}
              className="w-full inline-flex items-center justify-center gap-2 px-4 py-2 bg-teal-500 text-slate-950 font-medium rounded-md hover:bg-teal-400 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {runState === 'running' ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
              {runState === 'running' ? 'Running' : 'Run Routine'}
            </button>
          </div>
        </div>

        <div className="min-h-0 overflow-y-auto">
          <RoutineOutputPane
            projectPath={projectPath}
            dryRun={dryRun}
            result={result}
            runState={runState}
          />
        </div>
      </div>
    </div>
  );
}

function RoutinePicker({
  routines,
  selectedRoutineId,
  onSelect,
}: {
  routines: DesktopRoutine[];
  selectedRoutineId: string;
  onSelect: (routineId: string) => void;
}) {
  return (
    <Field label="Routine">
      <select
        value={selectedRoutineId}
        onChange={(event) => onSelect(event.target.value)}
        className={INPUT_CLASS}
      >
        {routines.length === 0 ? (
          <option value="">Loading routines</option>
        ) : (
          routines.map((routine) => (
            <option key={routine.id} value={routine.id}>
              {routine.name}
            </option>
          ))
        )}
      </select>
    </Field>
  );
}

function RoutineSummary({ routine }: { routine: DesktopRoutine }) {
  return (
    <div className="border border-slate-800 rounded-md p-3 bg-slate-950/40">
      <div className="flex items-center gap-2">
        <DatabaseZap className="w-4 h-4 text-teal-400" />
        <h3 className="text-sm font-medium text-slate-100">{routine.name}</h3>
      </div>
      <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-400">
        <span className="px-1.5 py-0.5 bg-slate-800 rounded">{routine.workflowTemplate || 'workflow'}</span>
        <span className="px-1.5 py-0.5 bg-slate-800 rounded">{routine.profile}</span>
        {(routine.integrations || []).map((integration) => (
          <span key={integration} className="px-1.5 py-0.5 bg-slate-800 rounded">
            {integration}
          </span>
        ))}
      </div>
    </div>
  );
}

function RoutineParameters({
  routine,
  values,
  onChange,
}: {
  routine: DesktopRoutine;
  values: Record<string, string>;
  onChange: (name: string, value: string) => void;
}) {
  const parameters = routine.parameters || [];
  if (parameters.length === 0) {
    return (
      <div className="border border-slate-800 rounded-md p-3 bg-slate-950/40 text-sm text-slate-500">
        No parameters
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {parameters.map((param) => (
        <Field key={param.name} label={parameterLabel(param.name)} required={param.required}>
          <input
            type={param.type === 'integer' ? 'number' : 'text'}
            value={values[param.name] || ''}
            onChange={(event) => onChange(param.name, event.target.value)}
            placeholder={param.default || param.description || ''}
            className={INPUT_CLASS}
          />
        </Field>
      ))}
    </div>
  );
}

function IntegrationsPanel({
  routine,
  statuses,
  onOpenSettings,
  onAuthenticateDatadog,
  datadogAuthPhase,
}: {
  routine: DesktopRoutine;
  statuses: IntegrationStatus[];
  onOpenSettings?: () => void;
  onAuthenticateDatadog?: () => void;
  datadogAuthPhase?: DatadogAuthPhase;
}) {
  const statusByName = new Map(statuses.map((status) => [status.name, status]));
  const integrations = routine.integrations || [];
  const authInProgress = datadogAuthPhase && datadogAuthPhase !== 'idle';

  return (
    <div className="border border-slate-800 rounded-md p-3 bg-slate-950/40">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 min-w-0">
          <DatabaseZap className="w-4 h-4 text-teal-400" />
          <span className="text-sm font-medium text-slate-200">Integrations</span>
        </div>
        {onOpenSettings && (
          <button
            onClick={onOpenSettings}
            className="inline-flex items-center gap-1.5 px-2 py-1 text-xs text-slate-300 border border-slate-700 rounded-md hover:bg-slate-800"
          >
            <Settings className="w-3.5 h-3.5" />
            Settings
          </button>
        )}
      </div>

      {integrations.length === 0 ? (
        <div className="mt-3 text-xs text-slate-500">None</div>
      ) : (
        <div className="mt-3 space-y-2">
          {integrations.map((name) => {
            const status = statusByName.get(name);
            const state = status?.state || 'not_checked';
            const isReady = state === 'connected' || state === 'ready';
            const missingEnv = status?.missingEnv || [];
            const canAuthenticate = name === 'datadog' && !isReady && onAuthenticateDatadog;
            return (
              <div key={name} className="border border-slate-800 rounded-md p-2 bg-slate-900/60">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2 min-w-0">
                    {isReady ? <CheckCircle2 className="w-4 h-4 text-emerald-400" /> : <AlertCircle className="w-4 h-4 text-amber-400" />}
                    <span className="text-sm text-slate-200">{name}</span>
                    <StatusBadge status={state} />
                  </div>
                  {canAuthenticate && (
                    <button
                      type="button"
                      onClick={onAuthenticateDatadog}
                      disabled={authInProgress}
                      className="inline-flex items-center gap-1.5 px-2 py-1 text-xs text-teal-200 border border-teal-700 rounded-md hover:bg-teal-950/40 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {authInProgress ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <KeyRound className="w-3.5 h-3.5" />}
                      {datadogAuthPhase === 'checking' ? 'Checking' : datadogAuthPhase === 'opening' ? 'Opening' : 'Authenticate'}
                    </button>
                  )}
                </div>
                {missingEnv.length > 0 && (
                  <div className="mt-2 text-xs text-amber-200">
                    Missing {missingEnv.join(', ')}
                  </div>
                )}
                {status?.message && (
                  <div className="mt-2 text-xs text-slate-500">
                    {status.message}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function RoutineOutputPane({
  projectPath,
  dryRun,
  result,
  runState,
}: {
  projectPath?: string;
  dryRun: RoutineDryRunResult | null;
  result: RoutineRunResult | null;
  runState: RunState;
}) {
  if (!projectPath) {
    return <EmptyState title="Open a project" body="Routines write reports and runtime records under the active project." />;
  }

  if (runState === 'running') {
    return <EmptyState title="Routine running" body="The report will appear here when the model finishes." loading />;
  }

  if (result) {
    return (
      <div className="p-4 space-y-4">
        <HeaderBlock
          title="Report"
          subtitle={result.reportPath || result.runId}
          icon={<FileText className="w-4 h-4 text-teal-400" />}
        />
        <RunMetrics result={result} />
        <pre className="whitespace-pre-wrap text-sm leading-6 text-slate-200 bg-slate-950 border border-slate-800 rounded-md p-4 overflow-x-auto">
          {result.report || 'No report text returned.'}
        </pre>
      </div>
    );
  }

  if (dryRun) {
    return (
      <div className="p-4 space-y-4">
        <HeaderBlock
          title="Dry Run"
          subtitle={dryRun.reportPath || dryRun.request.runId}
          icon={<CheckCircle2 className="w-4 h-4 text-cyan-400" />}
        />
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          <DetailBlock label="Provider" value={dryRun.request.provider} />
          <DetailBlock label="Model" value={dryRun.request.model || 'default'} />
          <DetailBlock label="MCP" value={(dryRun.request.mcpServerLabels || []).join(', ') || '-'} />
          <DetailBlock label="Reasoning" value={dryRun.request.reasoningEffort || '-'} />
        </div>
        <div className="border border-slate-800 rounded-md p-3 bg-slate-950/40">
          <div className="text-xs uppercase tracking-wide text-slate-500 mb-2">Prompt</div>
          <pre className="whitespace-pre-wrap text-xs leading-5 text-slate-300">{dryRun.request.firstMessagePreview}</pre>
        </div>
      </div>
    );
  }

  return <EmptyState title="No routine selected" body="Check or run a routine to see its request and report." />;
}

function HeaderBlock({ title, subtitle, icon }: { title: string; subtitle?: string; icon: ReactNode }) {
  return (
    <div className="flex items-start gap-2">
      <div className="mt-0.5">{icon}</div>
      <div className="min-w-0">
        <h3 className="text-sm font-semibold text-slate-100">{title}</h3>
        {subtitle && <p className="text-xs text-slate-500 truncate">{subtitle}</p>}
      </div>
    </div>
  );
}

function RunMetrics({ result }: { result: RoutineRunResult }) {
  const usage = result.usage;
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <DetailBlock label="Run" value={result.runId} />
      <DetailBlock label="Provider" value={result.provider} />
      <DetailBlock label="Tokens" value={usage ? `${usage.inputTokens || 0} / ${usage.outputTokens || 0}` : '-'} />
      <DetailBlock label="Cost" value={usage?.totalCostUsd ? `$${usage.totalCostUsd.toFixed(4)}` : '-'} />
    </div>
  );
}

function Field({ label, required, children }: { label: string; required?: boolean; children: ReactNode }) {
  return (
    <label className="block">
      <span className="flex items-center gap-1 text-xs font-medium text-slate-400 mb-1">
        {label}
        {required && <span className="text-teal-400">*</span>}
      </span>
      {children}
    </label>
  );
}

function DetailBlock({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-slate-800 rounded-md p-3 bg-slate-950/40 min-w-0">
      <div className="text-xs uppercase tracking-wide text-slate-500">{label}</div>
      <div className="mt-1 text-sm text-slate-200 truncate">{value}</div>
    </div>
  );
}

function EmptyState({ title, body, loading }: { title: string; body: string; loading?: boolean }) {
  return (
    <div className="h-full min-h-[280px] flex items-center justify-center p-6">
      <div className="text-center max-w-sm">
        <div className="w-12 h-12 mx-auto mb-4 rounded-md bg-slate-800 flex items-center justify-center">
          {loading ? <Loader2 className="w-6 h-6 text-teal-400 animate-spin" /> : <FileText className="w-6 h-6 text-slate-500" />}
        </div>
        <h3 className="text-sm font-semibold text-slate-200">{title}</h3>
        <p className="mt-2 text-sm text-slate-500">{body}</p>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const classes = status === 'connected' || status === 'ready'
    ? 'bg-emerald-950 text-emerald-300 border-emerald-900'
    : status === 'needs_configuration'
      ? 'bg-amber-950 text-amber-300 border-amber-900'
      : status === 'failed'
        ? 'bg-red-950 text-red-300 border-red-900'
        : 'bg-slate-800 text-slate-400 border-slate-700';

  return (
    <span className={`px-1.5 py-0.5 text-xs rounded border ${classes}`}>
      {status.replace('_', ' ')}
    </span>
  );
}

function integrationStatus(statuses: IntegrationStatus[], name: string) {
  return statuses.find((status) => status.name === name);
}

function isReadyIntegration(status?: IntegrationStatus) {
  return status?.state === 'connected' || status?.state === 'ready';
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message;
  if (typeof err === 'string' && err.trim()) return err;
  return fallback;
}

function defaultValuesForRoutine(routine: DesktopRoutine): Record<string, string> {
  const defaults: Record<string, string> = {};
  for (const param of routine.parameters || []) {
    defaults[param.name] = param.default || '';
  }
  return defaults;
}

function routineValues(routine: DesktopRoutine | null, values: Record<string, string>): Record<string, string> {
  const out: Record<string, string> = {};
  if (!routine) return out;
  for (const param of routine.parameters || []) {
    const value = String(values[param.name] || '').trim();
    if (value) {
      out[param.name] = value;
    }
  }
  return out;
}

function parameterLabel(name: string): string {
  return name
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}
