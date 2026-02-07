<svelte:options runes={true} />

<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { highlightText, prettyBody } from "$lib/utils";

	type Props = {
		title: string;
		value?: string;
		search?: string;
		copyMessage?: string;
		emptyLabel?: string;
	};

	let {
		title,
		value = "",
		search = "",
		copyMessage = "Copied",
		emptyLabel = "(empty)"
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		copy: {
			value: string;
			message: string;
		};
	}>();

	let expanded = $state(false);
	const formatted = $derived(prettyBody(value));
	const lines = $derived(formatted.text.split("\n").length);
	const isLong = $derived(formatted.text.length > 900 || lines > 18);
	const rendered = $derived(formatted.text || emptyLabel);
</script>

<section class="payload-section">
	<div class="payload-head">
		<h3 class="payload-title">{title}</h3>
		<div class="payload-actions">
			<button
				type="button"
				class="tiny-btn"
				onclick={() => dispatch("copy", { value: value ?? "", message: copyMessage })}
			>
				Copy
			</button>
			{#if isLong}
				<button type="button" class="tiny-btn" onclick={() => (expanded = !expanded)}>
					{expanded ? "Collapse" : "Expand"}
				</button>
			{/if}
		</div>
	</div>

	<pre class="payload-content {isLong && !expanded ? 'clamped' : ''}">{@html highlightText(rendered, search)}</pre>
</section>
