import { describe, it, expect, beforeEach, vi } from 'vitest';
import { act, render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RuntimeView } from './RuntimeView';
import {
  GetMemoryDocument,
  GetRuntimeRun,
  ListMemoryDocuments,
  ListRuntimeRuns,
} from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';

describe('RuntimeView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(EventsOn).mockReturnValue(vi.fn());
    vi.mocked(ListRuntimeRuns).mockResolvedValue([]);
    vi.mocked(GetRuntimeRun).mockResolvedValue(null as any);
    vi.mocked(ListMemoryDocuments).mockResolvedValue([]);
    vi.mocked(GetMemoryDocument).mockResolvedValue(null as any);
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
      request: {
        provider: 'claude-cli',
        role: 'planner',
        approvalPolicy: 'full_auto',
        reasoningEffort: 'high',
        messageCount: 1,
        toolNames: ['Read'],
        instructionsPreview: 'Plan carefully',
        firstMessagePreview: 'Build a runtime inspector',
      },
      events: [
        {
          type: 'run.completed',
          status: 'succeeded',
          message: 'done',
          timestamp: '2026-07-10T12:00:00Z',
        },
        {
          type: 'memory.loaded',
          status: 'completed',
          message: 'loaded memory',
          timestamp: '2026-07-10T12:00:01Z',
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

    expect((await screen.findAllByText('run-1')).length).toBeGreaterThan(0);
    await waitFor(() => expect(GetRuntimeRun).toHaveBeenCalledWith('/repo', 'run-1'));
    expect(await screen.findByText('plan.md')).toBeInTheDocument();
    expect(screen.getAllByText('run.completed').length).toBeGreaterThan(0);
    expect(screen.getByText('Plan carefully')).toBeInTheDocument();
    expect(screen.getByText('Read')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Filter event type'), { target: { value: 'memory.loaded' } });
    expect(screen.getByText('loaded memory')).toBeInTheDocument();
    expect(screen.queryByText('done')).not.toBeInTheDocument();
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

  it('refreshes and selects a routine runtime run when the backend emits an update', async () => {
    let runtimeHandler: ((event: any) => void | Promise<void>) | undefined;
    vi.mocked(EventsOn).mockImplementation((eventName, handler) => {
      if (eventName === 'runtime:run-updated') {
        runtimeHandler = handler;
      }
      return vi.fn();
    });
    vi.mocked(ListRuntimeRuns)
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        {
          runId: 'routine-run-1',
          provider: 'claude-cli',
          model: 'sonnet',
          role: 'routine',
          profile: 'routine.datadog-gql-slow-queries',
          status: 'succeeded',
          workDir: '/repo',
          updatedAt: '2026-07-10T12:00:00Z',
          eventCount: 4,
          artifactCount: 0,
        },
      ]);
    vi.mocked(GetRuntimeRun).mockResolvedValue({
      metadata: {
        runId: 'routine-run-1',
        provider: 'claude-cli',
        model: 'sonnet',
        role: 'routine',
        profile: 'routine.datadog-gql-slow-queries',
        status: 'succeeded',
        workDir: '/repo',
        eventCount: 4,
        artifactCount: 0,
      },
      request: {
        provider: 'claude-cli',
        role: 'routine',
        profile: 'routine.datadog-gql-slow-queries',
        approvalPolicy: 'suggest',
        reasoningEffort: 'high',
        messageCount: 1,
        toolNames: [],
        instructionsPreview: 'Investigate slow queries',
        firstMessagePreview: 'Find the slowest GraphQL queries',
      },
      events: [
        {
          type: 'run.completed',
          status: 'succeeded',
          message: 'done',
          timestamp: '2026-07-10T12:00:00Z',
        },
      ],
      artifacts: [],
    } as any);

    render(<RuntimeView projectPath="/repo" />);

    await waitFor(() => expect(runtimeHandler).toBeDefined());
    await act(async () => {
      await runtimeHandler?.({
        source: 'routine',
        projectPath: '/repo',
        runId: 'routine-run-1',
        status: 'completed',
      });
    });

    await waitFor(() => expect(GetRuntimeRun).toHaveBeenCalledWith('/repo', 'routine-run-1'));
    expect((await screen.findAllByText('routine-run-1')).length).toBeGreaterThan(0);
    expect(screen.getByText('Find the slowest GraphQL queries')).toBeInTheDocument();
  });
});
