// =============================================================================
// Agent Types
// =============================================================================

export type SessionStatus = 'idle' | 'running' | 'waiting' | 'error' | 'stopped';

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  metadata?: MessageMetadata;
}

export interface AgentInfo {
  agentId: string;
  agentType: string; // "main", "task", "explore", etc.
  parentAgentId?: string;
  description?: string;
}

export interface MessageMetadata {
  toolUse?: ToolUse;
  toolResult?: ToolResult;
  costInfo?: CostInfo;
  agent?: AgentInfo;
}

export interface ToolUse {
  toolName: string;
  toolId: string;
  input: unknown;
}

export interface ToolResult {
  toolId: string;
  content: string;
  isError: boolean;
}

export interface CostInfo {
  inputTokens: number;
  outputTokens: number;
  totalCost: number;
}

export interface Task {
  id: string;
  subject: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed';
  metadata?: Record<string, any>;
}

export interface AgentSession {
  id: string;
  projectPath: string;
  status: SessionStatus;
  createdAt: string;
  messages: Message[];
  tasks: Task[];
  tags?: string[];
  isFavorite?: boolean;
  mode?: string;
  modeConfig?: Record<string, any>;
}

// =============================================================================
// Project Types
// =============================================================================

export interface Project {
  id: string;
  name: string;
  path: string;
  description?: string;
  lastOpened: string;
  createdAt: string;
}

export interface WorkspaceInfo {
  path: string;
  name: string;
  isGitRepo: boolean;
  hasPackage: boolean;
  languages: string[];
}

// =============================================================================
// Git Types
// =============================================================================

export interface GitStatus {
  isRepo: boolean;
  branch: string;
  modified: string[];
  added: string[];
  deleted: string[];
  untracked: string[];
}

// =============================================================================
// Diff Types
// =============================================================================

export type LineType = 'context' | 'addition' | 'deletion';

export interface DiffLine {
  type: LineType;
  content: string;
  oldNum?: number;
  newNum?: number;
}

export interface DiffHunk {
  oldStart: number;
  oldLines: number;
  newStart: number;
  newLines: number;
  lines: DiffLine[];
  id?: string;
  approved?: boolean;
}

export interface FileDiff {
  oldPath: string;
  newPath: string;
  hunks: DiffHunk[];
  isNew: boolean;
  isDelete: boolean;
  isBinary: boolean;
  approved?: boolean;
  comments?: DiffComment[];
}

export interface SideBySideLine {
  leftNum?: number;
  leftContent?: string;
  rightNum?: number;
  rightContent?: string;
  type: 'context' | 'added' | 'deleted' | 'modified';
}

// Diff comment types
export interface DiffComment {
  id: string;
  lineNum: number;
  hunkId?: string;
  content: string;
  timestamp: string;
  author?: string;
}

export interface DiffSummary {
  totalFiles: number;
  filesAdded: number;
  filesDeleted: number;
  filesModified: number;
  linesAdded: number;
  linesDeleted: number;
  riskLevel: 'low' | 'medium' | 'high';
}

export interface HunkApprovalState {
  [fileKey: string]: {
    [hunkId: string]: boolean;
  };
}

export interface FileApprovalState {
  [fileKey: string]: boolean;
}

// =============================================================================
// Configuration Types
// =============================================================================

export type ApprovalMode = 'suggest' | 'auto-edit' | 'full-auto';
export type Theme = 'dark' | 'light';
export type AuthMethod = 'anthropic-api' | 'google-cloud';

export interface MCPServer {
  name: string;
  description?: string;
  command: string;
  args?: string[];
  env?: Record<string, string>;
  enabled: boolean;
}

export interface UserPreferences {
  apiKey: string;
  authMethod: AuthMethod;
  gcpProjectId?: string;
  gcpRegion?: string;
  approvalMode: ApprovalMode;
  defaultModel: string;
  theme: Theme;
  notificationsEnabled: boolean;
  mcpServers: MCPServer[];
  onboardingCompleted: boolean;

  // Memory management settings
  maxMessagesPerSession?: number;
  archiveOldMessages?: boolean;
  maxSessionAgeDays?: number;
  maxTotalSessions?: number;
  autoCleanupSessions?: boolean;
  maxAgentsPerSession?: number;
  keepCompletedAgents?: boolean;

  // Firefighter/Observability settings
  datadogAPIKey?: string;
  datadogAppKey?: string;
  datadogSite?: string;
  bugsnagAPIKey?: string;

  // Okta OAuth settings
  oktaDomain?: string;
  oktaClientID?: string;
  oktaClientSecret?: string;

  // Linear settings
  linearAPIKey?: string;
}

export interface ProjectPreferences {
  projectPath: string;
  approvalMode?: ApprovalMode;
  model?: string;
}

// =============================================================================
// UI State Types
// =============================================================================

export interface AppState {
  // Sessions
  sessions: AgentSession[];
  activeSessionId: string | null;

  // Projects
  projects: Project[];
  activeProjectId: string | null;

  // UI State
  sidebarOpen: boolean;
  settingsOpen: boolean;
  onboardingOpen: boolean;

  // Preferences
  preferences: UserPreferences | null;

  // Loading states
  loading: {
    sessions: boolean;
    projects: boolean;
    messages: boolean;
  };

  // Error state
  error: string | null;
}

// =============================================================================
// Event Types (from Wails)
// =============================================================================

export interface AgentMessageEvent {
  sessionId: string;
  message: Message;
}

export interface AgentTaskEvent {
  sessionId: string;
  task: Task;
}

export interface AgentStatusEvent {
  sessionId: string;
  status: SessionStatus;
}

export interface BoatmanModeEvent {
  type: string;
  id?: string;
  name?: string;
  description?: string;
  status?: string;
  message?: string;
  data?: Record<string, any>;
}

export interface BoatmanModeEventPayload {
  sessionId: string;
  event: BoatmanModeEvent;
}

export interface LinearTicket {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  priority?: number;
  state?: string;
  labels?: string[];
}

// =============================================================================
// Component Props Types
// =============================================================================

export interface ChatViewProps {
  sessionId: string;
  messages: Message[];
  onSendMessage: (content: string) => void;
  isLoading?: boolean;
}

export interface DiffViewProps {
  diff: FileDiff;
  viewMode: 'unified' | 'split';
  onAccept?: () => void;
  onReject?: () => void;
}

export interface TaskListProps {
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
}

export interface ApprovalBarProps {
  visible: boolean;
  onApprove: () => void;
  onReject: () => void;
  actionDescription?: string;
}

export interface SidebarProps {
  projects: Project[];
  sessions: AgentSession[];
  activeSessionId: string | null;
  onSessionSelect: (sessionId: string) => void;
  onProjectSelect: (projectId: string) => void;
  onNewSession: () => void;
  onOpenProject: () => void;
}
