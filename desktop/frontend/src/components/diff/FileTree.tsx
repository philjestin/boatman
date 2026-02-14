import { File, FilePlus, FileMinus, FileEdit, ChevronRight } from 'lucide-react';
import { diff } from '../../../wailsjs/go/models';

interface FileTreeProps {
  files: diff.FileDiff[];
  selectedFile: string | null;
  onFileSelect: (path: string) => void;
  selectedFiles?: Set<string>;
  onToggleFileSelection?: (path: string) => void;
}

function getFileIcon(fileDiff: diff.FileDiff) {
  if (fileDiff.isNew) return <FilePlus className="w-4 h-4 text-green-500" />;
  if (fileDiff.isDelete) return <FileMinus className="w-4 h-4 text-red-500" />;
  return <FileEdit className="w-4 h-4 text-yellow-500" />;
}

function getFileLabel(fileDiff: diff.FileDiff) {
  if (fileDiff.isNew) return 'Added';
  if (fileDiff.isDelete) return 'Deleted';
  return 'Modified';
}

export function FileTree({
  files,
  selectedFile,
  onFileSelect,
  selectedFiles,
  onToggleFileSelection
}: FileTreeProps) {
  const groupedFiles = files.reduce((acc, file) => {
    const path = file.newPath || file.oldPath;
    const parts = path.split('/');
    const fileName = parts.pop() || '';
    const directory = parts.join('/') || '.';

    if (!acc[directory]) {
      acc[directory] = [];
    }
    acc[directory].push({ ...(file as any), fileName });
    return acc;
  }, {} as Record<string, any[]>);

  return (
    <div className="bg-slate-900 border-r border-slate-700 w-64 overflow-y-auto">
      <div className="px-3 py-2 border-b border-slate-700">
        <h3 className="text-xs font-medium text-slate-400 uppercase tracking-wider">
          Changed Files ({files.length})
        </h3>
      </div>
      <div className="py-2">
        {Object.entries(groupedFiles).map(([directory, dirFiles]) => (
          <div key={directory}>
            <div className="px-3 py-1 flex items-center gap-1 text-xs text-slate-500">
              <ChevronRight className="w-3 h-3" />
              <span className="truncate">{directory}</span>
            </div>
            {dirFiles.map((file) => {
              const filePath = file.newPath || file.oldPath;
              const isSelected = selectedFile === filePath;
              const isChecked = selectedFiles?.has(filePath) || false;

              return (
                <div
                  key={filePath}
                  className={`w-full flex items-center gap-2 px-4 py-1.5 text-sm transition-colors ${
                    isSelected
                      ? 'bg-slate-700 text-slate-100'
                      : 'text-slate-300 hover:bg-slate-800'
                  }`}
                >
                  {onToggleFileSelection && (
                    <input
                      type="checkbox"
                      checked={isChecked}
                      onChange={() => onToggleFileSelection(filePath)}
                      className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-blue-500 focus:ring-blue-500 focus:ring-offset-slate-900"
                      onClick={(e) => e.stopPropagation()}
                    />
                  )}
                  <button
                    onClick={() => onFileSelect(filePath)}
                    className="flex-1 flex items-center gap-2 min-w-0"
                  >
                    {getFileIcon(file)}
                    <span className="flex-1 truncate text-left">{file.fileName}</span>
                    <span
                      className={`text-xs px-1.5 py-0.5 rounded flex-shrink-0 ${
                        file.isNew
                          ? 'bg-green-500/20 text-green-500'
                          : file.isDelete
                          ? 'bg-red-500/20 text-red-500'
                          : 'bg-yellow-500/20 text-yellow-500'
                      }`}
                    >
                      {getFileLabel(file)}
                    </span>
                  </button>
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}
