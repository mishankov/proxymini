<svelte:options runes={true} />

<script lang="ts">
	type Props = {
		summaryText?: string;
		activeFilters?: string;
	};

	let { summaryText = "", activeFilters = "" }: Props = $props();
</script>

<footer
	class="fixed right-2 bottom-2 left-2 z-10 flex flex-col items-start gap-1 rounded-lg bg-slate-900/90 px-3 py-2 font-mono text-[11px] text-slate-300 shadow-md backdrop-blur sm:flex-row sm:items-center sm:justify-between"
>
	<div>{summaryText}</div>
	<div class="w-full truncate text-left text-slate-400 sm:w-auto sm:max-w-[55%] sm:text-right">{activeFilters}</div>
</footer>
