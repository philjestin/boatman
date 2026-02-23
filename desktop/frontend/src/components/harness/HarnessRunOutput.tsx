import { useEffect, useRef } from 'react';
import { Square, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import type { HarnessRunState } from '../../types';

interface HarnessRunOutputProps {
  runState: HarnessRunState;
  onStop: () => void;
}

export function HarnessRunOutput({ runState, onStop }: HarnessRunOutputProps) {
  const outputRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [runState.output]);

  return (
    <div className="flex flex-col h-full">
      {/* Status bar */}
      <div className="flex items-center justify-between px-4 py-2 bg-slate-800 border-b border-slate-700">
        <div className="flex items-center gap-2">
          {runState.status === 'running' && (
            <>
              <Loader2 className="w-4 h-4 text-blue-400 animate-spin" />
              <span className="text-sm text-blue-400">Running</span>
            </>
          )}
          {runState.status === 'completed' && (
            <>
              <CheckCircle className="w-4 h-4 text-green-400" />
              <span className="text-sm text-green-400">Completed</span>
            </>
          )}
          {runState.status === 'error' && (
            <>
              <AlertCircle className="w-4 h-4 text-red-400" />
              <span className="text-sm text-red-400">Error</span>
            </>
          )}
        </div>

        {runState.status === 'running' && (
          <button
            onClick={onStop}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-red-500/20 text-red-400 hover:bg-red-500/30 rounded-md transition-colors"
          >
            <Square className="w-3 h-3" />
            Stop
          </button>
        )}
      </div>

      {/* Output */}
      <div
        ref={outputRef}
        className="flex-1 overflow-y-auto p-4 bg-slate-950 font-mono text-xs leading-relaxed"
      >
        {runState.output.length === 0 && runState.status === 'running' && (
          <p className="text-slate-600">Waiting for output...</p>
        )}
        {runState.output.map((line, i) => (
          <div
            key={i}
            className={`whitespace-pre-wrap ${
              line.toLowerCase().includes('error') || line.toLowerCase().includes('fatal')
                ? 'text-red-400'
                : line.startsWith('Status:') || line.startsWith('Step:') || line.startsWith('===')
                ? 'text-blue-400 font-semibold'
                : 'text-slate-300'
            }`}
          >
            {line}
          </div>
        ))}
        {runState.error && (
          <div className="mt-2 text-red-400">
            Error: {runState.error}
          </div>
        )}
      </div>
    </div>
  );
}
