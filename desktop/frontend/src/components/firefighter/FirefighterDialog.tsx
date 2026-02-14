import { useState } from 'react';
import { X, Flame, AlertTriangle } from 'lucide-react';

interface FirefighterDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onStart: (scope: string, enableMonitoring: boolean) => void;
  projectPath: string;
}

export function FirefighterDialog({ isOpen, onClose, onStart, projectPath }: FirefighterDialogProps) {
  const [scope, setScope] = useState('');
  const [enableMonitoring, setEnableMonitoring] = useState(true);

  if (!isOpen) return null;

  const handleStart = () => {
    onStart(scope, enableMonitoring);
    setScope('');
    setEnableMonitoring(true);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 no-drag">
      <div className="bg-slate-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 border border-slate-700">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700">
          <div className="flex items-center gap-2">
            <Flame className="w-5 h-5 text-red-500" />
            <h2 className="text-lg font-semibold text-slate-100">Start Firefighter Investigation</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-md hover:bg-slate-700 transition-colors"
            aria-label="Close"
          >
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          <div className="bg-slate-900/50 border border-slate-700 rounded-lg p-4">
            <h3 className="text-sm font-medium text-slate-200 mb-2">What is Firefighter Mode?</h3>
            <p className="text-sm text-slate-400 mb-3">
              Firefighter mode creates a specialized agent that actively monitors production and responds to incidents.
            </p>
            <ul className="text-sm text-slate-400 space-y-1 ml-4 list-disc">
              <li>Continuously monitors Bugsnag for new errors</li>
              <li>Watches Datadog for alerts and anomalies</li>
              <li>Automatically investigates high-severity issues</li>
              <li>Creates fixes in isolated git worktrees</li>
              <li>Runs tests and generates draft PRs</li>
              <li>Alerts you to new issues in real-time</li>
            </ul>
          </div>

          <div className="bg-blue-900/20 border border-blue-700/50 rounded-lg p-4">
            <label className="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={enableMonitoring}
                onChange={(e) => setEnableMonitoring(e.target.checked)}
                className="w-4 h-4 mt-1 rounded"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-blue-300">Enable Active Monitoring</span>
                  <span className="px-1.5 py-0.5 text-xs bg-blue-500/20 text-blue-400 rounded">Recommended</span>
                </div>
                <p className="text-xs text-blue-200/70 mt-1">
                  The agent will proactively check for new issues every 5 minutes and alert you immediately.
                  You can start/stop monitoring anytime from the session controls.
                </p>
              </div>
            </label>
          </div>

          {/* Warning if API keys not configured */}
          <div className="bg-amber-900/20 border border-amber-700/50 rounded-lg p-4 flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="text-sm font-medium text-amber-500 mb-1">API Configuration Required</h4>
              <p className="text-sm text-amber-200/80">
                Make sure you've configured Datadog and Bugsnag API keys in Settings &gt; Firefighter
                before starting. The agent won't be able to fetch data without valid credentials.
              </p>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-200 mb-2">
              Investigation Scope (Optional)
            </label>
            <input
              type="text"
              value={scope}
              onChange={(e) => setScope(e.target.value)}
              placeholder="e.g., payment-service, user-auth, checkout-flow"
              className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-transparent"
            />
            <p className="mt-1 text-xs text-slate-500">
              Optionally specify a service, component, or team to focus the investigation on.
            </p>
          </div>

          <div className="text-xs text-slate-500 bg-slate-900/50 rounded-md p-3 border border-slate-700">
            <span className="font-medium text-slate-400">Project:</span> {projectPath}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-4 border-t border-slate-700">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleStart}
            className="flex items-center gap-2 px-4 py-2 text-sm bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors"
          >
            <Flame className="w-4 h-4" />
            Start Investigation
          </button>
        </div>
      </div>
    </div>
  );
}
