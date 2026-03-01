import { useState } from 'react';
import { X, PlayCircle, AlertCircle, Ticket, MessageSquare, Settings, ChevronDown, ChevronUp } from 'lucide-react';

type ExecutionMode = 'ticket' | 'prompt';

export interface BoatmanModeConfig {
  maxIterations: number;
  baseBranch: string;
  autoPR: boolean;
  reviewSkill: string;
  timeout: number;
}

interface BoatmanModeDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onStart: (input: string, mode: ExecutionMode, config: BoatmanModeConfig) => void;
  projectPath: string;
  defaultConfig?: Partial<BoatmanModeConfig>;
}

export function BoatmanModeDialog({ isOpen, onClose, onStart, projectPath, defaultConfig }: BoatmanModeDialogProps) {
  const [mode, setMode] = useState<ExecutionMode>('prompt');
  const [input, setInput] = useState('');
  const [isStarting, setIsStarting] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Configuration state with defaults
  const [config, setConfig] = useState<BoatmanModeConfig>({
    maxIterations: defaultConfig?.maxIterations ?? 3,
    baseBranch: defaultConfig?.baseBranch ?? 'main',
    autoPR: defaultConfig?.autoPR ?? true,
    reviewSkill: defaultConfig?.reviewSkill ?? 'peer-review',
    timeout: defaultConfig?.timeout ?? 60,
  });

  if (!isOpen) return null;

  const handleStart = async () => {
    if (!input.trim()) return;

    setIsStarting(true);
    try {
      await onStart(input.trim(), mode, config);
      onClose();
      setInput('');
    } catch (err) {
      console.error('Failed to start boatmanmode:', err);
    } finally {
      setIsStarting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey && input.trim() && !isStarting && mode === 'ticket') {
      handleStart();
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-slate-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 border border-slate-700">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-purple-500/20 rounded-lg flex items-center justify-center">
              <PlayCircle className="w-6 h-6 text-purple-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-slate-100">Start Boatman Mode</h2>
              <p className="text-sm text-slate-400">Automated ticket execution with multi-agent orchestration</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-slate-700 rounded-md transition-colors"
          >
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Mode Tabs */}
          <div className="flex items-center gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-700">
            <button
              onClick={() => setMode('prompt')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
                mode === 'prompt'
                  ? 'bg-purple-500 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <MessageSquare className="w-4 h-4" />
              <span>Custom Prompt</span>
            </button>
            <button
              onClick={() => setMode('ticket')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
                mode === 'ticket'
                  ? 'bg-purple-500 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Ticket className="w-4 h-4" />
              <span>Linear Ticket</span>
            </button>
          </div>

          {/* What is Boatman Mode */}
          <div className="bg-slate-900/50 rounded-lg p-4 border border-slate-700">
            <h3 className="text-sm font-medium text-slate-200 mb-2">What is Boatman Mode?</h3>
            <ul className="text-sm text-slate-400 space-y-1.5">
              <li className="flex items-start gap-2">
                <span className="text-purple-400 mt-0.5">•</span>
                <span><strong className="text-slate-300">Automated workflow:</strong> Plan → Implement → Test → Peer Review → PR</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-400 mt-0.5">•</span>
                <span><strong className="text-slate-300">Multi-agent orchestration:</strong> Specialized agents handle different phases</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-400 mt-0.5">•</span>
                <span><strong className="text-slate-300">Git worktree isolation:</strong> Executes in separate worktree</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-400 mt-0.5">•</span>
                <span><strong className="text-slate-300">Iterative refinement:</strong> Refactors based on peer review feedback</span>
              </li>
            </ul>
          </div>

          {/* Input Field */}
          {mode === 'prompt' ? (
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Task Prompt
              </label>
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="e.g., Add health check endpoint to the API&#10;Refactor authentication middleware&#10;Fix bug in user registration flow"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none"
                rows={4}
                autoFocus
                disabled={isStarting}
              />
              <p className="text-xs text-slate-500 mt-1">
                Describe what you want the agent to build, fix, or refactor
              </p>
            </div>
          ) : (
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Linear Ticket ID
              </label>
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="e.g., TICKET-123, ENG-456"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                autoFocus
                disabled={isStarting}
              />
              <p className="text-xs text-slate-500 mt-1">
                Enter the Linear ticket ID to execute
              </p>
            </div>
          )}

          {/* Project Path */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              Project Path
            </label>
            <div className="px-3 py-2 bg-slate-900/50 border border-slate-700 rounded-md text-slate-400 text-sm font-mono">
              {projectPath || 'No project selected'}
            </div>
          </div>

          {/* Advanced Configuration */}
          <div className="border border-slate-700 rounded-lg">
            <button
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="w-full flex items-center justify-between p-3 text-sm text-slate-300 hover:bg-slate-900/50 rounded-lg transition-colors"
            >
              <div className="flex items-center gap-2">
                <Settings className="w-4 h-4" />
                <span>Advanced Configuration</span>
              </div>
              {showAdvanced ? (
                <ChevronUp className="w-4 h-4" />
              ) : (
                <ChevronDown className="w-4 h-4" />
              )}
            </button>

            {showAdvanced && (
              <div className="p-4 space-y-4 border-t border-slate-700">
                {/* Max Iterations */}
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    Max Review Iterations
                  </label>
                  <input
                    type="number"
                    min="1"
                    max="10"
                    value={config.maxIterations}
                    onChange={(e) => setConfig({ ...config, maxIterations: parseInt(e.target.value) || 3 })}
                    className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    disabled={isStarting}
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    Maximum number of review/refactor iterations (default: 3)
                  </p>
                </div>

                {/* Base Branch */}
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    Base Branch
                  </label>
                  <input
                    type="text"
                    value={config.baseBranch}
                    onChange={(e) => setConfig({ ...config, baseBranch: e.target.value })}
                    placeholder="main"
                    className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    disabled={isStarting}
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    Base branch for git worktree (default: main)
                  </p>
                </div>

                {/* Review Skill */}
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    Review Skill
                  </label>
                  <input
                    type="text"
                    value={config.reviewSkill}
                    onChange={(e) => setConfig({ ...config, reviewSkill: e.target.value })}
                    placeholder="peer-review"
                    className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    disabled={isStarting}
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    Claude skill/agent to use for code review (default: peer-review)
                  </p>
                </div>

                {/* Timeout */}
                <div>
                  <label className="block text-sm font-medium text-slate-300 mb-2">
                    Agent Timeout (minutes)
                  </label>
                  <input
                    type="number"
                    min="5"
                    max="180"
                    value={config.timeout}
                    onChange={(e) => setConfig({ ...config, timeout: parseInt(e.target.value) || 60 })}
                    className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    disabled={isStarting}
                  />
                  <p className="text-xs text-slate-500 mt-1">
                    Timeout in minutes for each Claude agent (default: 60)
                  </p>
                </div>

                {/* Auto PR */}
                <label className="flex items-center justify-between p-3 rounded-lg border border-slate-700 cursor-pointer hover:bg-slate-900/50 transition-colors">
                  <div>
                    <p className="text-sm text-slate-100">Auto-create Pull Request</p>
                    <p className="text-xs text-slate-400">
                      Automatically create PR when execution succeeds
                    </p>
                  </div>
                  <input
                    type="checkbox"
                    checked={config.autoPR}
                    onChange={(e) => setConfig({ ...config, autoPR: e.target.checked })}
                    className="w-4 h-4 rounded"
                    disabled={isStarting}
                  />
                </label>
              </div>
            )}
          </div>

          {/* Warning */}
          {!projectPath && (
            <div className="flex items-start gap-3 p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
              <AlertCircle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-yellow-200">
                <p className="font-medium mb-1">No project selected</p>
                <p className="text-yellow-300/80">
                  Please select a project before starting Boatman Mode.
                </p>
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
            disabled={!input.trim() || !projectPath || isStarting}
            className="flex items-center gap-2 px-4 py-2 text-sm bg-purple-500 text-white rounded-md hover:bg-purple-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <PlayCircle className="w-4 h-4" />
            <span>{isStarting ? 'Starting...' : 'Start Execution'}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
