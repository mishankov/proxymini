<svelte:options runes={true} />

<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { LOG_ROW_STATE_CLASSES, STATUS_PILL_CLASSES } from "$lib/ui-classes";
	import type { EnrichedLog } from "$lib/types";
	import { highlightText } from "$lib/utils";

	type Props = {
		logs?: EnrichedLog[];
		selectedLogId?: string | null;
		renderLimit?: number;
		totalCount?: number;
		search?: string;
		recentlyAddedIds?: Set<string>;
	};

	let {
		logs = [],
		selectedLogId = null,
		renderLimit = 500,
		totalCount = 0,
		search = "",
		recentlyAddedIds = new Set<string>()
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		select: string;
		showMore: void;
	}>();

	const rendered = $derived(logs.slice(0, renderLimit));

	function rowStateClass(log: EnrichedLog): string {
		if (selectedLogId === log.id) {
			return LOG_ROW_STATE_CLASSES.selected;
		}
		if (recentlyAddedIds.has(log.id)) {
			return LOG_ROW_STATE_CLASSES.fresh;
		}
		return LOG_ROW_STATE_CLASSES.idle;
	}
</script>

<section class="flex min-h-0 flex-col overflow-hidden rounded-xl bg-slate-900/80 shadow-lg">
	<div class="flex items-center justify-between gap-2 bg-slate-900/90 px-3 py-2">
		<h2 class="text-xs font-semibold uppercase tracking-[0.12em] text-slate-300">Event Stream</h2>
		<div class="font-mono text-[11px] text-slate-400">{logs.length} visible / {totalCount} total</div>
	</div>

	<div class="min-h-0 overflow-auto">
		{#if rendered.length === 0}
			<div class="px-6 py-10 text-center font-mono text-xs text-slate-400">No logs match active filters.</div>
		{:else}
			<ul class="flex flex-col gap-2 p-2">
				{#each rendered as log (log.id)}
					<li>
						<button
							type="button"
							class={`group grid min-h-16 w-full grid-cols-[auto_auto_minmax(0,1fr)] gap-2 rounded-lg px-3 py-2 text-left transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/40 md:grid-cols-[auto_auto_minmax(0,1fr)_auto] ${rowStateClass(log)}`}
							aria-label={"Select log " + log.id}
							onclick={() => dispatch("select", log.id)}
						>
							<span class="inline-flex min-w-12 items-center justify-center rounded-md bg-slate-700/60 px-2 py-1 font-mono text-[11px] font-semibold uppercase tracking-[0.06em] text-slate-100">
								{log.methodNormalized}
							</span>
							<span
								class={`inline-flex min-w-11 items-center justify-center rounded-md px-2 py-1 font-mono text-[11px] uppercase tracking-[0.06em] ${STATUS_PILL_CLASSES[log.statusClass]}`}
							>
								{log.status}
							</span>
							<div class="min-w-0">
								<p class="truncate font-mono text-xs text-slate-100">{@html highlightText(log.url || "(empty URL)", search)}</p>
								<p class="mt-1 truncate font-mono text-[11px] text-slate-400">
									{@html highlightText(log.proxyUrl || "(no proxy target)", search)}
								</p>
							</div>
							<span class="col-start-3 row-start-2 justify-self-end font-mono text-[11px] text-slate-400 md:col-start-4 md:row-start-1">
								{log.timeFormatted}
							</span>
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<div class="flex justify-center p-2">
		{#if logs.length > renderLimit}
			<button
				type="button"
				class="inline-flex min-w-48 items-center justify-center rounded-md bg-slate-800/80 px-3 py-1.5 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-100 transition hover:-translate-y-px hover:bg-slate-700/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50"
				onclick={() => dispatch("showMore")}
			>
				Show 250 More
			</button>
		{/if}
	</div>
</section>
