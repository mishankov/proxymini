<svelte:options runes={true} />

<script lang="ts">
	import { TINY_BUTTON_BASE_CLASSES } from "$lib/ui-classes";
	import { detectBodySyntax, formatBodyForDisplay, highlightText, renderPayloadHtml } from "$lib/utils";
	import { createEventDispatcher } from "svelte";

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

<section class="overflow-hidden rounded-lg bg-slate-800/50">
	<div class="flex items-center justify-between gap-2 bg-slate-800/80 px-3 py-2">
		<h3 class="font-mono text-xs uppercase tracking-[0.08em] text-slate-300">{title}</h3>
		<div class="flex flex-wrap gap-2">
			<button
				type="button"
				class={TINY_BUTTON_BASE_CLASSES}
				onclick={() => dispatch("copy", { value: value ?? "", message: copyMessage })}
			>
				Copy
			</button>
			{#if isLong}
				<button type="button" class={TINY_BUTTON_BASE_CLASSES} onclick={() => (expanded = !expanded)}>
					{expanded ? "Collapse" : "Expand"}
				</button>
			{/if}
		</div>
	</div>

	<pre
		class={`overflow-auto whitespace-pre-wrap break-words p-3 font-mono text-xs leading-6 text-slate-100 ${isLong && !expanded ? "max-h-48" : "max-h-[30rem]"}`}
	>{@html renderedHtml}</pre>
</section>
