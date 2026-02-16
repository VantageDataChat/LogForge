export namespace model {
	
	export class BatchProgress {
	    status: string;
	    current_file: string;
	    progress: number;
	    total_files: number;
	    processed: number;
	    failed: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.current_file = source["current_file"];
	        this.progress = source["progress"];
	        this.total_files = source["total_files"];
	        this.processed = source["processed"];
	        this.failed = source["failed"];
	        this.message = source["message"];
	    }
	}
	export class GenerateResult {
	    project_id: string;
	    code: string;
	    valid: boolean;
	    errors?: string[];
	
	    static createFrom(source: any = {}) {
	        return new GenerateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.code = source["code"];
	        this.valid = source["valid"];
	        this.errors = source["errors"];
	    }
	}
	export class LogFileSample {
	    file_name: string;
	    project_name: string;
	    sample_text: string;
	
	    static createFrom(source: any = {}) {
	        return new LogFileSample(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_name = source["file_name"];
	        this.project_name = source["project_name"];
	        this.sample_text = source["sample_text"];
	    }
	}
	export class LLMConfig {
	    base_url: string;
	    api_key: string;
	    model_name: string;
	
	    static createFrom(source: any = {}) {
	        return new LLMConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.base_url = source["base_url"];
	        this.api_key = source["api_key"];
	        this.model_name = source["model_name"];
	    }
	}
	export class Project {
	    id: string;
	    name: string;
	    sample_data: string;
	    code: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sample_data = source["sample_data"];
	        this.code = source["code"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.status = source["status"];
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
	export class Settings {
	    llm: LLMConfig;
	    uv_path: string;
	    default_input_dir: string;
	    default_output_dir: string;
	    sample_lines?: number;
	    show_wizard?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.llm = this.convertValues(source["llm"], LLMConfig);
	        this.uv_path = source["uv_path"];
	        this.default_input_dir = source["default_input_dir"];
	        this.default_output_dir = source["default_output_dir"];
	        this.sample_lines = source["sample_lines"];
	        this.show_wizard = source["show_wizard"];
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

export namespace pyenv {
	
	export class EnvStatus {
	    uv_available: boolean;
	    env_exists: boolean;
	    env_path: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uv_available = source["uv_available"];
	        this.env_exists = source["env_exists"];
	        this.env_path = source["env_path"];
	    }
	}

}

