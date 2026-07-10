import { useCallback, useState } from 'react';
import {
  GetMemoryDocument as GetMemoryDocumentBinding,
  GetRuntimeRun as GetRuntimeRunBinding,
  ListMemoryDocuments as ListMemoryDocumentsBinding,
  ListRuntimeRuns as ListRuntimeRunsBinding,
} from '../../wailsjs/go/main/App';
import type {
  MemoryDocumentDetail,
  MemoryDocumentSummary,
  RuntimeRunDetail,
  RuntimeRunSummary,
} from '../types';

export function useRuntimeInspector(projectPath?: string) {
  const [runs, setRuns] = useState<RuntimeRunSummary[]>([]);
  const [selectedRun, setSelectedRun] = useState<RuntimeRunDetail | null>(null);
  const [memoryDocuments, setMemoryDocuments] = useState<MemoryDocumentSummary[]>([]);
  const [selectedMemoryDocument, setSelectedMemoryDocument] = useState<MemoryDocumentDetail | null>(null);
  const [isLoadingRuns, setIsLoadingRuns] = useState(false);
  const [isLoadingMemory, setIsLoadingMemory] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const path = projectPath || '';

  const loadRuns = useCallback(async () => {
    try {
      setIsLoadingRuns(true);
      setError(null);
      const result = await ListRuntimeRunsBinding(path);
      setRuns(result || []);
      return result || [];
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load runtime runs');
      setRuns([]);
      return [];
    } finally {
      setIsLoadingRuns(false);
    }
  }, [path]);

  const loadRun = useCallback(async (runId: string) => {
    try {
      setIsLoadingRuns(true);
      setError(null);
      const result = await GetRuntimeRunBinding(path, runId);
      setSelectedRun(result);
      return result;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load runtime run');
      setSelectedRun(null);
      return null;
    } finally {
      setIsLoadingRuns(false);
    }
  }, [path]);

  const loadMemoryDocuments = useCallback(async () => {
    try {
      setIsLoadingMemory(true);
      setError(null);
      const result = await ListMemoryDocumentsBinding(path);
      setMemoryDocuments(result || []);
      return result || [];
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load memory documents');
      setMemoryDocuments([]);
      return [];
    } finally {
      setIsLoadingMemory(false);
    }
  }, [path]);

  const loadMemoryDocument = useCallback(async (id: string) => {
    try {
      setIsLoadingMemory(true);
      setError(null);
      const result = await GetMemoryDocumentBinding(path, id);
      setSelectedMemoryDocument(result);
      return result;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load memory document');
      setSelectedMemoryDocument(null);
      return null;
    } finally {
      setIsLoadingMemory(false);
    }
  }, [path]);

  return {
    runs,
    selectedRun,
    memoryDocuments,
    selectedMemoryDocument,
    isLoadingRuns,
    isLoadingMemory,
    error,
    loadRuns,
    loadRun,
    loadMemoryDocuments,
    loadMemoryDocument,
    setSelectedRun,
    setSelectedMemoryDocument,
  };
}
