<svelte:options runes={true} />

<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { detectBodySyntax, formatBodyForDisplay, highlightText, renderPayloadHtml } from "$lib/utils";

	type Props = {
		title: string;
		value?: string;
		contentType?: string;
		search?: string;
		copyMessage?: string;
		emptyLabel?: string;
	};

	let {
		title,
		value = "",
		contentType = "",
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
	const syntax = $derived(detectBodySyntax(value, contentType));
	const formattedText = $derived(formatBodyForDisplay(value, syntax));
	const lines = $derived(formattedText.split("\n").length);
	const isLong = $derived(formattedText.length > 900 || lines > 18);
	const renderedHtml = $derived(value ? renderPayloadHtml(value, search, contentType) : highlightText(emptyLabel, search));
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

	<pre class="payload-content {isLong && !expanded ? 'clamped' : ''}">{@html renderedHtml}</pre>
</section>
