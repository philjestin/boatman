import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type {
  AgentSession,
  Message,
  Task,
  Project,
  UserPreferences,
  SessionStatus,
} from '../types';

// =============================================================================
// Agent Slice
// =============================================================================

interface MessagePagination {
  page: number;
  hasMore: boolean;
  total: number;
}

interface AgentState {
  sessions: AgentSession[];
  activeSessionId: string | null;
  messagePagination: Map<string, MessagePagination>;
}

interface AgentActions {
  // Session management
  addSession: (session: AgentSession) => void;
  removeSession: (sessionId: string) => void;
  setActiveSession: (sessionId: string | null) => void;
  updateSessionStatus: (sessionId: string, status: SessionStatus) => void;

  // Messages
  addMessage: (sessionId: string, message: Message) => void;
  setMessages: (sessionId: string, messages: Message[]) => void;
  appendMessages: (sessionId: string, messages: Message[]) => void;

  // Pagination
  setMessagePagination: (sessionId: string, pagination: MessagePagination) => void;

  // Tasks
  updateTask: (sessionId: string, task: Task) => void;
  setTasks: (sessionId: string, tasks: Task[]) => void;
}

// =============================================================================
// Project Slice
// =============================================================================

interface ProjectState {
  projects: Project[];
  activeProjectId: string | null;
}

interface ProjectActions {
  setProjects: (projects: Project[]) => void;
  addProject: (project: Project) => void;
  removeProject: (projectId: string) => void;
  setActiveProject: (projectId: string | null) => void;
}

// =============================================================================
// UI Slice
// =============================================================================

interface UIState {
  sidebarOpen: boolean;
  settingsOpen: boolean;
  onboardingOpen: boolean;
  loading: {
    sessions: boolean;
    projects: boolean;
    messages: boolean;
  };
  error: string | null;
}

interface UIActions {
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;
  setOnboardingOpen: (open: boolean) => void;
  setLoading: (key: keyof UIState['loading'], loading: boolean) => void;
  setError: (error: string | null) => void;
}

// =============================================================================
// Preferences Slice
// =============================================================================

interface PreferencesState {
  preferences: UserPreferences | null;
}

interface PreferencesActions {
  setPreferences: (preferences: UserPreferences) => void;
  updatePreferences: (updates: Partial<UserPreferences>) => void;
}

// =============================================================================
// Combined Store
// =============================================================================

type StoreState = AgentState & ProjectState & UIState & PreferencesState;
type StoreActions = AgentActions & ProjectActions & UIActions & PreferencesActions;
type Store = StoreState & StoreActions;

const initialState: StoreState = {
  // Agent state
  sessions: [],
  activeSessionId: null,
  messagePagination: new Map(),

  // Project state
  projects: [],
  activeProjectId: null,

  // UI state
  sidebarOpen: true,
  settingsOpen: false,
  onboardingOpen: false,
  loading: {
    sessions: false,
    projects: false,
    messages: false,
  },
  error: null,

  // Preferences
  preferences: null,
};

export const useStore = create<Store>()(
  devtools(
    persist(
      (set, get) => ({
        ...initialState,

        // =============================================================================
        // Agent Actions
        // =============================================================================

        addSession: (session) =>
          set(
            (state) => ({
              sessions: [...state.sessions, session],
            }),
            false,
            'addSession'
          ),

        removeSession: (sessionId) =>
          set(
            (state) => ({
              sessions: state.sessions.filter((s) => s.id !== sessionId),
              activeSessionId:
                state.activeSessionId === sessionId ? null : state.activeSessionId,
            }),
            false,
            'removeSession'
          ),

        setActiveSession: (sessionId) =>
          set({ activeSessionId: sessionId }, false, 'setActiveSession'),

        updateSessionStatus: (sessionId, status) =>
          set(
            (state) => ({
              sessions: state.sessions.map((s) =>
                s.id === sessionId ? { ...s, status } : s
              ),
            }),
            false,
            'updateSessionStatus'
          ),

        addMessage: (sessionId, message) => {
          console.log('[STORE] addMessage called:', { sessionId, message });
          set(
            (state) => ({
              sessions: state.sessions.map((s) => {
                if (s.id !== sessionId) return s;

                // Check if message already exists (for streaming updates)
                const existingIndex = s.messages.findIndex((m) => m.id === message.id);
                if (existingIndex !== -1) {
                  // Update existing message
                  const newMessages = [...s.messages];
                  newMessages[existingIndex] = message;
                  return { ...s, messages: newMessages };
                } else {
                  // Add new message
                  return { ...s, messages: [...s.messages, message] };
                }
              }),
            }),
            false,
            'addMessage'
          );
        },

        setMessages: (sessionId, messages) =>
          set(
            (state) => ({
              sessions: state.sessions.map((s) =>
                s.id === sessionId ? { ...s, messages } : s
              ),
            }),
            false,
            'setMessages'
          ),

        appendMessages: (sessionId, messages) =>
          set(
            (state) => ({
              sessions: state.sessions.map((s) =>
                s.id === sessionId
                  ? { ...s, messages: [...messages, ...s.messages] }
                  : s
              ),
            }),
            false,
            'appendMessages'
          ),

        setMessagePagination: (sessionId, pagination) =>
          set(
            (state) => {
              const newPagination = new Map(state.messagePagination);
              newPagination.set(sessionId, pagination);
              return { messagePagination: newPagination };
            },
            false,
            'setMessagePagination'
          ),

        updateTask: (sessionId, task) =>
          set(
            (state) => ({
              sessions: state.sessions.map((s) => {
                if (s.id !== sessionId) return s;
                const taskIndex = s.tasks.findIndex((t) => t.id === task.id);
                if (taskIndex === -1) {
                  return { ...s, tasks: [...s.tasks, task] };
                }
                const newTasks = [...s.tasks];
                newTasks[taskIndex] = task;
                return { ...s, tasks: newTasks };
              }),
            }),
            false,
            'updateTask'
          ),

        setTasks: (sessionId, tasks) =>
          set(
            (state) => ({
              sessions: state.sessions.map((s) =>
                s.id === sessionId ? { ...s, tasks } : s
              ),
            }),
            false,
            'setTasks'
          ),

        // =============================================================================
        // Project Actions
        // =============================================================================

        setProjects: (projects) =>
          set({ projects }, false, 'setProjects'),

        addProject: (project) =>
          set(
            (state) => ({
              projects: [project, ...state.projects.filter((p) => p.id !== project.id)],
            }),
            false,
            'addProject'
          ),

        removeProject: (projectId) =>
          set(
            (state) => ({
              projects: state.projects.filter((p) => p.id !== projectId),
              activeProjectId:
                state.activeProjectId === projectId ? null : state.activeProjectId,
            }),
            false,
            'removeProject'
          ),

        setActiveProject: (projectId) =>
          set({ activeProjectId: projectId }, false, 'setActiveProject'),

        // =============================================================================
        // UI Actions
        // =============================================================================

        toggleSidebar: () =>
          set(
            (state) => ({ sidebarOpen: !state.sidebarOpen }),
            false,
            'toggleSidebar'
          ),

        setSidebarOpen: (open) =>
          set({ sidebarOpen: open }, false, 'setSidebarOpen'),

        setSettingsOpen: (open) =>
          set({ settingsOpen: open }, false, 'setSettingsOpen'),

        setOnboardingOpen: (open) =>
          set({ onboardingOpen: open }, false, 'setOnboardingOpen'),

        setLoading: (key, loading) =>
          set(
            (state) => ({
              loading: { ...state.loading, [key]: loading },
            }),
            false,
            'setLoading'
          ),

        setError: (error) =>
          set({ error }, false, 'setError'),

        // =============================================================================
        // Preferences Actions
        // =============================================================================

        setPreferences: (preferences) =>
          set({ preferences }, false, 'setPreferences'),

        updatePreferences: (updates) =>
          set(
            (state) => ({
              preferences: state.preferences
                ? { ...state.preferences, ...updates }
                : null,
            }),
            false,
            'updatePreferences'
          ),
      }),
      {
        name: 'boatman-store',
        partialize: (state) => ({
          // Only persist certain parts of state
          sidebarOpen: state.sidebarOpen,
          activeProjectId: state.activeProjectId,
        }),
      }
    ),
    { name: 'Boatman' }
  )
);

// =============================================================================
// Selectors
// =============================================================================

export const useActiveSession = () => {
  const { sessions, activeSessionId } = useStore();
  return sessions.find((s) => s.id === activeSessionId) ?? null;
};

export const useActiveProject = () => {
  const { projects, activeProjectId } = useStore();
  return projects.find((p) => p.id === activeProjectId) ?? null;
};

export const useSessionMessages = (sessionId: string) => {
  const { sessions } = useStore();
  const session = sessions.find((s) => s.id === sessionId);
  return session?.messages ?? [];
};

export const useSessionTasks = (sessionId: string) => {
  const { sessions } = useStore();
  const session = sessions.find((s) => s.id === sessionId);
  return session?.tasks ?? [];
};
