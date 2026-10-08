<script lang="ts">
	// A small, consistent stroke-icon set (stroke 1.6, 24×24, currentColor).
	// Kept intentionally minimal so the interface has one visual voice.
	const paths: Record<string, string[]> = {
		home: ['M4 11.5 12 4.5l8 7', 'M6.2 10.2V19h11.6v-8.8'],
		mic: [
			'M12 4.2a2.6 2.6 0 0 1 2.6 2.6v3.8a2.6 2.6 0 0 1-5.2 0V6.8A2.6 2.6 0 0 1 12 4.2Z',
			'M6.6 11a5.4 5.4 0 0 0 10.8 0',
			'M12 16.6V19.6'
		],
		user: ['M12 11.4a3.1 3.1 0 1 0 0-6.2 3.1 3.1 0 0 0 0 6.2Z', 'M5.8 19a6.2 6.2 0 0 1 12.4 0'],
		doc: ['M7 3.8h6.4L17.5 8v12.2H7z', 'M13.2 3.8V8h4.3'],
		check: ['M5 12.6l4.4 4.4L19 7.2'],
		back: ['M14 6l-6 6 6 6'],
		arrow: ['M6 12h12', 'M13 6l6 6-6 6'],
		chevron: ['M9 6l6 6-6 6'],
		close: ['M6.5 6.5l11 11', 'M17.5 6.5l-11 11'],
		plus: ['M12 5v14', 'M5 12h14'],
		bell: ['M7 10a5 5 0 0 1 10 0c0 4 1.4 5.4 1.4 5.4H5.6S7 14 7 10Z', 'M10.4 18.6a1.8 1.8 0 0 0 3.2 0'],
		shield: ['M12 3.6l6.5 2.4v5.2c0 4.2-3.2 7-6.5 8.2-3.3-1.2-6.5-4-6.5-8.2V6z'],
		globe: ['M12 4a8 8 0 1 0 0 16 8 8 0 0 0 0-16Z', 'M4.5 12h15', 'M12 4c2.4 2.4 2.4 13.6 0 16', 'M12 4c-2.4 2.4-2.4 13.6 0 16'],
		spark: ['M12 4v4', 'M12 16v4', 'M4 12h4', 'M16 12h4', 'M6.6 6.6l2.6 2.6', 'M14.8 14.8l2.6 2.6', 'M17.4 6.6l-2.6 2.6', 'M9.2 14.8l-2.6 2.6']
	};
	let { name, size = 24, stroke = 1.6 }: { name: string; size?: number; stroke?: number } = $props();
	let d = $derived(paths[name] ?? []);
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width={stroke}
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	{#each d as p}
		<path d={p}></path>
	{/each}
</svg>
