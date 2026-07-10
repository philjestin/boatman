import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RuntimeView } from './RuntimeView';
import {
  GetMemoryDocument,
  GetRuntimeRun,
  ListMemoryDocuments,
  ListRuntimeRuns,
} from '../../../wailsjs/go/main/App';

describe('RuntimeView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads runtime runs and renders selected run details', async () => {
    vi.mocked(ListRuntimeRuns).mockResolvedValue([
      {
        runId: 'run-1',
        provider: 'claude-cli',
        model: 'sonnet',
        role: 'planner',
        status: 'succeeded',
        workDir: '/repo',
        updatedAt: '2026-07-10T12:00:00Z',
        eventCount: 2,
        artifactCount: 1,
      },
    ]);
    vi.mocked(GetRuntimeRun).mockResolvedValue({
      metadata: {
        runId: 'run-1',
        provider: 'claude-cli',
        model: 'sonnet',
        role: 'planner',
        status: 'succeeded',
        workDir: '/repo',
        eventCount: 2,
        artifactCount: 1,
      },
      events: [
        {
          type: 'run.completed',
          status: 'succeeded',
          message: 'done',
          timestamp: '2026-07-10T12:00:00Z',
        },
      ],
      artifacts: [
        {
          kind: 'file',
          path: 'plan.md',
          message: 'plan written',
          eventCount: 1,
        },
      ],
    } as any);

    render(<RuntimeView projectPath="/repo" />);

    expect(await screen.findByText('run-1')).toBeInTheDocument();
    fireEvent.click(screen.getByText('run-1'));

    await waitFor(() => expect(GetRuntimeRun).toHaveBeenCalledWith('/repo', 'run-1'));
    expect(await screen.findByText('plan.md')).toBeInTheDocument();
    expect(screen.getByText('run.completed')).toBeInTheDocument();
  });

  it('loads memory documents and renders selected document body', async () => {
    vi.mocked(ListMemoryDocuments).mockResolvedValue([
      {
        id: 'domains/payments',
        scope: 'domain',
        title: 'Payments',
        sourceRunId: 'run-1',
        bodyPreview: 'Use gateway helpers',
        updatedAt: '2026-07-10T12:00:00Z',
      },
    ]);
    vi.mocked(GetMemoryDocument).mockResolvedValue({
      id: 'domains/payments',
      scope: 'domain',
      title: 'Payments',
      sourceRunId: 'run-1',
      path: '/repo/.boatman/memory/domains/payments.md',
      body: 'Use gateway helpers before adding new mocks.',
    });

    render(<RuntimeView projectPath="/repo" />);

    fireEvent.click(screen.getByRole('button', { name: /memory/i }));
    expect(await screen.findByText('Payments')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Payments'));

    await waitFor(() => expect(GetMemoryDocument).toHaveBeenCalledWith('/repo', 'domains/payments'));
    expect(await screen.findByText('Use gateway helpers before adding new mocks.')).toBeInTheDocument();
  });
});
