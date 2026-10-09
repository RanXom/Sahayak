<script lang="ts">
	import { onMount } from 'svelte';
	import { authenticated } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import {
		eligibility,
		loadEligibility,
		eligibilityLoading,
		eligibilityError
	} from '$lib/stores/eligibility';
	import type { EligibilityResult, Scheme } from '$lib/api/client';
	import { api } from '$lib/api/client';
	import StatusChip from '$lib/components/StatusChip.svelte';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let filter = $state<'all' | 'eligible' | 'potentially_eligible'>('all');
	let guestSchemes = $state<Scheme[]>([]);
	let guestError = $state('');

	onMount(async () => {
		if ($authenticated) {
			loadEligibility(false, lang);
			return;
		}
		// Guest browsing: published schemes, no eligibility.
		try {
			const res = await api.schemes();
			guestSchemes = res.schemes;
		} catch {
			guestError = 'Could not load schemes';
		}
	});

	function name(r: EligibilityResult) {
		return lang === 'hi' ? r.scheme.name_hi : r.scheme.name_en;
	}
	function chipKind(status: string) {
		return status === 'eligible' ? 'ok' : status === 'potentially_eligible' ? 'warn' : 'stop';
	}
	function statusLabel(status: string) {
		return status === 'eligible'
			? t('eligible')
			: status === 'potentially_eligible'
				? t('potentially_eligible')
				: t('not_eligible');
	}
	function fresh(iso?: string | null) {
		if (!iso) return '';
		const d = new Date(iso);
		return d.toLocaleDateString(lang === 'hi' ? 'hi-IN' : 'en-IN', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
	}

	let results = $derived($eligibility?.results ?? []);
	let filtered = $derived(results.filter((r) => (filter === 'all' ? true : r.status === filter)));
	let counts = $derived({
		eligible: $eligibility?.summary?.eligible ?? 0,
		potential: $eligibility?.summary?.potentially_eligible ?? 0
	});
</script>

<section class="shell page read">
	<header class="head">
		<div>
			<p class="eyebrow">{t('schemes')}</p>
			<h1 style="margin-top:8px">{$authenticated ? t('schemes_matched') : t('browse_schemes')}</h1>
		</div>
		{#if $authenticated}
			<button
				class="btn ghost sm"
				onclick={() => loadEligibility(true, lang)}
				disabled={$eligibilityLoading}
				aria-label="Refresh"
			>
				{$eligibilityLoading ? '…' : '↻'}
			</button>
		{/if}
	</header>

	{#if !$authenticated}
		<div class="panel soft pad" style="margin-top:18px">
			<p class="muted">{t('guest_schemes_cta')}</p>
			<a class="btn sm" href="/login" style="margin-top:14px">{t('login')}</a>
		</div>
		{#if guestError}
			<div class="banner" style="margin-top:16px">{guestError}</div>
		{:else}
			<div class="list" style="margin-top:8px">
				{#each guestSchemes as s (s.id)}
					<a class="scheme-row" href="/schemes/{s.id}">
						<div class="row-between">
							<h3 class="sname">{lang === 'hi' ? s.name_hi : s.name_en}</h3>
						</div>
						{#if s.benefit_amount}
							<div class="meta">
								<span class="muted sm">{s.benefit_amount}</span>
							</div>
						{/if}
					</a>
				{/each}
			</div>
		{/if}
	{:else if $eligibility}
		<p class="muted sm" style="margin-top:10px">{$eligibility.coverage_note}</p>
		<div class="summary">
			<span class="chip ok">{counts.eligible} {t('eligible')}</span>
			<span class="chip warn">{counts.potential} {t('potentially_eligible')}</span>
		</div>
	{/if}

	{#if $eligibilityError}
		<div class="banner" style="margin-top:16px">{$eligibilityError}</div>
	{/if}

	{#if $authenticated}
		<div class="filters" role="tablist">
			<button class:on={filter === 'all'} onclick={() => (filter = 'all')}>{t('schemes')}</button>
			<button class:on={filter === 'eligible'} onclick={() => (filter = 'eligible')}
				>{t('eligible')}</button
			>
			<button
				class:on={filter === 'potentially_eligible'}
				onclick={() => (filter = 'potentially_eligible')}>{t('potentially_eligible')}</button
			>
		</div>
	{/if}

	{#if $eligibilityLoading && !$eligibility}
		<div class="stack" style="--gap:14px; margin-top:20px">
			{#each [1, 2, 3, 4] as _}
				<div class="skeleton" style="height:88px"></div>
			{/each}
		</div>
	{:else if filtered.length}
		<div class="list">
			{#each filtered as r (r.scheme.id)}
				<a class="scheme-row" href="/schemes/{r.scheme.id}">
					<div class="row-between">
						<h3 class="sname">{name(r)}</h3>
						<StatusChip kind={chipKind(r.status)}>{statusLabel(r.status)}</StatusChip>
					</div>

					<div class="meta">
						{#if r.scheme.benefit_amount}
							<span class="muted sm">{r.scheme.benefit_amount}</span>
						{/if}
						{#if r.scheme.verified_at}
							<span class="chip accent sm-chip">
								<Icon name="shield" size={13} /> {t('verified')}
							</span>
						{:else}
							<span class="chip sm-chip">{t('unverified')}</span>
						{/if}
					</div>

					<p class="why muted sm">{r.explanation}</p>

					{#if r.documents && r.documents.missing_count > 0}
						<p class="sm" style="margin-top:6px">
							<span class="dot" aria-hidden="true"></span>
							{t('missing_docs')}: {r.documents.missing_count}
						</p>
					{/if}
				</a>
			{/each}
		</div>
	{:else if $eligibility}
		<div class="empty">
			<p class="muted">{t('no_schemes')}</p>
			<p class="dim sm" style="margin-top:8px">{t('no_schemes_note')}</p>
		</div>
	{/if}
</section>

<style>
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}
	.sm {
		font-size: var(--step--1);
	}
	.summary {
		display: flex;
		gap: 8px;
		margin-top: 12px;
	}
	.filters {
		display: flex;
		gap: 6px;
		margin: 22px 0 4px;
		overflow-x: auto;
		padding-bottom: 4px;
	}
	.filters button {
		border: 1px solid var(--line);
		background: transparent;
		color: var(--ink-2);
		border-radius: var(--r-pill);
		padding: 8px 14px;
		font-size: var(--step--1);
		font-weight: 600;
		white-space: nowrap;
		cursor: pointer;
		transition:
			background var(--dur-1) var(--ease),
			color var(--dur-1) var(--ease);
	}
	.filters button.on {
		background: var(--ink);
		color: var(--canvas);
		border-color: var(--ink);
	}
	.list {
		margin-top: 6px;
	}
	.row-between {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.sname {
		font-size: var(--step-1);
	}
	.meta {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 6px;
		flex-wrap: wrap;
	}
	.sm-chip {
		font-size: var(--step--2);
		padding: 3px 9px;
	}
	.why {
		margin-top: 8px;
		line-height: 1.5;
	}
	.dot {
		display: inline-block;
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--warn-ink);
		margin-right: 7px;
		vertical-align: middle;
	}
</style>
