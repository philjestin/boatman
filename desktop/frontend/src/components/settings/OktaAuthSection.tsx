import { useState } from 'react';
import { Shield, CheckCircle, XCircle, AlertCircle, ExternalLink, RefreshCw } from 'lucide-react';
import {
  OktaLogin,
  IsOktaAuthenticated,
  OktaRevoke,
} from '../../../wailsjs/go/main/App';

interface OktaAuthSectionProps {
  oktaDomain?: string;
  oktaClientID?: string;
  oktaClientSecret?: string;
  onOktaDomainChange: (domain: string) => void;
  onOktaClientIDChange: (clientID: string) => void;
  onOktaClientSecretChange: (secret: string) => void;
}

export function OktaAuthSection({
  oktaDomain,
  oktaClientID,
  oktaClientSecret,
  onOktaDomainChange,
  onOktaClientIDChange,
  onOktaClientSecretChange,
}: OktaAuthSectionProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const checkAuthStatus = async () => {
    if (!oktaDomain || !oktaClientID) {
      setIsAuthenticated(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      const authenticated = await IsOktaAuthenticated(
        oktaDomain,
        oktaClientID,
        oktaClientSecret || ''
      );
      setIsAuthenticated(authenticated);
    } catch (err: any) {
      console.error('Failed to check auth status:', err);
      setError(err.toString());
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogin = async () => {
    if (!oktaDomain || !oktaClientID) {
      setError('Please configure Okta domain and client ID first');
      return;
    }

    try {
      setIsAuthenticating(true);
      setError(null);
      await OktaLogin(oktaDomain, oktaClientID, oktaClientSecret || '');
      await checkAuthStatus();
    } catch (err: any) {
      console.error('Failed to authenticate:', err);
      setError('Authentication failed. Please try again.');
    } finally {
      setIsAuthenticating(false);
    }
  };

  const handleRevoke = async () => {
    if (!confirm('Are you sure you want to revoke Okta authentication?')) {
      return;
    }

    if (!oktaDomain || !oktaClientID) {
      return;
    }

    try {
      setIsLoading(true);
      await OktaRevoke(oktaDomain, oktaClientID, oktaClientSecret || '');
      await checkAuthStatus();
    } catch (err: any) {
      console.error('Failed to revoke:', err);
      setError('Failed to revoke authentication.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Okta Configuration */}
      <div>
        <label className="block text-sm font-medium text-slate-200 mb-2">
          Okta Domain
        </label>
        <input
          type="text"
          value={oktaDomain || ''}
          onChange={(e) => onOktaDomainChange(e.target.value)}
          placeholder="your-org.okta.com"
          className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500"
        />
        <p className="text-xs text-slate-500 mt-1">
          Your Okta organization domain (without https://)
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-200 mb-2">
          Client ID
        </label>
        <input
          type="text"
          value={oktaClientID || ''}
          onChange={(e) => onOktaClientIDChange(e.target.value)}
          placeholder="0oa..."
          className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-200 mb-2">
          Client Secret (Optional)
        </label>
        <input
          type="password"
          value={oktaClientSecret || ''}
          onChange={(e) => onOktaClientSecretChange(e.target.value)}
          placeholder="Optional for public clients"
          className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500"
        />
        <p className="text-xs text-slate-500 mt-1">
          Only required for confidential clients
        </p>
      </div>

      {/* Authentication Status */}
      {oktaDomain && oktaClientID && (
        <div
          className={`p-4 rounded-lg border ${
            isAuthenticated
              ? 'bg-green-900/20 border-green-700/50'
              : 'bg-slate-800 border-slate-700'
          }`}
        >
          <div className="flex items-start justify-between">
            <div className="flex items-start gap-3 flex-1">
              <Shield className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                isAuthenticated ? 'text-green-500' : 'text-slate-400'
              }`} />
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <h4 className="text-sm font-medium text-slate-100">
                    Okta OAuth
                  </h4>
                  {isAuthenticated ? (
                    <span className="flex items-center gap-1 px-2 py-0.5 text-xs bg-green-500/20 text-green-400 rounded-full">
                      <CheckCircle className="w-3 h-3" />
                      Authenticated
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 px-2 py-0.5 text-xs bg-slate-600 text-slate-400 rounded-full">
                      <XCircle className="w-3 h-3" />
                      Not Authenticated
                    </span>
                  )}
                </div>

                {!isAuthenticated && (
                  <p className="text-xs text-slate-400 mt-1">
                    Click "Sign In with Okta" to authenticate
                  </p>
                )}
              </div>
            </div>

            <div className="flex gap-2">
              {!isAuthenticated ? (
                <button
                  onClick={handleLogin}
                  disabled={isAuthenticating}
                  className="px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isAuthenticating ? 'Signing In...' : 'Sign In with Okta'}
                </button>
              ) : (
                <>
                  <button
                    onClick={checkAuthStatus}
                    disabled={isLoading}
                    className="px-3 py-1.5 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
                    title="Refresh status"
                  >
                    <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
                  </button>
                  <button
                    onClick={handleRevoke}
                    className="px-3 py-1.5 text-sm text-red-400 hover:text-red-300 hover:bg-red-900/20 rounded-md transition-colors"
                  >
                    Sign Out
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-sm text-red-400">
          {error}
        </div>
      )}

      {/* Info Box */}
      <div className="p-3 bg-blue-900/20 border border-blue-700/50 rounded-lg">
        <p className="text-xs text-blue-200/80">
          <strong className="text-blue-300">Okta OAuth:</strong> Authenticate with your Okta
          organization to access Datadog and Bugsnag APIs through SSO. Configure your Okta
          application to allow this redirect URI: <code className="text-blue-300">http://localhost:8484/callback</code>
        </p>
        <a
          href="https://developer.okta.com/docs/guides/implement-oauth-for-okta/main/"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 underline mt-2"
        >
          Okta OAuth Setup Guide
          <ExternalLink className="w-3 h-3" />
        </a>
      </div>
    </div>
  );
}
