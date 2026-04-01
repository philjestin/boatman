import { useState } from 'react';
import { X, Filter, AlertCircle } from 'lucide-react';
import type { TriageOptions } from '../../types';

interface TriageDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onStart: (opts: TriageOptions) => void;
  projectPath: string;
}

export function TriageDialog({ isOpen, onClose, onStart, projectPath }: TriageDialogProps) {
  const [teams, setTeams] = useState('');
  const [ticketIds, setTicketIds] = useState('');
  const [states, setStates] = useState('backlog,triage,unstarted');
  const [limit, setLimit] = useState(50);
  const [postComments, setPostComments] = useState(false);
  const [dryRun, setDryRun] = useState(true);
  const [isStarting, setIsStarting] = useState(false);
  const [useTicketIds, setUseTicketIds] = useState(false);
  const [generatePlans, setGeneratePlans] = useState(false);

  if (!isOpen) return null;

  const handleStart = async () => {
    setIsStarting(true);
    try {
      const opts: TriageOptions = {
        teams: useTicketIds ? [] : teams.split(',').map((t) => t.trim()).filter(Boolean),
        ticketIds: useTicketIds ? ticketIds.split(',').map((t) => t.trim()).filter(Boolean) : [],
        states: states.split(',').map((s) => s.trim()).filter(Boolean),
        limit,
        postComments,
        dryRun,
        outputDir: '',
        concurrency: 0,
        generatePlans,
        repoPath: generatePlans ? projectPath : '',
      };
      await onStart(opts);
      onClose();
    } catch (err) {
      console.error('Failed to start triage:', err);
    } finally {
      setIsStarting(false);
    }
  };

  const hasInput = useTicketIds ? ticketIds.trim().length > 0 : teams.trim().length > 0;

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 border border-slate-700">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-amber-500/20 rounded-lg flex items-center justify-center">
              <Filter className="w-6 h-6 text-amber-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-slate-100">Start Triage</h2>
              <p className="text-sm text-slate-400">Score and classify backlog tickets for AI execution</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-700 rounded-md">
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {/* Mode toggle */}
          <div className="flex items-center gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-700">
            <button
              onClick={() => setUseTicketIds(false)}
              className={`flex-1 px-4 py-2 text-sm rounded-md transition-colors ${
                !useTicketIds ? 'bg-amber-500 text-white' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              By Team
            </button>
            <button
              onClick={() => setUseTicketIds(true)}
              className={`flex-1 px-4 py-2 text-sm rounded-md transition-colors ${
                useTicketIds ? 'bg-amber-500 text-white' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              By Ticket IDs
            </button>
          </div>

          {/* Input */}
          {useTicketIds ? (
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Ticket IDs</label>
              <input
                type="text"
                value={ticketIds}
                onChange={(e) => setTicketIds(e.target.value)}
                placeholder="ENG-123, FE-456"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-amber-500"
                disabled={isStarting}
              />
            </div>
          ) : (
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Team Keys</label>
              <input
                type="text"
                value={teams}
                onChange={(e) => setTeams(e.target.value)}
                placeholder="ENG, FE, API"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-amber-500"
                disabled={isStarting}
              />
            </div>
          )}

          {/* States and Limit */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">States</label>
              <input
                type="text"
                value={states}
                onChange={(e) => setStates(e.target.value)}
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-amber-500 text-sm"
                disabled={isStarting}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Limit</label>
              <input
                type="number"
                value={limit}
                onChange={(e) => setLimit(parseInt(e.target.value) || 50)}
                min={1}
                max={200}
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 focus:outline-none focus:ring-2 focus:ring-amber-500 text-sm"
                disabled={isStarting}
              />
            </div>
          </div>

          {/* Toggles */}
          <div className="flex items-center gap-6">
            <label className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={dryRun}
                onChange={(e) => setDryRun(e.target.checked)}
                className="rounded border-slate-600 bg-slate-900 text-amber-500 focus:ring-amber-500"
              />
              Dry Run
            </label>
            <label className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={postComments}
                onChange={(e) => setPostComments(e.target.checked)}
                className="rounded border-slate-600 bg-slate-900 text-amber-500 focus:ring-amber-500"
              />
              Post Comments
            </label>
            <label className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={generatePlans}
                onChange={(e) => setGeneratePlans(e.target.checked)}
                className="rounded border-slate-600 bg-slate-900 text-amber-500 focus:ring-amber-500"
              />
              Generate Plans
            </label>
          </div>

          {/* Project Path */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Project Path</label>
            <div className="px-3 py-2 bg-slate-900/50 border border-slate-700 rounded-md text-slate-400 text-sm font-mono">
              {projectPath || 'No project selected'}
            </div>
          </div>

          {!projectPath && (
            <div className="flex items-start gap-3 p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
              <AlertCircle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-yellow-200">
                <p className="font-medium mb-1">No project selected</p>
                <p className="text-yellow-300/80">Please select a project before starting triage.</p>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-slate-700">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
            disabled={isStarting}
          >
            Cancel
          </button>
          <button
            onClick={handleStart}
            disabled={!hasInput || !projectPath || isStarting}
            className="flex items-center gap-2 px-4 py-2 text-sm bg-amber-500 text-white rounded-md hover:bg-amber-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Filter className="w-4 h-4" />
            <span>{isStarting ? 'Starting...' : 'Start Triage'}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
