import { useState, useEffect } from 'react';
import { Flame, Play, Pause, Activity, AlertCircle, CheckCircle, Clock } from 'lucide-react';

interface FirefighterMonitorProps {
  sessionId: string;
  isActive: boolean;
  onToggle: (active: boolean) => void;
}

interface MonitorStatus {
  active: boolean;
  checkInterval: string;
  lastCheck: string;
  seenIssues: number;
}

export function FirefighterMonitor({ sessionId, isActive, onToggle }: FirefighterMonitorProps) {
  const [status, setStatus] = useState<MonitorStatus | null>(null);
  const [recentAlerts, setRecentAlerts] = useState<number>(0);

  // In a real implementation, this would fetch from the backend
  useEffect(() => {
    // Mock status for now
    if (isActive) {
      setStatus({
        active: true,
        checkInterval: '5m',
        lastCheck: new Date().toISOString(),
        seenIssues: 0,
      });
    } else {
      setStatus(null);
    }
  }, [isActive]);

  if (!isActive && !status) {
    return null;
  }

  return (
    <div className="border-b border-slate-700 bg-gradient-to-r from-red-900/10 to-orange-900/10">
      <div className="px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="relative">
            <Flame className="w-5 h-5 text-red-500" />
            {isActive && (
              <span className="absolute -top-1 -right-1 w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-slate-100">Firefighter Monitoring</span>
              {isActive ? (
                <span className="px-2 py-0.5 text-xs bg-green-500/20 text-green-400 rounded-full flex items-center gap-1">
                  <Activity className="w-3 h-3" />
                  Active
                </span>
              ) : (
                <span className="px-2 py-0.5 text-xs bg-slate-600 text-slate-400 rounded-full">
                  Paused
                </span>
              )}
            </div>
            {status && (
              <div className="flex items-center gap-3 mt-1 text-xs text-slate-400">
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  Check every {status.checkInterval}
                </span>
                {status.seenIssues === 0 ? (
                  <span className="flex items-center gap-1 text-green-400">
                    <CheckCircle className="w-3 h-3" />
                    All clear
                  </span>
                ) : (
                  <span className="flex items-center gap-1 text-amber-400">
                    <AlertCircle className="w-3 h-3" />
                    {status.seenIssues} issue{status.seenIssues !== 1 ? 's' : ''} tracked
                  </span>
                )}
                {recentAlerts > 0 && (
                  <span className="flex items-center gap-1 text-red-400">
                    {recentAlerts} new alert{recentAlerts !== 1 ? 's' : ''}
                  </span>
                )}
              </div>
            )}
          </div>
        </div>

        <button
          onClick={() => onToggle(!isActive)}
          className={`flex items-center gap-2 px-3 py-1.5 text-sm rounded-md transition-colors ${
            isActive
              ? 'bg-amber-500/10 text-amber-400 hover:bg-amber-500/20 border border-amber-500/30'
              : 'bg-green-500/10 text-green-400 hover:bg-green-500/20 border border-green-500/30'
          }`}
        >
          {isActive ? (
            <>
              <Pause className="w-4 h-4" />
              Pause
            </>
          ) : (
            <>
              <Play className="w-4 h-4" />
              Resume
            </>
          )}
        </button>
      </div>

      {/* Alert banner for new issues */}
      {recentAlerts > 0 && (
        <div className="px-4 py-2 bg-red-900/20 border-t border-red-500/30 flex items-center justify-between">
          <div className="flex items-center gap-2 text-sm text-red-400">
            <AlertCircle className="w-4 h-4" />
            <span>{recentAlerts} new issue{recentAlerts !== 1 ? 's' : ''} detected - Check messages for details</span>
          </div>
          <button
            onClick={() => setRecentAlerts(0)}
            className="text-xs text-red-400 hover:text-red-300 underline"
          >
            Dismiss
          </button>
        </div>
      )}
    </div>
  );
}
