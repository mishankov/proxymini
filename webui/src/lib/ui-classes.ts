import type { StatusClass } from "$lib/types";

export const STATUS_PILL_CLASSES: Record<StatusClass, string> = {
	"2xx": "bg-emerald-500/15 text-emerald-200",
	"3xx": "bg-amber-500/15 text-amber-200",
	"4xx": "bg-orange-500/15 text-orange-200",
	"5xx": "bg-rose-500/15 text-rose-200",
	unknown: "bg-zinc-500/15 text-zinc-300"
};

export const STATUS_TEXT_CLASSES: Record<StatusClass, string> = {
	"2xx": "text-emerald-300",
	"3xx": "text-amber-300",
	"4xx": "text-orange-300",
	"5xx": "text-rose-300",
	unknown: "text-zinc-300"
};

export const FILTER_CHIP_STATE_CLASSES = {
	active: "bg-sky-500/20 text-sky-100",
	inactive: "bg-slate-800/70 text-slate-300 hover:bg-slate-700/80 hover:text-slate-100"
} as const;

export const PAUSE_BUTTON_STATE_CLASSES = {
	active: "bg-amber-500/20 text-amber-100 hover:bg-amber-400/25",
	inactive: "bg-slate-800/80 text-slate-100 hover:bg-slate-700/80"
} as const;

export const TOAST_STATE_CLASSES = {
	visible: "translate-y-0 opacity-100",
	hidden: "pointer-events-none -translate-y-1 opacity-0"
} as const;

export const TAB_STATE_CLASSES = {
	active: "bg-sky-500/20 text-sky-100",
	inactive: "bg-slate-800/70 text-slate-300 hover:bg-slate-700/80 hover:text-slate-100"
} as const;

export const LOG_ROW_STATE_CLASSES = {
	selected: "bg-sky-500/10",
	fresh: "bg-emerald-500/10 animate-pulse",
	idle: "bg-slate-800/60 hover:-translate-y-px hover:bg-slate-700/75"
} as const;

export const CONTROL_BUTTON_BASE_CLASSES =
	"inline-flex items-center justify-center rounded-md px-3 py-1.5 text-[11px] font-medium uppercase tracking-[0.08em] transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50 disabled:cursor-not-allowed disabled:opacity-50";

export const TINY_BUTTON_BASE_CLASSES =
	"inline-flex items-center justify-center rounded-md bg-slate-800/80 px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-200 transition hover:-translate-y-px hover:bg-slate-700/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50 disabled:cursor-not-allowed disabled:opacity-50";
