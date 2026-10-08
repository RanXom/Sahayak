<script lang="ts">
	let { value = 0, size = 64, label = '' }: { value?: number; size?: number; label?: string } =
		$props();

	const stroke = 4;
	let r = $derived((size - stroke) / 2);
	let c = $derived(2 * Math.PI * r);
	let offset = $derived(c * (1 - Math.max(0, Math.min(1, value))));
</script>

<div class="ring" style="--size:{size}px" role="img" aria-label={label}>
	<svg width={size} height={size} viewBox="0 0 {size} {size}">
		<circle cx={size / 2} cy={size / 2} {r} fill="none" stroke="var(--line-strong)" stroke-width={stroke} />
		<circle
			cx={size / 2}
			cy={size / 2}
			{r}
			fill="none"
			stroke="var(--accent)"
			stroke-width={stroke}
			stroke-linecap="round"
			stroke-dasharray={c}
			stroke-dashoffset={offset}
			transform="rotate(-90 {size / 2} {size / 2})"
		/>
	</svg>
	<div class="pct mono">{Math.round(value * 100)}<span>%</span></div>
</div>

<style>
	.ring {
		position: relative;
		width: var(--size);
		height: var(--size);
		display: grid;
		place-items: center;
	}
	svg {
		position: absolute;
		inset: 0;
	}
	.pct {
		font-size: 0.95rem;
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	.pct span {
		font-size: 0.6em;
		color: var(--ink-3);
	}
</style>
