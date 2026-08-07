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

export namespace agentruntime {

	export class ModelProfile {
	    plan?: string;
	    implementation?: string;
	    skills?: string;

	    static createFrom(source: any = {}) {
	        return new ModelProfile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plan = source["plan"];
	        this.implementation = source["implementation"];
	        this.skills = source["skills"];
	    }
	}
	export class Usage {
	    inputTokens?: number;
	    outputTokens?: number;
	    cacheReadTokens?: number;
	    cacheWriteTokens?: number;
	    totalCostUsd?: number;

	    static createFrom(source: any = {}) {
	        return new Usage(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputTokens = source["inputTokens"];
	        this.outputTokens = source["outputTokens"];
	        this.cacheReadTokens = source["cacheReadTokens"];
	        this.cacheWriteTokens = source["cacheWriteTokens"];
	        this.totalCostUsd = source["totalCostUsd"];
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
	    slackAlertChannels?: string;

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
	        this.slackAlertChannels = source["slackAlertChannels"];
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

export namespace frontendqueue {

	export class Batch {
	    id: string;
	    name: string;
	    ticketIds: string[];
	    targetKeys: string[];
	    parallelism: number;
	    rationale: string;

	    static createFrom(source: any = {}) {
	        return new Batch(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ticketIds = source["ticketIds"];
	        this.targetKeys = source["targetKeys"];
	        this.parallelism = source["parallelism"];
	        this.rationale = source["rationale"];
	    }
	}
	export class ValidationStep {
	    kind: string;
	    command?: string;
	    description: string;
	    required: boolean;

	    static createFrom(source: any = {}) {
	        return new ValidationStep(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.command = source["command"];
	        this.description = source["description"];
	        this.required = source["required"];
	    }
	}
	export class TargetHint {
	    kind: string;
	    value: string;
	    confidence: number;

	    static createFrom(source: any = {}) {
	        return new TargetHint(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.value = source["value"];
	        this.confidence = source["confidence"];
	    }
	}
	export class Ticket {
	    id: string;
	    title: string;
	    description?: string;
	    url?: string;
	    branchName?: string;
	    status?: string;
	    statusType?: string;
	    team?: string;
	    project?: string;
	    parentId?: string;
	    labels?: string[];
	    estimate?: number;
	    priority?: number;
	    // Go type: time
	    createdAt?: any;
	    // Go type: time
	    updatedAt?: any;

	    static createFrom(source: any = {}) {
	        return new Ticket(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.url = source["url"];
	        this.branchName = source["branchName"];
	        this.status = source["status"];
	        this.statusType = source["statusType"];
	        this.team = source["team"];
	        this.project = source["project"];
	        this.parentId = source["parentId"];
	        this.labels = source["labels"];
	        this.estimate = source["estimate"];
	        this.priority = source["priority"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class TicketPlan {
	    ticket: Ticket;
	    type: string;
	    policy: string;
	    confidence: number;
	    riskScore: number;
	    targetKey: string;
	    targetHints?: TargetHint[];
	    figmaRefs?: string[];
	    validation?: ValidationStep[];
	    reasons?: string[];
	    blockers?: string[];
	    workerPrompt: string;
	    passKCandidates?: number;

	    static createFrom(source: any = {}) {
	        return new TicketPlan(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ticket = this.convertValues(source["ticket"], Ticket);
	        this.type = source["type"];
	        this.policy = source["policy"];
	        this.confidence = source["confidence"];
	        this.riskScore = source["riskScore"];
	        this.targetKey = source["targetKey"];
	        this.targetHints = this.convertValues(source["targetHints"], TargetHint);
	        this.figmaRefs = source["figmaRefs"];
	        this.validation = this.convertValues(source["validation"], ValidationStep);
	        this.reasons = source["reasons"];
	        this.blockers = source["blockers"];
	        this.workerPrompt = source["workerPrompt"];
	        this.passKCandidates = source["passKCandidates"];
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
	export class Stats {
	    totalTickets: number;
	    autoPrCount: number;
	    draftPrCount: number;
	    planOnlyCount: number;
	    blockedCount: number;
	    batchCount: number;
	    maxParallel: number;
	    byType: Record<string, number>;
	    byTarget: Record<string, number>;
	    eligibleCount: number;
	    passKCandidateCount: number;

	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalTickets = source["totalTickets"];
	        this.autoPrCount = source["autoPrCount"];
	        this.draftPrCount = source["draftPrCount"];
	        this.planOnlyCount = source["planOnlyCount"];
	        this.blockedCount = source["blockedCount"];
	        this.batchCount = source["batchCount"];
	        this.maxParallel = source["maxParallel"];
	        this.byType = source["byType"];
	        this.byTarget = source["byTarget"];
	        this.eligibleCount = source["eligibleCount"];
	        this.passKCandidateCount = source["passKCandidateCount"];
	    }
	}
	export class PlanOptions {
	    projectPath?: string;
	    linearProject?: string;
	    maxParallel?: number;
	    allowedPolicies?: string[];
	    includePlanOnly?: boolean;
	    defaultBaseBranch?: string;

	    static createFrom(source: any = {}) {
	        return new PlanOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.linearProject = source["linearProject"];
	        this.maxParallel = source["maxParallel"];
	        this.allowedPolicies = source["allowedPolicies"];
	        this.includePlanOnly = source["includePlanOnly"];
	        this.defaultBaseBranch = source["defaultBaseBranch"];
	    }
	}
	export class Plan {
	    // Go type: time
	    generatedAt: any;
	    options: PlanOptions;
	    stats: Stats;
	    tickets: TicketPlan[];
	    batches: Batch[];

	    static createFrom(source: any = {}) {
	        return new Plan(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = this.convertValues(source["generatedAt"], null);
	        this.options = this.convertValues(source["options"], PlanOptions);
	        this.stats = this.convertValues(source["stats"], Stats);
	        this.tickets = this.convertValues(source["tickets"], TicketPlan);
	        this.batches = this.convertValues(source["batches"], Batch);
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

export namespace harnessui {

	export class HarnessInfo {
	    name: string;
	    path: string;
	    hasGoMod: boolean;
	    hasMain: boolean;

	    static createFrom(source: any = {}) {
	        return new HarnessInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.hasGoMod = source["hasGoMod"];
	        this.hasMain = source["hasMain"];
	    }
	}
	export class RunRequest {
	    harnessPath: string;
	    workDir: string;
	    taskTitle: string;
	    taskDescription: string;
	    model?: string;
	    models?: agentruntime.ModelProfile;
	    envVars: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new RunRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.harnessPath = source["harnessPath"];
	        this.workDir = source["workDir"];
	        this.taskTitle = source["taskTitle"];
	        this.taskDescription = source["taskDescription"];
	        this.model = source["model"];
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
	        this.envVars = source["envVars"];
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
	export class ScaffoldRequest {
	    projectName: string;
	    outputDir: string;
	    provider: string;
	    projectLang: string;
	    includePlanner: boolean;
	    includeTester: boolean;
	    includeCostTracking: boolean;
	    maxIterations: number;
	    reviewCriteria: string;

	    static createFrom(source: any = {}) {
	        return new ScaffoldRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectName = source["projectName"];
	        this.outputDir = source["outputDir"];
	        this.provider = source["provider"];
	        this.projectLang = source["projectLang"];
	        this.includePlanner = source["includePlanner"];
	        this.includeTester = source["includeTester"];
	        this.includeCostTracking = source["includeCostTracking"];
	        this.maxIterations = source["maxIterations"];
	        this.reviewCriteria = source["reviewCriteria"];
	    }
	}
	export class ScaffoldResponse {
	    outputDir: string;
	    filesCreated: string[];

	    static createFrom(source: any = {}) {
	        return new ScaffoldResponse(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outputDir = source["outputDir"];
	        this.filesCreated = source["filesCreated"];
	    }
	}

}

export namespace integrations {

	export class Status {
	    name: string;
	    state: string;
	    message?: string;
	    missingEnv?: string[];
	    // Go type: time
	    lastChecked: any;
	    metadata?: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.state = source["state"];
	        this.message = source["message"];
	        this.missingEnv = source["missingEnv"];
	        this.lastChecked = this.convertValues(source["lastChecked"], null);
	        this.metadata = source["metadata"];
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

export namespace main {

	export class AgentSessionInfo {
	    id: string;
	    projectPath: string;
	    status: string;
	    createdAt: string;
	    tags?: string[];
	    isFavorite?: boolean;
	    model?: string;
	    reasoningEffort?: string;
	    mode?: string;
	    modeConfig?: Record<string, any>;

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
	        this.model = source["model"];
	        this.reasoningEffort = source["reasoningEffort"];
	        this.mode = source["mode"];
	        this.modeConfig = source["modeConfig"];
	    }
	}
	export class DatadogMCPAuthResult {
	    mcpName: string;
	    command: string;
	    message?: string;
	    output?: string;
	    interactive: boolean;
	    launched: boolean;

	    static createFrom(source: any = {}) {
	        return new DatadogMCPAuthResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mcpName = source["mcpName"];
	        this.command = source["command"];
	        this.message = source["message"];
	        this.output = source["output"];
	        this.interactive = source["interactive"];
	        this.launched = source["launched"];
	    }
	}
	export class RoutineOutput {
	    format: string;
	    defaultPath?: string;

	    static createFrom(source: any = {}) {
	        return new RoutineOutput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.defaultPath = source["defaultPath"];
	    }
	}
	export class RoutineParameter {
	    name: string;
	    type: string;
	    description?: string;
	    default?: string;
	    required?: boolean;

	    static createFrom(source: any = {}) {
	        return new RoutineParameter(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.default = source["default"];
	        this.required = source["required"];
	    }
	}
	export class DesktopRoutine {
	    id: string;
	    name: string;
	    description?: string;
	    schedule?: string;
	    workflowTemplate?: string;
	    role: string;
	    profile: string;
	    models?: agentruntime.ModelProfile;
	    integrations?: string[];
	    parameters?: RoutineParameter[];
	    output: RoutineOutput;
	    metadata?: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new DesktopRoutine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.schedule = source["schedule"];
	        this.workflowTemplate = source["workflowTemplate"];
	        this.role = source["role"];
	        this.profile = source["profile"];
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
	        this.integrations = source["integrations"];
	        this.parameters = this.convertValues(source["parameters"], RoutineParameter);
	        this.output = this.convertValues(source["output"], RoutineOutput);
	        this.metadata = source["metadata"];
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
	export class FrontendTicketQueueRequest {
	    projectPath?: string;
	    linearProject?: string;
	    query?: string;
	    limit?: number;
	    states?: string[];
	    includeCanceled?: boolean;
	    maxParallel?: number;
	    tickets?: frontendqueue.Ticket[];

	    static createFrom(source: any = {}) {
	        return new FrontendTicketQueueRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.linearProject = source["linearProject"];
	        this.query = source["query"];
	        this.limit = source["limit"];
	        this.states = source["states"];
	        this.includeCanceled = source["includeCanceled"];
	        this.maxParallel = source["maxParallel"];
	        this.tickets = this.convertValues(source["tickets"], frontendqueue.Ticket);
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
	export class FrontendTicketQueueResult {
	    source: string;
	    linearProject?: string;
	    tickets: frontendqueue.Ticket[];
	    plan: frontendqueue.Plan;
	    warning?: string;

	    static createFrom(source: any = {}) {
	        return new FrontendTicketQueueResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.linearProject = source["linearProject"];
	        this.tickets = this.convertValues(source["tickets"], frontendqueue.Ticket);
	        this.plan = this.convertValues(source["plan"], frontendqueue.Plan);
	        this.warning = source["warning"];
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
	export class FrontendTicketQueueRunRequest {
	    projectPath?: string;
	    linearProject?: string;
	    plan: frontendqueue.Plan;
	    batchId?: string;
	    ticketIds?: string[];
	    runAll?: boolean;
	    baseBranch?: string;
	    models?: agentruntime.ModelProfile;

	    static createFrom(source: any = {}) {
	        return new FrontendTicketQueueRunRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.linearProject = source["linearProject"];
	        this.plan = this.convertValues(source["plan"], frontendqueue.Plan);
	        this.batchId = source["batchId"];
	        this.ticketIds = source["ticketIds"];
	        this.runAll = source["runAll"];
	        this.baseBranch = source["baseBranch"];
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
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
	export class FrontendTicketQueueWorkerSession {
	    ticketId: string;
	    sessionId?: string;
	    policy: string;
	    branchName?: string;
	    status: string;
	    message?: string;

	    static createFrom(source: any = {}) {
	        return new FrontendTicketQueueWorkerSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ticketId = source["ticketId"];
	        this.sessionId = source["sessionId"];
	        this.policy = source["policy"];
	        this.branchName = source["branchName"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class FrontendTicketQueueRunResult {
	    runId: string;
	    batchId?: string;
	    projectPath: string;
	    startedAt: string;
	    sessions: FrontendTicketQueueWorkerSession[];
	    startedCount: number;
	    skippedCount: number;

	    static createFrom(source: any = {}) {
	        return new FrontendTicketQueueRunResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.batchId = source["batchId"];
	        this.projectPath = source["projectPath"];
	        this.startedAt = source["startedAt"];
	        this.sessions = this.convertValues(source["sessions"], FrontendTicketQueueWorkerSession);
	        this.startedCount = source["startedCount"];
	        this.skippedCount = source["skippedCount"];
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
	export class MemoryDocumentDetail {
	    id: string;
	    scope?: string;
	    title?: string;
	    bodyPreview?: string;
	    provenance?: string;
	    sourceRunId?: string;
	    metadata?: Record<string, string>;
	    generatedAt?: string;
	    updatedAt?: string;
	    expiresAt?: string;
	    expired?: boolean;
	    path?: string;
	    body?: string;

	    static createFrom(source: any = {}) {
	        return new MemoryDocumentDetail(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scope = source["scope"];
	        this.title = source["title"];
	        this.bodyPreview = source["bodyPreview"];
	        this.provenance = source["provenance"];
	        this.sourceRunId = source["sourceRunId"];
	        this.metadata = source["metadata"];
	        this.generatedAt = source["generatedAt"];
	        this.updatedAt = source["updatedAt"];
	        this.expiresAt = source["expiresAt"];
	        this.expired = source["expired"];
	        this.path = source["path"];
	        this.body = source["body"];
	    }
	}
	export class MemoryDocumentSummary {
	    id: string;
	    scope?: string;
	    title?: string;
	    bodyPreview?: string;
	    provenance?: string;
	    sourceRunId?: string;
	    metadata?: Record<string, string>;
	    generatedAt?: string;
	    updatedAt?: string;
	    expiresAt?: string;
	    expired?: boolean;
	    path?: string;

	    static createFrom(source: any = {}) {
	        return new MemoryDocumentSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scope = source["scope"];
	        this.title = source["title"];
	        this.bodyPreview = source["bodyPreview"];
	        this.provenance = source["provenance"];
	        this.sourceRunId = source["sourceRunId"];
	        this.metadata = source["metadata"];
	        this.generatedAt = source["generatedAt"];
	        this.updatedAt = source["updatedAt"];
	        this.expiresAt = source["expiresAt"];
	        this.expired = source["expired"];
	        this.path = source["path"];
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
	export class RoutineRequestPreview {
	    runId: string;
	    provider: string;
	    model?: string;
	    role: string;
	    profile?: string;
	    workDir?: string;
	    approvalPolicy?: string;
	    reasoningEffort?: string;
	    mcpServerLabels?: string[];
	    metadata?: Record<string, string>;
	    instructionsPreview?: string;
	    firstMessagePreview?: string;
	    firstMessageCharacterCount?: number;

	    static createFrom(source: any = {}) {
	        return new RoutineRequestPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.role = source["role"];
	        this.profile = source["profile"];
	        this.workDir = source["workDir"];
	        this.approvalPolicy = source["approvalPolicy"];
	        this.reasoningEffort = source["reasoningEffort"];
	        this.mcpServerLabels = source["mcpServerLabels"];
	        this.metadata = source["metadata"];
	        this.instructionsPreview = source["instructionsPreview"];
	        this.firstMessagePreview = source["firstMessagePreview"];
	        this.firstMessageCharacterCount = source["firstMessageCharacterCount"];
	    }
	}
	export class RoutineDryRunResult {
	    routine: DesktopRoutine;
	    values: Record<string, string>;
	    request: RoutineRequestPreview;
	    models?: agentruntime.ModelProfile;
	    integrations?: integrations.Status[];
	    reportPath?: string;
	    command?: string;

	    static createFrom(source: any = {}) {
	        return new RoutineDryRunResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routine = this.convertValues(source["routine"], DesktopRoutine);
	        this.values = source["values"];
	        this.request = this.convertValues(source["request"], RoutineRequestPreview);
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
	        this.integrations = this.convertValues(source["integrations"], integrations.Status);
	        this.reportPath = source["reportPath"];
	        this.command = source["command"];
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


	export class RoutineRemediationResult {
	    id?: string;
	    title: string;
	    status: string;
	    prUrl?: string;
	    branchName?: string;
	    message?: string;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new RoutineRemediationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.prUrl = source["prUrl"];
	        this.branchName = source["branchName"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}

	export class RoutineRunRequest {
	    routineId: string;
	    projectPath: string;
	    values?: Record<string, string>;
	    provider?: string;
	    model?: string;
	    models?: agentruntime.ModelProfile;
	    runId?: string;
	    reportOut?: string;

	    static createFrom(source: any = {}) {
	        return new RoutineRunRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routineId = source["routineId"];
	        this.projectPath = source["projectPath"];
	        this.values = source["values"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
	        this.runId = source["runId"];
	        this.reportOut = source["reportOut"];
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
	export class RoutineRunResult {
	    routineId: string;
	    runId: string;
	    sessionId?: string;
	    provider: string;
	    model?: string;
	    models?: agentruntime.ModelProfile;
	    values?: Record<string, string>;
	    reportPath?: string;
	    integrations?: integrations.Status[];
	    usage?: agentruntime.Usage;
	    report?: string;
	    remediations?: RoutineRemediationResult[];

	    static createFrom(source: any = {}) {
	        return new RoutineRunResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routineId = source["routineId"];
	        this.runId = source["runId"];
	        this.sessionId = source["sessionId"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.models = this.convertValues(source["models"], agentruntime.ModelProfile);
	        this.values = source["values"];
	        this.reportPath = source["reportPath"];
	        this.integrations = this.convertValues(source["integrations"], integrations.Status);
	        this.usage = this.convertValues(source["usage"], agentruntime.Usage);
	        this.report = source["report"];
	        this.remediations = this.convertValues(source["remediations"], RoutineRemediationResult);
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
	export class RuntimeArtifactSummary {
	    kind: string;
	    path?: string;
	    url?: string;
	    diff?: string;
	    message?: string;
	    phaseId?: string;
	    taskId?: string;
	    eventType?: string;
	    firstSeenAt?: string;
	    lastSeenAt?: string;
	    eventCount: number;

	    static createFrom(source: any = {}) {
	        return new RuntimeArtifactSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.path = source["path"];
	        this.url = source["url"];
	        this.diff = source["diff"];
	        this.message = source["message"];
	        this.phaseId = source["phaseId"];
	        this.taskId = source["taskId"];
	        this.eventType = source["eventType"];
	        this.firstSeenAt = source["firstSeenAt"];
	        this.lastSeenAt = source["lastSeenAt"];
	        this.eventCount = source["eventCount"];
	    }
	}
	export class RuntimeEventSummary {
	    type: string;
	    status?: string;
	    role?: string;
	    provider?: string;
	    model?: string;
	    phaseId?: string;
	    taskId?: string;
	    name?: string;
	    message?: string;
	    timestamp?: string;
	    toolName?: string;
	    toolError?: boolean;
	    artifactKind?: string;
	    artifactPath?: string;
	    artifactUrl?: string;
	    rawPreview?: string;

	    static createFrom(source: any = {}) {
	        return new RuntimeEventSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.status = source["status"];
	        this.role = source["role"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.phaseId = source["phaseId"];
	        this.taskId = source["taskId"];
	        this.name = source["name"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	        this.toolName = source["toolName"];
	        this.toolError = source["toolError"];
	        this.artifactKind = source["artifactKind"];
	        this.artifactPath = source["artifactPath"];
	        this.artifactUrl = source["artifactUrl"];
	        this.rawPreview = source["rawPreview"];
	    }
	}
	export class RuntimeRequestSummary {
	    provider?: string;
	    model?: string;
	    role?: string;
	    profile?: string;
	    workDir?: string;
	    approvalPolicy?: string;
	    reasoningEffort?: string;
	    background?: boolean;
	    messageCount: number;
	    toolNames?: string[];
	    mcpServerLabels?: string[];
	    outputSchema?: string;
	    instructionsPreview?: string;
	    firstMessagePreview?: string;
	    metadata?: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new RuntimeRequestSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.role = source["role"];
	        this.profile = source["profile"];
	        this.workDir = source["workDir"];
	        this.approvalPolicy = source["approvalPolicy"];
	        this.reasoningEffort = source["reasoningEffort"];
	        this.background = source["background"];
	        this.messageCount = source["messageCount"];
	        this.toolNames = source["toolNames"];
	        this.mcpServerLabels = source["mcpServerLabels"];
	        this.outputSchema = source["outputSchema"];
	        this.instructionsPreview = source["instructionsPreview"];
	        this.firstMessagePreview = source["firstMessagePreview"];
	        this.metadata = source["metadata"];
	    }
	}
	export class RuntimeRunSummary {
	    runId: string;
	    provider?: string;
	    model?: string;
	    role?: string;
	    profile?: string;
	    workDir?: string;
	    status?: string;
	    startedAt?: string;
	    updatedAt?: string;
	    endedAt?: string;
	    eventCount: number;
	    artifactCount?: number;

	    static createFrom(source: any = {}) {
	        return new RuntimeRunSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.role = source["role"];
	        this.profile = source["profile"];
	        this.workDir = source["workDir"];
	        this.status = source["status"];
	        this.startedAt = source["startedAt"];
	        this.updatedAt = source["updatedAt"];
	        this.endedAt = source["endedAt"];
	        this.eventCount = source["eventCount"];
	        this.artifactCount = source["artifactCount"];
	    }
	}
	export class RuntimeRunDetail {
	    metadata: RuntimeRunSummary;
	    request?: RuntimeRequestSummary;
	    events: RuntimeEventSummary[];
	    artifacts: RuntimeArtifactSummary[];

	    static createFrom(source: any = {}) {
	        return new RuntimeRunDetail(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metadata = this.convertValues(source["metadata"], RuntimeRunSummary);
	        this.request = this.convertValues(source["request"], RuntimeRequestSummary);
	        this.events = this.convertValues(source["events"], RuntimeEventSummary);
	        this.artifacts = this.convertValues(source["artifacts"], RuntimeArtifactSummary);
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

export namespace services {

	export class AutoDistillResult {
	    domain: string;
	    brainId: string;
	    path: string;
	    signals: number;
	    usedLlm: boolean;

	    static createFrom(source: any = {}) {
	        return new AutoDistillResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.brainId = source["brainId"];
	        this.path = source["path"];
	        this.signals = source["signals"];
	        this.usedLlm = source["usedLlm"];
	    }
	}
	export class BrainReference {
	    path: string;
	    description: string;
	    checksum: string;

	    static createFrom(source: any = {}) {
	        return new BrainReference(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.description = source["description"];
	        this.checksum = source["checksum"];
	    }
	}
	export class BrainSection {
	    title: string;
	    content: string;

	    static createFrom(source: any = {}) {
	        return new BrainSection(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.content = source["content"];
	    }
	}
	export class BrainDetail {
	    id: string;
	    name: string;
	    description: string;
	    confidence: number;
	    version: number;
	    lastUpdated: string;
	    keywords: string[];
	    entities: string[];
	    filePatterns: string[];
	    sections: BrainSection[];
	    references: BrainReference[];

	    static createFrom(source: any = {}) {
	        return new BrainDetail(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.confidence = source["confidence"];
	        this.version = source["version"];
	        this.lastUpdated = source["lastUpdated"];
	        this.keywords = source["keywords"];
	        this.entities = source["entities"];
	        this.filePatterns = source["filePatterns"];
	        this.sections = this.convertValues(source["sections"], BrainSection);
	        this.references = this.convertValues(source["references"], BrainReference);
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
	export class BrainEntry {
	    id: string;
	    name: string;
	    description: string;
	    confidence: number;
	    version: number;
	    lastUpdated: string;
	    keywords: string[];
	    entities: string[];
	    filePatterns: string[];

	    static createFrom(source: any = {}) {
	        return new BrainEntry(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.confidence = source["confidence"];
	        this.version = source["version"];
	        this.lastUpdated = source["lastUpdated"];
	        this.keywords = source["keywords"];
	        this.entities = source["entities"];
	        this.filePatterns = source["filePatterns"];
	    }
	}


	export class StaleRefResult {
	    path: string;
	    reason: string;

	    static createFrom(source: any = {}) {
	        return new StaleRefResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.reason = source["reason"];
	    }
	}
	export class BrainValidationResult {
	    valid: boolean;
	    errors: string[];
	    stale: StaleRefResult[];

	    static createFrom(source: any = {}) {
	        return new BrainValidationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.errors = source["errors"];
	        this.stale = this.convertValues(source["stale"], StaleRefResult);
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
	export class SignalEntry {
	    type: string;
	    domain: string;
	    details: string;
	    filePaths: string[];
	    count: number;
	    firstSeen: string;
	    lastSeen: string;

	    static createFrom(source: any = {}) {
	        return new SignalEntry(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.domain = source["domain"];
	        this.details = source["details"];
	        this.filePaths = source["filePaths"];
	        this.count = source["count"];
	        this.firstSeen = source["firstSeen"];
	        this.lastSeen = source["lastSeen"];
	    }
	}

}

export namespace triage {

	export class TriageOptions {
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

	    static createFrom(source: any = {}) {
	        return new TriageOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.teams = source["teams"];
	        this.states = source["states"];
	        this.limit = source["limit"];
	        this.ticketIds = source["ticketIds"];
	        this.postComments = source["postComments"];
	        this.dryRun = source["dryRun"];
	        this.outputDir = source["outputDir"];
	        this.concurrency = source["concurrency"];
	        this.generatePlans = source["generatePlans"];
	        this.repoPath = source["repoPath"];
	    }
	}

}

