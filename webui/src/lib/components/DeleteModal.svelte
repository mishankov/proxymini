<svelte:options runes={true} />

<script lang="ts">
	import { CONTROL_BUTTON_BASE_CLASSES } from "$lib/ui-classes";
	import { createEventDispatcher, tick } from "svelte";

	type Props = {
		open?: boolean;
	};

	let { open = false }: Props = $props();
	let confirmation = $state("");
	let inputRef = $state<HTMLInputElement | null>(null);

	const dispatch = createEventDispatcher<{
		confirm: string;
		cancel: void;
	}>();

	$effect(() => {
		if (!open) {
			return;
		}

		confirmation = "";
		void tick().then(() => inputRef?.focus());
	});

	function onBackdropClick(event: MouseEvent): void {
		if (event.target === event.currentTarget) {
			dispatch("cancel");
		}
	}
</script>

{#if open}
	<div
		class="fixed inset-0 z-30 flex items-center justify-center bg-slate-950/80 p-5 backdrop-blur-sm"
		role="dialog"
		aria-modal="true"
		aria-label="Delete logs confirmation"
		tabindex="-1"
		onclick={onBackdropClick}
		onkeydown={(event) => {
			if (event.key === "Escape") {
				dispatch("cancel");
			}
		}}
	>
		<div class="w-full max-w-md rounded-xl bg-slate-900/95 p-4 shadow-xl">
			<h2 class="mb-2 text-sm font-semibold uppercase tracking-[0.08em] text-rose-200">Confirm Destructive Action</h2>
			<p class="mb-3 text-xs leading-6 text-slate-300">
				This will permanently delete all request logs. Type <strong>DELETE</strong> to proceed.
			</p>
			<label for="deleteConfirmInput" class="sr-only">Type DELETE to confirm</label>
			<input
				id="deleteConfirmInput"
				type="text"
				autocomplete="off"
				placeholder="Type DELETE"
				class="mb-3 w-full rounded-md bg-rose-500/10 px-3 py-2 font-mono text-sm text-rose-100 placeholder:text-rose-200/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-rose-300/50"
				bind:value={confirmation}
				bind:this={inputRef}
				onkeydown={(event) => {
					if (event.key === "Enter") {
						dispatch("confirm", confirmation.trim());
					}
					if (event.key === "Escape") {
						dispatch("cancel");
					}
				}}
			/>
			<div class="flex justify-end gap-2">
				<button
					type="button"
					class={`${CONTROL_BUTTON_BASE_CLASSES} bg-slate-800/80 text-slate-100 hover:bg-slate-700/80`}
					onclick={() => dispatch("cancel")}
				>
					Cancel
				</button>
				<button
					type="button"
					class={`${CONTROL_BUTTON_BASE_CLASSES} bg-rose-500/25 text-rose-100 hover:bg-rose-400/30`}
					onclick={() => dispatch("confirm", confirmation.trim())}
				>
					Delete Logs
				</button>
			</div>
		</div>
	</div>
{/if}
