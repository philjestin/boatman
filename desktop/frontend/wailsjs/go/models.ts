export namespace agent {
	
	export class AgentInfo {
	    agentId: string;
	    agentType: string;
	    parentAgentId?: string;
	    description?: string;
	    status?: string;
	    // Go type: time
	    completedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.agentType = source["agentType"];
	        this.parentAgentId = source["parentAgentId"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.completedAt = this.convertValues(source["completedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CostInfo {
	    inputTokens: number;
	    outputTokens: number;
	    totalCost: number;
	
	    static createFrom(source: any = {}) {
	        return new CostInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputTokens = source["inputTokens"];
	        this.outputTokens = source["outputTokens"];
	        this.totalCost = source["totalCost"];
	    }
	}
	export class ToolResult {
	    toolId: string;
	    content: string;
	    isError: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolId = source["toolId"];
	        this.content = source["content"];
	        this.isError = source["isError"];
	    }
	}
	export class ToolUse {
	    toolName: string;
	    toolId: string;
	    input: number[];
	
	    static createFrom(source: any = {}) {
	        return new ToolUse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolName = source["toolName"];
	        this.toolId = source["toolId"];
	        this.input = source["input"];
	    }
	}
	export class MessageMetadata {
	    toolUse?: ToolUse;
	    toolResult?: ToolResult;
	    costInfo?: CostInfo;
	    agent?: AgentInfo;
	
	    static createFrom(source: any = {}) {
	        return new MessageMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolUse = this.convertValues(source["toolUse"], ToolUse);
	        this.toolResult = this.convertValues(source["toolResult"], ToolResult);
	        this.costInfo = this.convertValues(source["costInfo"], CostInfo);
	        this.agent = this.convertValues(source["agent"], AgentInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Message {
	    id: string;
	    role: string;
	    content: string;
	    // Go type: time
	    timestamp: any;
	    metadata?: MessageMetadata;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.metadata = this.convertValues(source["metadata"], MessageMetadata);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Task {
	    id: string;
	    subject: string;
	    description: string;
	    status: string;
	    metadata?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.subject = source["subject"];
	        this.description = source["description"];
	        this.status = source["status"];
	        this.metadata = source["metadata"];
	    }
	}
	

}

export namespace config {
	
	export class MCPServer {
	    name: string;
	    command: string;
	    args: string[];
	    env?: Record<string, string>;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MCPServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.enabled = source["enabled"];
	    }
	}
	export class UserPreferences {
	    apiKey: string;
	    authMethod: string;
	    gcpProjectId?: string;
	    gcpRegion?: string;
	    approvalMode: string;
	    defaultModel: string;
	    theme: string;
	    notificationsEnabled: boolean;
	    mcpServers: MCPServer[];
	    onboardingCompleted: boolean;
	    maxMessagesPerSession: number;
	    archiveOldMessages: boolean;
	    maxSessionAgeDays: number;
	    maxTotalSessions: number;
	    autoCleanupSessions: boolean;
	    maxAgentsPerSession: number;
	    keepCompletedAgents: boolean;
	    datadogAPIKey?: string;
	    datadogAppKey?: string;
	    datadogSite?: string;
	    bugsnagAPIKey?: string;
	    oktaDomain?: string;
	    oktaClientID?: string;
	    oktaClientSecret?: string;
	    linearAPIKey?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserPreferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.authMethod = source["authMethod"];
	        this.gcpProjectId = source["gcpProjectId"];
	        this.gcpRegion = source["gcpRegion"];
	        this.approvalMode = source["approvalMode"];
	        this.defaultModel = source["defaultModel"];
	        this.theme = source["theme"];
	        this.notificationsEnabled = source["notificationsEnabled"];
	        this.mcpServers = this.convertValues(source["mcpServers"], MCPServer);
	        this.onboardingCompleted = source["onboardingCompleted"];
	        this.maxMessagesPerSession = source["maxMessagesPerSession"];
	        this.archiveOldMessages = source["archiveOldMessages"];
	        this.maxSessionAgeDays = source["maxSessionAgeDays"];
	        this.maxTotalSessions = source["maxTotalSessions"];
	        this.autoCleanupSessions = source["autoCleanupSessions"];
	        this.maxAgentsPerSession = source["maxAgentsPerSession"];
	        this.keepCompletedAgents = source["keepCompletedAgents"];
	        this.datadogAPIKey = source["datadogAPIKey"];
	        this.datadogAppKey = source["datadogAppKey"];
	        this.datadogSite = source["datadogSite"];
	        this.bugsnagAPIKey = source["bugsnagAPIKey"];
	        this.oktaDomain = source["oktaDomain"];
	        this.oktaClientID = source["oktaClientID"];
	        this.oktaClientSecret = source["oktaClientSecret"];
	        this.linearAPIKey = source["linearAPIKey"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace diff {
	
	export class DiffComment {
	    id: string;
	    lineNum: number;
	    hunkId?: string;
	    content: string;
	    timestamp: string;
	    author?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffComment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.lineNum = source["lineNum"];
	        this.hunkId = source["hunkId"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.author = source["author"];
	    }
	}
	export class Line {
	    type: string;
	    content: string;
	    oldNum?: number;
	    newNum?: number;
	
	    static createFrom(source: any = {}) {
	        return new Line(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	        this.oldNum = source["oldNum"];
	        this.newNum = source["newNum"];
	    }
	}
	export class Hunk {
	    id?: string;
	    oldStart: number;
	    oldLines: number;
	    newStart: number;
	    newLines: number;
	    lines: Line[];
	    approved?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Hunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.oldStart = source["oldStart"];
	        this.oldLines = source["oldLines"];
	        this.newStart = source["newStart"];
	        this.newLines = source["newLines"];
	        this.lines = this.convertValues(source["lines"], Line);
	        this.approved = source["approved"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileDiff {
	    oldPath: string;
	    newPath: string;
	    hunks: Hunk[];
	    isNew: boolean;
	    isDelete: boolean;
	    isBinary: boolean;
	    approved?: boolean;
	    comments?: DiffComment[];
	
	    static createFrom(source: any = {}) {
	        return new FileDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldPath = source["oldPath"];
	        this.newPath = source["newPath"];
	        this.hunks = this.convertValues(source["hunks"], Hunk);
	        this.isNew = source["isNew"];
	        this.isDelete = source["isDelete"];
	        this.isBinary = source["isBinary"];
	        this.approved = source["approved"];
	        this.comments = this.convertValues(source["comments"], DiffComment);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class SideBySideLine {
	    leftNum?: number;
	    leftContent?: string;
	    rightNum?: number;
	    rightContent?: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new SideBySideLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.leftNum = source["leftNum"];
	        this.leftContent = source["leftContent"];
	        this.rightNum = source["rightNum"];
	        this.rightContent = source["rightContent"];
	        this.type = source["type"];
	    }
	}

}

export namespace main {
	
	export class AgentSessionInfo {
	    id: string;
	    projectPath: string;
	    status: string;
	    createdAt: string;
	    tags?: string[];
	    isFavorite?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentSessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectPath = source["projectPath"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.tags = source["tags"];
	        this.isFavorite = source["isFavorite"];
	    }
	}
	export class GitStatus {
	    isRepo: boolean;
	    branch: string;
	    modified: string[];
	    added: string[];
	    deleted: string[];
	    untracked: string[];
	
	    static createFrom(source: any = {}) {
	        return new GitStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isRepo = source["isRepo"];
	        this.branch = source["branch"];
	        this.modified = source["modified"];
	        this.added = source["added"];
	        this.deleted = source["deleted"];
	        this.untracked = source["untracked"];
	    }
	}
	export class MessagePage {
	    messages: agent.Message[];
	    total: number;
	    page: number;
	    pageSize: number;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MessagePage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messages = this.convertValues(source["messages"], agent.Message);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.hasMore = source["hasMore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SearchSessionsRequest {
	    query: string;
	    tags: string[];
	    projectPath: string;
	    isFavorite?: boolean;
	    fromDate: string;
	    toDate: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchSessionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.tags = source["tags"];
	        this.projectPath = source["projectPath"];
	        this.isFavorite = source["isFavorite"];
	        this.fromDate = source["fromDate"];
	        this.toDate = source["toDate"];
	    }
	}
	export class SearchSessionsResponse {
	    sessionId: string;
	    projectPath: string;
	    createdAt: string;
	    updatedAt: string;
	    tags: string[];
	    isFavorite: boolean;
	    messageCount: number;
	    score: number;
	    matchReasons: string[];
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchSessionsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.projectPath = source["projectPath"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.tags = source["tags"];
	        this.isFavorite = source["isFavorite"];
	        this.messageCount = source["messageCount"];
	        this.score = source["score"];
	        this.matchReasons = source["matchReasons"];
	        this.status = source["status"];
	    }
	}

}

export namespace mcp {
	
	export class Server {
	    name: string;
	    description?: string;
	    command: string;
	    args?: string[];
	    env?: Record<string, string>;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace project {
	
	export class Project {
	    id: string;
	    name: string;
	    path: string;
	    description?: string;
	    // Go type: time
	    lastOpened: any;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.description = source["description"];
	        this.lastOpened = this.convertValues(source["lastOpened"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkspaceInfo {
	    path: string;
	    name: string;
	    isGitRepo: boolean;
	    hasPackage: boolean;
	    languages: string[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.isGitRepo = source["isGitRepo"];
	        this.hasPackage = source["hasPackage"];
	        this.languages = source["languages"];
	    }
	}

}

