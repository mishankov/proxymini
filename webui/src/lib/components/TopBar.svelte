<svelte:options runes={true} />

<script lang="ts">
	import { METHOD_OPTIONS, STATUS_OPTIONS } from "$lib/constants";
	import {
		CONTROL_BUTTON_BASE_CLASSES,
		FILTER_CHIP_STATE_CLASSES,
		PAUSE_BUTTON_STATE_CLASSES
	} from "$lib/ui-classes";
	import type { SortOption, StatusFilter } from "$lib/types";
	import { createEventDispatcher } from "svelte";

	type Props = {
		search?: string;
		selectedMethods?: Set<string>;
		selectedStatuses?: Set<StatusFilter>;
		sort: SortOption;
		isPaused?: boolean;
		incomingCount?: number;
	};

	let {
		search = "",
		selectedMethods = new Set<string>(),
		selectedStatuses = new Set<StatusFilter>(),
		sort,
		isPaused = false,
		incomingCount = 0
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		searchChange: string;
		toggleMethod: string;
		toggleStatus: StatusFilter;
		togglePause: void;
		refresh: void;
		sortChange: SortOption;
		openDelete: void;
		applyIncoming: void;
	}>();

	function emitSort(event: Event): void {
		const select = event.currentTarget as HTMLSelectElement;
		dispatch("sortChange", select.value as SortOption);
	}

	const chipBaseClass =
		"inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50";
</script>

<header
	class="grid gap-3 rounded-xl bg-slate-900/80 p-3 shadow-lg backdrop-blur lg:grid-cols-[minmax(18rem,1.1fr)_minmax(16rem,0.9fr)_auto]"
>
	<div class="min-w-0 space-y-3">
		<div class="flex min-w-0 items-center gap-3">
			<h1 class="truncate text-lg font-semibold uppercase tracking-[0.12em] text-slate-50">ProxyMini</h1>
			{#if incomingCount > 0}
				<button
					type="button"
					class="inline-flex items-center rounded-full bg-sky-500/15 px-2.5 py-1 text-[11px] font-mono uppercase tracking-[0.08em] text-sky-100 transition hover:-translate-y-px hover:bg-sky-400/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50"
					aria-label="Apply incoming logs"
					onclick={() => dispatch("applyIncoming")}
				>
					{incomingCount} incoming
				</button>
			{/if}
		</div>
		<div class="flex items-center gap-2 rounded-lg bg-slate-800/80 px-3 py-2 focus-within:ring-2 focus-within:ring-sky-300/40">
			<span class="font-mono text-xs text-slate-400">/</span>
			<input
				id="searchInput"
				type="text"
				placeholder="Search URL, proxy, headers, body"
				aria-label="Search logs"
				class="w-full bg-transparent font-mono text-sm text-slate-100 placeholder:text-slate-500 focus:outline-none"
				value={search}
				oninput={(event) => dispatch("searchChange", (event.currentTarget as HTMLInputElement).value)}
			/>
		</div>
	</div>

	<div class="grid min-w-0 gap-2">
		<div class="flex min-w-0 flex-wrap items-start gap-2">
			<span class="pt-1 font-mono text-[11px] uppercase tracking-[0.1em] text-slate-400">Method</span>
			<div class="flex flex-wrap gap-2" aria-label="Method filters">
				{#each METHOD_OPTIONS as method}
					<button
						type="button"
						class={`${chipBaseClass} ${selectedMethods.has(method) ? FILTER_CHIP_STATE_CLASSES.active : FILTER_CHIP_STATE_CLASSES.inactive}`}
						aria-label={"Toggle " + method + " filter"}
						aria-pressed={selectedMethods.has(method)}
						onclick={() => dispatch("toggleMethod", method)}
					>
						{method}
					</button>
				{/each}
			</div>
		</div>

		<div class="flex min-w-0 flex-wrap items-start gap-2">
			<span class="pt-1 font-mono text-[11px] uppercase tracking-[0.1em] text-slate-400">Status</span>
			<div class="flex flex-wrap gap-2" aria-label="Status filters">
				{#each STATUS_OPTIONS as status}
					<button
						type="button"
						class={`${chipBaseClass} ${selectedStatuses.has(status) ? FILTER_CHIP_STATE_CLASSES.active : FILTER_CHIP_STATE_CLASSES.inactive}`}
						aria-label={"Toggle " + status + " filter"}
						aria-pressed={selectedStatuses.has(status)}
						onclick={() => dispatch("toggleStatus", status)}
					>
						{status}
					</button>
				{/each}
			</div>
		</div>
	</div>

	<div class="flex flex-wrap items-start justify-start gap-2 lg:justify-end">
		<button
			type="button"
			class={`${CONTROL_BUTTON_BASE_CLASSES} ${isPaused ? PAUSE_BUTTON_STATE_CLASSES.active : PAUSE_BUTTON_STATE_CLASSES.inactive}`}
			aria-label={isPaused ? "Resume auto refresh" : "Pause auto refresh"}
			onclick={() => dispatch("togglePause")}
		>
			{isPaused ? "Resume" : "Pause"}
		</button>
		<button
			type="button"
			class={`${CONTROL_BUTTON_BASE_CLASSES} bg-sky-500/20 text-sky-100 hover:bg-sky-400/25`}
			aria-label="Refresh logs now"
			onclick={() => dispatch("refresh")}
		>
			Refresh
		</button>
		<label for="sortSelect" class="sr-only">Sort logs</label>
		<select
			id="sortSelect"
			class="rounded-md bg-slate-800/80 px-2.5 py-1.5 font-mono text-xs text-slate-100 transition hover:bg-slate-700/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50"
			aria-label="Sort logs"
			value={sort}
			onchange={emitSort}
		>
			<option value="timeDesc">Newest first</option>
			<option value="timeAsc">Oldest first</option>
			<option value="statusDesc">Highest status first</option>
		</select>
		<button
			type="button"
			class={`${CONTROL_BUTTON_BASE_CLASSES} bg-rose-500/20 text-rose-100 hover:bg-rose-400/25`}
			aria-label="Delete all logs"
			onclick={() => dispatch("openDelete")}
		>
			Delete All
		</button>
	</div>
</header>
