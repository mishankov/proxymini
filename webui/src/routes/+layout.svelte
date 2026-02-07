<svelte:options runes={true} />

<script lang="ts">
	import "../app.css";

	let { children } = $props();
</script>

{@render children()}
