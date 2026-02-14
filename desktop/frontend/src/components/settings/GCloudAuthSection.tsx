import { useState, useEffect } from 'react';
import { Cloud, CheckCircle, XCircle, AlertCircle, ExternalLink, RefreshCw } from 'lucide-react';
import {
  IsGCloudInstalled,
  IsGCloudAuthenticated,
  GetGCloudAuthInfo,
  GCloudLoginApplicationDefault,
  GCloudGetAvailableProjects,
  GCloudSetProject,
  GCloudRevoke,
} from '../../../wailsjs/go/main/App';

interface GCloudAuthSectionProps {
  projectId?: string;
  region?: string;
  onProjectChange: (projectId: string) => void;
  onRegionChange: (region: string) => void;
}

export function GCloudAuthSection({
  projectId,
  region,
  onProjectChange,
  onRegionChange,
}: GCloudAuthSectionProps) {
  const [isInstalled, setIsInstalled] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [authInfo, setAuthInfo] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [availableProjects, setAvailableProjects] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  const checkAuthStatus = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const installed = await IsGCloudInstalled();
      setIsInstalled(installed);

      if (installed) {
        const authenticated = await IsGCloudAuthenticated();
        setIsAuthenticated(authenticated);

        const info = await GetGCloudAuthInfo();
        setAuthInfo(info);

        if (authenticated) {
          const projects = await GCloudGetAvailableProjects();
          setAvailableProjects(projects);
        }
      }
    } catch (err: any) {
      console.error('Failed to check auth status:', err);
      setError(err.toString());
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const handleLogin = async () => {
    try {
      setIsAuthenticating(true);
      setError(null);
      await GCloudLoginApplicationDefault();
      await checkAuthStatus();
    } catch (err: any) {
      console.error('Failed to authenticate:', err);
      setError('Authentication failed. Please try again.');
    } finally {
      setIsAuthenticating(false);
    }
  };

  const handleRevoke = async () => {
    if (!confirm('Are you sure you want to revoke Google Cloud authentication?')) {
      return;
    }

    try {
      setIsLoading(true);
      await GCloudRevoke();
      await checkAuthStatus();
    } catch (err: any) {
      console.error('Failed to revoke:', err);
      setError('Failed to revoke authentication.');
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <RefreshCw className="w-5 h-5 text-slate-400 animate-spin" />
      </div>
    );
  }

  if (!isInstalled) {
    return (
      <div className="p-4 bg-amber-900/20 border border-amber-700/50 rounded-lg">
        <div className="flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
          <div>
            <h4 className="text-sm font-medium text-amber-500 mb-1">gcloud CLI Not Installed</h4>
            <p className="text-sm text-amber-200/80 mb-3">
              You need to install the Google Cloud CLI to use OAuth authentication.
            </p>
            <a
              href="https://cloud.google.com/sdk/docs/install"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-sm text-amber-400 hover:text-amber-300 underline"
            >
              Install gcloud CLI
              <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Authentication Status */}
      <div
        className={`p-4 rounded-lg border ${
          isAuthenticated
            ? 'bg-green-900/20 border-green-700/50'
            : 'bg-slate-800 border-slate-700'
        }`}
      >
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3 flex-1">
            <Cloud className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
              isAuthenticated ? 'text-green-500' : 'text-slate-400'
            }`} />
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <h4 className="text-sm font-medium text-slate-100">
                  Google Cloud OAuth
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

              {authInfo && isAuthenticated && (
                <div className="mt-2 space-y-1 text-xs text-slate-400">
                  <div>
                    <span className="text-slate-500">Account:</span>{' '}
                    <span className="text-slate-300">{authInfo.account || 'N/A'}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Project:</span>{' '}
                    <span className="text-slate-300">{authInfo.project || 'N/A'}</span>
                  </div>
                </div>
              )}

              {!isAuthenticated && (
                <p className="text-xs text-slate-400 mt-1">
                  Click "Sign In" to authenticate with your Google account
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
                {isAuthenticating ? 'Signing In...' : 'Sign In with OAuth'}
              </button>
            ) : (
              <>
                <button
                  onClick={checkAuthStatus}
                  className="px-3 py-1.5 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
                  title="Refresh status"
                >
                  <RefreshCw className="w-4 h-4" />
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

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-sm text-red-400">
          {error}
        </div>
      )}

      {/* Project Selection */}
      {isAuthenticated && (
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-2">
            GCP Project
          </label>
          <select
            value={projectId || ''}
            onChange={(e) => onProjectChange(e.target.value)}
            className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
          >
            <option value="">Select a project...</option>
            {availableProjects.map((project) => (
              <option key={project} value={project}>
                {project}
              </option>
            ))}
          </select>
          <p className="text-xs text-slate-500 mt-2">
            Select the GCP project where Claude/Vertex AI is enabled
          </p>
        </div>
      )}

      {/* Region Selection */}
      {isAuthenticated && projectId && (
        <div>
          <label className="block text-sm font-medium text-slate-200 mb-2">
            Region
          </label>
          <select
            value={region || 'us-east5'}
            onChange={(e) => onRegionChange(e.target.value)}
            className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-blue-500"
          >
            <option value="us-east5">us-east5</option>
            <option value="us-central1">us-central1</option>
            <option value="europe-west1">europe-west1</option>
            <option value="asia-southeast1">asia-southeast1</option>
          </select>
        </div>
      )}

      {/* Info Box */}
      <div className="p-3 bg-blue-900/20 border border-blue-700/50 rounded-lg">
        <p className="text-xs text-blue-200/80">
          <strong className="text-blue-300">OAuth Authentication:</strong> Uses your Google account credentials
          instead of API keys. More secure and supports organization policies.
        </p>
      </div>
    </div>
  );
}
