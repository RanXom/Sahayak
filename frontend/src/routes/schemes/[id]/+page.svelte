<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { authenticated } from '$lib/stores/auth';
	import { api, ApiError, type Scheme, type DocumentStatus } from '$lib/api/client';
	import { language, translate, type Lang } from '$lib/i18n';
	import { eligibility } from '$lib/stores/eligibility';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let scheme = $state<Scheme | null>(null);
	let docs = $state<DocumentStatus[]>([]);
	let error = $state('');
	let openDoc = $state<number | null>(null);

	let match = $derived(($eligibility?.results ?? []).find((r) => r.scheme.id === page.params.id));

	onMount(async () => {
		const id = page.params.id as string;
		try {
			scheme = await api.scheme(id);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not load scheme';
			return;
		}
		// Document checklist needs a profile; guests see scheme info + login CTA.
		try {
			const res = await api.schemeDocuments(id);
			docs = res.documents;
		} catch {
			docs = [];
		}
	});

	function name(s: Scheme) {
		return lang === 'hi' ? s.name_hi : s.name_en;
	}
	function desc(s: Scheme) {
		return lang === 'hi' ? s.description_hi : s.description_en;
	}
	function fresh(iso?: string | null) {
		if (!iso) return '';
		return new Date(iso).toLocaleDateString(lang === 'hi' ? 'hi-IN' : 'en-IN', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		});
	}
</script>

<section class="shell page read">
	<button class="back" onclick={() => goto('/schemes')} aria-label={t('back')}>
		<Icon name="back" size={18} /> {t('back')}
	</button>

	{#if error}
		<div class="banner" style="margin-top:16px">{error}</div>
	{:else if scheme}
		<header style="margin-top:18px">
			<div class="row-between">
				<p class="eyebrow">{scheme.benefit_type.replace('_', ' ')}</p>
				{#if scheme.verified_at}
					<span class="chip accent sm-chip"><Icon name="shield" size={13} /> {t('verified')}</span>
				{:else}
					<span class="chip sm-chip">{t('unverified')}</span>
				{/if}
			</div>
			<h1 style="margin-top:12px">{name(scheme)}</h1>
			{#if scheme.benefit_amount}
				<p class="benefit">{scheme.benefit_amount}</p>
			{/if}
			{#if desc(scheme)}
				<p class="muted" style="margin-top:14px; line-height:1.6">{desc(scheme)}</p>
			{/if}
			{#if scheme.ministry}
				<p class="dim sm" style="margin-top:10px">{scheme.ministry}</p>
			{/if}
		</header>

		{#if match}
			<div class="panel pad" style="margin-top:22px">
				<p class="eyebrow">{t('why_eligible')}</p>
				<p style="margin-top:8px; line-height:1.55">{match.explanation}</p>
			</div>
		{/if}

		<div class="cta" style="margin-top:18px">
			<a class="btn block" href="/apply/{scheme.id}">
				{t('apply')}
				<span class="disc"><Icon name="arrow" size={15} /></span>
			</a>
			{#if scheme.official_url}
				<a class="btn ghost block" style="margin-top:10px" href={scheme.official_url} target="_blank"
					>{t('official_website')}</a
				>
			{/if}
		</div>

		{#if !docs.length && !$authenticated}
			<div class="panel soft pad" style="margin-top:22px">
				<p class="muted">{t('guest_schemes_cta')}</p>
				<a class="btn sm" href="/login" style="margin-top:14px">{t('login')}</a>
			</div>
		{/if}

		{#if docs.length}
			<section style="margin-top:36px">
				<h3>{t('documents')}</h3>
				<div style="margin-top:8px">
					{#each docs as d, i}
						<div class="doc">
							<div class="doc-main">
								<span class="tick" class:has={d.user_has} aria-hidden="true">
									{#if d.user_has}<Icon name="check" size={14} />{/if}
								</span>
								<span class="doc-name">{d.name}{#if d.mandatory}<span class="dim"> ·</span>{/if}</span>
								{#if !d.user_has && d.acquisition}
									<button
										class="btn quiet sm"
										onclick={() => (openDoc = openDoc === i ? null : i)}
										aria-expanded={openDoc === i}
									>
										{t('how_to_get')}
									</button>
								{/if}
							</div>
							{#if openDoc === i && d.acquisition}
								<div class="acquire">
									{#if d.acquisition.issuing_authority}
										<p><span class="dim">{t('issuing_authority')}:</span> {d.acquisition.issuing_authority}</p>
									{/if}
									{#if d.acquisition.steps}
										<p class="steps">{d.acquisition.steps}</p>
									{/if}
									<div class="acquire-meta">
										{#if d.acquisition.typical_time}<span>{t('typical_time')}: {d.acquisition.typical_time}</span>{/if}
										{#if d.acquisition.typical_fee}<span>{t('typical_fee')}: {d.acquisition.typical_fee}</span>{/if}
										{#if d.acquisition.source_url}
											<a class="quiet-link" href={d.acquisition.source_url} target="_blank">source</a>
										{/if}
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</section>
		{/if}

		{#if scheme.verified_at}
			<p class="dim sm" style="margin-top:28px">
				{t('verified')} · {fresh(scheme.verified_at)}
			</p>
		{/if}
	{/if}
</section>

<style>
	.back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		border: 0;
		background: transparent;
		color: var(--ink-2);
		font-weight: 600;
		font-size: var(--step--1);
		padding: 8px 0;
		cursor: pointer;
	}
	.back:hover {
		color: var(--ink);
	}
	.row-between {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.sm {
		font-size: var(--step--1);
	}
	.sm-chip {
		font-size: var(--step--2);
		padding: 3px 9px;
	}
	.benefit {
		margin-top: 10px;
		font-size: var(--step-2);
		font-weight: 600;
		color: var(--accent-strong);
	}
	.doc {
		border-bottom: 1px solid var(--line);
	}
	.doc:last-child {
		border-bottom: 0;
	}
	.doc-main {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 13px 0;
	}
	.tick {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		border: 1.5px solid var(--line-strong);
		display: grid;
		place-items: center;
		color: #fff;
		flex: none;
	}
	.tick.has {
		background: var(--accent);
		border-color: var(--accent);
	}
	.doc-name {
		font-weight: 500;
		flex: 1;
	}
	.acquire {
		padding: 0 0 16px 34px;
		display: grid;
		gap: 8px;
		font-size: var(--step--1);
		color: var(--ink-2);
	}
	.steps {
		white-space: pre-wrap;
		line-height: 1.55;
	}
	.acquire-meta {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
	}
	.quiet-link {
		display: inline-block;
		color: var(--accent-strong);
		font-weight: 600;
		padding-block: 4px;
	}
</style>
