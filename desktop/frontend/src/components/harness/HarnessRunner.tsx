import { useState, useEffect } from 'react';
import { ArrowLeft, Play, FolderOpen, Plus, Trash2 } from 'lucide-react';
import type { HarnessInfo, RunRequest, HarnessRunState } from '../../types';
import { HarnessList } from './HarnessList';
import { HarnessRunOutput } from './HarnessRunOutput';
import { SelectFolder } from '../../../wailsjs/go/main/App';

interface HarnessRunnerProps {
  harnesses: HarnessInfo[];
  isLoading: boolean;
  runState: HarnessRunState;
  onLoadHarnesses: () => void;
  onRun: (req: RunRequest) => void;
  onStop: () => void;
  onResetRun: () => void;
}

type RunnerView = 'list' | 'config' | 'output';

export function HarnessRunner({
  harnesses,
  isLoading,
  runState,
  onLoadHarnesses,
  onRun,
  onStop,
  onResetRun,
}: HarnessRunnerProps) {
  const [view, setView] = useState<RunnerView>('list');
  const [selectedHarness, setSelectedHarness] = useState<HarnessInfo | null>(null);
  const [workDir, setWorkDir] = useState('');
  const [taskTitle, setTaskTitle] = useState('');
  const [taskDescription, setTaskDescription] = useState('');
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>([]);

  // Load harnesses on mount
  useEffect(() => {
    onLoadHarnesses();
  }, [onLoadHarnesses]);

  // Switch to output view when run starts
  useEffect(() => {
    if (runState.status === 'running') {
      setView('output');
    }
  }, [runState.status]);

  const handleSelect = (harness: HarnessInfo) => {
    setSelectedHarness(harness);
    setView('config');
  };

  const handleBrowseWorkDir = async () => {
    try {
      const folder = await SelectFolder();
      if (folder) setWorkDir(folder);
    } catch (err) {
      console.error('Failed to select folder:', err);
    }
  };

  const handleAddEnvVar = () => {
    setEnvVars((prev) => [...prev, { key: '', value: '' }]);
  };

  const handleRemoveEnvVar = (index: number) => {
    setEnvVars((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpdateEnvVar = (index: number, field: 'key' | 'value', val: string) => {
    setEnvVars((prev) =>
      prev.map((ev, i) => (i === index ? { ...ev, [field]: val } : ev))
    );
  };

  const handleRun = () => {
    if (!selectedHarness) return;

    const envMap: Record<string, string> = {};
    envVars.forEach((ev) => {
      if (ev.key.trim()) {
        envMap[ev.key.trim()] = ev.value;
      }
    });

    onRun({
      harnessPath: selectedHarness.path,
      workDir,
      taskTitle,
      taskDescription,
      envVars: envMap,
    });
  };

  const handleBack = () => {
    if (view === 'output') {
      onResetRun();
      setView('config');
    } else if (view === 'config') {
      setSelectedHarness(null);
      setView('list');
    }
  };

  if (view === 'output') {
    return (
      <div className="flex flex-col h-full">
        <div className="px-4 py-2 border-b border-slate-700 bg-slate-800/50">
          <button
            onClick={handleBack}
            disabled={runState.status === 'running'}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200 disabled:opacity-50 transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            Back to Config
          </button>
        </div>
        <div className="flex-1 overflow-hidden">
          <HarnessRunOutput runState={runState} onStop={onStop} />
        </div>
      </div>
    );
  }

  if (view === 'config' && selectedHarness) {
    return (
      <div className="flex flex-col h-full">
        <div className="px-4 py-2 border-b border-slate-700 bg-slate-800/50">
          <button
            onClick={handleBack}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200 transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            Back to List
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-5">
          {/* Selected harness */}
          <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-3">
            <div className="text-sm font-medium text-slate-200">{selectedHarness.name}</div>
            <div className="text-xs font-mono text-slate-500 mt-1">{selectedHarness.path}</div>
          </div>

          {/* Work directory */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Work Directory</label>
            <div className="flex gap-2">
              <input
                type="text"
                value={workDir}
                onChange={(e) => setWorkDir(e.target.value)}
                placeholder="Project directory for the harness to work in"
                className="flex-1 px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
              />
              <button
                onClick={handleBrowseWorkDir}
                className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-slate-300 rounded-md transition-colors"
              >
                <FolderOpen className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Task title */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Task Title</label>
            <input
              type="text"
              value={taskTitle}
              onChange={(e) => setTaskTitle(e.target.value)}
              placeholder="e.g., Add user authentication"
              className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          {/* Task description */}
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Task Description</label>
            <textarea
              value={taskDescription}
              onChange={(e) => setTaskDescription(e.target.value)}
              placeholder="Describe the task for the harness agent..."
              className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              rows={4}
            />
          </div>

          {/* Environment variables */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-slate-300">Environment Variables</label>
              <button
                onClick={handleAddEnvVar}
                className="flex items-center gap-1 px-2 py-1 text-xs text-slate-400 hover:text-slate-200 hover:bg-slate-700 rounded transition-colors"
              >
                <Plus className="w-3 h-3" />
                Add
              </button>
            </div>
            {envVars.length > 0 && (
              <div className="space-y-2">
                {envVars.map((ev, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <input
                      type="text"
                      value={ev.key}
                      onChange={(e) => handleUpdateEnvVar(i, 'key', e.target.value)}
                      placeholder="KEY"
                      className="w-1/3 px-2 py-1.5 bg-slate-900 border border-slate-700 rounded text-sm font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                    <span className="text-slate-600">=</span>
                    <input
                      type="text"
                      value={ev.value}
                      onChange={(e) => handleUpdateEnvVar(i, 'value', e.target.value)}
                      placeholder="value"
                      className="flex-1 px-2 py-1.5 bg-slate-900 border border-slate-700 rounded text-sm font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                    <button
                      onClick={() => handleRemoveEnvVar(i)}
                      className="p-1.5 text-slate-500 hover:text-red-400 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Run button */}
          <button
            onClick={handleRun}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 text-sm bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors"
          >
            <Play className="w-4 h-4" />
            Run Harness
          </button>
        </div>
      </div>
    );
  }

  // List view
  return (
    <div className="p-6 overflow-y-auto h-full">
      <HarnessList
        harnesses={harnesses}
        onSelect={handleSelect}
        onRefresh={onLoadHarnesses}
        isLoading={isLoading}
      />
    </div>
  );
}
