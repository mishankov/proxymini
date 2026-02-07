<svelte:options runes={true} />

<script lang="ts">
	import { METHOD_OPTIONS, STATUS_OPTIONS } from "$lib/constants";
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
</script>

<header class="topbar">
	<div class="title-block">
		<div class="title-row">
			<h1 class="title">ProxyMini Ops Console</h1>
			{#if incomingCount > 0}
				<button
					type="button"
					class="badge new-logs"
					aria-label="Apply incoming logs"
					onclick={() => dispatch("applyIncoming")}
				>
					{incomingCount} incoming
				</button>
			{/if}
		</div>
		<div class="subtitle">request telemetry · sveltekit static · tailwind css</div>
		<div class="search-wrap">
			<span>/</span>
			<input
				id="searchInput"
				type="text"
				placeholder="Search URL, proxy, headers, body"
				aria-label="Search logs"
				value={search}
				oninput={(event) => dispatch("searchChange", (event.currentTarget as HTMLInputElement).value)}
			/>
		</div>
	</div>

	<div class="filter-blocks">
		<div class="filter-row">
			<span class="filter-label">Method</span>
			<div class="chip-group" aria-label="Method filters">
				{#each METHOD_OPTIONS as method}
					<button
						type="button"
						class:active={selectedMethods.has(method)}
						class="chip"
						aria-label={"Toggle " + method + " filter"}
						aria-pressed={selectedMethods.has(method)}
						onclick={() => dispatch("toggleMethod", method)}
					>
						{method}
					</button>
				{/each}
			</div>
		</div>

		<div class="filter-row">
			<span class="filter-label">Status</span>
			<div class="chip-group" aria-label="Status filters">
				{#each STATUS_OPTIONS as status}
					<button
						type="button"
						class:active={selectedStatuses.has(status)}
						class="chip"
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

	<div class="actions">
		<button
			type="button"
			class:active={isPaused}
			class="btn pause"
			aria-label={isPaused ? "Resume auto refresh" : "Pause auto refresh"}
			onclick={() => dispatch("togglePause")}
		>
			{isPaused ? "Resume" : "Pause"}
		</button>
		<button type="button" class="btn primary" aria-label="Refresh logs now" onclick={() => dispatch("refresh")}>
			Refresh
		</button>
		<label for="sortSelect" class="sr-only">Sort logs</label>
		<select id="sortSelect" class="select" aria-label="Sort logs" value={sort} onchange={emitSort}>
			<option value="timeDesc">Newest first</option>
			<option value="timeAsc">Oldest first</option>
			<option value="statusDesc">Highest status first</option>
		</select>
		<button type="button" class="btn warning" aria-label="Delete all logs" onclick={() => dispatch("openDelete")}>
			Delete All
		</button>
	</div>
</header>
