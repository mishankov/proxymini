<svelte:options runes={true} />

<script lang="ts">
	import { TAB_OPTIONS } from "$lib/constants";
	import PayloadPanel from "$lib/components/PayloadPanel.svelte";
	import { STATUS_TEXT_CLASSES, TAB_STATE_CLASSES, TINY_BUTTON_BASE_CLASSES } from "$lib/ui-classes";
	import type { EnrichedLog, InspectorTab } from "$lib/types";
	import { createEventDispatcher } from "svelte";
	import { highlightText, safeParseJSON } from "$lib/utils";

	type Props = {
		selected?: EnrichedLog | null;
		activeTab?: InspectorTab;
		search?: string;
	};

	let { selected = null, activeTab = "overview", search = "" }: Props = $props();

	const dispatch = createEventDispatcher<{
		tabChange: InspectorTab;
		copyLog: void;
		copyValue: { value: string; message: string };
	}>();

	const canonicalRaw = $derived.by(() => {
		if (!selected) {
			return "";
		}

		return JSON.stringify(
			{
				id: selected.id,
				time: selected.time,
				method: selected.method,
				proxyUrl: selected.proxyUrl,
				url: selected.url,
				requestHeaders: safeParseJSON(selected.requestHeaders) ?? selected.requestHeaders,
				requestBody: selected.requestBody,
				status: selected.status,
				responseHeaders: safeParseJSON(selected.responseHeaders) ?? selected.responseHeaders,
				responseBody: selected.responseBody
			},
			null,
			2
		);
	});

	function findHeaderValue(entries: Array<{ key: string; value: string }>, headerName: string): string {
		const lookup = headerName.toLowerCase();
		const matched = entries.find((entry) => entry.key.toLowerCase() === lookup);
		return matched?.value ?? "";
	}

	const requestContentType = $derived(selected ? findHeaderValue(selected.requestHeadersEntries, "content-type") : "");
	const responseContentType = $derived(selected ? findHeaderValue(selected.responseHeadersEntries, "content-type") : "");

	function requestBodySize(log: EnrichedLog | null): number {
		return (log?.requestBody ?? "").length;
	}

	function responseBodySize(log: EnrichedLog | null): number {
		return (log?.responseBody ?? "").length;
	}

	const tabBaseClass =
		"inline-flex items-center rounded-md px-3 py-1.5 text-xs capitalize tracking-[0.03em] transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-300/50";
</script>

<section
	class="grid min-h-0 grid-rows-[auto_auto_1fr] overflow-hidden rounded-xl bg-slate-900/80 shadow-lg"
>
	<div class="flex items-start justify-between gap-3 bg-slate-900/90 px-3 py-3">
		<div class="min-w-0">
			<h2 class="truncate text-sm font-semibold tracking-[0.03em] text-slate-100">
				{#if selected}
					{selected.methodNormalized} {selected.url}
				{:else}
					Select a request log
				{/if}
			</h2>
			{#if selected}
				<div class="mt-3 grid gap-2 sm:grid-cols-2">
					<dl class="min-w-0">
						<dt class="font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Time</dt>
						<dd class="mt-1 font-mono text-xs text-slate-200">{selected.timeFormatted}</dd>
					</dl>
					<dl class="min-w-0">
						<dt class="font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Status</dt>
						<dd class={`mt-1 font-mono text-xs ${STATUS_TEXT_CLASSES[selected.statusClass]}`}>{selected.status}</dd>
					</dl>
					<dl class="min-w-0">
						<dt class="font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Proxy URL</dt>
						<dd class="mt-1 break-all font-mono text-xs text-slate-200">
							{@html highlightText(selected.proxyUrl || "-", search)}
						</dd>
					</dl>
					<dl class="min-w-0">
						<dt class="font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">ID</dt>
						<dd class="mt-1 break-all font-mono text-xs text-slate-200">{selected.id}</dd>
					</dl>
				</div>
			{/if}
		</div>
		<div class="flex shrink-0 gap-2">
			<button
				type="button"
				class={TINY_BUTTON_BASE_CLASSES}
				disabled={!selected}
				onclick={() => dispatch("copyLog")}
			>
				Copy Log JSON
			</button>
		</div>
	</div>

	<div class="flex flex-wrap gap-2 bg-slate-900/85 px-3 py-2" role="tablist" aria-label="Inspector tabs">
		{#each TAB_OPTIONS as tab}
			<button
				type="button"
				class={`${tabBaseClass} ${activeTab === tab ? TAB_STATE_CLASSES.active : TAB_STATE_CLASSES.inactive}`}
				role="tab"
				aria-selected={activeTab === tab}
				onclick={() => dispatch("tabChange", tab)}
			>
				{tab}
			</button>
		{/each}
	</div>

	<div class="min-h-0 overflow-auto p-3">
		{#if !selected}
			<div class="px-6 py-10 text-center font-mono text-xs text-slate-400">No log selected.</div>
		{:else if activeTab === "overview"}
			<section class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Route</p>
					<p class="break-all font-mono text-xs text-slate-100">{@html highlightText(selected.url || "-", search)}</p>
				</div>
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Target</p>
					<p class="break-all font-mono text-xs text-slate-100">{@html highlightText(selected.proxyUrl || "-", search)}</p>
				</div>
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Request Body Size</p>
					<p class="font-mono text-xs text-slate-100">{requestBodySize(selected)} bytes</p>
				</div>
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Response Body Size</p>
					<p class="font-mono text-xs text-slate-100">{responseBodySize(selected)} bytes</p>
				</div>
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Request Headers</p>
					<p class="font-mono text-xs text-slate-100">{selected.requestHeadersEntries.length} keys</p>
				</div>
				<div class="rounded-lg bg-slate-800/50 p-2">
					<p class="mb-1 font-mono text-[11px] uppercase tracking-[0.08em] text-slate-400">Response Headers</p>
					<p class="font-mono text-xs text-slate-100">{selected.responseHeadersEntries.length} keys</p>
				</div>
			</section>
		{:else if activeTab === "request"}
			<section>
				<PayloadPanel
					title="Request Body"
					value={selected.requestBody || ""}
					contentType={requestContentType}
					search={search}
					copyMessage="Request body copied"
					on:copy={(event) => dispatch("copyValue", event.detail)}
				/>
			</section>
		{:else if activeTab === "response"}
			<section>
				<PayloadPanel
					title="Response Body"
					value={selected.responseBody || ""}
					contentType={responseContentType}
					search={search}
					copyMessage="Response body copied"
					on:copy={(event) => dispatch("copyValue", event.detail)}
				/>
			</section>
		{:else if activeTab === "headers"}
			<section class="space-y-3">
				<div class="overflow-hidden rounded-lg bg-slate-800/50">
					<div class="flex items-center justify-between gap-2 bg-slate-800/80 px-3 py-2">
						<h3 class="font-mono text-xs uppercase tracking-[0.08em] text-slate-300">Request Headers</h3>
						<div class="flex gap-2">
							<button
								type="button"
								class={TINY_BUTTON_BASE_CLASSES}
								onclick={() =>
									dispatch("copyValue", {
										value: selected.requestHeaders || "",
										message: "Request headers copied"
									})}
							>
								Copy
							</button>
						</div>
					</div>
					<div class="grid gap-2 p-3">
						{#if selected.requestHeadersEntries.length === 0}
							<div class="px-4 py-8 text-center font-mono text-xs text-slate-400">No request headers.</div>
						{:else}
							{#each selected.requestHeadersEntries as header}
								<div class="rounded-lg bg-slate-800/60 p-2">
									<p class="mb-1 break-all font-mono text-[11px] text-sky-200">{@html highlightText(header.key, search)}</p>
									<p class="break-all font-mono text-xs text-slate-100">{@html highlightText(header.value, search)}</p>
								</div>
							{/each}
						{/if}
					</div>
				</div>

				<div class="overflow-hidden rounded-lg bg-slate-800/50">
					<div class="flex items-center justify-between gap-2 bg-slate-800/80 px-3 py-2">
						<h3 class="font-mono text-xs uppercase tracking-[0.08em] text-slate-300">Response Headers</h3>
						<div class="flex gap-2">
							<button
								type="button"
								class={TINY_BUTTON_BASE_CLASSES}
								onclick={() =>
									dispatch("copyValue", {
										value: selected.responseHeaders || "",
										message: "Response headers copied"
									})}
							>
								Copy
							</button>
						</div>
					</div>
					<div class="grid gap-2 p-3">
						{#if selected.responseHeadersEntries.length === 0}
							<div class="px-4 py-8 text-center font-mono text-xs text-slate-400">No response headers.</div>
						{:else}
							{#each selected.responseHeadersEntries as header}
								<div class="rounded-lg bg-slate-800/60 p-2">
									<p class="mb-1 break-all font-mono text-[11px] text-sky-200">{@html highlightText(header.key, search)}</p>
									<p class="break-all font-mono text-xs text-slate-100">{@html highlightText(header.value, search)}</p>
								</div>
							{/each}
						{/if}
					</div>
				</div>
			</section>
		{:else}
			<section>
				<PayloadPanel
					title="Canonical JSON"
					value={canonicalRaw}
					search={search}
					copyMessage="Log JSON copied"
					on:copy={(event) => dispatch("copyValue", event.detail)}
				/>
			</section>
		{/if}
	</div>
</section>
