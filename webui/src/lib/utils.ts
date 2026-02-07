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

export type BodySyntax = "plain" | "json" | "xml";

const JSON_TOKEN_REGEX =
	/"(?:\\.|[^"\\])*"(?=\s*:)|"(?:\\.|[^"\\])*"|\b-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?\b|\btrue\b|\bfalse\b|\bnull\b|[{}\[\]:,]/g;
const XML_NODE_REGEX = /<!--[\s\S]*?-->|<\?[\s\S]*?\?>|<!DOCTYPE[\s\S]*?>|<\/?[A-Za-z_][\w:.-]*(?:\s+[^<>]*?)?\/?>/gi;
const JSON_TOKEN_CLASSES = {
	key: "text-sky-300",
	string: "text-lime-300",
	number: "text-amber-300",
	boolean: "text-fuchsia-300",
	null: "text-rose-300",
	punctuation: "text-slate-300"
} as const;
const XML_TOKEN_CLASSES = {
	bracket: "text-sky-200",
	tag: "text-sky-300",
	attrName: "text-cyan-200",
	attrValue: "text-lime-300",
	comment: "text-slate-500",
	prolog: "text-slate-500"
} as const;

type ParseJSONResult =
	| {
			ok: true;
			value: unknown;
	  }
	| {
			ok: false;
	  };

function parseJSON(value: string): ParseJSONResult {
	try {
		return { ok: true, value: JSON.parse(value) };
	} catch {
		return { ok: false };
	}
}

function highlightEscapedSegment(value: string, query: string): string {
	const trimmed = query.trim();
	if (!trimmed) {
		return escapeHtml(value);
	}

	const matcher = new RegExp(`(${escapeRegExp(trimmed)})`, "ig");
	const parts = value.split(matcher);
	if (parts.length === 0) {
		return "";
	}

	return parts
		.map((part) => {
			if (!part) {
				return "";
			}
			if (part.toLowerCase() === trimmed.toLowerCase()) {
				return `<mark>${escapeHtml(part)}</mark>`;
			}
			return escapeHtml(part);
		})
		.join("");
}

function wrapToken(className: string, value: string, query: string): string {
	return `<span class="${className}">${highlightEscapedSegment(value, query)}</span>`;
}

function renderTokenized(
	value: string,
	query: string,
	pattern: RegExp,
	renderToken: (token: string, index: number, source: string) => string
): string {
	pattern.lastIndex = 0;

	let rendered = "";
	let cursor = 0;
	let match: RegExpExecArray | null;

	while ((match = pattern.exec(value)) !== null) {
		const token = match[0];
		const start = match.index;

		rendered += highlightEscapedSegment(value.slice(cursor, start), query);
		rendered += renderToken(token, start, value);
		cursor = start + token.length;
	}

	rendered += highlightEscapedSegment(value.slice(cursor), query);
	return rendered;
}

function isJSONContentType(contentType: string): boolean {
	return contentType === "application/json" || contentType.endsWith("+json");
}

function isXMLContentType(contentType: string): boolean {
	return contentType === "application/xml" || contentType === "text/xml" || contentType.endsWith("+xml");
}

function isLikelyXML(value: string): boolean {
	const trimmed = value.trim();
	if (!trimmed || !trimmed.startsWith("<") || !trimmed.endsWith(">")) {
		return false;
	}

	XML_NODE_REGEX.lastIndex = 0;

	const stack: string[] = [];
	let cursor = 0;
	let foundElement = false;
	let match: RegExpExecArray | null;

	while ((match = XML_NODE_REGEX.exec(trimmed)) !== null) {
		const token = match[0];
		const gap = trimmed.slice(cursor, match.index);
		if (gap.includes("<") || gap.includes(">")) {
			return false;
		}

		if (token.startsWith("</")) {
			const nameMatch = token.match(/^<\/([A-Za-z_][\w:.-]*)/);
			const name = nameMatch?.[1];
			if (!name) {
				return false;
			}

			const current = stack.pop();
			if (current !== name) {
				return false;
			}
		} else if (token.startsWith("<!") || token.startsWith("<?")) {
			// Ignore declarations and comments in structural matching.
		} else {
			const nameMatch = token.match(/^<([A-Za-z_][\w:.-]*)/);
			const name = nameMatch?.[1];
			if (!name) {
				return false;
			}

			foundElement = true;
			if (!token.endsWith("/>")) {
				stack.push(name);
			}
		}

		cursor = match.index + token.length;
	}

	const tail = trimmed.slice(cursor);
	if (tail.includes("<") || tail.includes(">")) {
		return false;
	}

	return foundElement && stack.length === 0;
}

function renderJSONHtml(value: string, query: string): string {
	return renderTokenized(value, query, JSON_TOKEN_REGEX, (token, index, source) => {
		if (token.startsWith('"')) {
			let cursor = index + token.length;
			while (cursor < source.length && /\s/.test(source[cursor])) {
				cursor += 1;
			}

			const className = source[cursor] === ":" ? JSON_TOKEN_CLASSES.key : JSON_TOKEN_CLASSES.string;
			return wrapToken(className, token, query);
		}

		if (token === "true" || token === "false") {
			return wrapToken(JSON_TOKEN_CLASSES.boolean, token, query);
		}

		if (token === "null") {
			return wrapToken(JSON_TOKEN_CLASSES.null, token, query);
		}

		if (/^[{}\[\]:,]$/.test(token)) {
			return wrapToken(JSON_TOKEN_CLASSES.punctuation, token, query);
		}

		return wrapToken(JSON_TOKEN_CLASSES.number, token, query);
	});
}

function renderXMLAttributes(attributes: string, query: string): string {
	if (!attributes) {
		return "";
	}

	const ATTR_REGEX = /([^\s=/>]+)(\s*=\s*(?:"[^"]*"|'[^']*'|[^\s/>]+))?/g;
	ATTR_REGEX.lastIndex = 0;

	let rendered = "";
	let cursor = 0;
	let match: RegExpExecArray | null;

	while ((match = ATTR_REGEX.exec(attributes)) !== null) {
		const [raw, name, assignment = ""] = match;
		const start = match.index;

		rendered += highlightEscapedSegment(attributes.slice(cursor, start), query);
		rendered += wrapToken(XML_TOKEN_CLASSES.attrName, name, query);

		if (assignment) {
			const assignmentMatch = assignment.match(/^(\s*=\s*)([\s\S]+)$/);
			if (assignmentMatch) {
				rendered += highlightEscapedSegment(assignmentMatch[1], query);
				rendered += wrapToken(XML_TOKEN_CLASSES.attrValue, assignmentMatch[2], query);
			} else {
				rendered += highlightEscapedSegment(assignment, query);
			}
		}

		cursor = start + raw.length;
	}

	rendered += highlightEscapedSegment(attributes.slice(cursor), query);
	return rendered;
}

function renderXMLTagToken(token: string, query: string): string {
	if (!token.startsWith("<")) {
		return highlightEscapedSegment(token, query);
	}

	let openBracket = "<";
	let closeBracket = ">";
	let inner = token.slice(1, -1);

	if (token.startsWith("</")) {
		openBracket = "</";
		inner = token.slice(2, -1);
	} else if (token.endsWith("/>")) {
		closeBracket = "/>";
		inner = token.slice(1, -2);
	}

	const nameMatch = inner.match(/^(\s*)([^\s/>]+)([\s\S]*)$/);
	if (!nameMatch) {
		return wrapToken(XML_TOKEN_CLASSES.tag, token, query);
	}

	const [, leadingSpace, name, attributes] = nameMatch;

	return (
		wrapToken(XML_TOKEN_CLASSES.bracket, openBracket, query) +
		highlightEscapedSegment(leadingSpace, query) +
		wrapToken(XML_TOKEN_CLASSES.tag, name, query) +
		renderXMLAttributes(attributes, query) +
		wrapToken(XML_TOKEN_CLASSES.bracket, closeBracket, query)
	);
}

function renderXMLHtml(value: string, query: string): string {
	return renderTokenized(value, query, XML_NODE_REGEX, (token) => {
		if (token.startsWith("<!--")) {
			return wrapToken(XML_TOKEN_CLASSES.comment, token, query);
		}

		if (token.startsWith("<?") || token.toUpperCase().startsWith("<!DOCTYPE")) {
			return wrapToken(XML_TOKEN_CLASSES.prolog, token, query);
		}

		return renderXMLTagToken(token, query);
	});
}

export function normalizeContentType(contentType = ""): string {
	return contentType.split(";")[0]?.trim().toLowerCase() ?? "";
}

export function detectBodySyntax(body: string, contentType = ""): BodySyntax {
	if (!body) {
		return "plain";
	}

	const normalizedContentType = normalizeContentType(contentType);
	if (isJSONContentType(normalizedContentType)) {
		return "json";
	}

	if (isXMLContentType(normalizedContentType)) {
		return "xml";
	}

	if (parseJSON(body).ok) {
		return "json";
	}

	if (isLikelyXML(body)) {
		return "xml";
	}

	return "plain";
}

export function formatBodyForDisplay(body: string, syntax: BodySyntax): string {
	if (!body) {
		return "";
	}

	if (syntax !== "json") {
		return body;
	}

	const parsed = parseJSON(body);
	if (!parsed.ok) {
		return body;
	}

	return JSON.stringify(parsed.value, null, 2);
}

export function renderPayloadHtml(body: string, query: string, contentType = ""): string {
	if (!body) {
		return "";
	}

	const syntax = detectBodySyntax(body, contentType);
	const formattedBody = formatBodyForDisplay(body, syntax);

	if (syntax === "json") {
		return renderJSONHtml(formattedBody, query);
	}

	if (syntax === "xml") {
		return renderXMLHtml(formattedBody, query);
	}

	return highlightText(formattedBody, query);
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

export function formatElapsed(elapsedMs: number): string {
	if (elapsedMs < 1000) {
		return `${elapsedMs} ms`;
	}

	return `${(elapsedMs / 1000).toFixed(2)} s`;
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
			String(log.elapsedMs),
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
		elapsedFormatted: formatElapsed(Number(log.elapsedMs)),
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
