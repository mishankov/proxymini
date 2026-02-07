<svelte:options runes={true} />

<script lang="ts">
	import DeleteModal from "$lib/components/DeleteModal.svelte";
	import Inspector from "$lib/components/Inspector.svelte";
	import LogList from "$lib/components/LogList.svelte";
	import StatusStrip from "$lib/components/StatusStrip.svelte";
	import TopBar from "$lib/components/TopBar.svelte";
	import { DEFAULT_SORT, INITIAL_RENDER_LIMIT, POLL_INTERVAL_MS, RENDER_STEP, TAB_OPTIONS } from "$lib/constants";
	import { TOAST_STATE_CLASSES } from "$lib/ui-classes";
	import type { EnrichedLog, InspectorTab, RequestLog, SortOption, StatusFilter } from "$lib/types";
	import { dedupeByID, enrichLog, normalizeText } from "$lib/utils";
	import { onMount } from "svelte";

	let allLogs = $state<EnrichedLog[]>([]);
	let visibleLogs = $state<EnrichedLog[]>([]);
	let selectedLogId = $state<string | null>(null);
	let sort = $state<SortOption>(DEFAULT_SORT);
	let isPaused = $state(false);
	let incomingLogs = $state<EnrichedLog[]>([]);
	let recentlyAddedIds = $state(new Set<string>());
	let renderLimit = $state(INITIAL_RENDER_LIMIT);
	let activeTab = $state<InspectorTab>("overview");
	let lastSeenTime = $state(0);

	let showDeleteModal = $state(false);
	let toastVisible = $state(false);
	let toastMessage = $state("");

	let searchQuery = $state("");
	let selectedMethods = $state(new Set<string>());
	let selectedStatuses = $state(new Set<StatusFilter>());

	let pollTimer: ReturnType<typeof setInterval> | undefined;
	let toastTimer: ReturnType<typeof setTimeout> | undefined;

	const selectedLog = $derived(visibleLogs.find((log) => log.id === selectedLogId) ?? null);
	const incomingCount = $derived(incomingLogs.length);
	const visibleCount = $derived(visibleLogs.length);
	const showingCount = $derived(Math.min(visibleCount, renderLimit));
	const summaryText = $derived(
		`${visibleCount} matched · showing ${showingCount} · polling ${isPaused ? "paused" : "every 3s"}`
	);
	const activeFilters = $derived.by(() => {
		const parts: string[] = [];
		if (searchQuery.trim()) {
			parts.push(`q="${searchQuery.trim()}"`);
		}
		if (selectedMethods.size > 0) {
			parts.push(`method:${Array.from(selectedMethods.values()).join(", ")}`);
		}
		if (selectedStatuses.size > 0) {
			parts.push(`status:${Array.from(selectedStatuses.values()).join(", ")}`);
		}
		return parts.length > 0 ? parts.join(" | ") : "No active filters";
	});

	function keepSelectionValid(): void {
		if (visibleLogs.length === 0) {
			selectedLogId = null;
			return;
		}

		if (!selectedLogId) {
			selectedLogId = visibleLogs[0].id;
			return;
		}

		const exists = visibleLogs.some((log) => log.id === selectedLogId);
		if (!exists) {
			selectedLogId = visibleLogs[0].id;
		}
	}

	function applyFiltersAndSort(): void {
		const search = normalizeText(searchQuery.trim());

		let filtered = allLogs.filter((log) => {
			if (selectedMethods.size > 0 && !selectedMethods.has(log.methodNormalized)) {
				return false;
			}

			if (selectedStatuses.size > 0) {
				if (log.statusClass === "unknown") {
					return false;
				}
				if (!selectedStatuses.has(log.statusClass)) {
					return false;
				}
			}

			if (search && !log.searchBlob.includes(search)) {
				return false;
			}

			return true;
		});

		if (sort === "timeAsc") {
			filtered = filtered.sort((a, b) => a.time - b.time);
		} else if (sort === "statusDesc") {
			filtered = filtered.sort((a, b) => b.status - a.status);
		} else {
			filtered = filtered.sort((a, b) => b.time - a.time);
		}

		visibleLogs = filtered;
		keepSelectionValid();
	}

	function mergeFetchedLogs(enriched: EnrichedLog[], manual: boolean): void {
		const currentIds = new Set(allLogs.map((log) => log.id));
		const incoming = enriched.filter((log) => !currentIds.has(log.id));

		if (isPaused && !manual) {
			incomingLogs = incoming;
			return;
		}

		const previousIDs = new Set(allLogs.map((log) => log.id));
		recentlyAddedIds = new Set(enriched.filter((log) => !previousIDs.has(log.id)).map((log) => log.id));

		allLogs = enriched;
		incomingLogs = [];
		lastSeenTime = enriched.reduce((max, log) => Math.max(max, log.time), lastSeenTime);
	}

	async function fetchLogs(manual = false): Promise<void> {
		try {
			const response = await fetch("/api/logs");
			if (!response.ok) {
				throw new Error(`Network response was not ok ${response.statusText}`);
			}

			const payload = await response.json();
			const logs = (Array.isArray(payload) ? payload : []) as RequestLog[];
			const enriched = logs.map(enrichLog);

			mergeFetchedLogs(enriched, manual);
			applyFiltersAndSort();

			if (manual) {
				showToast("Logs refreshed");
			}
		} catch (error) {
			console.error("Failed to fetch logs", error);
			showToast("Failed to fetch logs");
		}
	}

	function setSearch(value: string): void {
		searchQuery = value;
		renderLimit = INITIAL_RENDER_LIMIT;
		applyFiltersAndSort();
	}

	function toggleMethod(value: string): void {
		const next = new Set(selectedMethods);
		if (next.has(value)) {
			next.delete(value);
		} else {
			next.add(value);
		}

		selectedMethods = next;
		renderLimit = INITIAL_RENDER_LIMIT;
		applyFiltersAndSort();
	}

	function toggleStatus(value: StatusFilter): void {
		const next = new Set(selectedStatuses);
		if (next.has(value)) {
			next.delete(value);
		} else {
			next.add(value);
		}

		selectedStatuses = next;
		renderLimit = INITIAL_RENDER_LIMIT;
		applyFiltersAndSort();
	}

	function setSort(nextSort: SortOption): void {
		sort = nextSort;
		applyFiltersAndSort();
	}

	function setActiveTab(tab: InspectorTab): void {
		activeTab = tab;
	}

	function moveSelection(delta: number): void {
		if (visibleLogs.length === 0) {
			return;
		}

		const currentIndex = visibleLogs.findIndex((log) => log.id === selectedLogId);
		const index = currentIndex < 0 ? 0 : currentIndex;
		const next = Math.max(0, Math.min(visibleLogs.length - 1, index + delta));
		selectedLogId = visibleLogs[next].id;

		if (next >= renderLimit) {
			renderLimit = Math.min(visibleLogs.length, renderLimit + RENDER_STEP);
		}
	}

	function cycleTab(direction: number): void {
		const current = TAB_OPTIONS.indexOf(activeTab);
		const next = (current + direction + TAB_OPTIONS.length) % TAB_OPTIONS.length;
		activeTab = TAB_OPTIONS[next];
	}

	function showMore(): void {
		renderLimit += RENDER_STEP;
	}

	function togglePause(): void {
		isPaused = !isPaused;

		if (!isPaused && incomingLogs.length > 0) {
			allLogs = dedupeByID([...incomingLogs, ...allLogs]);
			incomingLogs = [];
			applyFiltersAndSort();
		}
	}

	function applyIncoming(): void {
		if (incomingLogs.length === 0) {
			return;
		}

		allLogs = dedupeByID([...incomingLogs, ...allLogs]);
		incomingLogs = [];
		applyFiltersAndSort();
	}

	function openDeleteModal(): void {
		showDeleteModal = true;
	}

	function closeDeleteModal(): void {
		showDeleteModal = false;
	}

	async function confirmDelete(confirmation: string): Promise<void> {
		if (confirmation !== "DELETE") {
			showToast("Type DELETE to confirm");
			return;
		}

		try {
			const response = await fetch("/api/logs", { method: "DELETE" });
			if (!response.ok) {
				throw new Error(`Network response was not ok ${response.statusText}`);
			}

			allLogs = [];
			visibleLogs = [];
			selectedLogId = null;
			incomingLogs = [];
			recentlyAddedIds = new Set<string>();
			lastSeenTime = 0;
			showDeleteModal = false;
			showToast("Logs deleted");
			await fetchLogs(true);
		} catch (error) {
			console.error("Failed to delete logs", error);
			showToast("Failed to delete logs");
		}
	}

	async function copyToClipboard(value: string, message: string): Promise<void> {
		try {
			await navigator.clipboard.writeText(value);
			showToast(message);
		} catch (error) {
			console.error("Copy failed", error);
			showToast("Copy failed");
		}
	}

	function showToast(message: string): void {
		toastMessage = message;
		toastVisible = true;

		if (toastTimer) {
			clearTimeout(toastTimer);
		}
		toastTimer = setTimeout(() => {
			toastVisible = false;
		}, 1800);
	}

	function installKeyboardShortcuts(): () => void {
		const handler = (event: KeyboardEvent): void => {
			const target = event.target as HTMLElement | null;
			const tag = target?.tagName.toLowerCase() ?? "";
			const typing = tag === "input" || tag === "textarea";

			if (event.key === "Escape" && showDeleteModal) {
				closeDeleteModal();
				return;
			}

			if (!typing && event.key === "/") {
				event.preventDefault();
				const searchInput = document.getElementById("searchInput") as HTMLInputElement | null;
				searchInput?.focus();
				searchInput?.select();
				return;
			}

			if (typing) {
				return;
			}

			if (event.key === "j") {
				event.preventDefault();
				moveSelection(1);
			} else if (event.key === "k") {
				event.preventDefault();
				moveSelection(-1);
			} else if (event.key === "[") {
				event.preventDefault();
				cycleTab(-1);
			} else if (event.key === "]") {
				event.preventDefault();
				cycleTab(1);
			} else if (event.key.toLowerCase() === "p") {
				event.preventDefault();
				togglePause();
			}
		};

		window.addEventListener("keydown", handler);
		return () => window.removeEventListener("keydown", handler);
	}

	onMount(() => {
		void fetchLogs(true);
		pollTimer = setInterval(() => {
			void fetchLogs(false);
		}, POLL_INTERVAL_MS);

		const uninstallHotkeys = installKeyboardShortcuts();

		return () => {
			uninstallHotkeys();
			if (pollTimer) {
				clearInterval(pollTimer);
			}
			if (toastTimer) {
				clearTimeout(toastTimer);
			}
		};
	});
</script>

<div class="relative z-10 grid min-h-screen grid-rows-[auto_1fr_auto] gap-3 p-2 pb-20 sm:p-4 sm:pb-24">
	<TopBar
		search={searchQuery}
		selectedMethods={selectedMethods}
		selectedStatuses={selectedStatuses}
		{sort}
		{isPaused}
		{incomingCount}
		on:searchChange={(event) => setSearch(event.detail)}
		on:toggleMethod={(event) => toggleMethod(event.detail)}
		on:toggleStatus={(event) => toggleStatus(event.detail)}
		on:togglePause={togglePause}
		on:refresh={() => fetchLogs(true)}
		on:sortChange={(event) => setSort(event.detail)}
		on:openDelete={openDeleteModal}
		on:applyIncoming={applyIncoming}
	/>

	<main class="grid min-h-0 gap-3 xl:grid-cols-[minmax(22rem,0.95fr)_minmax(28rem,1.05fr)]">
		<LogList
			logs={visibleLogs}
			{selectedLogId}
			{renderLimit}
			totalCount={allLogs.length}
			search={searchQuery}
			{recentlyAddedIds}
			on:select={(event) => (selectedLogId = event.detail)}
			on:showMore={showMore}
		/>

		<Inspector
			selected={selectedLog}
			activeTab={activeTab}
			search={searchQuery}
			on:tabChange={(event) => setActiveTab(event.detail)}
			on:copyLog={() => selectedLog && copyToClipboard(JSON.stringify(selectedLog, null, 2), "Full log copied")}
			on:copyValue={(event) => copyToClipboard(event.detail.value, event.detail.message)}
		/>
	</main>

	<StatusStrip {summaryText} activeFilters={activeFilters} />

	<DeleteModal open={showDeleteModal} on:cancel={closeDeleteModal} on:confirm={(event) => confirmDelete(event.detail)} />

	<div
		class={`fixed top-3 right-3 z-40 rounded-md bg-slate-900/95 px-3 py-2 font-mono text-xs text-slate-100 shadow-md transition ${toastVisible ? TOAST_STATE_CLASSES.visible : TOAST_STATE_CLASSES.hidden}`}
		role="status"
		aria-live="polite"
	>
		{toastMessage}
	</div>
</div>
