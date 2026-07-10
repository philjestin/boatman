import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock scrollIntoView for jsdom
Element.prototype.scrollIntoView = vi.fn();

// Mock Wails runtime bindings
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
}));

// Mock Go bindings - these would be injected by Wails at runtime
vi.mock('../../wailsjs/go/main/App', () => ({
  GetPreferences: vi.fn().mockResolvedValue({
    apiKey: '',
    approvalMode: 'suggest',
    defaultModel: 'claude-sonnet-4-20250514',
    theme: 'dark',
    notificationsEnabled: true,
    mcpServers: [],
    onboardingCompleted: true,
  }),
  SetPreferences: vi.fn().mockResolvedValue(undefined),
  IsOnboardingCompleted: vi.fn().mockResolvedValue(true),
  CompleteOnboarding: vi.fn().mockResolvedValue(undefined),
  ListProjects: vi.fn().mockResolvedValue([]),
  ListAgentSessions: vi.fn().mockResolvedValue([]),
  CreateAgentSession: vi.fn().mockResolvedValue({ id: 'test-session', projectPath: '/test', status: 'idle', createdAt: new Date().toISOString() }),
  StartAgentSession: vi.fn().mockResolvedValue(undefined),
  StopAgentSession: vi.fn().mockResolvedValue(undefined),
  DeleteAgentSession: vi.fn().mockResolvedValue(undefined),
  SendAgentMessage: vi.fn().mockResolvedValue(undefined),
  GetAgentMessages: vi.fn().mockResolvedValue([]),
  GetAgentMessagesPaginated: vi.fn().mockResolvedValue({
    messages: [],
    total: 0,
    page: 1,
    pageSize: 50,
    hasMore: false
  }),
  GetAgentTasks: vi.fn().mockResolvedValue([]),
  SelectFolder: vi.fn().mockResolvedValue('/test/path'),
  OpenProject: vi.fn().mockResolvedValue({ id: 'test-project', name: 'Test', path: '/test/path', lastOpened: new Date().toISOString(), createdAt: new Date().toISOString() }),
  SearchSessions: vi.fn().mockResolvedValue([]),
  GetAllTags: vi.fn().mockResolvedValue([]),
  AddSessionTag: vi.fn().mockResolvedValue(undefined),
  RemoveSessionTag: vi.fn().mockResolvedValue(undefined),
  SetSessionFavorite: vi.fn().mockResolvedValue(undefined),
  CleanupOldSessions: vi.fn().mockResolvedValue(0),
  GetSessionStats: vi.fn().mockResolvedValue({
    total: 0,
    oldestDate: '',
    newestDate: ''
  }),
  IsGCloudInstalled: vi.fn().mockResolvedValue(true),
  IsGCloudAuthenticated: vi.fn().mockResolvedValue(true),
  GetGCloudAuthInfo: vi.fn().mockResolvedValue({
    account: 'test@example.com',
    project: 'my-project'
  }),
  GCloudLoginApplicationDefault: vi.fn().mockResolvedValue(undefined),
  GCloudGetAvailableProjects: vi.fn().mockResolvedValue(['my-project', 'new-project']),
  GCloudSetProject: vi.fn().mockResolvedValue(undefined),
  GCloudRevoke: vi.fn().mockResolvedValue(undefined),
  GetMCPServers: vi.fn().mockResolvedValue([]),
  GetMCPPresets: vi.fn().mockResolvedValue([]),
  GetIntegrationStatuses: vi.fn().mockResolvedValue([
    {
      name: 'linear',
      state: 'needs_configuration',
      message: 'integration is missing required configuration',
      missingEnv: ['LINEAR_API_KEY'],
      lastChecked: new Date().toISOString()
    },
    {
      name: 'slack',
      state: 'disabled',
      message: 'integration is disabled',
      lastChecked: new Date().toISOString()
    }
  ]),
  AddMCPServer: vi.fn().mockResolvedValue(undefined),
  RemoveMCPServer: vi.fn().mockResolvedValue(undefined),
  UpdateMCPServer: vi.fn().mockResolvedValue(undefined),
  ListRuntimeRuns: vi.fn().mockResolvedValue([]),
  GetRuntimeRun: vi.fn().mockResolvedValue(null),
  ListMemoryDocuments: vi.fn().mockResolvedValue([]),
  GetMemoryDocument: vi.fn().mockResolvedValue(null),
  ListProjectRoutines: vi.fn().mockResolvedValue([]),
  ListRoutines: vi.fn().mockResolvedValue([]),
  DryRunRoutine: vi.fn().mockResolvedValue(null),
  RunRoutine: vi.fn().mockResolvedValue(null),
}));
