export namespace present {
	
	export class ApiErrorDTO {
	    error_code: string;
	    message: string;
	    detail?: string;
	    target_path?: string;
	    hint?: string;
	
	    static createFrom(source: any = {}) {
	        return new ApiErrorDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	        this.target_path = source["target_path"];
	        this.hint = source["hint"];
	    }
	}
	export class AttachmentUploadDTO {
	    source_path: string;
	    original_file_name: string;
	    mime_type: string;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentUploadDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source_path = source["source_path"];
	        this.original_file_name = source["original_file_name"];
	        this.mime_type = source["mime_type"];
	    }
	}
	export class CommentCreateDTO {
	    body: string;
	    author_name: string;
	    attachments: AttachmentUploadDTO[];
	
	    static createFrom(source: any = {}) {
	        return new CommentCreateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.body = source["body"];
	        this.author_name = source["author_name"];
	        this.attachments = this.convertValues(source["attachments"], AttachmentUploadDTO);
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
	export class IssueCreateDTO {
	    title: string;
	    description: string;
	    due_date: string;
	    priority: string;
	    assignee: string;
	
	    static createFrom(source: any = {}) {
	        return new IssueCreateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.due_date = source["due_date"];
	        this.priority = source["priority"];
	        this.assignee = source["assignee"];
	    }
	}
	export class IssueListQueryDTO {
	    page: number;
	    page_size: number;
	    sort_by: string;
	    sort_order: string;
	
	    static createFrom(source: any = {}) {
	        return new IssueListQueryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.sort_by = source["sort_by"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class IssueUpdateDTO {
	    title: string;
	    description: string;
	    due_date: string;
	    priority: string;
	    status: string;
	    assignee: string;
	
	    static createFrom(source: any = {}) {
	        return new IssueUpdateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.due_date = source["due_date"];
	        this.priority = source["priority"];
	        this.status = source["status"];
	        this.assignee = source["assignee"];
	    }
	}
	export class Response {
	    ok: boolean;
	    data?: any;
	    error?: ApiErrorDTO;
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.data = source["data"];
	        this.error = this.convertValues(source["error"], ApiErrorDTO);
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

