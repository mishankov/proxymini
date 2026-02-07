<svelte:options runes={true} />

<script lang="ts">
	type Props = {
		summaryText?: string;
		activeFilters?: string;
	};

	let { summaryText = "", activeFilters = "" }: Props = $props();
</script>

<footer class="status-strip">
	<div>{summaryText}</div>
	<div class="status-active-filters">{activeFilters}</div>
</footer>
