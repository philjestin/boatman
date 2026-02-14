import { useState } from 'react';
import { X, Server, Plus } from 'lucide-react';
import type { MCPServer } from '../../types';

interface MCPServerDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onAdd: (server: MCPServer) => void;
  presets: MCPServer[];
}

export function MCPServerDialog({ isOpen, onClose, onAdd, presets }: MCPServerDialogProps) {
  const [mode, setMode] = useState<'preset' | 'custom'>('preset');
  const [selectedPreset, setSelectedPreset] = useState<MCPServer | null>(null);
  const [customServer, setCustomServer] = useState<MCPServer>({
    name: '',
    description: '',
    command: 'npx',
    args: [],
    env: {},
    enabled: true,
  });

  if (!isOpen) return null;

  const handleAddPreset = () => {
    if (selectedPreset) {
      onAdd(selectedPreset);
      onClose();
    }
  };

  const handleAddCustom = () => {
    if (customServer.name && customServer.command) {
      onAdd(customServer);
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 no-drag">
      <div className="bg-slate-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 border border-slate-700">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700">
          <div className="flex items-center gap-2">
            <Server className="w-5 h-5 text-blue-500" />
            <h2 className="text-lg font-semibold text-slate-100">Add MCP Server</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-md hover:bg-slate-700 transition-colors"
            aria-label="Close"
          >
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        {/* Mode Selection */}
        <div className="p-4 border-b border-slate-700">
          <div className="flex gap-2">
            <button
              onClick={() => setMode('preset')}
              className={`flex-1 px-4 py-2 rounded-md text-sm transition-colors ${
                mode === 'preset'
                  ? 'bg-blue-500 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
              }`}
            >
              From Preset
            </button>
            <button
              onClick={() => setMode('custom')}
              className={`flex-1 px-4 py-2 rounded-md text-sm transition-colors ${
                mode === 'custom'
                  ? 'bg-blue-500 text-white'
                  : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
              }`}
            >
              Custom Server
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 max-h-96 overflow-y-auto">
          {mode === 'preset' ? (
            <div className="space-y-3">
              <p className="text-sm text-slate-400 mb-4">
                Select a pre-configured MCP server to add
              </p>
              {presets.map((preset) => (
                <label
                  key={preset.name}
                  className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                    selectedPreset?.name === preset.name
                      ? 'border-blue-500 bg-blue-500/10'
                      : 'border-slate-700 hover:border-slate-600'
                  }`}
                >
                  <input
                    type="radio"
                    name="preset"
                    checked={selectedPreset?.name === preset.name}
                    onChange={() => setSelectedPreset(preset)}
                    className="mt-1"
                  />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Server className="w-4 h-4 text-slate-400" />
                      <span className="text-sm font-medium text-slate-100">
                        {preset.name}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 mt-1">{preset.description}</p>
                    {preset.env && Object.keys(preset.env).length > 0 && (
                      <div className="mt-2 text-xs text-amber-400">
                        Requires: {Object.keys(preset.env).join(', ')}
                      </div>
                    )}
                  </div>
                </label>
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-200 mb-2">
                  Server Name *
                </label>
                <input
                  type="text"
                  value={customServer.name}
                  onChange={(e) => setCustomServer({ ...customServer, name: e.target.value })}
                  placeholder="my-mcp-server"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-200 mb-2">
                  Description
                </label>
                <input
                  type="text"
                  value={customServer.description}
                  onChange={(e) => setCustomServer({ ...customServer, description: e.target.value })}
                  placeholder="What this server does"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-200 mb-2">
                  Command *
                </label>
                <input
                  type="text"
                  value={customServer.command}
                  onChange={(e) => setCustomServer({ ...customServer, command: e.target.value })}
                  placeholder="npx"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-200 mb-2">
                  Arguments (comma-separated)
                </label>
                <input
                  type="text"
                  value={customServer.args?.join(', ') || ''}
                  onChange={(e) =>
                    setCustomServer({
                      ...customServer,
                      args: e.target.value.split(',').map((s) => s.trim()).filter(Boolean),
                    })
                  }
                  placeholder="-y, @my/mcp-server"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          )}
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
            onClick={mode === 'preset' ? handleAddPreset : handleAddCustom}
            disabled={mode === 'preset' ? !selectedPreset : !customServer.name || !customServer.command}
            className="flex items-center gap-2 px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus className="w-4 h-4" />
            Add Server
          </button>
        </div>
      </div>
    </div>
  );
}
