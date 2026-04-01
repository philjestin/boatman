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
  status?: 'active' | 'completed';
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
  model?: string;
  reasoningEffort?: string;
}

export const MODEL_OPTIONS = [
  { value: 'sonnet', label: 'Claude Sonnet 4' },
  { value: 'opus', label: 'Claude Opus 4' },
  { value: 'haiku', label: 'Claude Haiku 3.5' },
] as const;

export const REASONING_EFFORT_OPTIONS = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
] as const;

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

  // Slack monitoring settings
  slackAlertChannels?: string;
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

export interface Incident {
  id: string;
  source: 'linear' | 'bugsnag' | 'datadog' | 'slack';
  severity: 'urgent' | 'high' | 'medium' | 'low';
  title: string;
  description: string;
  status: 'new' | 'investigating' | 'fixing' | 'testing' | 'resolved' | 'failed';
  firstSeen: string;
  lastUpdated: string;
  messageIds: string[];
  linearId?: string;
  slackThread?: string;
  url?: string;
  prNumber?: string;
}

// =============================================================================
// Triage Types
// =============================================================================

export type TriageCategory = 'AI_DEFINITE' | 'AI_LIKELY' | 'HUMAN_REVIEW_REQUIRED' | 'HUMAN_ONLY';

export interface TriageRubricScores {
  clarity: number;
  codeLocality: number;
  patternMatch: number;
  validationStrength: number;
  dependencyRisk: number;
  productAmbiguity: number;
  blastRadius: number;
}

export interface TriageGateResult {
  gate: string;
  passed: boolean;
  reason?: string;
}

export interface TriageSignals {
  mentionsFiles: string[];
  domains: string[];
  labels: string[];
  dependencies: string[];
  acceptanceCriteriaPresent: boolean;
  acceptanceCriteriaExplicit: boolean;
  hasDesignSpec: boolean;
  commentCount: number;
  teamKey: string;
  projectName: string;
}

export interface TriageNormalizedTicket {
  ticketId: string;
  title: string;
  summary: string;
  description: string;
  ingestedAt: string;
  staleAfter: string;
  signals: TriageSignals;
}

export interface TriageClassification {
  ticketId: string;
  category: TriageCategory;
  rubric: TriageRubricScores;
  uncertainAxes: string[];
  reasons: string[];
  hardStops: string[] | null;
  gateResults: TriageGateResult[];
}

export interface TriageCostCeiling {
  maxTokensPerTicket: number;
  maxAgentMinutesPerTicket: number;
}

export interface TriageCluster {
  clusterId: string;
  tickets: string[];
  repoAreas: string[];
  rationale: string;
}

export interface TriageContextDoc {
  clusterId: string;
  rationale: string;
  tickets: string[];
  repoAreas: string[];
  knownPatterns: string[];
  validationPlan: string[];
  risks: string[];
  costCeiling: TriageCostCeiling;
}

export interface TriageStats {
  totalTickets: number;
  aiDefiniteCount: number;
  aiLikelyCount: number;
  humanReviewCount: number;
  humanOnlyCount: number;
  clusterCount: number;
  totalTokensUsed: number;
  totalCostUsd: number;
}

export interface TriageResult {
  tickets: TriageNormalizedTicket[];
  classifications: TriageClassification[];
  clusters: TriageCluster[];
  contextDocs: TriageContextDoc[];
  stats: TriageStats;
  plans?: PlanResult[];
  planStats?: PlanStats;
}

// =============================================================================
// Plan Types (Stage 4)
// =============================================================================

export interface TicketPlan {
  ticketId: string;
  approach: string;
  candidateFiles: string[];
  newFiles: string[];
  deletedFiles: string[];
  validation: string[];
  rollback: string;
  stopConditions: string[];
  uncertainties: string[];
}

export interface PlanGateResult {
  gate: string;
  passed: boolean;
  reason?: string;
}

export interface PlanValidation {
  passed: boolean;
  gateResults: PlanGateResult[];
  validatedFiles: string[];
  missingFiles: string[];
  outOfScopeFiles: string[];
}

export interface PlanResult {
  ticketId: string;
  plan: TicketPlan | null;
  validation: PlanValidation | null;
  usage?: { input_tokens: number; output_tokens: number; total_cost_usd: number };
  error?: string;
}

export interface PlanStats {
  total: number;
  passed: number;
  failed: number;
  totalTokensUsed: number;
  totalCostUsd: number;
}

export interface TriageOptions {
  teams: string[];
  states: string[];
  limit: number;
  ticketIds: string[];
  postComments: boolean;
  dryRun: boolean;
  outputDir: string;
  concurrency: number;
  generatePlans: boolean;
  repoPath: string;
}

export interface TriageEvent {
  type: string;
  id?: string;
  name?: string;
  status?: string;
  message?: string;
  data?: Record<string, any>;
}

export interface TriageEventPayload {
  sessionId: string;
  event: TriageEvent;
}

// =============================================================================
// Harness Types
// =============================================================================

export type LLMProvider = 'claude' | 'openai' | 'ollama' | 'generic';
export type ProjectLanguage = 'go' | 'typescript' | 'python' | 'ruby' | 'generic';

export interface ScaffoldRequest {
  projectName: string;
  outputDir: string;
  provider: LLMProvider;
  projectLang: ProjectLanguage;
  includePlanner: boolean;
  includeTester: boolean;
  includeCostTracking: boolean;
  maxIterations: number;
  reviewCriteria: string;
}

export interface ScaffoldResponse {
  outputDir: string;
  filesCreated: string[];
}

export interface HarnessInfo {
  name: string;
  path: string;
  hasGoMod: boolean;
  hasMain: boolean;
}

export interface RunRequest {
  harnessPath: string;
  workDir: string;
  taskTitle: string;
  taskDescription: string;
  envVars: Record<string, string>;
}

export type HarnessRunStatus = 'idle' | 'running' | 'completed' | 'error';

export interface HarnessRunState {
  status: HarnessRunStatus;
  runId: string | null;
  output: string[];
  error: string | null;
}

// =============================================================================
// Brain Types
// =============================================================================

export interface BrainEntry {
  id: string;
  name: string;
  description: string;
  confidence: number;
  version: number;
  lastUpdated: string;
  keywords: string[];
  entities: string[];
  filePatterns: string[];
}

export interface BrainDetail extends BrainEntry {
  sections: BrainSection[];
  references: BrainReference[];
}

export interface BrainSection {
  title: string;
  content: string;
}

export interface BrainReference {
  path: string;
  description: string;
  checksum: string;
}

export interface BrainValidationResult {
  valid: boolean;
  errors: string[];
  stale: StaleRefResult[];
}

export interface StaleRefResult {
  path: string;
  reason: string;
}

export interface SignalEntry {
  type: string;
  domain: string;
  details: string;
  filePaths: string[];
  count: number;
  firstSeen: string;
  lastSeen: string;
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
