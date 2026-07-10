import { useMemo, useState } from 'react';
import { FolderOpen, GitBranch, Plus, Trash2, X } from 'lucide-react';
import type { Project } from '../../types';

type ProjectMode = 'single' | 'multi';

interface ProjectDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onOpenSingleProject: () => Promise<Project | null>;
  onCreateMultiRepoProject: (name: string, repositoryPaths: string[]) => Promise<Project | null>;
  onSelectFolder: () => Promise<string | null>;
}

function baseName(path: string): string {
  return path.split(/[\\/]/).filter(Boolean).pop() || path;
}

export function ProjectDialog({
  isOpen,
  onClose,
  onOpenSingleProject,
  onCreateMultiRepoProject,
  onSelectFolder,
}: ProjectDialogProps) {
  const [mode, setMode] = useState<ProjectMode>('single');
  const [name, setName] = useState('');
  const [repositoryPaths, setRepositoryPaths] = useState<string[]>([]);
  const [isWorking, setIsWorking] = useState(false);

  const resolvedName = useMemo(() => {
    const trimmed = name.trim();
    if (trimmed) return trimmed;
    if (repositoryPaths.length === 0) return '';
    if (repositoryPaths.length === 1) return baseName(repositoryPaths[0]);
    return `${baseName(repositoryPaths[0])} workspace`;
  }, [name, repositoryPaths]);

  if (!isOpen) return null;

  const reset = () => {
    setMode('single');
    setName('');
    setRepositoryPaths([]);
    setIsWorking(false);
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const handleOpenSingleProject = async () => {
    setIsWorking(true);
    try {
      const project = await onOpenSingleProject();
      if (project) handleClose();
    } finally {
      setIsWorking(false);
    }
  };

  const handleAddRepository = async () => {
    const selectedPath = await onSelectFolder();
    if (!selectedPath) return;
    setRepositoryPaths((current) => {
      if (current.includes(selectedPath)) return current;
      return [...current, selectedPath];
    });
  };

  const handleRemoveRepository = (path: string) => {
    setRepositoryPaths((current) => current.filter((candidate) => candidate !== path));
  };

  const handleCreateMultiRepoProject = async () => {
    if (repositoryPaths.length === 0) return;
    setIsWorking(true);
    try {
      const project = await onCreateMultiRepoProject(resolvedName, repositoryPaths);
      if (project) handleClose();
    } finally {
      setIsWorking(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 no-drag">
      <div className="bg-slate-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 border border-slate-700">
        <div className="flex items-center justify-between p-6 border-b border-slate-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-500/20 rounded-lg flex items-center justify-center">
              <FolderOpen className="w-6 h-6 text-blue-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-slate-100">Open Project</h2>
              <p className="text-sm text-slate-400">Select one repo or create a multi-repo workspace</p>
            </div>
          </div>
          <button
            onClick={handleClose}
            className="p-2 hover:bg-slate-700 rounded-md transition-colors"
            aria-label="Close"
            disabled={isWorking}
          >
            <X className="w-5 h-5 text-slate-400" />
          </button>
        </div>

        <div className="p-6 space-y-5">
          <div className="flex items-center gap-2 p-1 bg-slate-900/50 rounded-lg border border-slate-700">
            <button
              onClick={() => setMode('single')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
                mode === 'single'
                  ? 'bg-blue-500 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
              disabled={isWorking}
            >
              <FolderOpen className="w-4 h-4" />
              Single Repo
            </button>
            <button
              onClick={() => setMode('multi')}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm rounded-md transition-colors ${
                mode === 'multi'
                  ? 'bg-blue-500 text-white'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
              disabled={isWorking}
            >
              <GitBranch className="w-4 h-4" />
              Multi Repo
            </button>
          </div>

          {mode === 'single' ? (
            <div className="flex items-center justify-between gap-4 rounded-lg border border-slate-700 bg-slate-900/50 p-4">
              <div className="min-w-0">
                <h3 className="text-sm font-medium text-slate-200">Repository Folder</h3>
                <p className="mt-1 text-sm text-slate-400">Open an existing repository as the active project.</p>
              </div>
              <button
                onClick={handleOpenSingleProject}
                className="flex items-center gap-2 px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={isWorking}
              >
                <FolderOpen className="w-4 h-4" />
                Choose Folder
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Project Name
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder={resolvedName || 'Employer Graph'}
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-md text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  disabled={isWorking}
                />
              </div>

              <div>
                <div className="flex items-center justify-between gap-3 mb-2">
                  <label className="block text-sm font-medium text-slate-300">
                    Repositories
                  </label>
                  <button
                    onClick={handleAddRepository}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-slate-200 bg-slate-700 hover:bg-slate-600 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    disabled={isWorking}
                  >
                    <Plus className="w-4 h-4" />
                    Add Repo
                  </button>
                </div>

                <div className="overflow-hidden rounded-lg border border-slate-700 bg-slate-900/50">
                  {repositoryPaths.length === 0 ? (
                    <button
                      onClick={handleAddRepository}
                      className="w-full flex items-center justify-center gap-2 px-4 py-8 text-sm text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition-colors"
                      disabled={isWorking}
                    >
                      <FolderOpen className="w-4 h-4" />
                      Choose First Repo
                    </button>
                  ) : (
                    <div className="divide-y divide-slate-700">
                      {repositoryPaths.map((path, index) => (
                        <div key={path} className="flex items-center gap-3 px-3 py-3">
                          <GitBranch className="w-4 h-4 text-slate-500 flex-shrink-0" />
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <p className="text-sm text-slate-200 truncate">{baseName(path)}</p>
                              {index === 0 && (
                                <span className="px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-blue-200 bg-blue-500/20 border border-blue-500/30 rounded">
                                  Primary
                                </span>
                              )}
                            </div>
                            <p className="mt-0.5 text-xs text-slate-500 font-mono truncate">{path}</p>
                          </div>
                          <button
                            onClick={() => handleRemoveRepository(path)}
                            className="p-1.5 text-slate-500 hover:text-red-400 hover:bg-slate-800 rounded-md transition-colors"
                            aria-label={`Remove ${baseName(path)}`}
                            disabled={isWorking}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="flex items-center justify-end gap-3 p-6 border-t border-slate-700">
          <button
            onClick={handleClose}
            className="px-4 py-2 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors"
            disabled={isWorking}
          >
            Cancel
          </button>
          {mode === 'multi' && (
            <button
              onClick={handleCreateMultiRepoProject}
              disabled={repositoryPaths.length === 0 || isWorking}
              className="flex items-center gap-2 px-4 py-2 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <GitBranch className="w-4 h-4" />
              {isWorking ? 'Creating...' : 'Create Project'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
