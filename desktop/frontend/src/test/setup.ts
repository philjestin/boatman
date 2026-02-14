import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock scrollIntoView for jsdom
Element.prototype.scrollIntoView = vi.fn();

// Mock Wails runtime bindings
vi.mock('../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
}));

// Mock Go bindings - these would be injected by Wails at runtime
vi.mock('../wailsjs/go/main/App', () => ({
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
}));
