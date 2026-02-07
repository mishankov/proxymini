import type { EnrichedLog, HeaderEntry, RequestLog, StatusClass } from "$lib/types";

export function escapeHtml(value: string): string {
	return value
		.replaceAll("&", "&amp;")
		.replaceAll("<", "&lt;")
		.replaceAll(">", "&gt;")
		.replaceAll('"', "&quot;")
		.replaceAll("'", "&#39;");
}

export function escapeRegExp(value: string): string {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function highlightText(value: string, query: string): string {
	const escapedValue = escapeHtml(value);
	const trimmed = query.trim();
	if (!trimmed) {
		return escapedValue;
	}

	const matcher = new RegExp(`(${escapeRegExp(trimmed)})`, "ig");
	return escapedValue.replace(matcher, "<mark>$1</mark>");
}

export function safeParseJSON(value: string): unknown | null {
	if (!value) {
		return null;
	}

	try {
		return JSON.parse(value);
	} catch {
		return null;
	}
}

export function formatTimestamp(unixSeconds: number): string {
	const date = new Date(unixSeconds * 1000);
	return date.toISOString().replace("T", " ").substring(0, 19);
}

export function statusClassOf(status: number): StatusClass {
	if (status >= 200 && status < 300) {
		return "2xx";
	}
	if (status >= 300 && status < 400) {
		return "3xx";
	}
	if (status >= 400 && status < 500) {
		return "4xx";
	}
	if (status >= 500 && status < 600) {
		return "5xx";
	}
	return "unknown";
}

export function normalizeMethod(method: string): string {
	const normalized = method.toUpperCase().trim();
	return normalized || "UNKNOWN";
}

export function normalizeText(value: string): string {
	return value.toLowerCase();
}

export function toHeaderEntries(data: unknown): HeaderEntry[] {
	if (!data || typeof data !== "object") {
		return [];
	}

	return Object.entries(data as Record<string, unknown>).map(([key, raw]) => {
		if (Array.isArray(raw)) {
			return { key, value: raw.join(", ") };
		}
		return { key, value: String(raw ?? "") };
	});
}

export function prettyBody(value: string): { text: string; isJSON: boolean } {
	if (!value) {
		return { text: "", isJSON: false };
	}

	const parsed = safeParseJSON(value);
	if (parsed === null) {
		return { text: value, isJSON: false };
	}

	return { text: JSON.stringify(parsed, null, 2), isJSON: true };
}

export function buildSearchBlob(log: RequestLog): string {
	return normalizeText(
		[
			log.id,
			log.method,
			String(log.status),
			log.proxyUrl,
			log.url,
			log.requestHeaders,
			log.responseHeaders,
			log.requestBody,
			log.responseBody
		].join(" ")
	);
}

export function enrichLog(log: RequestLog): EnrichedLog {
	const requestHeadersParsed = safeParseJSON(log.requestHeaders);
	const responseHeadersParsed = safeParseJSON(log.responseHeaders);

	return {
		...log,
		methodNormalized: normalizeMethod(log.method),
		statusClass: statusClassOf(Number(log.status)),
		timeFormatted: formatTimestamp(Number(log.time)),
		requestHeadersEntries: toHeaderEntries(requestHeadersParsed),
		responseHeadersEntries: toHeaderEntries(responseHeadersParsed),
		searchBlob: buildSearchBlob(log)
	};
}

export function dedupeByID(logs: EnrichedLog[]): EnrichedLog[] {
	const seen = new Set<string>();
	const result: EnrichedLog[] = [];

	for (const log of logs) {
		if (seen.has(log.id)) {
			continue;
		}
		seen.add(log.id);
		result.push(log);
	}

	return result;
}
