import { diff } from '../../wailsjs/go/models';

export interface DiffSummary {
  totalFiles: number;
  filesAdded: number;
  filesDeleted: number;
  filesModified: number;
  linesAdded: number;
  linesDeleted: number;
  riskLevel: 'low' | 'medium' | 'high';
}

export function calculateDiffSummary(diffs: diff.FileDiff[]): DiffSummary {
  let filesAdded = 0;
  let filesDeleted = 0;
  let filesModified = 0;
  let linesAdded = 0;
  let linesDeleted = 0;

  for (const diff of diffs) {
    if (diff.isNew) {
      filesAdded++;
    } else if (diff.isDelete) {
      filesDeleted++;
    } else {
      filesModified++;
    }

    for (const hunk of diff.hunks) {
      for (const line of hunk.lines) {
        if (line.type === 'addition') {
          linesAdded++;
        } else if (line.type === 'deletion') {
          linesDeleted++;
        }
      }
    }
  }

  // Calculate risk level based on total changes
  const totalChanges = linesAdded + linesDeleted;
  let riskLevel: 'low' | 'medium' | 'high' = 'low';

  if (totalChanges > 500 || filesDeleted > 5) {
    riskLevel = 'high';
  } else if (totalChanges > 100 || filesDeleted > 2) {
    riskLevel = 'medium';
  }

  return {
    totalFiles: diffs.length,
    filesAdded,
    filesDeleted,
    filesModified,
    linesAdded,
    linesDeleted,
    riskLevel,
  };
}

export function generateCommentId(): string {
  return `comment_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}
