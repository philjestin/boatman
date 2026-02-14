import { useEffect, useCallback } from 'react';
import { useStore } from '../store';
import type { AgentSession, Message, Task, SessionStatus, BoatmanModeEventPayload } from '../types';

// Import Wails bindings (will be generated)
import {
  CreateAgentSession,
  CreateFirefighterSession,
  CreateBoatmanModeSession,
  StartAgentSession,
  StopAgentSession,
  DeleteAgentSession,
  SendAgentMessage,
  ApproveAgentAction,
  RejectAgentAction,
  GetAgentMessages,
  GetAgentMessagesPaginated,
  GetAgentTasks,
  ListAgentSessions,
  StartFirefighterMonitoring,
  StopFirefighterMonitoring,
  IsFirefighterMonitoringActive,
  HandleBoatmanModeEvent,
  StreamBoatmanModeExecution,
} from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export function useAgent() {
  const {
    sessions,
    activeSessionId,
    messagePagination,
    addSession,
    removeSession,
    setActiveSession,
    updateSessionStatus,
    addMessage,
    setMessages,
    appendMessages,
    setMessagePagination,
    updateTask,
    setTasks,
    setLoading,
    setError,
  } = useStore();

  // Subscribe to agent events
  useEffect(() => {
    const messageHandler = (data: { sessionId: string; message: Message }) => {
      console.log('[FRONTEND] Received message event:', data);
      addMessage(data.sessionId, data.message);
    };

    const taskHandler = (data: { sessionId: string; task: Task }) => {
      console.log('[FRONTEND] Received task event:', data);
      updateTask(data.sessionId, data.task);
    };

    const statusHandler = (data: { sessionId: string; status: SessionStatus }) => {
      console.log('[FRONTEND] Received status event:', data);
      updateSessionStatus(data.sessionId, data.status);
    };

    const boatmanModeEventHandler = async (data: BoatmanModeEventPayload) => {
      console.log('[FRONTEND] Received boatmanmode event:', data);
      try {
        // Let backend handle the event and update tasks
        await HandleBoatmanModeEvent(data.sessionId, data.event.type, data.event);
      } catch (err) {
        console.error('Failed to handle boatmanmode event:', err);
      }
    };

    console.log('[FRONTEND] Subscribing to agent events...');
    EventsOn('agent:message', messageHandler);
    EventsOn('agent:task', taskHandler);
    EventsOn('agent:status', statusHandler);
    EventsOn('boatmanmode:event', boatmanModeEventHandler);

    return () => {
      console.log('[FRONTEND] Unsubscribing from agent events...');
      EventsOff('agent:message');
      EventsOff('agent:task');
      EventsOff('agent:status');
      EventsOff('boatmanmode:event');
    };
  }, [addMessage, updateTask, updateSessionStatus]);

  // Load existing sessions on mount
  useEffect(() => {
    const loadSessions = async () => {
      try {
        // Check if Wails bindings are available
        if (typeof ListAgentSessions !== 'function') {
          console.warn('Wails bindings not available yet');
          return;
        }

        setLoading('sessions', true);
        const sessionInfos = await ListAgentSessions();
        // Convert to full sessions (messages/tasks will be loaded on select)
        sessionInfos.forEach((info) => {
          addSession({
            id: info.id,
            projectPath: info.projectPath,
            status: info.status as SessionStatus,
            createdAt: info.createdAt,
            messages: [],
            tasks: [],
            tags: info.tags || [],
            isFavorite: info.isFavorite || false,
          });
        });
      } catch (err) {
        console.error('Failed to load sessions:', err);
        setError('Failed to load sessions');
      } finally {
        setLoading('sessions', false);
      }
    };

    loadSessions();
  }, []);

  // Create a new session
  const createSession = useCallback(async (projectPath: string): Promise<string | null> => {
    try {
      setLoading('sessions', true);
      const info = await CreateAgentSession(projectPath);

      const session: AgentSession = {
        id: info.id,
        projectPath: info.projectPath,
        status: info.status as SessionStatus,
        createdAt: info.createdAt,
        messages: [],
        tasks: [],
        tags: info.tags || [],
        isFavorite: info.isFavorite || false,
      };

      addSession(session);
      setActiveSession(session.id);

      // Start the session
      await StartAgentSession(session.id);

      return session.id;
    } catch (err) {
      setError('Failed to create session');
      return null;
    } finally {
      setLoading('sessions', false);
    }
  }, [addSession, setActiveSession, setLoading, setError]);

  // Create a firefighter session
  const createFirefighterSession = useCallback(async (projectPath: string, scope: string): Promise<string | null> => {
    try {
      setLoading('sessions', true);
      const info = await CreateFirefighterSession(projectPath, scope);

      const session: AgentSession = {
        id: info.id,
        projectPath: info.projectPath,
        status: info.status as SessionStatus,
        createdAt: info.createdAt,
        messages: [],
        tasks: [],
        tags: info.tags || [],
        isFavorite: info.isFavorite || false,
        mode: 'firefighter',
        modeConfig: { scope },
      };

      addSession(session);
      setActiveSession(session.id);

      // Start the session
      await StartAgentSession(session.id);

      return session.id;
    } catch (err) {
      setError('Failed to create firefighter session: ' + err);
      return null;
    } finally {
      setLoading('sessions', false);
    }
  }, [addSession, setActiveSession, setLoading, setError]);

  // Create a boatmanmode session
  // mode can be "ticket" or "prompt"
  const createBoatmanModeSession = useCallback(async (projectPath: string, input: string, mode: string, linearAPIKey: string): Promise<string | null> => {
    try {
      setLoading('sessions', true);
      const info = await CreateBoatmanModeSession(projectPath, input, mode);

      const session: AgentSession = {
        id: info.id,
        projectPath: info.projectPath,
        status: info.status as SessionStatus,
        createdAt: info.createdAt,
        messages: [],
        tasks: [],
        tags: info.tags || [],
        isFavorite: info.isFavorite || false,
        mode: 'boatmanmode',
        modeConfig: { input, mode },
      };

      addSession(session);
      setActiveSession(session.id);

      // Start the session
      await StartAgentSession(session.id);

      // Start streaming execution
      await StreamBoatmanModeExecution(session.id, input, mode, linearAPIKey, projectPath);

      return session.id;
    } catch (err) {
      setError('Failed to create boatmanmode session: ' + err);
      return null;
    } finally {
      setLoading('sessions', false);
    }
  }, [addSession, setActiveSession, setLoading, setError]);

  // Start a session
  const startSession = useCallback(async (sessionId: string) => {
    try {
      await StartAgentSession(sessionId);
      updateSessionStatus(sessionId, 'running');
    } catch (err) {
      setError('Failed to start session');
    }
  }, [updateSessionStatus, setError]);

  // Stop a session
  const stopSession = useCallback(async (sessionId: string) => {
    try {
      await StopAgentSession(sessionId);
      updateSessionStatus(sessionId, 'stopped');
    } catch (err) {
      setError('Failed to stop session');
    }
  }, [updateSessionStatus, setError]);

  // Delete a session
  const deleteSession = useCallback(async (sessionId: string) => {
    try {
      await DeleteAgentSession(sessionId);
      removeSession(sessionId);
    } catch (err) {
      setError('Failed to delete session');
    }
  }, [removeSession, setError]);

  // Send a message
  const sendMessage = useCallback(async (sessionId: string, content: string) => {
    try {
      setLoading('messages', true);
      await SendAgentMessage(sessionId, content);
    } catch (err) {
      setError('Failed to send message');
    } finally {
      setLoading('messages', false);
    }
  }, [setLoading, setError]);

  // Approve an action
  const approveAction = useCallback(async (sessionId: string, actionId: string) => {
    try {
      await ApproveAgentAction(sessionId, actionId);
    } catch (err) {
      setError('Failed to approve action');
    }
  }, [setError]);

  // Reject an action
  const rejectAction = useCallback(async (sessionId: string, actionId: string) => {
    try {
      await RejectAgentAction(sessionId, actionId);
    } catch (err) {
      setError('Failed to reject action');
    }
  }, [setError]);

  // Load messages for a session
  const loadMessages = useCallback(async (sessionId: string) => {
    try {
      const messages = await GetAgentMessages(sessionId);
      setMessages(sessionId, messages as unknown as Message[]);
    } catch (err) {
      console.error('Failed to load messages:', err);
    }
  }, [setMessages]);

  // Load messages with pagination
  const loadMessagesPaginated = useCallback(
    async (sessionId: string, page: number, pageSize: number = 50) => {
      try {
        setLoading('messages', true);
        const result = await GetAgentMessagesPaginated(sessionId, page, pageSize);

        // If it's the first page, replace messages; otherwise append
        if (page === 0) {
          setMessages(sessionId, result.messages as unknown as Message[]);
        } else {
          appendMessages(sessionId, result.messages as unknown as Message[]);
        }

        // Update pagination state
        setMessagePagination(sessionId, {
          page: result.page,
          hasMore: result.hasMore,
          total: result.total,
        });
      } catch (err) {
        console.error('Failed to load paginated messages:', err);
        setError('Failed to load messages');
      } finally {
        setLoading('messages', false);
      }
    },
    [setMessages, appendMessages, setMessagePagination, setLoading, setError]
  );

  // Load tasks for a session
  const loadTasks = useCallback(async (sessionId: string) => {
    try {
      const tasks = await GetAgentTasks(sessionId);
      setTasks(sessionId, tasks as unknown as Task[]);
    } catch (err) {
      console.error('Failed to load tasks:', err);
    }
  }, [setTasks]);

  // Select a session
  const selectSession = useCallback(async (sessionId: string) => {
    setActiveSession(sessionId);
    await Promise.all([loadMessages(sessionId), loadTasks(sessionId)]);
  }, [setActiveSession, loadMessages, loadTasks]);

  // Get active session
  const activeSession = sessions.find((s) => s.id === activeSessionId) ?? null;

  // Toggle firefighter monitoring
  const toggleFirefighterMonitoring = useCallback(async (sessionId: string, active: boolean) => {
    try {
      if (active) {
        await StartFirefighterMonitoring(sessionId);
      } else {
        await StopFirefighterMonitoring(sessionId);
      }
    } catch (err) {
      console.error('Failed to toggle firefighter monitoring:', err);
      setError('Failed to toggle monitoring');
    }
  }, [setError]);

  // Check if monitoring is active
  const isMonitoringActive = useCallback(async (sessionId: string): Promise<boolean> => {
    try {
      return await IsFirefighterMonitoringActive(sessionId);
    } catch (err) {
      console.error('Failed to check monitoring status:', err);
      return false;
    }
  }, []);

  return {
    sessions,
    activeSession,
    activeSessionId,
    messagePagination,
    createSession,
    createFirefighterSession,
    createBoatmanModeSession,
    startSession,
    stopSession,
    deleteSession,
    selectSession,
    sendMessage,
    approveAction,
    rejectAction,
    loadMessagesPaginated,
    toggleFirefighterMonitoring,
    isMonitoringActive,
  };
}
