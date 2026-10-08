<script lang="ts">
	import { onMount } from 'svelte';
	import { Orb as OrbEngine, type OrbState } from '$lib/voice/orb';

	let {
		mode = 'idle',
		analyser = null,
		size = 300,
		label = 'Voice assistant'
	}: {
		mode?: OrbState;
		analyser?: AnalyserNode | null;
		size?: number;
		label?: string;
	} = $props();

	let canvas = $state<HTMLCanvasElement | null>(null);
	let engine: OrbEngine | null = null;

	onMount(() => {
		if (!canvas) return;
		engine = new OrbEngine(canvas);
		engine.setState(mode);
		engine.setAnalyser(analyser);
		engine.start();
		return () => engine?.destroy();
	});

	$effect(() => {
		engine?.setState(mode);
	});
	$effect(() => {
		engine?.setAnalyser(analyser);
	});
</script>

<div class="orb-wrap" style="--size:{size}px" role="img" aria-label={label}>
	<canvas bind:this={canvas}></canvas>
</div>

<style>
	.orb-wrap {
		position: relative;
		width: var(--size);
		height: var(--size);
		max-width: 78vw;
		max-height: 78vw;
		display: grid;
		place-items: center;
		overflow: hidden;
		border-radius: 50%;
	}
	canvas {
		position: absolute;
		inset: 0;
		border-radius: 50%;
	}
</style>
