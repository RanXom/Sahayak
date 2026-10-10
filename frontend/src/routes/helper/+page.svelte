<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, ApiError, type AssistRequest, type AssistSession, type HelperProfile } from '$lib/api/client';
	import { handleAuthError, user } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let profile = $state<HelperProfile | null>(null);
	let queue = $state<AssistRequest[]>([]);
	let sessions = $state<AssistSession[]>([]);
	let busyAccept = $state('');
	let error = $state('');
	let notice = $state('');

	onMount(async () => {
		if ($user && $user.role !== 'helper' && $user.role !== 'admin') {
			goto('/helper/register');
			return;
		}
		await reload();
	});

	async function reload() {
		error = '';
		try {
			profile = await api.getHelperProfile();
		} catch {
			goto('/helper/register');
			return;
		}
		try {
			const q = await api.assistQueue();
			queue = q.requests;
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
		}
		try {
			const s = await api.myAssistSessions();
			sessions = s.sessions.filter((x) => x.status === 'active');
		} catch {
			/* non-critical */
		}
	}

	async function toggleAvailability() {
		if (!profile) return;
		const next = !profile.is_available;
		try {
			const res = await api.setAvailability(next);
			profile.is_available = res.is_available;
			await reload();
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
		}
	}

	async function accept(id: string) {
		busyAccept = id;
		error = '';
		notice = '';
		try {
			const sess = await api.acceptAssistRequest(id);
			goto(`/assist/${sess.id}`);
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
			await reload();
		} finally {
			busyAccept = '';
		}
	}

	function waiting(s: number | undefined): string {
		const m = Math.floor((s ?? 0) / 60);
		if (m < 1) return t('just_now');
		if (m < 60) return `${m} ${t('minutes')}`;
		return `${Math.floor(m / 60)} ${t('hours')}`;
	}
</script>

<section class="shell page read">
	<header class="head">
		<div>
			<p class="eyebrow">{t('helper_dashboard')}</p>
			<h1 style="margin-top:8px">{profile?.display_name || t('helper_dashboard')}</h1>
			{#if profile && !profile.verified}
				<p class="muted sm" style="margin-top:6px">{t('helper_unverified_note')}</p>
			{/if}
		</div>
		{#if profile}
			<button
				class="btn sm"
				class:ghost={profile.is_available}
				onclick={toggleAvailability}
				aria-pressed={profile.is_available}
			>
				{profile.is_available ? t('helper_available_on') : t('helper_available_off')}
			</button>
		{/if}
	</header>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}
	{#if notice}<p class="muted sm" style="margin-top:12px">{notice}</p>{/if}

	{#if sessions.length > 0}
		<h3 style="margin-top:26px">{t('helper_active_session')}</h3>
		{#each sessions as s (s.id)}
			<a class="scheme-row" href="/assist/{s.id}">
				<div class="row-between">
					<span class="scheme-name">{t('helper_session')} · {s.status}</span>
					<Icon name="arrow" size={16} />
				</div>
			</a>
		{/each}
	{/if}

	<h3 style="margin-top:26px">{t('helper_queue')}</h3>
	{#if !profile?.is_available}
		<p class="muted sm" style="margin-top:8px">{t('helper_queue_paused')}</p>
	{:else if queue.length === 0}
		<div class="empty"><p class="muted">{t('helper_queue_empty')}</p></div>
	{:else}
		<div class="list">
			{#each queue as q (q.id)}
				<div class="panel soft pad req">
					<div class="row-between">
						<span class="chip accent">{q.language}</span>
						<span class="muted sm">{waiting(q.waiting_seconds)} · {t('helper_cat_' + q.category)}</span>
					</div>
					{#if q.details}
						<p class="sm" style="margin-top:10px">{q.details}</p>
					{/if}
					<div class="row-between" style="margin-top:12px">
						<span class="muted sm">
							{q.share_profile_facts ? t('share_facts_on') : t('share_facts_off')}
						</span>
						<button class="btn sm" onclick={() => accept(q.id)} disabled={busyAccept === q.id}>
							{busyAccept === q.id ? '…' : t('helper_accept')}
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<footer class="home-foot">
		<a class="quiet-link" href="/helper/register">{t('helper_edit_profile')}</a>
	</footer>
</section>

<style>
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}
	.req {
		margin-top: 12px;
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
	.list {
		margin-top: 6px;
	}
	.scheme-row {
		display: block;
		padding: 18px 0;
		border-bottom: 1px solid var(--line);
	}
	.scheme-name {
		font-weight: 600;
	}
	.home-foot {
		margin-top: 32px;
		padding-top: 20px;
		border-top: 1px solid var(--line);
	}
	.quiet-link {
		display: inline-block;
		font-size: var(--step--1);
		font-weight: 600;
		color: var(--accent-strong);
		padding-block: 4px;
	}
	.empty {
		text-align: center;
		padding: 32px 16px;
	}
</style>
