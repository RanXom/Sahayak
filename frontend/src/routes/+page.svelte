<script lang="ts">
	import { onMount } from 'svelte';
	import { authenticated, user } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import { eligibility, loadEligibility, eligibilityLoading } from '$lib/stores/eligibility';
	import { api, ApiError, type ProfileView } from '$lib/api/client';
	import Orb from '$lib/components/Orb.svelte';
	import Ring from '$lib/components/Ring.svelte';
	import StatusChip from '$lib/components/StatusChip.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import InstallButton from '$lib/components/InstallButton.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let profile = $state<ProfileView | null>(null);
	let appCount = $state(0);

	const coreKeys = [
		'age',
		'gender',
		'marital_status',
		'income_annual',
		'occupation',
		'caste_category',
		'residence_type',
		'ration_card_type',
		'land_acres'
	];

	let completeness = $derived.by(() => {
		const facts = profile?.facts ?? {};
		const present = coreKeys.filter((k) => facts[k] !== undefined && facts[k] !== null).length;
		const stateBonus = profile?.state ? 1 : 0;
		return Math.min(1, (present + stateBonus) / (coreKeys.length + 1));
	});

	let matches = $derived(
		($eligibility?.results ?? []).filter((r) => r.status === 'eligible').slice(0, 3)
	);
	let eligibleCount = $derived($eligibility?.summary?.eligible ?? 0);

	onMount(async () => {
		if (!$authenticated) return;
		try {
			profile = await api.getProfile();
		} catch (e) {
			if (!(e instanceof ApiError)) throw e;
		}
		try {
			const apps = await api.listApplications();
			appCount = apps.applications.length;
		} catch {
			/* non-critical */
		}
		if (profile?.completed) loadEligibility(false, lang);
	});

	function schemeName(r: { scheme: { name_en: string; name_hi: string } }) {
		return lang === 'hi' ? r.scheme.name_hi : r.scheme.name_en;
	}
</script>

{#if !$authenticated}
	<section class="shell read" style="padding: 8vh 20px 40px">
		<p class="eyebrow">{t('app_name')}</p>
		<h1 style="margin-top:14px">{t('tagline')}</h1>
		<p class="muted" style="margin-top:16px; max-width: 46ch">
			Speak or type in your language. Sahayak understands your situation and finds the government
			welfare schemes you may be eligible for.
		</p>
		<a class="btn" href="/login" style="margin-top:28px">
			{t('login')}
			<span class="disc"><Icon name="arrow" size={15} /></span>
		</a>
		<InstallButton />
	</section>
{:else}
	<section class="shell page">
		<header class="greet">
			<p class="eyebrow">{t('greeting')}</p>
			<h1 style="margin-top:8px">{t('dashboard_question')}</h1>
		</header>

		<a class="talk-hero panel" href="/talk">
			<div class="orb-mini" aria-hidden="true"><Orb size={72} /></div>
			<div class="talk-copy">
				<span class="talk-title">{t('talk_to_sahayak')}</span>
				<span class="muted">{t('tap_to_speak')}</span>
			</div>
			<span class="go"><Icon name="arrow" size={18} /></span>
		</a>

		<a class="help-hero panel soft" href="/help/new">
			<div class="talk-copy">
				<span class="talk-title">{t('get_help_title')}</span>
				<span class="muted">{t('get_help_sub')}</span>
			</div>
			<span class="go"><Icon name="arrow" size={18} /></span>
		</a>

		<div class="stats">
			<a class="stat" href="/profile">
				<Ring value={completeness} size={62} label={t('profile_completeness')} />
				<div>
					<div class="stat-label">{t('profile_completeness')}</div>
					<div class="muted sm">
						{completeness >= 1 ? t('ready') : t('complete_profile')}
					</div>
				</div>
			</a>
			<a class="stat" href="/schemes">
				<div class="value mono">{eligibilityLoading ? '—' : eligibleCount}</div>
				<div class="stat-label">{t('schemes_matched')}</div>
			</a>
			<a class="stat" href="/status">
				<div class="value mono">{appCount}</div>
				<div class="stat-label">{t('applications_count')}</div>
			</a>
		</div>

		<section class="recent">
			<div class="recent-head">
				<h3>{t('recent_matches')}</h3>
				{#if matches.length}
					<a class="quiet-link" href="/schemes">{t('view_all')}</a>
				{/if}
			</div>

			{#if matches.length}
				{#each matches as r (r.scheme.id)}
					<a class="scheme-row" href="/schemes/{r.scheme.id}">
						<div class="row-between">
							<span class="scheme-name">{schemeName(r)}</span>
							<StatusChip kind="ok">{t('eligible')}</StatusChip>
						</div>
						{#if r.scheme.benefit_amount}
							<p class="muted sm" style="margin-top:4px">
								{t('benefit')} · {r.scheme.benefit_amount}
							</p>
						{/if}
					</a>
				{/each}
			{:else if profile && !profile.completed}
				<div class="panel soft pad empty">
					<p class="muted">{t('no_matches_yet')}</p>
					<a class="btn ghost sm" href="/profile" style="margin-top:14px">
						{t('complete_profile')}
					</a>
				</div>
			{:else}
				<div class="empty">
					<p class="muted">{t('no_matches_yet')}</p>
				</div>
			{/if}
		</section>

		<footer class="home-foot">
			<a class="quiet-link" href="/helper">{t('helper')}</a>
			<span class="dim">·</span>
			<a class="quiet-link" href="/status">{t('notifications')}</a>
		</footer>
	</section>
{/if}

<style>
	.greet h1 {
		max-width: 22ch;
	}
	.help-hero {
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 18px 22px;
		margin-top: 14px;
	}
	.help-hero .talk-title {
		font-size: var(--step-1);
	}
	.talk-hero {
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 20px 22px;
		margin-top: 24px;
		transition:
			transform var(--dur-2) var(--ease),
			border-color var(--dur-2) var(--ease),
			box-shadow var(--dur-2) var(--ease);
	}
	.talk-hero:hover {
		transform: translateY(-2px);
		border-color: var(--line-strong);
		box-shadow: var(--shadow);
	}
	.orb-mini {
		flex: none;
		width: 72px;
		height: 72px;
		display: grid;
		place-items: center;
		overflow: hidden;
		border-radius: 50%;
	}
	.talk-copy {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.talk-title {
		font-size: var(--step-2);
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	.go {
		margin-left: auto;
		width: 42px;
		height: 42px;
		border-radius: var(--r-pill);
		display: grid;
		place-items: center;
		background: var(--accent-soft);
		color: var(--accent-strong);
		flex: none;
	}
	.stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
		margin-top: 16px;
	}
	.stats .stat:first-child {
		grid-column: 1 / -1;
		flex-direction: row;
		align-items: center;
		gap: 18px;
	}
	.stat-label {
		font-size: var(--step--1);
		font-weight: 600;
		color: var(--ink-2);
	}
	.sm {
		font-size: var(--step--1);
	}
	.recent {
		margin-top: 36px;
	}
	.recent-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		margin-bottom: 10px;
	}
	.row-between {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.scheme-name {
		font-weight: 600;
		letter-spacing: -0.01em;
	}
	.quiet-link {
		font-size: var(--step--1);
		font-weight: 600;
		color: var(--accent-strong);
	}
	.quiet-link:hover {
		text-decoration: underline;
	}
	.home-foot {
		display: flex;
		gap: 10px;
		align-items: center;
		margin-top: 32px;
		padding-top: 20px;
		border-top: 1px solid var(--line);
	}
	@media (min-width: 860px) {
		.stats {
			grid-template-columns: 1.4fr 1fr 1fr;
		}
		.stats .stat:first-child {
			grid-column: auto;
		}
	}
</style>
