<svelte:options runes={true} />

<script lang="ts">
	import { TAB_OPTIONS } from "$lib/constants";
	import type { EnrichedLog, InspectorTab } from "$lib/types";
	import { createEventDispatcher } from "svelte";
	import PayloadPanel from "$lib/components/PayloadPanel.svelte";
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
</script>

<section class="panel inspector">
	<div class="inspector-head">
		<div>
			<h2 class="inspector-title">
				{#if selected}
					{selected.methodNormalized} {selected.url}
				{:else}
					Select a request log
				{/if}
			</h2>
			{#if selected}
				<div class="inspector-grid">
					<dl class="kv">
						<dt>Time</dt>
						<dd>{selected.timeFormatted}</dd>
					</dl>
					<dl class="kv">
						<dt>Status</dt>
						<dd class="status-{selected.statusClass}">{selected.status}</dd>
					</dl>
					<dl class="kv">
						<dt>Proxy URL</dt>
						<dd>{@html highlightText(selected.proxyUrl || "-", search)}</dd>
					</dl>
					<dl class="kv">
						<dt>ID</dt>
						<dd>{selected.id}</dd>
					</dl>
				</div>
			{/if}
		</div>
		<div class="payload-actions">
			<button type="button" class="tiny-btn" disabled={!selected} onclick={() => dispatch("copyLog")}>Copy Log JSON</button>
		</div>
	</div>

	<div class="tabs" role="tablist" aria-label="Inspector tabs">
		{#each TAB_OPTIONS as tab}
			<button
				type="button"
				class="tab {activeTab === tab ? 'active' : ''}"
				role="tab"
				aria-selected={activeTab === tab}
				onclick={() => dispatch("tabChange", tab)}
			>
				{tab}
			</button>
		{/each}
	</div>

	<div class="tab-panel-wrap">
		{#if !selected}
			<div class="empty-state">No log selected.</div>
		{:else if activeTab === "overview"}
			<section class="tab-panel active">
				<div class="headers-grid">
					<div class="header-item">
						<p class="header-key">Route</p>
						<p class="header-value">{@html highlightText(selected.url || "-", search)}</p>
					</div>
					<div class="header-item">
						<p class="header-key">Target</p>
						<p class="header-value">{@html highlightText(selected.proxyUrl || "-", search)}</p>
					</div>
					<div class="header-item">
						<p class="header-key">Request Body Size</p>
						<p class="header-value">{requestBodySize(selected)} bytes</p>
					</div>
					<div class="header-item">
						<p class="header-key">Response Body Size</p>
						<p class="header-value">{responseBodySize(selected)} bytes</p>
					</div>
					<div class="header-item">
						<p class="header-key">Request Headers</p>
						<p class="header-value">{selected.requestHeadersEntries.length} keys</p>
					</div>
					<div class="header-item">
						<p class="header-key">Response Headers</p>
						<p class="header-value">{selected.responseHeadersEntries.length} keys</p>
					</div>
				</div>
			</section>
		{:else if activeTab === "request"}
			<section class="tab-panel active">
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
			<section class="tab-panel active">
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
			<section class="tab-panel active">
				<div class="payload-section">
					<div class="payload-head">
						<h3 class="payload-title">Request Headers</h3>
						<div class="payload-actions">
							<button
								type="button"
								class="tiny-btn"
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
					<div class="headers-grid pad">
						{#if selected.requestHeadersEntries.length === 0}
							<div class="empty-state">No request headers.</div>
						{:else}
							{#each selected.requestHeadersEntries as header}
								<div class="header-item">
									<p class="header-key">{@html highlightText(header.key, search)}</p>
									<p class="header-value">{@html highlightText(header.value, search)}</p>
								</div>
							{/each}
						{/if}
					</div>
				</div>

				<div class="payload-section">
					<div class="payload-head">
						<h3 class="payload-title">Response Headers</h3>
						<div class="payload-actions">
							<button
								type="button"
								class="tiny-btn"
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
					<div class="headers-grid pad">
						{#if selected.responseHeadersEntries.length === 0}
							<div class="empty-state">No response headers.</div>
						{:else}
							{#each selected.responseHeadersEntries as header}
								<div class="header-item">
									<p class="header-key">{@html highlightText(header.key, search)}</p>
									<p class="header-value">{@html highlightText(header.value, search)}</p>
								</div>
							{/each}
						{/if}
					</div>
				</div>
			</section>
		{:else}
			<section class="tab-panel active">
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
