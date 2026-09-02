export namespace main {
	
	export class AgentMessageRequest {
	    conversationId: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentMessageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversationId = source["conversationId"];
	        this.content = source["content"];
	    }
	}
	export class ChatMessage {
	    id: number;
	    conversationId: number;
	    role: string;
	    content: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.conversationId = source["conversationId"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class AgentReply {
	    conversationId: number;
	    messages: ChatMessage[];
	
	    static createFrom(source: any = {}) {
	        return new AgentReply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversationId = source["conversationId"];
	        this.messages = this.convertValues(source["messages"], ChatMessage);
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
	export class AgentStatus {
	    oaMode: string;
	    oaConfigured: boolean;
	    toolCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AgentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oaMode = source["oaMode"];
	        this.oaConfigured = source["oaConfigured"];
	        this.toolCount = source["toolCount"];
	    }
	}
	
	export class Conversation {
	    id: number;
	    title: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CreateSkillRequest {
	    name: string;
	    description: string;
	    tags: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateSkillRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	    }
	}
	export class ImportResult {
	    imported: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.message = source["message"];
	    }
	}
	export class Run {
	    id: number;
	    skillName: string;
	    skillVersion: string;
	    operator: string;
	    dataLevel: string;
	    executorAlias: string;
	    status: string;
	    output: string;
	    errorMessage: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Run(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.skillName = source["skillName"];
	        this.skillVersion = source["skillVersion"];
	        this.operator = source["operator"];
	        this.dataLevel = source["dataLevel"];
	        this.executorAlias = source["executorAlias"];
	        this.status = source["status"];
	        this.output = source["output"];
	        this.errorMessage = source["errorMessage"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Session {
	    userId: number;
	    account: string;
	    displayName: string;
	    role: string;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userId = source["userId"];
	        this.account = source["account"];
	        this.displayName = source["displayName"];
	        this.role = source["role"];
	    }
	}
	export class Skill {
	    id: number;
	    name: string;
	    description: string;
	    owner: string;
	    version: string;
	    status: string;
	    enabled: boolean;
	    tags: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.owner = source["owner"];
	        this.version = source["version"];
	        this.status = source["status"];
	        this.enabled = source["enabled"];
	        this.tags = source["tags"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class StartRunRequest {
	    skillId: number;
	    dataLevel: string;
	    inputSummary: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.dataLevel = source["dataLevel"];
	        this.inputSummary = source["inputSummary"];
	    }
	}

}

