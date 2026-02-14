import { Flame, CheckCircle, AlertTriangle, Copy } from 'lucide-react';
import { OktaAuthSection } from './OktaAuthSection';

interface FirefighterSettingsProps {
  oktaDomain?: string;
  oktaClientID?: string;
  oktaClientSecret?: string;
  linearAPIKey?: string;
  onOktaDomainChange: (domain: string) => void;
  onOktaClientIDChange: (clientID: string) => void;
  onOktaClientSecretChange: (secret: string) => void;
  onLinearAPIKeyChange: (key: string) => void;
}

export function FirefighterSettings({
  oktaDomain,
  oktaClientID,
  oktaClientSecret,
  linearAPIKey,
  onOktaDomainChange,
  onOktaClientIDChange,
  onOktaClientSecretChange,
  onLinearAPIKeyChange,
}: FirefighterSettingsProps) {
  const isConfigured = oktaDomain && oktaClientID;

  const configExample = `{
  "mcpServers": {
    "datadog-okta": {
      "command": "./mcp-servers/datadog-okta/datadog-okta",
      "args": [],
      "env": {
        "OKTA_ACCESS_TOKEN": "[automatically-injected]",
        "DD_SITE": "datadoghq.com"
      }
    },
    "bugsnag-okta": {
      "command": "./mcp-servers/bugsnag-okta/bugsnag-okta",
      "args": [],
      "env": {
        "OKTA_ACCESS_TOKEN": "[automatically-injected]"
      }
    }
  }
}`;

  const copyConfig = () => {
    navigator.clipboard.writeText(configExample);
  };

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-slate-100 mb-2 flex items-center gap-2">
          <Flame className="w-4 h-4 text-red-500" />
          Firefighter Mode Configuration
        </h3>
        <p className="text-xs text-slate-400 mb-4">
          Configure API credentials for production monitoring and incident response
        </p>
      </div>

      {/* Status Overview */}
      <div className={`p-4 rounded-lg border ${
        isConfigured
          ? 'bg-green-900/20 border-green-700/50'
          : 'bg-amber-900/20 border-amber-700/50'
      }`}>
        <div className="flex items-center gap-2 mb-2">
          {isConfigured ? (
            <>
              <CheckCircle className="w-5 h-5 text-green-500" />
              <span className="text-sm font-medium text-green-400">Okta OAuth Configured</span>
            </>
          ) : (
            <>
              <AlertTriangle className="w-5 h-5 text-amber-500" />
              <span className="text-sm font-medium text-amber-400">Okta Configuration Required</span>
            </>
          )}
        </div>
        <p className="text-xs text-slate-400">
          {isConfigured
            ? 'Sign in with Okta below to enable Firefighter mode with Datadog and Bugsnag access'
            : 'Configure your Okta domain and client ID to enable OAuth authentication'}
        </p>
      </div>

      {/* Okta OAuth Configuration */}
      <OktaAuthSection
        oktaDomain={oktaDomain}
        oktaClientID={oktaClientID}
        oktaClientSecret={oktaClientSecret}
        onOktaDomainChange={onOktaDomainChange}
        onOktaClientIDChange={onOktaClientIDChange}
        onOktaClientSecretChange={onOktaClientSecretChange}
      />

      {/* Linear API Key */}
      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">
          Linear API Key
        </label>
        <input
          type="password"
          value={linearAPIKey || ''}
          onChange={(e) => onLinearAPIKeyChange(e.target.value)}
          placeholder="lin_api_..."
          className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <p className="text-xs text-slate-400 mt-1">
          Required for Boatman Mode ticket execution and Firefighter Mode Linear integration.
          Get your API key from Linear Settings → API → Personal API keys
        </p>
      </div>

      {/* MCP Configuration Instructions */}
      <div className="p-4 bg-blue-900/20 border border-blue-700/50 rounded-lg">
        <h4 className="text-sm font-medium text-blue-300 mb-2">MCP Configuration</h4>
        <p className="text-xs text-blue-200/80 mb-3">
          After saving these settings, add the MCP servers via Settings → MCP Servers tab.
          The servers will use these credentials automatically.
        </p>
        <div className="relative">
          <pre className="text-xs bg-slate-900/50 p-3 rounded-md overflow-x-auto text-slate-300 border border-slate-700">
            {configExample}
          </pre>
          <button
            onClick={copyConfig}
            className="absolute top-2 right-2 p-1.5 bg-slate-800 hover:bg-slate-700 rounded border border-slate-600 transition-colors"
            title="Copy to clipboard"
          >
            <Copy className="w-3 h-3 text-slate-400" />
          </button>
        </div>
        <p className="text-xs text-blue-200/70 mt-2">
          Save to: <code className="text-blue-300">~/.claude/claude_mcp_config.json</code>
        </p>
      </div>
    </div>
  );
}
