export namespace config {
	
	export class FolderBookmark {
	    id: string;
	    label: string;
	    path: string;
	    isDefault: boolean;
	    lastUsed: number;
	
	    static createFrom(source: any = {}) {
	        return new FolderBookmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.path = source["path"];
	        this.isDefault = source["isDefault"];
	        this.lastUsed = source["lastUsed"];
	    }
	}
	export class RoutingRule {
	    pattern: string;
	    targetDir: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new RoutingRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.targetDir = source["targetDir"];
	        this.label = source["label"];
	    }
	}
	export class Settings {
	    hfToken: string;
	    defaultDownloadDir: string;
	    autoDownload: boolean;
	    maxConcurrentFiles: number;
	    maxConnectionsPerFile: number;
	    bookmarks: FolderBookmark[];
	    routingRules: RoutingRule[];
	    recentPaths: string[];
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hfToken = source["hfToken"];
	        this.defaultDownloadDir = source["defaultDownloadDir"];
	        this.autoDownload = source["autoDownload"];
	        this.maxConcurrentFiles = source["maxConcurrentFiles"];
	        this.maxConnectionsPerFile = source["maxConnectionsPerFile"];
	        this.bookmarks = this.convertValues(source["bookmarks"], FolderBookmark);
	        this.routingRules = this.convertValues(source["routingRules"], RoutingRule);
	        this.recentPaths = source["recentPaths"];
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

export namespace downloader {
	
	export class DownloadItem {
	    id: string;
	    repoId: string;
	    revision: string;
	    remotePath: string;
	    destinationDir: string;
	    finalFilename: string;
	    size: number;
	    expectedSha256: string;
	    status: string;
	    autoStart: boolean;
	    downloadedBytes: number;
	    progress: number;
	    speedBps: number;
	    speedFormatted: string;
	    etaSeconds: number;
	    errorMessage?: string;
	    createdAt: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.repoId = source["repoId"];
	        this.revision = source["revision"];
	        this.remotePath = source["remotePath"];
	        this.destinationDir = source["destinationDir"];
	        this.finalFilename = source["finalFilename"];
	        this.size = source["size"];
	        this.expectedSha256 = source["expectedSha256"];
	        this.status = source["status"];
	        this.autoStart = source["autoStart"];
	        this.downloadedBytes = source["downloadedBytes"];
	        this.progress = source["progress"];
	        this.speedBps = source["speedBps"];
	        this.speedFormatted = source["speedFormatted"];
	        this.etaSeconds = source["etaSeconds"];
	        this.errorMessage = source["errorMessage"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class VerificationResult {
	    exists: boolean;
	    valid: boolean;
	    actualSize: number;
	    actualSha256?: string;
	    expectedSha256?: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new VerificationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.valid = source["valid"];
	        this.actualSize = source["actualSize"];
	        this.actualSha256 = source["actualSha256"];
	        this.expectedSha256 = source["expectedSha256"];
	        this.message = source["message"];
	    }
	}

}

export namespace hfapi {
	
	export class LFSInfo {
	    oid: string;
	    size: number;
	    pointerSize?: number;
	    sha256?: string;
	
	    static createFrom(source: any = {}) {
	        return new LFSInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oid = source["oid"];
	        this.size = source["size"];
	        this.pointerSize = source["pointerSize"];
	        this.sha256 = source["sha256"];
	    }
	}
	export class FileNode {
	    type: string;
	    oid: string;
	    size: number;
	    path: string;
	    lfs?: LFSInfo;
	    sha256?: string;
	    downloadUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.oid = source["oid"];
	        this.size = source["size"];
	        this.path = source["path"];
	        this.lfs = this.convertValues(source["lfs"], LFSInfo);
	        this.sha256 = source["sha256"];
	        this.downloadUrl = source["downloadUrl"];
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
	
	export class ParsedTarget {
	    rawInput: string;
	    repoId: string;
	    revision: string;
	    type: string;
	    subpath: string;
	
	    static createFrom(source: any = {}) {
	        return new ParsedTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rawInput = source["rawInput"];
	        this.repoId = source["repoId"];
	        this.revision = source["revision"];
	        this.type = source["type"];
	        this.subpath = source["subpath"];
	    }
	}

}

export namespace main {
	
	export class InspectResponse {
	    target?: hfapi.ParsedTarget;
	    files: hfapi.FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new InspectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = this.convertValues(source["target"], hfapi.ParsedTarget);
	        this.files = this.convertValues(source["files"], hfapi.FileNode);
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

export namespace platform {
	
	export class DiskSpaceInfo {
	    path: string;
	    freeBytes: number;
	    totalBytes: number;
	    availableBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskSpaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.freeBytes = source["freeBytes"];
	        this.totalBytes = source["totalBytes"];
	        this.availableBytes = source["availableBytes"];
	    }
	}

}

