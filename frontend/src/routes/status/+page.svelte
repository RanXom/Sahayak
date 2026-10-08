<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api/client';
	import { authenticated } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import StatusChip from '$lib/components/StatusChip.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	type AppRow = { id: string; scheme_id: string; status: string; has_pdf: boolean };
	let apps = $state<AppRow[]>([]);
	let prefs = $state<Array<{ channel: string; enabled: boolean }>>([]);
	let refs = $state<Record<string, string>>({});
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		if (!$authenticated) return;
		try {
			const [a, p] = await Promise.all([api.listApplications(), api.getPreferences()]);
			apps = a.applications as AppRow[];
			prefs = p.preferences;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not load applications';
		} finally {
			loading = false;
		}
	});

	async function submit(appId: string) {
		try {
			await api.recordSubmission(appId, refs[appId] ?? '');
			const a = await api.listApplications();
			apps = a.applications as AppRow[];
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not record submission';
		}
	}

	async function toggle(channel: string, enabled: boolean) {
		try {
			prefs = (await api.updatePreferences(channel, enabled)).preferences;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not update preferences';
		}
	}
	function prefEnabled(channel: string) {
		return prefs.find((p) => p.channel === channel)?.enabled ?? false;
	}
	function statusKind(s: string) {
		if (s === 'completed' || s === 'ready') return 'ok';
		if (s === 'rejected') return 'stop';
		return 'info';
	}
</script>

<section class="shell page read">
	<header>
		<p class="eyebrow">{t('applications')}</p>
		<h1 style="margin-top:8px">{t('applications_count')}</h1>
	</header>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	{#if loading}
		<div class="stack" style="--gap:12px; margin-top:20px">
			{#each [1, 2] as _}<div class="skeleton" style="height:96px"></div>{/each}
		</div>
	{:else if apps.length === 0}
		<div class="empty">
			<p class="muted">{t('applications_empty')}</p>
			<a class="btn ghost sm" href="/schemes" style="margin-top:14px">{t('schemes')}</a>
		</div>
	{:else}
		<div style="margin-top:10px">
			{#each apps as a (a.id)}
				<article class="app">
					<div class="row-between">
						<code class="ref dim">#{a.id.slice(0, 8)}</code>
						<StatusChip kind={statusKind(a.status)}>{a.status}</StatusChip>
					</div>
					<div class="refrow">
						<input class="input" placeholder={t('reference')} bind:value={refs[a.id]} />
						<button class="btn sm" onclick={() => submit(a.id)}>{t('record_reference')}</button>
					</div>
				</article>
			{/each}
		</div>
	{/if}

	<section class="notif">
		<h3>{t('notifications')}</h3>
		<p class="muted sm">{t('enable_notifications')}</p>
		<div class="checkline">
			<span>{t('whatsapp')}</span>
			<input
				type="checkbox"
				checked={prefEnabled('whatsapp')}
				onchange={(e) => toggle('whatsapp', e.currentTarget.checked)}
			/>
		</div>
		<div class="checkline">
			<span>{t('sms')}</span>
			<input
				type="checkbox"
				checked={prefEnabled('sms')}
				onchange={(e) => toggle('sms', e.currentTarget.checked)}
			/>
		</div>
	</section>
</section>

<style>
	.sm {
		font-size: var(--step--1);
	}
	.app {
		padding: 18px 0;
		border-bottom: 1px solid var(--line);
	}
	.row-between {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.ref {
		font-size: var(--step--1);
		font-variant-numeric: tabular-nums;
	}
	.refrow {
		display: flex;
		gap: 8px;
		margin-top: 12px;
	}
	.notif {
		margin-top: 40px;
		padding-top: 24px;
		border-top: 1px solid var(--line);
	}
	.notif h3 {
		margin-bottom: 4px;
	}
</style>
