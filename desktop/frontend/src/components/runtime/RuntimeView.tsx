import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import {
  Activity,
  AlertCircle,
  Archive,
  Bot,
  Braces,
  CheckCircle2,
  ChevronRight,
  Clock,
  Database,
  FileText,
  RefreshCw,
  Search,
  Server,
  Wrench,
} from 'lucide-react';
import { useRuntimeInspector } from '../../hooks/useRuntimeInspector';
import type {
  MemoryDocumentDetail,
  MemoryDocumentSummary,
  RuntimeArtifactSummary,
  RuntimeEventSummary,
  RuntimeRequestSummary,
  RuntimeRunDetail,
  RuntimeRunSummary,
} from '../../types';

type RuntimeTab = 'runs' | 'memory';

interface RuntimeViewProps {
  projectPath?: string;
}

export function RuntimeView({ projectPath }: RuntimeViewProps) {
  const {
    runs,
    selectedRun,
    memoryDocuments,
    selectedMemoryDocument,
    isLoadingRuns,
    isLoadingMemory,
    error,
    loadRuns,
    loadRun,
    loadMemoryDocuments,
    loadMemoryDocument,
  } = useRuntimeInspector(projectPath);

  const [activeTab, setActiveTab] = useState<RuntimeTab>('runs');
  const [query, setQuery] = useState('');

  useEffect(() => {
    loadRuns();
    loadMemoryDocuments();
  }, [loadRuns, loadMemoryDocuments]);

  const filteredRuns = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return runs;
    return runs.filter((run) =>
      [
        run.runId,
        run.provider,
        run.model,
        run.role,
        run.profile,
        run.status,
        run.workDir,
      ].some((value) => value?.toLowerCase().includes(q))
    );
  }, [runs, query]);

  const filteredMemory = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return memoryDocuments;
    return memoryDocuments.filter((doc) =>
      [
        doc.id,
        doc.title,
        doc.scope,
        doc.provenance,
        doc.sourceRunId,
        doc.bodyPreview,
      ].some((value) => value?.toLowerCase().includes(q))
    );
  }, [memoryDocuments, query]);

  const refresh = () => {
    if (activeTab === 'runs') {
      loadRuns();
      if (selectedRun) loadRun(selectedRun.metadata.runId);
    } else {
      loadMemoryDocuments();
      if (selectedMemoryDocument) loadMemoryDocument(selectedMemoryDocument.id);
    }
  };

  const isLoading = activeTab === 'runs' ? isLoadingRuns : isLoadingMemory;

  return (
    <div className="flex flex-col h-full bg-slate-900">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-cyan-400" />
          <h2 className="text-sm font-semibold text-slate-200">Runtime</h2>
          <span className="px-1.5 py-0.5 text-xs bg-slate-700 rounded-full text-slate-400">
            {runs.length} runs
          </span>
          <span className="px-1.5 py-0.5 text-xs bg-slate-700 rounded-full text-slate-400">
            {memoryDocuments.length} memories
          </span>
        </div>
        <button
          onClick={refresh}
          className="p-1.5 text-slate-400 hover:text-slate-200 transition-colors"
          title="Refresh runtime data"
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {error && (
        <div className="flex items-center gap-2 px-4 py-2 bg-red-950/50 border-b border-red-900/60 text-xs text-red-200">
          <AlertCircle className="w-3.5 h-3.5 text-red-400" />
          {error}
        </div>
      )}

      <div className="flex border-b border-slate-700">
        <button
          onClick={() => setActiveTab('runs')}
          className={`flex items-center gap-1.5 px-4 py-2 text-xs transition-colors border-b-2 ${
            activeTab === 'runs'
              ? 'border-cyan-500 text-slate-100'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          <Archive className="w-3.5 h-3.5" />
          Runs
          {runs.length > 0 && (
            <span className="px-1 py-0.5 text-xs bg-cyan-900/40 text-cyan-300 rounded">
              {runs.length}
            </span>
          )}
        </button>
        <button
          onClick={() => setActiveTab('memory')}
          className={`flex items-center gap-1.5 px-4 py-2 text-xs transition-colors border-b-2 ${
            activeTab === 'memory'
              ? 'border-emerald-500 text-slate-100'
              : 'border-transparent text-slate-400 hover:text-slate-200'
          }`}
        >
          <Database className="w-3.5 h-3.5" />
          Memory
          {memoryDocuments.length > 0 && (
            <span className="px-1 py-0.5 text-xs bg-emerald-900/40 text-emerald-300 rounded">
              {memoryDocuments.length}
            </span>
          )}
        </button>
      </div>

      <div className="px-3 py-2 border-b border-slate-800">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
          <input
            type="text"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={activeTab === 'runs' ? 'Search runs...' : 'Search memory...'}
            className="w-full pl-8 pr-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>
      </div>

      <div className="flex-1 min-h-0">
        {activeTab === 'runs' ? (
          <RunsPane
            runs={filteredRuns}
            selectedRun={selectedRun}
            onSelectRun={loadRun}
            isLoading={isLoadingRuns}
          />
        ) : (
          <MemoryPane
            documents={filteredMemory}
            selectedDocument={selectedMemoryDocument}
            onSelectDocument={loadMemoryDocument}
            isLoading={isLoadingMemory}
          />
        )}
      </div>

      <div className="px-3 py-2 border-t border-slate-800 text-xs text-slate-500">
        {projectPath ? (
          <span className="truncate block">{projectPath}/.boatman</span>
        ) : (
          <span>Open a project to inspect its runtime store and memory documents.</span>
        )}
      </div>
    </div>
  );
}

interface RunsPaneProps {
  runs: RuntimeRunSummary[];
  selectedRun: RuntimeRunDetail | null;
  onSelectRun: (runId: string) => void;
  isLoading: boolean;
}

function RunsPane({ runs, selectedRun, onSelectRun, isLoading }: RunsPaneProps) {
  if (runs.length === 0 && !isLoading) {
    return (
      <EmptyState
        icon={<Archive className="w-12 h-12 opacity-30" />}
        title="No runtime runs recorded"
        body="Enable the runtime store or run a Boatman workflow to capture provider-neutral events."
      />
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[minmax(300px,360px)_1fr] h-full min-h-0">
      <div className="border-b lg:border-b-0 lg:border-r border-slate-800 overflow-y-auto min-h-[220px] lg:min-h-0">
        <div className="divide-y divide-slate-800">
          {runs.map((run) => (
            <RunListItem
              key={run.runId}
              run={run}
              selected={selectedRun?.metadata.runId === run.runId}
              onClick={() => onSelectRun(run.runId)}
            />
          ))}
        </div>
      </div>
      <div className="overflow-y-auto">
        {selectedRun ? <RunDetail run={selectedRun} /> : <SelectPrompt label="Select a run to inspect events and artifacts." />}
      </div>
    </div>
  );
}

function RunListItem({ run, selected, onClick }: { run: RuntimeRunSummary; selected: boolean; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-3 transition-colors ${
        selected ? 'bg-cyan-950/40' : 'hover:bg-slate-800/50'
      }`}
    >
      <div className="flex items-center gap-2">
        <ChevronRight className={`w-4 h-4 ${selected ? 'text-cyan-300' : 'text-slate-500'}`} />
        <span className="text-sm font-medium text-slate-200 truncate">{run.runId}</span>
        <StatusBadge status={run.status} />
      </div>
      <div className="mt-1 ml-6 flex flex-wrap gap-x-3 gap-y-1 text-xs text-slate-500">
        {run.provider && <span>{run.provider}</span>}
        {run.model && <span>{run.model}</span>}
        {run.role && <span>{run.role}</span>}
      </div>
      <div className="mt-1 ml-6 flex items-center gap-3 text-xs text-slate-500">
        <span>{run.eventCount} events</span>
        <span>{run.artifactCount || 0} artifacts</span>
        {run.updatedAt && <span>{formatRelative(run.updatedAt)}</span>}
      </div>
    </button>
  );
}

function RunDetail({ run }: { run: RuntimeRunDetail }) {
  const [eventQuery, setEventQuery] = useState('');
  const [eventType, setEventType] = useState('all');
  const [eventStatus, setEventStatus] = useState('all');

  const eventTypes = useMemo(() => uniqueValues(run.events.map((event) => event.type)), [run.events]);
  const eventStatuses = useMemo(() => uniqueValues(run.events.map((event) => event.status).filter(Boolean) as string[]), [run.events]);
  const filteredEvents = useMemo(() => {
    const q = eventQuery.trim().toLowerCase();
    return run.events.filter((event) => {
      if (eventType !== 'all' && event.type !== eventType) return false;
      if (eventStatus !== 'all' && event.status !== eventStatus) return false;
      if (!q) return true;
      return [
        event.type,
        event.status,
        event.role,
        event.provider,
        event.model,
        event.phaseId,
        event.taskId,
        event.name,
        event.message,
        event.toolName,
        event.artifactPath,
        event.artifactUrl,
        event.rawPreview,
      ].some((value) => value?.toLowerCase().includes(q));
    });
  }, [run.events, eventQuery, eventType, eventStatus]);

  return (
    <div className="p-4 space-y-4">
      <section>
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <h3 className="text-sm font-semibold text-slate-100 truncate">{run.metadata.runId}</h3>
            <p className="text-xs text-slate-500 truncate">{run.metadata.workDir || 'No workdir recorded'}</p>
          </div>
          <StatusBadge status={run.metadata.status} />
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 mt-3">
          <Metric label="Provider" value={run.metadata.provider || '-'} icon={<Server className="w-3.5 h-3.5" />} />
          <Metric label="Role" value={run.metadata.role || '-'} icon={<Bot className="w-3.5 h-3.5" />} />
          <Metric label="Events" value={String(run.events.length)} icon={<Activity className="w-3.5 h-3.5" />} />
          <Metric label="Artifacts" value={String(run.artifacts.length)} icon={<FileText className="w-3.5 h-3.5" />} />
        </div>
      </section>

      {run.request && <RequestSummary request={run.request} />}

      {run.artifacts.length > 0 && (
        <section>
          <h4 className="text-xs font-semibold text-slate-300 mb-2">Artifacts</h4>
          <div className="divide-y divide-slate-800 border border-slate-800 rounded">
            {run.artifacts.map((artifact) => (
              <ArtifactRow key={`${artifact.kind}:${artifact.path || artifact.url}`} artifact={artifact} />
            ))}
          </div>
        </section>
      )}

      <section>
        <div className="flex flex-col gap-2 mb-2">
          <div className="flex items-center justify-between gap-2">
            <h4 className="text-xs font-semibold text-slate-300">Event Stream</h4>
            <span className="text-xs text-slate-500">{filteredEvents.length} / {run.events.length}</span>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-[1fr_160px_160px] gap-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
              <input
                type="text"
                value={eventQuery}
                onChange={(event) => setEventQuery(event.target.value)}
                placeholder="Search events..."
                className="w-full pl-8 pr-3 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
              />
            </div>
            <select
              value={eventType}
              onChange={(event) => setEventType(event.target.value)}
              className="px-2 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-200 focus:outline-none focus:border-cyan-500"
              aria-label="Filter event type"
            >
              <option value="all">All types</option>
              {eventTypes.map((type) => (
                <option key={type} value={type}>{type}</option>
              ))}
            </select>
            <select
              value={eventStatus}
              onChange={(event) => setEventStatus(event.target.value)}
              className="px-2 py-1.5 text-xs bg-slate-800 border border-slate-700 rounded text-slate-200 focus:outline-none focus:border-cyan-500"
              aria-label="Filter event status"
            >
              <option value="all">All statuses</option>
              {eventStatuses.map((status) => (
                <option key={status} value={status}>{status}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="divide-y divide-slate-800 border border-slate-800 rounded">
          {filteredEvents.length === 0 ? (
            <div className="px-3 py-6 text-xs text-center text-slate-500">No events match the current filters.</div>
          ) : filteredEvents.map((event, index) => (
            <RuntimeEventRow key={`${event.timestamp}:${event.type}:${index}`} event={event} />
          ))}
        </div>
      </section>
    </div>
  );
}

function RequestSummary({ request }: { request: RuntimeRequestSummary }) {
  const tools = request.toolNames?.join(', ') || '-';
  const mcpServers = request.mcpServerLabels?.join(', ') || '-';
  const metadataEntries = Object.entries(request.metadata || {}).filter(([key]) => !key.toLowerCase().includes('key'));

  return (
    <section>
      <h4 className="text-xs font-semibold text-slate-300 mb-2">Request</h4>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-2">
        <Metric label="Approval" value={request.approvalPolicy || '-'} icon={<CheckCircle2 className="w-3.5 h-3.5" />} />
        <Metric label="Reasoning" value={request.reasoningEffort || '-'} icon={<Bot className="w-3.5 h-3.5" />} />
        <Metric label="Messages" value={String(request.messageCount)} icon={<FileText className="w-3.5 h-3.5" />} />
        <Metric label="Schema" value={request.outputSchema || '-'} icon={<Braces className="w-3.5 h-3.5" />} />
      </div>
      <div className="mt-3 grid grid-cols-1 lg:grid-cols-2 gap-2">
        <DetailBlock label="Tools" value={tools} />
        <DetailBlock label="MCP" value={mcpServers} />
      </div>
      {(request.instructionsPreview || request.firstMessagePreview || metadataEntries.length > 0) && (
        <div className="mt-3 space-y-2">
          {request.instructionsPreview && <DetailBlock label="Instructions" value={request.instructionsPreview} multiline />}
          {request.firstMessagePreview && <DetailBlock label="First Message" value={request.firstMessagePreview} multiline />}
          {metadataEntries.length > 0 && (
            <DetailBlock
              label="Metadata"
              value={metadataEntries.map(([key, value]) => `${key}: ${value}`).join('\n')}
              multiline
            />
          )}
        </div>
      )}
    </section>
  );
}

function DetailBlock({ label, value, multiline = false }: { label: string; value: string; multiline?: boolean }) {
  return (
    <div className="border border-slate-800 rounded px-3 py-2 min-w-0">
      <div className="text-xs text-slate-500 mb-1">{label}</div>
      <div className={`text-xs text-slate-300 ${multiline ? 'whitespace-pre-wrap break-words max-h-48 overflow-y-auto' : 'truncate'}`}>
        {value || '-'}
      </div>
    </div>
  );
}

function ArtifactRow({ artifact }: { artifact: RuntimeArtifactSummary }) {
  const location = artifact.path || artifact.url || artifact.kind;
  return (
    <div className="px-3 py-2 text-xs">
      <div className="flex items-center gap-2">
        <FileText className="w-3.5 h-3.5 text-cyan-400" />
        <span className="font-medium text-slate-200 truncate">{location}</span>
        <span className="px-1.5 py-0.5 bg-slate-800 text-slate-400 rounded">{artifact.kind}</span>
      </div>
      <div className="mt-1 ml-5 text-slate-500 truncate">
        {artifact.message || artifact.eventType || 'Artifact recorded'}
      </div>
    </div>
  );
}

function RuntimeEventRow({ event }: { event: RuntimeEventSummary }) {
  const detail =
    event.message ||
    event.toolName ||
    event.artifactPath ||
    event.artifactUrl ||
    event.rawPreview ||
    event.name ||
    '';

  return (
    <div className="px-3 py-2 text-xs">
      <div className="flex items-center gap-2">
        <EventIcon event={event} />
        <span className="font-medium text-slate-200">{event.type}</span>
        {event.status && <span className="text-slate-500">{event.status}</span>}
        {event.timestamp && (
          <span className="ml-auto text-slate-500">{new Date(event.timestamp).toLocaleString()}</span>
        )}
      </div>
      {detail && <p className="mt-1 ml-5 text-slate-400 whitespace-pre-wrap break-words">{detail}</p>}
      {(event.phaseId || event.taskId || event.role) && (
        <div className="mt-1 ml-5 flex flex-wrap gap-2 text-slate-500">
          {event.role && <span>{event.role}</span>}
          {event.phaseId && <span>phase {event.phaseId}</span>}
          {event.taskId && <span>task {event.taskId}</span>}
        </div>
      )}
    </div>
  );
}

function EventIcon({ event }: { event: RuntimeEventSummary }) {
  if (event.toolName) return <Wrench className="w-3.5 h-3.5 text-amber-400" />;
  if (event.artifactKind) return <FileText className="w-3.5 h-3.5 text-cyan-400" />;
  if (event.status === 'failed' || event.toolError) return <AlertCircle className="w-3.5 h-3.5 text-red-400" />;
  if (event.status === 'completed' || event.status === 'succeeded') return <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />;
  return <Clock className="w-3.5 h-3.5 text-slate-500" />;
}

interface MemoryPaneProps {
  documents: MemoryDocumentSummary[];
  selectedDocument: MemoryDocumentDetail | null;
  onSelectDocument: (id: string) => void;
  isLoading: boolean;
}

function MemoryPane({ documents, selectedDocument, onSelectDocument, isLoading }: MemoryPaneProps) {
  if (documents.length === 0 && !isLoading) {
    return (
      <EmptyState
        icon={<Database className="w-12 h-12 opacity-30" />}
        title="No memory documents found"
        body="Memory documents are Markdown files under .boatman/memory that future sessions can load."
      />
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[minmax(300px,360px)_1fr] h-full min-h-0">
      <div className="border-b lg:border-b-0 lg:border-r border-slate-800 overflow-y-auto min-h-[220px] lg:min-h-0">
        <div className="divide-y divide-slate-800">
          {documents.map((doc) => (
            <MemoryListItem
              key={doc.id}
              document={doc}
              selected={selectedDocument?.id === doc.id}
              onClick={() => onSelectDocument(doc.id)}
            />
          ))}
        </div>
      </div>
      <div className="overflow-y-auto">
        {selectedDocument ? (
          <MemoryDetail document={selectedDocument} />
        ) : (
          <SelectPrompt label="Select a memory document to inspect its full context." />
        )}
      </div>
    </div>
  );
}

function MemoryListItem({
  document,
  selected,
  onClick,
}: {
  document: MemoryDocumentSummary;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-3 transition-colors ${
        selected ? 'bg-emerald-950/35' : 'hover:bg-slate-800/50'
      }`}
    >
      <div className="flex items-center gap-2">
        <ChevronRight className={`w-4 h-4 ${selected ? 'text-emerald-300' : 'text-slate-500'}`} />
        <span className="text-sm font-medium text-slate-200 truncate">{document.title || document.id}</span>
        {document.expired && <span className="px-1.5 py-0.5 text-xs bg-amber-900/40 text-amber-300 rounded">expired</span>}
      </div>
      <div className="mt-1 ml-6 flex flex-wrap gap-x-3 gap-y-1 text-xs text-slate-500">
        {document.scope && <span>{document.scope}</span>}
        {document.sourceRunId && <span>{document.sourceRunId}</span>}
        {document.updatedAt && <span>{formatRelative(document.updatedAt)}</span>}
      </div>
      {document.bodyPreview && (
        <p className="mt-1 ml-6 text-xs text-slate-400 line-clamp-2">{document.bodyPreview}</p>
      )}
    </button>
  );
}

function MemoryDetail({ document }: { document: MemoryDocumentDetail }) {
  return (
    <div className="p-4 space-y-4">
      <section>
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h3 className="text-sm font-semibold text-slate-100 truncate">{document.title || document.id}</h3>
            <p className="text-xs text-slate-500 truncate">{document.path}</p>
          </div>
          {document.scope && <span className="px-2 py-1 text-xs bg-emerald-900/40 text-emerald-300 rounded">{document.scope}</span>}
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-2 mt-3">
          <Metric label="Updated" value={formatDate(document.updatedAt)} icon={<Clock className="w-3.5 h-3.5" />} />
          <Metric label="Source Run" value={document.sourceRunId || '-'} icon={<Activity className="w-3.5 h-3.5" />} />
          <Metric label="Provenance" value={document.provenance || '-'} icon={<Braces className="w-3.5 h-3.5" />} />
          <Metric label="Expires" value={document.expiresAt ? formatDate(document.expiresAt) : '-'} icon={<AlertCircle className="w-3.5 h-3.5" />} />
        </div>
      </section>

      <section>
        <h4 className="text-xs font-semibold text-slate-300 mb-2">Context</h4>
        <pre className="text-xs text-slate-300 leading-relaxed whitespace-pre-wrap break-words border border-slate-800 rounded p-3 bg-slate-950">
          {document.body || 'No body content'}
        </pre>
      </section>
    </div>
  );
}

function Metric({ label, value, icon }: { label: string; value: string; icon: ReactNode }) {
  return (
    <div className="border border-slate-800 rounded px-3 py-2 min-w-0">
      <div className="flex items-center gap-1.5 text-xs text-slate-500">
        {icon}
        {label}
      </div>
      <div className="mt-1 text-xs font-medium text-slate-200 truncate">{value}</div>
    </div>
  );
}

function StatusBadge({ status }: { status?: string }) {
  const normalized = status || 'unknown';
  const color =
    normalized === 'failed' || normalized === 'error'
      ? 'bg-red-900/40 text-red-300'
      : normalized === 'succeeded' || normalized === 'completed'
        ? 'bg-emerald-900/40 text-emerald-300'
        : normalized === 'running' || normalized === 'started'
          ? 'bg-blue-900/40 text-blue-300'
          : 'bg-slate-800 text-slate-400';
  return <span className={`px-1.5 py-0.5 text-xs rounded ${color}`}>{normalized}</span>;
}

function SelectPrompt({ label }: { label: string }) {
  return (
    <div className="h-full flex items-center justify-center text-xs text-slate-500 p-8">
      {label}
    </div>
  );
}

function EmptyState({ icon, title, body }: { icon: ReactNode; title: string; body: string }) {
  return (
    <div className="flex flex-col items-center justify-center h-full text-slate-500 p-8">
      {icon}
      <p className="text-sm font-medium mt-3 mb-1">{title}</p>
      <p className="text-xs text-center max-w-md">{body}</p>
    </div>
  );
}

function formatRelative(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString();
}

function formatDate(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function uniqueValues(values: string[]) {
  return Array.from(new Set(values.filter(Boolean))).sort((a, b) => a.localeCompare(b));
}
