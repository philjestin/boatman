import { useState, useEffect, useCallback } from 'react';
import type { HarnessInfo, ScaffoldRequest, ScaffoldResponse, RunRequest, HarnessRunState } from '../types';
import {
  ScaffoldHarness,
  ListHarnesses as ListHarnessesBinding,
  RunHarness as RunHarnessBinding,
  StopHarness as StopHarnessBinding,
  SelectHarnessFolder,
} from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export function useHarness() {
  const [harnesses, setHarnesses] = useState<HarnessInfo[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [runState, setRunState] = useState<HarnessRunState>({
    status: 'idle',
    runId: null,
    output: [],
    error: null,
  });

  // Subscribe to harness events
  useEffect(() => {
    const outputHandler = (data: { runId: string; line: string; stream: string }) => {
      setRunState((prev) => {
        if (prev.runId !== data.runId) return prev;
        return { ...prev, output: [...prev.output, data.line] };
      });
    };

    const errorHandler = (data: { runId: string; error: string }) => {
      setRunState((prev) => {
        if (prev.runId !== data.runId) return prev;
        return { ...prev, status: 'error', error: data.error };
      });
    };

    const completeHandler = (data: { runId: string }) => {
      setRunState((prev) => {
        if (prev.runId !== data.runId) return prev;
        return { ...prev, status: 'completed' };
      });
    };

    const startedHandler = (data: { runId: string }) => {
      setRunState((prev) => {
        if (prev.runId !== data.runId) return prev;
        return { ...prev, status: 'running' };
      });
    };

    EventsOn('harness:output', outputHandler);
    EventsOn('harness:error', errorHandler);
    EventsOn('harness:complete', completeHandler);
    EventsOn('harness:started', startedHandler);

    return () => {
      EventsOff('harness:output');
      EventsOff('harness:error');
      EventsOff('harness:complete');
      EventsOff('harness:started');
    };
  }, []);

  const loadHarnesses = useCallback(async () => {
    try {
      setIsLoading(true);
      const result = await ListHarnessesBinding();
      setHarnesses(result || []);
    } catch (err) {
      console.error('Failed to load harnesses:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const scaffoldHarness = useCallback(async (req: ScaffoldRequest): Promise<ScaffoldResponse | null> => {
    try {
      setIsLoading(true);
      const result = await ScaffoldHarness(req);
      // Refresh list after scaffolding
      await loadHarnesses();
      return result;
    } catch (err) {
      console.error('Failed to scaffold harness:', err);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [loadHarnesses]);

  const runHarness = useCallback(async (req: RunRequest) => {
    const runId = crypto.randomUUID();
    setRunState({
      status: 'running',
      runId,
      output: [],
      error: null,
    });

    try {
      await RunHarnessBinding(runId, req);
    } catch (err) {
      setRunState((prev) => ({
        ...prev,
        status: 'error',
        error: String(err),
      }));
    }
  }, []);

  const stopHarness = useCallback(async () => {
    if (!runState.runId) return;
    try {
      await StopHarnessBinding(runState.runId);
    } catch (err) {
      console.error('Failed to stop harness:', err);
    }
  }, [runState.runId]);

  const resetRunState = useCallback(() => {
    setRunState({
      status: 'idle',
      runId: null,
      output: [],
      error: null,
    });
  }, []);

  const selectFolder = useCallback(async (): Promise<string> => {
    try {
      return await SelectHarnessFolder();
    } catch (err) {
      console.error('Failed to select folder:', err);
      return '';
    }
  }, []);

  return {
    harnesses,
    isLoading,
    runState,
    loadHarnesses,
    scaffoldHarness,
    runHarness,
    stopHarness,
    resetRunState,
    selectFolder,
  };
}
