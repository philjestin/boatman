import { describe, it, expect } from 'vitest';
import { calculateDiffSummary, generateCommentId } from './diffUtils';
import { diff } from '../../wailsjs/go/models';

describe('diffUtils', () => {
  describe('calculateDiffSummary', () => {
    it('should calculate summary for empty diffs', () => {
      const summary = calculateDiffSummary([]);

      expect(summary).toEqual({
        totalFiles: 0,
        filesAdded: 0,
        filesDeleted: 0,
        filesModified: 0,
        linesAdded: 0,
        linesDeleted: 0,
        riskLevel: 'low',
      });
    });

    it('should calculate summary for new files', () => {
      const diffs = [
        {
          oldPath: '/dev/null',
          newPath: 'file1.txt',
          hunks: [
            {
              oldStart: 0,
              oldLines: 0,
              newStart: 1,
              newLines: 3,
              lines: [
                { type: 'addition', content: 'line 1', newNum: 1 },
                { type: 'addition', content: 'line 2', newNum: 2 },
                { type: 'addition', content: 'line 3', newNum: 3 },
              ],
            },
          ],
          isNew: true,
          isDelete: false,
          isBinary: false,
        },
      ];

      const summary = calculateDiffSummary(diffs as any);

      expect(summary.totalFiles).toBe(1);
      expect(summary.filesAdded).toBe(1);
      expect(summary.filesDeleted).toBe(0);
      expect(summary.filesModified).toBe(0);
      expect(summary.linesAdded).toBe(3);
      expect(summary.linesDeleted).toBe(0);
      expect(summary.riskLevel).toBe('low');
    });

    it('should calculate summary for deleted files', () => {
      const diffs = [
        {
          oldPath: 'file1.txt',
          newPath: '/dev/null',
          hunks: [
            {
              oldStart: 1,
              oldLines: 3,
              newStart: 0,
              newLines: 0,
              lines: [
                { type: 'deletion', content: 'line 1', oldNum: 1 },
                { type: 'deletion', content: 'line 2', oldNum: 2 },
                { type: 'deletion', content: 'line 3', oldNum: 3 },
              ],
            },
          ],
          isNew: false,
          isDelete: true,
          isBinary: false,
        },
      ];

      const summary = calculateDiffSummary(diffs as any);

      expect(summary.filesDeleted).toBe(1);
      expect(summary.linesDeleted).toBe(3);
    });

    it('should calculate summary for modified files', () => {
      const diffs = [
        {
          oldPath: 'file1.txt',
          newPath: 'file1.txt',
          hunks: [
            {
              oldStart: 1,
              oldLines: 3,
              newStart: 1,
              newLines: 3,
              lines: [
                { type: 'context', content: 'line 1', oldNum: 1, newNum: 1 },
                { type: 'deletion', content: 'line 2', oldNum: 2 },
                { type: 'addition', content: 'modified line 2', newNum: 2 },
                { type: 'context', content: 'line 3', oldNum: 3, newNum: 3 },
              ],
            },
          ],
          isNew: false,
          isDelete: false,
          isBinary: false,
        },
      ];

      const summary = calculateDiffSummary(diffs as any);

      expect(summary.filesModified).toBe(1);
      expect(summary.linesAdded).toBe(1);
      expect(summary.linesDeleted).toBe(1);
    });

    it('should calculate medium risk for 100-500 line changes', () => {
      const lines = Array.from({ length: 150 }, (_, i) => ({
        type: 'addition' as const,
        content: `line ${i}`,
        newNum: i + 1,
      }));

      const diffs = [
        {
          oldPath: 'file1.txt',
          newPath: 'file1.txt',
          hunks: [{ oldStart: 1, oldLines: 0, newStart: 1, newLines: 150, lines }],
          isNew: false,
          isDelete: false,
          isBinary: false,
        },
      ];

      const summary = calculateDiffSummary(diffs as any);
      expect(summary.riskLevel).toBe('medium');
    });

    it('should calculate high risk for >500 line changes', () => {
      const lines = Array.from({ length: 600 }, (_, i) => ({
        type: 'addition' as const,
        content: `line ${i}`,
        newNum: i + 1,
      }));

      const diffs = [
        {
          oldPath: 'file1.txt',
          newPath: 'file1.txt',
          hunks: [{ oldStart: 1, oldLines: 0, newStart: 1, newLines: 600, lines }],
          isNew: false,
          isDelete: false,
          isBinary: false,
        },
      ];

      const summary = calculateDiffSummary(diffs as any);
      expect(summary.riskLevel).toBe('high');
    });

    it('should calculate high risk for many deleted files', () => {
      const diffs = Array.from({ length: 6 }, (_, i) => ({
        oldPath: `file${i}.txt`,
        newPath: '/dev/null',
        hunks: [],
        isNew: false,
        isDelete: true,
        isBinary: false,
      }));

      const summary = calculateDiffSummary(diffs as any);
      expect(summary.riskLevel).toBe('high');
      expect(summary.filesDeleted).toBe(6);
    });
  });

  describe('generateCommentId', () => {
    it('should generate unique IDs', () => {
      const id1 = generateCommentId();
      const id2 = generateCommentId();

      expect(id1).toBeTruthy();
      expect(id2).toBeTruthy();
      expect(id1).not.toBe(id2);
    });

    it('should start with comment_ prefix', () => {
      const id = generateCommentId();
      expect(id).toMatch(/^comment_/);
    });
  });
});
