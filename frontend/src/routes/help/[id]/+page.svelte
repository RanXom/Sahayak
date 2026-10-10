<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, ApiError, type AssistRequest } from '$lib/api/client';
	import { handleAuthError } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let req = $state<AssistRequest | null>(null);
	let sessionId = $state('');
	let error = $state('');
	let busy = $state(false);
	let shareFacts = $state(true);
	let shareDocs = $state(false);
	let timer: ReturnType<typeof setInterval> | null = null;

	async function load() {
		try {
			req = await api.getAssistRequest(page.params.id as string);
			shareFacts = req.share_profile_facts;
			shareDocs = req.share_documents;
			if (req.status === 'in_session') {
				try {
					const s = await api.myAssistSessions();
					const found = s.sessions.find((x) => x.request_id === req?.id && x.status === 'active');
					if (found) sessionId = found.id;
				} catch {
					/* non-critical */
				}
			}
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
		}
	}

	onMount(() => {
		load();
		timer = setInterval(() => {
			if (req && (req.status === 'pending' || req.status === 'accepted')) load();
		}, 5000);
		return () => {
			if (timer) clearInterval(timer);
		};
	});

	async function join() {
		if (!req) return;
		busy = true;
		error = '';
		try {
			const sess = await api.joinAssistRequest(req.id, {
				share_profile_facts: shareFacts,
				share_documents: shareDocs
			});
			goto(`/assist/${sess.id}`);
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}

	async function cancel() {
		if (!req) return;
		busy = true;
		error = '';
		try {
			req = await api.cancelAssistRequest(req.id);
		} catch (e) {
			if (handleAuthError(e)) return;
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}
</script>

<section class="shell read page">
	<button class="back" onclick={() => goto('/')} aria-label={t('back')}>
		<Icon name="back" size={18} /> {t('back')}
	</button>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	{#if req}
		<p class="eyebrow" style="margin-top:14px">{t('help_status_title')}</p>
		<h1 style="margin-top:8px">{t('help_status_' + req.status)}</h1>
		<p class="muted sm" style="margin-top:10px">
			{req.language} · {t('helper_cat_' + req.category)}
			{#if req.details} · {req.details}{/if}
		</p>

		{#if req.status === 'accepted' && req.helper_display_name}
			<div class="panel soft pad" style="margin-top:18px">
				<p class="eyebrow">{t('help_assigned')}</p>
				<p style="margin-top:8px; font-weight:700">{req.helper_display_name}</p>
				<p class="muted sm" style="margin-top:8px">{t('help_join_note')}</p>
				<div class="field" style="margin-top:12px">
					<label class="checkline">
						<span class="sm">{t('help_share_facts')}</span>
						<input type="checkbox" bind:checked={shareFacts} />
					</label>
					<label class="checkline">
						<span class="sm">{t('help_share_docs')}</span>
						<input type="checkbox" bind:checked={shareDocs} />
					</label>
				</div>
				<button class="btn block" style="margin-top:12px" onclick={join} disabled={busy}>
					{t('help_join')}
				</button>
			</div>
		{/if}

		{#if req.status === 'pending'}
			<p class="muted sm" style="margin-top:18px">{t('help_waiting')}</p>
			<button class="btn quiet block" style="margin-top:10px" onclick={cancel} disabled={busy}>
				{t('help_cancel')}
			</button>
		{/if}

		{#if req.status === 'in_session' && sessionId}
			<a class="btn block" style="margin-top:18px" href="/assist/{sessionId}">
				{t('help_open_workspace')}
			</a>
		{/if}
	{:else if !error}
		<div class="skeleton" style="height:220px; margin-top:24px"></div>
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
	.sm {
		font-size: var(--step--1);
	}
</style>
