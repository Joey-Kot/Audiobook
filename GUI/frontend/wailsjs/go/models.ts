export namespace batch {
	
	export class BatchProgress {
	    fileIndex: number;
	    fileName: string;
	    charCount: number;
	    percent: number;
	    status: string;
	    message?: string;
	    output?: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileIndex = source["fileIndex"];
	        this.fileName = source["fileName"];
	        this.charCount = source["charCount"];
	        this.percent = source["percent"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.output = source["output"];
	    }
	}

}

export namespace config {
	
	export class Config {
	    API_BASE_URL: string;
	    API_TOKEN: string;
	    MODEL: string;
	    VOICE_JSON: string;
	    SPLIT_THRESHOLD: number;
	    OUTPUT_DIR: string;
	    CONCURRENCY: number;
	    REQUEST_TIMEOUT: number;
	    FFMPEG_PATH: string;
	    OUTPUT_BITRATE_KB: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.API_BASE_URL = source["API_BASE_URL"];
	        this.API_TOKEN = source["API_TOKEN"];
	        this.MODEL = source["MODEL"];
	        this.VOICE_JSON = source["VOICE_JSON"];
	        this.SPLIT_THRESHOLD = source["SPLIT_THRESHOLD"];
	        this.OUTPUT_DIR = source["OUTPUT_DIR"];
	        this.CONCURRENCY = source["CONCURRENCY"];
	        this.REQUEST_TIMEOUT = source["REQUEST_TIMEOUT"];
	        this.FFMPEG_PATH = source["FFMPEG_PATH"];
	        this.OUTPUT_BITRATE_KB = source["OUTPUT_BITRATE_KB"];
	    }
	}

}

export namespace main {
	
	export class AppState {
	    running: boolean;
	    paused: boolean;
	    files: batch.BatchProgress[];
	    config: config.Config;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.paused = source["paused"];
	        this.files = this.convertValues(source["files"], batch.BatchProgress);
	        this.config = this.convertValues(source["config"], config.Config);
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
	export class ConfigPayload {
	    path: string;
	    json: string;
	    data: config.Config;
	
	    static createFrom(source: any = {}) {
	        return new ConfigPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.json = source["json"];
	        this.data = this.convertValues(source["data"], config.Config);
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

