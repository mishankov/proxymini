<svelte:options runes={true} />

<script lang="ts">
	import { createEventDispatcher } from "svelte";
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
</script>

<section class="panel">
	<div class="panel-head">
		<h2 class="panel-title">Event Stream</h2>
		<div class="panel-meta">{logs.length} visible / {totalCount} total</div>
	</div>

	<div class="list-wrap">
		{#if rendered.length === 0}
			<div class="empty-state">No logs match active filters.</div>
		{:else}
			<ul class="log-list">
				{#each rendered as log, index (log.id)}
					<li>
						<button
							type="button"
							class="log-item {selectedLogId === log.id ? 'selected' : ''} {recentlyAddedIds.has(log.id) ? 'new-entry' : ''}"
							style={`animation-delay: ${Math.min(index, 18) * 18}ms;`}
							aria-label={"Select log " + log.id}
							onclick={() => dispatch("select", log.id)}
						>
							<span class="method-pill">{log.methodNormalized}</span>
							<span class="status-pill status-{log.statusClass}">{log.status}</span>
							<div class="log-main">
								<p class="log-url">{@html highlightText(log.url || "(empty URL)", search)}</p>
								<p class="log-sub">{@html highlightText(log.proxyUrl || "(no proxy target)", search)}</p>
							</div>
							<span class="log-time">{log.timeFormatted}</span>
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<div class="list-foot">
		{#if logs.length > renderLimit}
			<button type="button" class="btn show-more" onclick={() => dispatch("showMore")}>Show 250 More</button>
		{/if}
	</div>
</section>
