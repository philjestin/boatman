import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DiffSummaryCard } from './DiffSummaryCard';
import type { DiffSummary } from '../../types';

describe('DiffSummaryCard', () => {
  it('should render all file counts', () => {
    const summary: DiffSummary = {
      totalFiles: 10,
      filesAdded: 3,
      filesDeleted: 2,
      filesModified: 5,
      linesAdded: 100,
      linesDeleted: 50,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('3')).toBeInTheDocument(); // files added
    expect(screen.getByText('5')).toBeInTheDocument(); // files modified
    expect(screen.getByText('2')).toBeInTheDocument(); // files deleted
  });

  it('should render line counts', () => {
    const summary: DiffSummary = {
      totalFiles: 1,
      filesAdded: 0,
      filesDeleted: 0,
      filesModified: 1,
      linesAdded: 100,
      linesDeleted: 50,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('+100')).toBeInTheDocument();
    expect(screen.getByText('-50')).toBeInTheDocument();
  });

  it('should display low risk level correctly', () => {
    const summary: DiffSummary = {
      totalFiles: 1,
      filesAdded: 1,
      filesDeleted: 0,
      filesModified: 0,
      linesAdded: 10,
      linesDeleted: 0,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('Low Risk')).toBeInTheDocument();
  });

  it('should display medium risk level correctly', () => {
    const summary: DiffSummary = {
      totalFiles: 1,
      filesAdded: 0,
      filesDeleted: 0,
      filesModified: 1,
      linesAdded: 200,
      linesDeleted: 100,
      riskLevel: 'medium',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('Medium Risk')).toBeInTheDocument();
  });

  it('should display high risk level correctly', () => {
    const summary: DiffSummary = {
      totalFiles: 10,
      filesAdded: 0,
      filesDeleted: 5,
      filesModified: 5,
      linesAdded: 600,
      linesDeleted: 400,
      riskLevel: 'high',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('High Risk')).toBeInTheDocument();
  });

  it('should render summary title', () => {
    const summary: DiffSummary = {
      totalFiles: 0,
      filesAdded: 0,
      filesDeleted: 0,
      filesModified: 0,
      linesAdded: 0,
      linesDeleted: 0,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('Change Summary')).toBeInTheDocument();
  });

  it('should render file type labels', () => {
    const summary: DiffSummary = {
      totalFiles: 3,
      filesAdded: 1,
      filesDeleted: 1,
      filesModified: 1,
      linesAdded: 10,
      linesDeleted: 5,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('Added')).toBeInTheDocument();
    expect(screen.getByText('Modified')).toBeInTheDocument();
    expect(screen.getByText('Deleted')).toBeInTheDocument();
  });

  it('should render Lines Changed label', () => {
    const summary: DiffSummary = {
      totalFiles: 1,
      filesAdded: 0,
      filesDeleted: 0,
      filesModified: 1,
      linesAdded: 50,
      linesDeleted: 25,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    expect(screen.getByText('Lines Changed')).toBeInTheDocument();
  });

  it('should handle zero values', () => {
    const summary: DiffSummary = {
      totalFiles: 0,
      filesAdded: 0,
      filesDeleted: 0,
      filesModified: 0,
      linesAdded: 0,
      linesDeleted: 0,
      riskLevel: 'low',
    };

    render(<DiffSummaryCard summary={summary} />);

    const zeros = screen.getAllByText('0');
    expect(zeros.length).toBeGreaterThan(0);
  });
});
