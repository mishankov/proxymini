export interface RequestLog {
	id: string;
	time: number;
	method: string;
	proxyUrl: string;
	url: string;
	requestHeaders: string;
	requestBody: string;
	status: number;
	responseHeaders: string;
	responseBody: string;
}

export type StatusClass = "2xx" | "3xx" | "4xx" | "5xx" | "unknown";
export type StatusFilter = Exclude<StatusClass, "unknown">;
export type SortOption = "timeDesc" | "timeAsc" | "statusDesc";
export type InspectorTab = "overview" | "request" | "response" | "headers" | "raw";

export interface HeaderEntry {
	key: string;
	value: string;
}

export interface Filters {
	search: string;
	methods: Set<string>;
	statuses: Set<StatusFilter>;
}

export interface EnrichedLog extends RequestLog {
	methodNormalized: string;
	statusClass: StatusClass;
	timeFormatted: string;
	requestHeadersEntries: HeaderEntry[];
	responseHeadersEntries: HeaderEntry[];
	searchBlob: string;
}

