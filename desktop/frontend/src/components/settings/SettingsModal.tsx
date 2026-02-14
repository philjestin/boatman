import { useState, useEffect } from 'react';
import { X, Moon, Sun, Bell, BellOff, Shield, Zap, Bot, Server, Key, Eye, EyeOff, Database, Trash2, Plus, Flame } from 'lucide-react';
import type { UserPreferences, ApprovalMode, Theme, MCPServer } from '../../types';
import { CleanupOldSessions, GetSessionStats, GetMCPServers, GetMCPPresets, AddMCPServer, RemoveMCPServer, UpdateMCPServer } from '../../../wailsjs/go/main/App';
import { MCPServerDialog } from './MCPServerDialog';
import { GCloudAuthSection } from './GCloudAuthSection';
import { FirefighterSettings } from './FirefighterSettings';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  preferences: UserPreferences;
  onSave: (preferences: UserPreferences) => void;
}

type SettingsTab = 'general' | 'approval' | 'memory' | 'mcp' | 'firefighter' | 'about';

export function SettingsModal({ isOpen, onClose, preferences, onSave }: SettingsModalProps) {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general');
  const [localPrefs, setLocalPrefs] = useState<UserPreferences>(preferences);

  if (!isOpen) return null;

  const handleSave = () => {
    onSave(localPrefs);
    onClose();
  };

  const tabs: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
    { id: 'general', label: 'General', icon: <Sun className="w-4 h-4" /> },
    { id: 'approval', label: 'Approval', icon: <Shield className="w-4 h-4" /> },
    { id: 'memory', label: 'Memory', icon: <Database className="w-4 h-4" /> },
    { id: 'mcp', label: 'MCP Servers', icon: <Server className="w-4 h-4" /> },
    { id: 'firefighter', label: 'Firefighter', icon: <Flame className="w-4 h-4" /> },
    { id: 'about', label: 'About', icon: <Bot className="w-4 h-4" /> },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative bg-slate-900 rounded-xl shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden border border-slate-700">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-700">
          <h2 className="text-lg font-semibold text-slate-100">Settings</h2>
          <button
            onClick={onClose}
            className="p-1 text-slate-400 hover:text-slate-200 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex h-[60vh]">
          {/* Sidebar */}
          <div className="w-48 border-r border-slate-700 py-4">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full flex items-center gap-2 px-4 py-2 text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'bg-slate-800 text-slate-100 border-r-2 border-blue-500'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                }`}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-6">
            {activeTab === 'general' && (
              <GeneralSettings
                preferences={localPrefs}
                onChange={setLocalPrefs}
              />
            )}
            {activeTab === 'approval' && (
              <ApprovalSettings
                preferences={localPrefs}
                onChange={setLocalPrefs}
              />
            )}
            {activeTab === 'memory' && (
              <MemorySettings
                preferences={localPrefs}
                onChange={setLocalPrefs}
              />
            )}
            {activeTab === 'mcp' && (
              <MCPSettings
                servers={localPrefs.mcpServers}
                onChange={(servers) =>
                  setLocalPrefs({ ...localPrefs, mcpServers: servers })
                }
              />
            )}
            {activeTab === 'firefighter' && (
              <FirefighterSettings
                oktaDomain={localPrefs.oktaDomain}
                oktaClientID={localPrefs.oktaClientID}
                oktaClientSecret={localPrefs.oktaClientSecret}
                linearAPIKey={localPrefs.linearAPIKey}
                onOktaDomainChange={(domain) =>
                  setLocalPrefs({ ...localPrefs, oktaDomain: domain })
                }
                onOktaClientIDChange={(clientID) =>
                  setLocalPrefs({ ...localPrefs, oktaClientID: clientID })
                }
                onOktaClientSecretChange={(secret) =>
                  setLocalPrefs({ ...localPrefs, oktaClientSecret: secret })
                }
                onLinearAPIKeyChange={(key) =>
                  setLocalPrefs({ ...localPrefs, linearAPIKey: key })
                }
              />
            )}
            {activeTab === 'about' && <AboutSettings />}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-700">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-slate-300 hover:text-slate-100 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}

// General Settings Tab
function GeneralSettings({
  preferences,
  onChange,
}: {
  preferences: UserPreferences;
  onChange: (prefs: UserPreferences) => void;
}) {
  const [showApiKey, setShowApiKey] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-2">Authentication Method</h3>
        <p className="text-xs text-slate-400 mb-3">
          Choose how to authenticate with Claude
        </p>
        <div className="flex items-center gap-3 mb-4">
          <button
            onClick={() => onChange({ ...preferences, authMethod: 'anthropic-api' })}
            className={`flex-1 flex items-center justify-center gap-2 p-3 rounded-lg border transition-colors ${
              preferences.authMethod === 'anthropic-api'
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <Key className="w-4 h-4" />
            <span className="text-sm">Anthropic API</span>
          </button>
          <button
            onClick={() => onChange({ ...preferences, authMethod: 'google-cloud' })}
            className={`flex-1 flex items-center justify-center gap-2 p-3 rounded-lg border transition-colors ${
              preferences.authMethod === 'google-cloud'
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <Server className="w-4 h-4" />
            <span className="text-sm">Google Cloud</span>
          </button>
        </div>

        {preferences.authMethod === 'anthropic-api' ? (
          <div>
            <h3 className="text-sm font-medium text-slate-100 mb-2">API Key</h3>
            <p className="text-xs text-slate-400 mb-3">
              Your Anthropic API key for Claude access
            </p>
            <div className="relative">
              <Key className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                type={showApiKey ? 'text' : 'password'}
                value={preferences.apiKey || ''}
                onChange={(e) => onChange({ ...preferences, apiKey: e.target.value })}
                placeholder="sk-ant-..."
                className="w-full pl-10 pr-10 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
              <button
                type="button"
                onClick={() => setShowApiKey(!showApiKey)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
              >
                {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            <p className="text-xs text-slate-500 mt-2">
              Get your API key from{' '}
              <a href="https://console.anthropic.com" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline">
                console.anthropic.com
              </a>
            </p>
          </div>
        ) : (
          <GCloudAuthSection
            projectId={preferences.gcpProjectId}
            region={preferences.gcpRegion}
            onProjectChange={(projectId) => onChange({ ...preferences, gcpProjectId: projectId })}
            onRegionChange={(region) => onChange({ ...preferences, gcpRegion: region })}
          />
        )}
      </div>

      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-4">Appearance</h3>
        <div className="flex items-center gap-3">
          <button
            onClick={() => onChange({ ...preferences, theme: 'dark' })}
            className={`flex-1 flex items-center justify-center gap-2 p-4 rounded-lg border transition-colors ${
              preferences.theme === 'dark'
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <Moon className="w-5 h-5" />
            <span className="text-sm">Dark</span>
          </button>
          <button
            onClick={() => onChange({ ...preferences, theme: 'light' })}
            className={`flex-1 flex items-center justify-center gap-2 p-4 rounded-lg border transition-colors ${
              preferences.theme === 'light'
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-700 hover:border-slate-600'
            }`}
          >
            <Sun className="w-5 h-5" />
            <span className="text-sm">Light</span>
          </button>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-4">Notifications</h3>
        <label className="flex items-center justify-between p-4 rounded-lg border border-slate-700 cursor-pointer hover:bg-slate-800 transition-colors">
          <div className="flex items-center gap-3">
            {preferences.notificationsEnabled ? (
              <Bell className="w-5 h-5 text-blue-500" />
            ) : (
              <BellOff className="w-5 h-5 text-slate-500" />
            )}
            <div>
              <p className="text-sm text-slate-100">Desktop Notifications</p>
              <p className="text-xs text-slate-400">
                Get notified when tasks complete
              </p>
            </div>
          </div>
          <input
            type="checkbox"
            checked={preferences.notificationsEnabled}
            onChange={(e) =>
              onChange({ ...preferences, notificationsEnabled: e.target.checked })
            }
            className="w-4 h-4 rounded"
          />
        </label>
      </div>

      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-4">Default Model</h3>
        <select
          value={preferences.defaultModel}
          onChange={(e) =>
            onChange({ ...preferences, defaultModel: e.target.value })
          }
          className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
        >
          <option value="sonnet">Claude Sonnet 4</option>
          <option value="opus">Claude Opus 4</option>
          <option value="claude-opus-4-6">Claude Opus 4.6 (Latest)</option>
          <option value="haiku">Claude Haiku 4</option>
          <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</option>
        </select>
      </div>
    </div>
  );
}

// Approval Settings Tab
function ApprovalSettings({
  preferences,
  onChange,
}: {
  preferences: UserPreferences;
  onChange: (prefs: UserPreferences) => void;
}) {
  const modes: { value: ApprovalMode; label: string; description: string; icon: React.ReactNode }[] = [
    {
      value: 'suggest',
      label: 'Suggest Mode',
      description: 'Claude suggests changes, you approve everything',
      icon: <Shield className="w-5 h-5" />,
    },
    {
      value: 'auto-edit',
      label: 'Auto-Edit Mode',
      description: 'Claude can edit files, but asks for bash commands',
      icon: <Zap className="w-5 h-5" />,
    },
    {
      value: 'full-auto',
      label: 'Full Auto Mode',
      description: 'Claude has full control (use with caution)',
      icon: <Bot className="w-5 h-5" />,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-2">Approval Mode</h3>
        <p className="text-xs text-slate-400 mb-4">
          Control how much autonomy Claude has when making changes
        </p>
        <div className="space-y-3">
          {modes.map((mode) => (
            <label
              key={mode.value}
              className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                preferences.approvalMode === mode.value
                  ? 'border-blue-500 bg-blue-500/10'
                  : 'border-slate-700 hover:border-slate-600'
              }`}
            >
              <input
                type="radio"
                name="approvalMode"
                value={mode.value}
                checked={preferences.approvalMode === mode.value}
                onChange={() =>
                  onChange({ ...preferences, approvalMode: mode.value })
                }
                className="mt-1"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  {mode.icon}
                  <span className="text-sm font-medium text-slate-100">
                    {mode.label}
                  </span>
                </div>
                <p className="text-xs text-slate-400 mt-1">{mode.description}</p>
              </div>
            </label>
          ))}
        </div>
      </div>

      {preferences.approvalMode === 'full-auto' && (
        <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
          <p className="text-sm text-amber-500">
            Warning: Full auto mode gives Claude complete control over your
            system. Use this only in trusted environments.
          </p>
        </div>
      )}
    </div>
  );
}

// Memory Settings Tab
function MemorySettings({
  preferences,
  onChange,
}: {
  preferences: UserPreferences;
  onChange: (prefs: UserPreferences) => void;
}) {
  const [isCleaningUp, setIsCleaningUp] = useState(false);
  const [cleanupMessage, setCleanupMessage] = useState('');

  const handleManualCleanup = async () => {
    try {
      setIsCleaningUp(true);
      const count = await CleanupOldSessions();
      setCleanupMessage(`Cleaned up ${count} old session(s)`);
      setTimeout(() => setCleanupMessage(''), 3000);
    } catch (error) {
      setCleanupMessage('Failed to cleanup sessions');
      setTimeout(() => setCleanupMessage(''), 3000);
    } finally {
      setIsCleaningUp(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-2">Message Management</h3>
        <p className="text-xs text-slate-400 mb-4">
          Control how messages are stored and archived
        </p>

        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-300 mb-2">
              Max Messages Per Session
            </label>
            <input
              type="number"
              min="100"
              max="10000"
              value={preferences.maxMessagesPerSession || 1000}
              onChange={(e) =>
                onChange({
                  ...preferences,
                  maxMessagesPerSession: parseInt(e.target.value) || 1000,
                })
              }
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
            />
            <p className="text-xs text-slate-500 mt-1">
              Keep the most recent messages (default: 1000)
            </p>
          </div>

          <label className="flex items-center justify-between p-4 rounded-lg border border-slate-700 cursor-pointer hover:bg-slate-800 transition-colors">
            <div>
              <p className="text-sm text-slate-100">Archive Old Messages</p>
              <p className="text-xs text-slate-400">
                Save trimmed messages to disk instead of deleting them
              </p>
            </div>
            <input
              type="checkbox"
              checked={preferences.archiveOldMessages ?? true}
              onChange={(e) =>
                onChange({ ...preferences, archiveOldMessages: e.target.checked })
              }
              className="w-4 h-4 rounded"
            />
          </label>
        </div>
      </div>

      <div className="pt-6 border-t border-slate-700">
        <h3 className="text-sm font-medium text-slate-100 mb-2">Session Cleanup</h3>
        <p className="text-xs text-slate-400 mb-4">
          Automatically remove old sessions to save disk space
        </p>

        <div className="space-y-4">
          <label className="flex items-center justify-between p-4 rounded-lg border border-slate-700 cursor-pointer hover:bg-slate-800 transition-colors">
            <div>
              <p className="text-sm text-slate-100">Auto-Cleanup Sessions</p>
              <p className="text-xs text-slate-400">
                Automatically clean up sessions on app startup
              </p>
            </div>
            <input
              type="checkbox"
              checked={preferences.autoCleanupSessions ?? true}
              onChange={(e) =>
                onChange({ ...preferences, autoCleanupSessions: e.target.checked })
              }
              className="w-4 h-4 rounded"
            />
          </label>

          <div>
            <label className="block text-sm text-slate-300 mb-2">
              Max Session Age (days)
            </label>
            <input
              type="number"
              min="1"
              max="365"
              value={preferences.maxSessionAgeDays || 30}
              onChange={(e) =>
                onChange({
                  ...preferences,
                  maxSessionAgeDays: parseInt(e.target.value) || 30,
                })
              }
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
            />
            <p className="text-xs text-slate-500 mt-1">
              Delete sessions older than this (default: 30 days)
            </p>
          </div>

          <div>
            <label className="block text-sm text-slate-300 mb-2">
              Max Total Sessions
            </label>
            <input
              type="number"
              min="10"
              max="1000"
              value={preferences.maxTotalSessions || 100}
              onChange={(e) =>
                onChange({
                  ...preferences,
                  maxTotalSessions: parseInt(e.target.value) || 100,
                })
              }
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
            />
            <p className="text-xs text-slate-500 mt-1">
              Keep only the most recent sessions (default: 100)
            </p>
          </div>

          <div className="pt-2">
            <button
              onClick={handleManualCleanup}
              disabled={isCleaningUp}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 hover:bg-red-500/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Trash2 className="w-4 h-4" />
              {isCleaningUp ? 'Cleaning up...' : 'Clean Up Old Sessions Now'}
            </button>
            {cleanupMessage && (
              <p className="text-xs text-center text-green-400 mt-2">
                {cleanupMessage}
              </p>
            )}
          </div>
        </div>
      </div>

      <div className="pt-6 border-t border-slate-700">
        <h3 className="text-sm font-medium text-slate-100 mb-2">Agent Management</h3>
        <p className="text-xs text-slate-400 mb-4">
          Control how sub-agents are tracked and stored
        </p>

        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-300 mb-2">
              Max Agents Per Session
            </label>
            <input
              type="number"
              min="5"
              max="100"
              value={preferences.maxAgentsPerSession || 20}
              onChange={(e) =>
                onChange({
                  ...preferences,
                  maxAgentsPerSession: parseInt(e.target.value) || 20,
                })
              }
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
            />
            <p className="text-xs text-slate-500 mt-1">
              Maximum number of agents to track (default: 20)
            </p>
          </div>

          <label className="flex items-center justify-between p-4 rounded-lg border border-slate-700 cursor-pointer hover:bg-slate-800 transition-colors">
            <div>
              <p className="text-sm text-slate-100">Keep Completed Agents</p>
              <p className="text-xs text-slate-400">
                Retain completed agents (may increase memory usage)
              </p>
            </div>
            <input
              type="checkbox"
              checked={preferences.keepCompletedAgents ?? false}
              onChange={(e) =>
                onChange({ ...preferences, keepCompletedAgents: e.target.checked })
              }
              className="w-4 h-4 rounded"
            />
          </label>
        </div>
      </div>
    </div>
  );
}

// MCP Settings Tab
function MCPSettings({
  servers,
  onChange,
}: {
  servers: MCPServer[];
  onChange: (servers: MCPServer[]) => void;
}) {
  const [localServers, setLocalServers] = useState<MCPServer[]>([]);
  const [presets, setPresets] = useState<MCPServer[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load servers from backend on mount
  useEffect(() => {
    loadServers();
    loadPresets();
  }, []);

  const loadServers = async () => {
    try {
      setIsLoading(true);
      const mcpServers = await GetMCPServers();
      setLocalServers(mcpServers);
      setError(null);
    } catch (err) {
      console.error('Failed to load MCP servers:', err);
      setError('Failed to load MCP servers');
    } finally {
      setIsLoading(false);
    }
  };

  const loadPresets = async () => {
    try {
      const mcpPresets = await GetMCPPresets();
      // Filter out presets that are already added
      setPresets(mcpPresets);
    } catch (err) {
      console.error('Failed to load MCP presets:', err);
    }
  };

  const handleAddServer = async (server: MCPServer) => {
    try {
      await AddMCPServer(server);
      await loadServers();
      setDialogOpen(false);
    } catch (err) {
      console.error('Failed to add MCP server:', err);
      setError('Failed to add server');
    }
  };

  const handleRemoveServer = async (name: string) => {
    if (!confirm(`Are you sure you want to remove the "${name}" server?`)) {
      return;
    }

    try {
      await RemoveMCPServer(name);
      await loadServers();
    } catch (err) {
      console.error('Failed to remove MCP server:', err);
      setError('Failed to remove server');
    }
  };

  const handleToggleServer = async (server: MCPServer) => {
    try {
      const updated = { ...server, enabled: !server.enabled };
      await UpdateMCPServer(updated);
      await loadServers();
    } catch (err) {
      console.error('Failed to update MCP server:', err);
      setError('Failed to update server');
    }
  };

  // Filter presets to exclude already added servers
  const availablePresets = presets.filter(
    (preset) => !localServers.some((server) => server.name === preset.name)
  );

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-2">MCP Servers</h3>
        <p className="text-xs text-slate-400 mb-4">
          Configure Model Context Protocol servers for extended capabilities
        </p>

        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-sm text-red-400">
            {error}
          </div>
        )}

        {isLoading ? (
          <div className="text-center py-8">
            <div className="animate-spin w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full mx-auto mb-2"></div>
            <p className="text-sm text-slate-400">Loading servers...</p>
          </div>
        ) : localServers.length === 0 ? (
          <div className="text-center py-8 border border-dashed border-slate-700 rounded-lg">
            <Server className="w-8 h-8 text-slate-600 mx-auto mb-2" />
            <p className="text-sm text-slate-400">No MCP servers configured</p>
            <button
              onClick={() => setDialogOpen(true)}
              className="mt-3 flex items-center gap-1 mx-auto text-sm text-blue-500 hover:underline"
            >
              <Plus className="w-4 h-4" />
              Add Server
            </button>
          </div>
        ) : (
          <>
            <div className="space-y-3 mb-4">
              {localServers.map((server) => (
                <div
                  key={server.name}
                  className="flex items-center justify-between p-4 rounded-lg border border-slate-700"
                >
                  <div className="flex items-center gap-3 flex-1">
                    <Server className="w-5 h-5 text-slate-400 flex-shrink-0" />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm text-slate-100">{server.name}</p>
                        {server.enabled && (
                          <span className="px-1.5 py-0.5 text-xs bg-green-500/20 text-green-400 rounded">
                            Active
                          </span>
                        )}
                      </div>
                      {server.description && (
                        <p className="text-xs text-slate-500 mt-0.5">{server.description}</p>
                      )}
                      <p className="text-xs text-slate-600 mt-1">{server.command} {server.args?.join(' ')}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 flex-shrink-0">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <span className="text-xs text-slate-400">
                        {server.enabled ? 'Enabled' : 'Disabled'}
                      </span>
                      <input
                        type="checkbox"
                        checked={server.enabled}
                        onChange={() => handleToggleServer(server)}
                        className="w-4 h-4 rounded"
                      />
                    </label>
                    <button
                      onClick={() => handleRemoveServer(server.name)}
                      className="p-1.5 hover:bg-slate-600 rounded transition-colors"
                      aria-label="Remove server"
                    >
                      <Trash2 className="w-4 h-4 text-slate-400 hover:text-red-500" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
            <button
              onClick={() => setDialogOpen(true)}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-500/10 border border-blue-500/20 rounded-lg text-blue-500 hover:bg-blue-500/20 transition-colors"
            >
              <Plus className="w-4 h-4" />
              Add Server
            </button>
          </>
        )}
      </div>

      <MCPServerDialog
        isOpen={dialogOpen}
        onClose={() => setDialogOpen(false)}
        onAdd={handleAddServer}
        presets={availablePresets}
      />
    </div>
  );
}

// About Tab
function AboutSettings() {
  return (
    <div className="space-y-6">
      <div className="text-center py-8">
        <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-blue-500/20 flex items-center justify-center">
          <Bot className="w-8 h-8 text-blue-500" />
        </div>
        <h3 className="text-lg font-semibold text-slate-100">Boatman</h3>
        <p className="text-sm text-slate-400">Claude Code Desktop App</p>
        <p className="text-xs text-slate-500 mt-2">Version 0.1.0</p>
      </div>

      <div className="space-y-2 text-sm text-slate-400">
        <p>
          Built with Wails and React for a native desktop experience.
        </p>
        <p>
          Powered by Claude, Anthropic's AI assistant.
        </p>
      </div>

      <div className="pt-4 border-t border-slate-700">
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-blue-500 hover:underline"
        >
          View on GitHub
        </a>
      </div>
    </div>
  );
}
