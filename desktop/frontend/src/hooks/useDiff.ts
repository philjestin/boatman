import { useState, useCallback } from 'react';
import { GetGitDiff, ParseDiff, GetSideBySideDiff } from '../../wailsjs/go/main/App';
import { diff } from '../../wailsjs/go/models';
import { useStore } from '../store';

export function useDiff() {
  const [diffs, setDiffs] = useState<diff.FileDiff[]>([]);
  const [sideBySideData, setSideBySideData] = useState<Record<string, diff.SideBySideLine[]>>({});
  const { setError } = useStore();

  // Load diffs from git status
  const loadDiffs = useCallback(async (projectPath: string) => {
    try {
      // Get git diff for the entire project (empty string means all changes)
      const diffText = await GetGitDiff(projectPath, '');

      if (diffText && diffText.trim()) {
        const parsedDiffs = await ParseDiff(diffText);
        setDiffs(parsedDiffs);
      } else {
        setDiffs([]);
      }
    } catch (err) {
      console.error('Failed to load diffs:', err);
      setError('Failed to load changes');
      setDiffs([]);
    }
  }, [setError]);

  // Load side-by-side data for a specific file
  const loadSideBySide = useCallback(async (fileDiff: diff.FileDiff) => {
    try {
      const filePath = fileDiff.newPath || fileDiff.oldPath;
      const sideBySideLines = await GetSideBySideDiff(fileDiff);

      setSideBySideData((prev) => ({
        ...prev,
        [filePath]: sideBySideLines,
      }));
    } catch (err) {
      console.error('Failed to load side-by-side diff:', err);
      setError('Failed to load side-by-side view');
    }
  }, [setError]);

  // Accept a file (mark as approved)
  const acceptFile = useCallback((filePath: string) => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        if ((diff.newPath || diff.oldPath) === filePath) {
          diff.approved = true;
        }
        return diff;
      })
    );
  }, []);

  // Reject a file (mark as not approved)
  const rejectFile = useCallback((filePath: string) => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        if ((diff.newPath || diff.oldPath) === filePath) {
          diff.approved = false;
        }
        return diff;
      })
    );
  }, []);

  // Accept all files
  const acceptAll = useCallback(() => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        diff.approved = true;
        return diff;
      })
    );
  }, []);

  // Reject all files
  const rejectAll = useCallback(() => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        diff.approved = false;
        return diff;
      })
    );
  }, []);

  // Update comments for a file
  const updateComments = useCallback((filePath: string, comments: diff.DiffComment[]) => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((d) => {
        if ((d.newPath || d.oldPath) === filePath) {
          d.comments = comments;
        }
        return d;
      })
    );
  }, []);

  // Approve a specific hunk
  const approveHunk = useCallback((filePath: string, hunkId: string) => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        if ((diff.newPath || diff.oldPath) === filePath) {
          diff.hunks.forEach((hunk) => {
            if (hunk.id === hunkId) {
              hunk.approved = true;
            }
          });
        }
        return diff;
      })
    );
  }, []);

  // Reject a specific hunk
  const rejectHunk = useCallback((filePath: string, hunkId: string) => {
    setDiffs((prevDiffs) =>
      prevDiffs.map((diff) => {
        if ((diff.newPath || diff.oldPath) === filePath) {
          diff.hunks.forEach((hunk) => {
            if (hunk.id === hunkId) {
              hunk.approved = false;
            }
          });
        }
        return diff;
      })
    );
  }, []);

  return {
    diffs,
    sideBySideData,
    loadDiffs,
    loadSideBySide,
    acceptFile,
    rejectFile,
    acceptAll,
    rejectAll,
    updateComments,
    approveHunk,
    rejectHunk,
  };
}
