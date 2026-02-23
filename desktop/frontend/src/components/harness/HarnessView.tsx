import { useState } from 'react';
import { Wrench, Play } from 'lucide-react';
import { useHarness } from '../../hooks/useHarness';
import { HarnessBuilder } from './HarnessBuilder';
import { HarnessRunner } from './HarnessRunner';

type HarnessMode = 'build' | 'run';

export function HarnessView() {
  const [mode, setMode] = useState<HarnessMode>('build');

  const {
    harnesses,
    isLoading,
    runState,
    loadHarnesses,
    scaffoldHarness,
    runHarness,
    stopHarness,
    resetRunState,
    selectFolder,
  } = useHarness();

  return (
    <div className="flex flex-col h-full">
      {/* Mode Toggle */}
      <div className="px-4 py-3 border-b border-slate-700 bg-slate-800">
        <div className="flex items-center gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-700 max-w-xs">
          <button
            onClick={() => setMode('build')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
              mode === 'build'
                ? 'bg-blue-500 text-white'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Wrench className="w-4 h-4" />
            Build
          </button>
          <button
            onClick={() => setMode('run')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
              mode === 'run'
                ? 'bg-blue-500 text-white'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Play className="w-4 h-4" />
            Run
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {mode === 'build' ? (
          <HarnessBuilder
            onScaffold={scaffoldHarness}
            onSelectFolder={selectFolder}
            onSwitchToRun={() => {
              setMode('run');
              loadHarnesses();
            }}
          />
        ) : (
          <HarnessRunner
            harnesses={harnesses}
            isLoading={isLoading}
            runState={runState}
            onLoadHarnesses={loadHarnesses}
            onRun={runHarness}
            onStop={stopHarness}
            onResetRun={resetRunState}
          />
        )}
      </div>
    </div>
  );
}
