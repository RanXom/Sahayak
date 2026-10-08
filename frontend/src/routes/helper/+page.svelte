<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { language, translate, type Lang } from '$lib/i18n';
	import { apiBase } from '$lib/api/client';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let token = $state('');
	let sessionId = $state('');
	let message = $state('');
	let error = $state('');

	onMount(() => {
		token = localStorage.getItem('sahayak_access') ?? '';
	});

	async function call(path: string, body: unknown) {
		error = '';
		message = '';
		const res = await fetch(`${apiBase()}${path}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
			body: JSON.stringify(body)
		});
		const data = await res.json();
		if (!res.ok) {
			error = data?.error?.message ?? 'Request failed';
			return null;
		}
		return data;
	}

	async function register() {
		const data = await call('/helper/register', {
			helper_type: 'field_worker',
			organization: 'Sahayak',
			id_number: ''
		});
		if (data) message = 'Helper profile saved. Awaiting admin verification.';
	}
	async function startSession() {
		const data = await call('/helper/sessions', { purpose: 'scheme_discovery' });
		if (data) {
			sessionId = data.id;
			message = 'Session started. Record consent before entering citizen data.';
		}
	}
	async function consent() {
		const data = await call(`/helper/sessions/${sessionId}/consent`, {});
		if (data) message = 'Consent recorded';
	}
	async function endSession() {
		const data = await call(`/helper/sessions/${sessionId}/end`, {});
		if (data) {
			message = 'Session ended. Citizen data is sealed.';
			sessionId = '';
		}
	}
</script>

<section class="shell page read">
	<button class="back" onclick={() => goto('/')} aria-label={t('back')}>
		<Icon name="back" size={18} /> {t('back')}
	</button>

	<header style="margin-top:16px">
		<p class="eyebrow">{t('helper')}</p>
		<h1 style="margin-top:8px">{t('helper')}</h1>
		<p class="muted" style="margin-top:10px; line-height:1.6">
			Assist citizens with scheme discovery and applications. Verified helper accounts only. Every
			action is audit-logged and explicit consent is required.
		</p>
	</header>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}
	{#if message}<p class="saved" role="status">{message}</p>{/if}

	<div class="actions">
		{#if !sessionId}
			<button class="btn ghost block" onclick={register}>Register as helper</button>
			<button class="btn block" onclick={startSession}>
				Start new session
				<span class="disc"><Icon name="arrow" size={15} /></span>
			</button>
		{:else}
			<p class="dim sm">Session <code>{sessionId.slice(0, 8)}</code></p>
			<button class="btn block" onclick={consent}>Record consent</button>
			<button class="btn ghost block" onclick={endSession}>End session</button>
		{/if}
	</div>
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
	.saved {
		margin-top: 14px;
		color: var(--accent-strong);
		font-weight: 600;
		font-size: var(--step--1);
	}
	.actions {
		display: grid;
		gap: 10px;
		margin-top: 26px;
	}
</style>
