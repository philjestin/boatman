import { useState, useMemo } from 'react';
import { Check, X, Columns, Rows, File } from 'lucide-react';
import { FileTree } from './FileTree';
import { DiffLine, SideBySideLine } from './DiffLine';
import { DiffCommentThread } from './DiffCommentThread';
import { BatchApprovalBar } from './BatchApprovalBar';
import { DiffSummaryCard } from './DiffSummaryCard';
import { calculateDiffSummary, generateCommentId } from '../../utils/diffUtils';
import { diff } from '../../../wailsjs/go/models';

interface DiffViewProps {
  diffs: diff.FileDiff[];
  sideBySideData?: Record<string, diff.SideBySideLine[]>;
  onAccept?: (filePath: string) => void;
  onReject?: (filePath: string) => void;
  onAcceptAll?: () => void;
  onRejectAll?: () => void;
  onUpdateComments?: (filePath: string, comments: diff.DiffComment[]) => void;
  onApproveHunk?: (filePath: string, hunkId: string) => void;
  onRejectHunk?: (filePath: string, hunkId: string) => void;
}

type ViewMode = 'unified' | 'split';

export function DiffView({
  diffs,
  sideBySideData = {},
  onAccept,
  onReject,
  onAcceptAll,
  onRejectAll,
  onUpdateComments,
  onApproveHunk,
  onRejectHunk,
}: DiffViewProps) {
  const [selectedFile, setSelectedFile] = useState<string | null>(
    diffs.length > 0 ? diffs[0].newPath || diffs[0].oldPath : null
  );
  const [viewMode, setViewMode] = useState<ViewMode>('unified');
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());
  const [showComments, setShowComments] = useState<Record<string, boolean>>({});

  const selectedDiff = diffs.find(
    (d) => (d.newPath || d.oldPath) === selectedFile
  );

  const sideBySideLines = selectedFile ? sideBySideData[selectedFile] : undefined;

  const summary = useMemo(() => calculateDiffSummary(diffs), [diffs]);

  const handleToggleFileSelection = (filePath: string) => {
    const newSelection = new Set(selectedFiles);
    if (newSelection.has(filePath)) {
      newSelection.delete(filePath);
    } else {
      newSelection.add(filePath);
    }
    setSelectedFiles(newSelection);
  };

  const handleApproveSelected = () => {
    selectedFiles.forEach((filePath) => {
      onAccept?.(filePath);
    });
    setSelectedFiles(new Set());
  };

  const handleRejectSelected = () => {
    selectedFiles.forEach((filePath) => {
      onReject?.(filePath);
    });
    setSelectedFiles(new Set());
  };

  const handleAddComment = (lineNum: number, hunkId: string | undefined, content: string) => {
    if (!selectedFile || !onUpdateComments) return;

    const fileDiff = diffs.find((d) => (d.newPath || d.oldPath) === selectedFile);
    if (!fileDiff) return;

    const newComment: diff.DiffComment = {
      id: generateCommentId(),
      lineNum,
      hunkId,
      content,
      timestamp: new Date().toISOString(),
    };

    const updatedComments = [...(fileDiff.comments || []), newComment];
    onUpdateComments(selectedFile, updatedComments);
  };

  const handleDeleteComment = (commentId: string) => {
    if (!selectedFile || !onUpdateComments) return;

    const fileDiff = diffs.find((d) => (d.newPath || d.oldPath) === selectedFile);
    if (!fileDiff) return;

    const updatedComments = (fileDiff.comments || []).filter((c) => c.id !== commentId);
    onUpdateComments(selectedFile, updatedComments);
  };

  const toggleCommentThread = (lineKey: string) => {
    setShowComments((prev) => ({
      ...prev,
      [lineKey]: !prev[lineKey],
    }));
  };

  return (
    <div className="flex flex-col h-full bg-slate-950">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 bg-slate-900 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode('unified')}
            className={`flex items-center gap-1.5 px-2 py-1 text-sm rounded transition-colors ${
              viewMode === 'unified'
                ? 'bg-slate-700 text-slate-100'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Rows className="w-4 h-4" />
            Unified
          </button>
          <button
            onClick={() => setViewMode('split')}
            className={`flex items-center gap-1.5 px-2 py-1 text-sm rounded transition-colors ${
              viewMode === 'split'
                ? 'bg-slate-700 text-slate-100'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Columns className="w-4 h-4" />
            Split
          </button>
        </div>
        <div className="flex items-center gap-2">
          {onRejectAll && (
            <button
              onClick={onRejectAll}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-red-500 hover:bg-red-500/10 rounded transition-colors"
            >
              <X className="w-4 h-4" />
              Reject All
            </button>
          )}
          {onAcceptAll && (
            <button
              onClick={onAcceptAll}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
            >
              <Check className="w-4 h-4" />
              Accept All
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Sidebar: File tree + Summary */}
        <div className="flex flex-col border-r border-slate-700 w-64">
          <FileTree
            files={diffs}
            selectedFile={selectedFile}
            onFileSelect={setSelectedFile}
            selectedFiles={selectedFiles}
            onToggleFileSelection={handleToggleFileSelection}
          />
          <div className="p-3 border-t border-slate-700">
            <DiffSummaryCard summary={summary} />
          </div>
        </div>

        {/* Diff content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedDiff ? (
            <>
              {/* File header */}
              <div className="flex items-center justify-between px-4 py-2 bg-slate-800 border-b border-slate-700">
                <div className="flex items-center gap-2">
                  <File className="w-4 h-4 text-slate-400" />
                  <span className="text-sm text-slate-200">
                    {selectedDiff.newPath || selectedDiff.oldPath}
                  </span>
                  {selectedDiff.isNew && (
                    <span className="text-xs px-1.5 py-0.5 bg-green-500/20 text-green-500 rounded">
                      New
                    </span>
                  )}
                  {selectedDiff.isDelete && (
                    <span className="text-xs px-1.5 py-0.5 bg-red-500/20 text-red-500 rounded">
                      Deleted
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {onReject && selectedFile && (
                    <button
                      onClick={() => onReject(selectedFile)}
                      className="flex items-center gap-1 px-2 py-1 text-xs text-red-500 hover:bg-red-500/10 rounded transition-colors"
                    >
                      <X className="w-3 h-3" />
                      Reject
                    </button>
                  )}
                  {onAccept && selectedFile && (
                    <button
                      onClick={() => onAccept(selectedFile)}
                      className="flex items-center gap-1 px-2 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
                    >
                      <Check className="w-3 h-3" />
                      Accept
                    </button>
                  )}
                </div>
              </div>

              {/* Diff lines */}
              <div className="flex-1 overflow-auto bg-slate-950">
                {selectedDiff.isBinary ? (
                  <div className="flex items-center justify-center h-full text-slate-400">
                    Binary file - cannot display diff
                  </div>
                ) : viewMode === 'split' && sideBySideLines ? (
                  <div className="min-w-max">
                    {sideBySideLines.map((line, index) => (
                      <SideBySideLine
                        key={index}
                        leftNum={line.leftNum}
                        leftContent={line.leftContent}
                        rightNum={line.rightNum}
                        rightContent={line.rightContent}
                        type={line.type as any}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="min-w-max">
                    {selectedDiff.hunks.map((hunk, hunkIndex) => (
                      <div key={hunkIndex}>
                        <div className="px-4 py-1 bg-slate-800 text-xs text-slate-400 font-mono">
                          @@ -{hunk.oldStart},{hunk.oldLines} +{hunk.newStart},{hunk.newLines} @@
                        </div>
                        {hunk.lines.map((line, lineIndex) => (
                          <DiffLine
                            key={`${hunkIndex}-${lineIndex}`}
                            type={line.type as any}
                            content={line.content}
                            oldNum={line.oldNum}
                            newNum={line.newNum}
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-slate-400">
              Select a file to view changes
            </div>
          )}
        </div>
      </div>

      {/* Batch Approval Bar */}
      <BatchApprovalBar
        selectedCount={selectedFiles.size}
        totalCount={diffs.length}
        onApproveSelected={handleApproveSelected}
        onRejectSelected={handleRejectSelected}
        onClearSelection={() => setSelectedFiles(new Set())}
      />
    </div>
  );
}
