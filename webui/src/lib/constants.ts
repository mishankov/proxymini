import type { InspectorTab, SortOption, StatusFilter } from "$lib/types";

export const METHOD_OPTIONS = ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"] as const;
export const STATUS_OPTIONS: readonly StatusFilter[] = ["2xx", "3xx", "4xx", "5xx"];
export const TAB_OPTIONS: readonly InspectorTab[] = ["overview", "request", "response", "headers", "raw"];

export const POLL_INTERVAL_MS = 3000;
export const INITIAL_RENDER_LIMIT = 500;
export const RENDER_STEP = 250;
export const DEFAULT_SORT: SortOption = "timeDesc";
