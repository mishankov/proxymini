<svelte:options runes={true} />

<script lang="ts">
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
		class="modal-wrap open"
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
		<div class="modal">
			<h2>Confirm Destructive Action</h2>
			<p>This will permanently delete all request logs. Type <strong>DELETE</strong> to proceed.</p>
			<label for="deleteConfirmInput" class="sr-only">Type DELETE to confirm</label>
			<input
				id="deleteConfirmInput"
				type="text"
				autocomplete="off"
				placeholder="Type DELETE"
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
			<div class="modal-actions">
				<button type="button" class="btn" onclick={() => dispatch("cancel")}>Cancel</button>
				<button type="button" class="btn warning" onclick={() => dispatch("confirm", confirmation.trim())}>
					Delete Logs
				</button>
			</div>
		</div>
	</div>
{/if}
