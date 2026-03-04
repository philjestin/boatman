import { useState } from 'react';
import { X, Server, Plus, Trash2 } from 'lucide-react';
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
  const [presetEnvValues, setPresetEnvValues] = useState<Record<string, string>>({});
  const [customServer, setCustomServer] = useState<MCPServer>({
    name: '',
    description: '',
    command: 'npx',
    args: [],
    env: {},
    enabled: true,
  });
  const [customEnvEntries, setCustomEnvEntries] = useState<{ key: string; value: string }[]>([]);

  if (!isOpen) return null;

  const handleSelectPreset = (preset: MCPServer) => {
    setSelectedPreset(preset);
    // Initialize env values with empty strings for each required key
    const envDefaults: Record<string, string> = {};
    if (preset.env) {
      for (const key of Object.keys(preset.env)) {
        envDefaults[key] = preset.env[key] || '';
      }
    }
    setPresetEnvValues(envDefaults);
  };

  const handleAddPreset = () => {
    if (selectedPreset) {
      onAdd({
        ...selectedPreset,
        env: { ...presetEnvValues },
        enabled: true,
      });
      setSelectedPreset(null);
      setPresetEnvValues({});
      onClose();
    }
  };

  const handleAddCustom = () => {
    if (customServer.name && customServer.command) {
      const env: Record<string, string> = {};
      for (const entry of customEnvEntries) {
        if (entry.key.trim()) {
          env[entry.key.trim()] = entry.value;
        }
      }
      onAdd({ ...customServer, env });
      setCustomServer({ name: '', description: '', command: 'npx', args: [], env: {}, enabled: true });
      setCustomEnvEntries([]);
      onClose();
    }
  };

  const addCustomEnvEntry = () => {
    setCustomEnvEntries([...customEnvEntries, { key: '', value: '' }]);
  };

  const updateCustomEnvEntry = (index: number, field: 'key' | 'value', val: string) => {
    const updated = [...customEnvEntries];
    updated[index] = { ...updated[index], [field]: val };
    setCustomEnvEntries(updated);
  };

  const removeCustomEnvEntry = (index: number) => {
    setCustomEnvEntries(customEnvEntries.filter((_, i) => i !== index));
  };

  const presetEnvKeys = selectedPreset?.env ? Object.keys(selectedPreset.env) : [];
  const hasEmptyRequiredEnv = presetEnvKeys.some((key) => !presetEnvValues[key]?.trim());

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 no-drag">
      <div className="bg-slate-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 border border-slate-700 max-h-[85vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700 flex-shrink-0">
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
        <div className="p-4 border-b border-slate-700 flex-shrink-0">
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
        <div className="p-6 overflow-y-auto flex-1">
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
                    onChange={() => handleSelectPreset(preset)}
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
                  </div>
                </label>
              ))}

              {/* Env var inputs for selected preset */}
              {selectedPreset && presetEnvKeys.length > 0 && (
                <div className="mt-4 p-4 bg-slate-900/50 border border-slate-700 rounded-lg space-y-3">
                  <h4 className="text-sm font-medium text-slate-200">
                    Configuration for {selectedPreset.name}
                  </h4>
                  <p className="text-xs text-slate-400">
                    Enter the required environment variables to connect this server.
                  </p>
                  {presetEnvKeys.map((key) => (
                    <div key={key}>
                      <label className="block text-xs font-medium text-slate-300 mb-1">
                        {key}
                      </label>
                      <input
                        type={key.toLowerCase().includes('token') || key.toLowerCase().includes('secret') || key.toLowerCase().includes('key') ? 'password' : 'text'}
                        value={presetEnvValues[key] || ''}
                        onChange={(e) =>
                          setPresetEnvValues({ ...presetEnvValues, [key]: e.target.value })
                        }
                        placeholder={getPlaceholder(key)}
                        className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md text-slate-100 placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  ))}
                </div>
              )}
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

              {/* Environment Variables */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-slate-200">
                    Environment Variables
                  </label>
                  <button
                    onClick={addCustomEnvEntry}
                    className="flex items-center gap-1 px-2 py-1 text-xs text-blue-400 hover:text-blue-300 hover:bg-slate-700 rounded transition-colors"
                  >
                    <Plus className="w-3 h-3" />
                    Add Variable
                  </button>
                </div>
                {customEnvEntries.length === 0 && (
                  <p className="text-xs text-slate-500">
                    No environment variables configured. Click "Add Variable" to add API keys, tokens, etc.
                  </p>
                )}
                <div className="space-y-2">
                  {customEnvEntries.map((entry, index) => (
                    <div key={index} className="flex items-center gap-2">
                      <input
                        type="text"
                        value={entry.key}
                        onChange={(e) => updateCustomEnvEntry(index, 'key', e.target.value)}
                        placeholder="VARIABLE_NAME"
                        className="w-2/5 px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      <input
                        type={entry.key.toLowerCase().includes('token') || entry.key.toLowerCase().includes('secret') || entry.key.toLowerCase().includes('key') ? 'password' : 'text'}
                        value={entry.value}
                        onChange={(e) => updateCustomEnvEntry(index, 'value', e.target.value)}
                        placeholder="value"
                        className="flex-1 px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      <button
                        onClick={() => removeCustomEnvEntry(index)}
                        className="p-2 text-slate-400 hover:text-red-400 hover:bg-slate-700 rounded transition-colors"
                        aria-label="Remove variable"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-4 border-t border-slate-700 flex-shrink-0">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={mode === 'preset' ? handleAddPreset : handleAddCustom}
            disabled={
              mode === 'preset'
                ? !selectedPreset || hasEmptyRequiredEnv
                : !customServer.name || !customServer.command
            }
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

/** Returns a helpful placeholder for known env var names */
function getPlaceholder(key: string): string {
  const lower = key.toLowerCase();
  if (lower.includes('slack_bot_token')) return 'xoxb-...';
  if (lower.includes('slack_team_id')) return 'T0123456789';
  if (lower.includes('linear_api_key')) return 'lin_api_...';
  if (lower.includes('api_key') || lower.includes('apikey')) return 'your-api-key';
  if (lower.includes('app_key')) return 'your-app-key';
  if (lower.includes('token')) return 'your-token';
  if (lower.includes('secret')) return 'your-secret';
  if (lower.includes('site')) return 'e.g., datadoghq.com';
  return 'value';
}
