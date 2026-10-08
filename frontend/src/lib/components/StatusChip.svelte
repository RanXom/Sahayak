<script lang="ts">
	type Kind = 'ok' | 'warn' | 'stop' | 'info' | 'accent' | 'plain';
	let { kind = 'plain', children }: { kind?: Kind; children?: import('svelte').Snippet } = $props();
</script>

<span class="chip {kind}">{@render children?.()}</span>
