import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { RoutineView } from './RoutineView';
import {
  AuthenticateDatadogMCP,
  DryRunRoutine,
  ListProjectRoutines,
  ListRoutines,
  RunRoutine,
} from '../../../wailsjs/go/main/App';

const routine = {
  id: 'datadog-gql-slow-queries',
  name: 'Datadog GraphQL Slow Queries',
  description: 'Investigate slow GraphQL operations',
  schedule: '0 8 * * *',
  workflowTemplate: 'research',
  role: 'routine',
  profile: 'routine.datadog-gql-slow-queries',
  integrations: ['datadog'],
  parameters: [
    { name: 'graph_area', type: 'string', description: 'Graph area', required: true, default: '' },
    { name: 'top_n', type: 'integer', description: 'Top N', required: false, default: '20' },
    { name: 'lookback', type: 'duration', description: 'Lookback', required: false, default: '24h' },
    { name: 'environment', type: 'string', description: 'Environment', required: false, default: 'prod' },
    { name: 'service', type: 'string', description: 'Service', required: false, default: '' },
  ],
  output: { format: 'markdown', defaultPath: '.boatman/routines/datadog-gql-slow-queries' },
  metadata: { cadence: 'daily' },
};

describe('RoutineView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(ListProjectRoutines).mockResolvedValue([routine] as any);
    vi.mocked(ListRoutines).mockResolvedValue([routine] as any);
    vi.mocked(AuthenticateDatadogMCP).mockResolvedValue({
      mcpName: 'plugin:datadog:datadog-mcp',
      command: 'claude mcp login plugin:datadog:datadog-mcp',
      message: 'Datadog MCP auth opened in an interactive terminal. Complete the browser flow, then click Check.',
      interactive: true,
      launched: true,
    } as any);
    vi.mocked(DryRunRoutine).mockResolvedValue({
      routine,
      values: {
        graph_area: 'employer',
        top_n: '5',
        lookback: '12h',
        environment: 'prod',
      },
      request: {
        runId: 'run-1',
        provider: 'claude-cli',
        model: 'sonnet',
        role: 'routine',
        profile: 'routine.datadog-gql-slow-queries',
        workDir: '/repo',
        approvalPolicy: 'suggest',
        reasoningEffort: 'high',
        mcpServerLabels: ['datadog'],
        firstMessagePreview: 'Investigate the top 5 slowest GraphQL operations',
      },
      integrations: [
        {
          name: 'datadog',
          state: 'connected',
          message: 'integration connection handle is ready',
          missingEnv: [],
          lastChecked: new Date().toISOString(),
        },
      ],
      reportPath: '/repo/.boatman/routines/datadog-gql-slow-queries/run-1.md',
    } as any);
    vi.mocked(RunRoutine).mockResolvedValue({
      routineId: 'datadog-gql-slow-queries',
      runId: 'run-1',
      provider: 'claude-cli',
      model: 'sonnet',
      values: { graph_area: 'employer' },
      reportPath: '/repo/.boatman/routines/datadog-gql-slow-queries/run-1.md',
      integrations: [],
      report: '## Executive summary\n\np95 latency regressed.',
      usage: { inputTokens: 10, outputTokens: 20, totalCostUsd: 0.12 },
    } as any);
  });

  it('dry-runs with the selected routine parameters', async () => {
    const user = userEvent.setup();
    render(<RoutineView projectPath="/repo" />);

    await screen.findByRole('heading', { name: 'Datadog GraphQL Slow Queries' });
    await waitFor(() => {
      expect(ListProjectRoutines).toHaveBeenCalledWith('/repo');
    });
    await user.type(screen.getByLabelText(/Graph Area/i), 'employer');
    await user.clear(screen.getByLabelText(/Top N/i));
    await user.type(screen.getByLabelText(/Top N/i), '5');
    await user.clear(screen.getByLabelText(/Lookback/i));
    await user.type(screen.getByLabelText(/Lookback/i), '12h');
    await user.click(screen.getByRole('button', { name: /Check/i }));

    await waitFor(() => {
      expect(DryRunRoutine).toHaveBeenCalledWith(expect.objectContaining({
        routineId: 'datadog-gql-slow-queries',
        projectPath: '/repo',
        values: expect.objectContaining({
          graph_area: 'employer',
          top_n: '5',
          lookback: '12h',
          environment: 'prod',
        }),
      }));
    });
    expect(await screen.findByText('Dry Run')).toBeInTheDocument();
    expect(screen.getAllByText('datadog').length).toBeGreaterThan(0);
  });

  it('runs the routine and renders the report', async () => {
    const user = userEvent.setup();
    render(<RoutineView projectPath="/repo" />);

    await screen.findByRole('heading', { name: 'Datadog GraphQL Slow Queries' });
    await user.type(screen.getByLabelText(/Graph Area/i), 'employer');
    await user.click(screen.getByRole('button', { name: /Run Routine/i }));

    await waitFor(() => {
      expect(RunRoutine).toHaveBeenCalledWith(expect.objectContaining({
        routineId: 'datadog-gql-slow-queries',
        projectPath: '/repo',
      }));
    });
    expect(await screen.findByText('Report')).toBeInTheDocument();
    expect(screen.getByText(/p95 latency regressed/i)).toBeInTheDocument();
  });

  it('opens Datadog MCP authentication from the integration panel', async () => {
    const user = userEvent.setup();
    render(<RoutineView projectPath="/repo" />);

    await screen.findByRole('heading', { name: 'Datadog GraphQL Slow Queries' });
    await user.click(screen.getByRole('button', { name: /Authenticate/i }));

    await waitFor(() => {
      expect(AuthenticateDatadogMCP).toHaveBeenCalled();
    });
    expect(await screen.findByText(/auth opened in an interactive terminal/i)).toBeInTheDocument();
    expect(DryRunRoutine).not.toHaveBeenCalled();
  });
});
