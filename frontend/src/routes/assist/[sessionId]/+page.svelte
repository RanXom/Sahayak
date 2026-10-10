<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		api,
		ApiError,
		type AssistContext,
		type AssistMessage,
		type AssistSession,
		type DraftValue
	} from '$lib/api/client';
	import { user } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';
	import AssistCall from '$lib/components/AssistCall.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	const role = $derived($user?.role === 'helper' ? 'helper' : 'citizen');

	let session = $state<AssistSession | null>(null);
	let ctx = $state<AssistContext | null>(null);
	let drafts = $state<DraftValue[]>([]);
	let messages = $state<AssistMessage[]>([]);
	let lastMsgSeen = $state('');
	let chatInput = $state('');
	let fieldName = $state('');
	let fieldValue = $state('');
	let error = $state('');
	let busy = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	function sid(): string {
		return page.params.sessionId as string;
	}

	async function loadAll() {
		try {
			session = await api.getAssistSession(sid());
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
			return;
		}
		try {
			ctx = await api.assistContext(sid());
			drafts = ctx.drafts ?? [];
		} catch {
			/* context may be unavailable before consent */
		}
		await pollChat();
	}

	async function pollChat() {
		try {
			const res = await api.listAssistMessages(sid(), lastMsgSeen || undefined);
			if (res.messages.length) {
				messages = [...messages, ...res.messages];
				lastMsgSeen = res.messages[res.messages.length - 1].created_at;
			}
		} catch {
			/* polling is best-effort */
		}
	}

	onMount(() => {
		loadAll();
		pollTimer = setInterval(pollChat, 3000);
		return () => {
			if (pollTimer) clearInterval(pollTimer);
		};
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
	});

	function newClientId(): string {
		return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
	}

	async function sendChat() {
		const text = chatInput.trim();
		if (!text) return;
		chatInput = '';
		try {
			const m = await api.sendAssistMessage(sid(), text, newClientId());
			messages = [...messages, { ...m, mine: true }];
			lastMsgSeen = m.created_at;
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
			chatInput = text;
		}
	}

	async function propose() {
		if (!fieldName.trim()) return;
		busy = true;
		error = '';
		try {
			const d = await api.proposeDraftValue(sid(), fieldName.trim(), fieldValue);
			drafts = [...drafts.filter((x) => x.field_name !== d.field_name), d];
			fieldName = '';
			fieldValue = '';
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}

	async function confirm(name: string) {
		busy = true;
		error = '';
		try {
			const d = await api.confirmDraftValue(sid(), name);
			drafts = [...drafts.filter((x) => x.field_name !== d.field_name), d];
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}

	async function applyAll() {
		busy = true;
		error = '';
		try {
			await api.applyDraftValues(sid());
			const c = await api.assistContext(sid());
			ctx = c;
			drafts = c.drafts ?? [];
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}

	async function endSession() {
		busy = true;
		error = '';
		try {
			session = await api.endAssistSession(sid());
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}

	function fmtTime(iso: string): string {
		try {
			return new Date(iso).toLocaleTimeString(lang === 'hi' ? 'hi-IN' : 'en-IN', {
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return '';
		}
	}
</script>

<section class="shell page read">
	<button class="back" onclick={() => goto(role === 'helper' ? '/helper' : '/')} aria-label={t('back')}>
		<Icon name="back" size={18} /> {t('back')}
	</button>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	{#if session}
		<p class="eyebrow" style="margin-top:14px">{t('workspace_title')}</p>
		<h1 style="margin-top:8px">{t('workspace_status_' + session.status)}</h1>

		{#if session.status !== 'active'}
			<div class="panel soft pad" style="margin-top:16px">
				<p class="muted">{t('workspace_closed')}</p>
			</div>
		{:else}
			<AssistCall sessionId={sid()} />

			<section style="margin-top:24px">
				<h3>{t('workspace_shared')}</h3>
				{#if ctx?.application}
					<p class="muted sm" style="margin-top:6px">
						{t('workspace_application')}: {ctx.application.id.slice(0, 8)} · {ctx.application.status}
					</p>
				{/if}
				{#if ctx && 'profile_facts' in ctx && ctx.profile_facts}
					<div class="facts">
						{#each Object.entries(ctx.profile_facts) as [k, v]}
							<div class="fact"><span class="dim sm">{k}</span><span>{String(v)}</span></div>
						{/each}
					</div>
				{/if}
				{#if ctx?.documents && ctx.documents.length}
					<ul class="docs">
						{#each ctx.documents as d}
							<li class="sm">{d.document_type}: {d.has_document ? '✓' : '—'}</li>
						{/each}
					</ul>
				{/if}
			</section>

			<section style="margin-top:24px">
				<h3>{t('workspace_drafts')}</h3>
				{#if drafts.length === 0}
					<p class="muted sm" style="margin-top:8px">{t('workspace_drafts_empty')}</p>
				{:else}
					{#each drafts as d (d.field_name)}
						<div class="panel soft pad draft">
							<div class="row-between">
								<strong>{d.field_name}</strong>
								<span class="chip">{d.status}</span>
							</div>
							<p style="margin-top:6px">{String(d.value ?? '')}</p>
							{#if role === 'citizen' && d.status === 'proposed'}
								<button class="btn sm" style="margin-top:10px" onclick={() => confirm(d.field_name)} disabled={busy}>
									{t('workspace_confirm')}
								</button>
							{/if}
						</div>
					{/each}
					{#if role === 'citizen'}
						<button class="btn ghost block" style="margin-top:12px" onclick={applyAll} disabled={busy}>
							{t('workspace_apply')}
						</button>
					{/if}
				{/if}
				{#if role === 'helper'}
					<div class="field" style="margin-top:14px">
						<label for="dfield">{t('workspace_propose')}</label>
						<input id="dfield" class="input" bind:value={fieldName} placeholder="full_name" />
					</div>
					<div class="field">
						<input class="input" bind:value={fieldValue} placeholder={t('workspace_value')} />
					</div>
					<button class="btn block" onclick={propose} disabled={busy}>{t('workspace_propose_btn')}</button>
				{/if}
			</section>

			<section style="margin-top:24px">
				<h3>{t('workspace_chat')}</h3>
				<div class="chat" aria-live="polite">
					{#each messages as m (m.id)}
						<div class="msg" class:mine={m.mine}>
							<p>{m.body}</p>
							<span class="dim sm">{fmtTime(m.created_at)}</span>
						</div>
					{/each}
					{#if messages.length === 0}
						<p class="muted sm">{t('workspace_chat_empty')}</p>
					{/if}
				</div>
				<form
					class="chatbox"
					onsubmit={(e) => {
						e.preventDefault();
						sendChat();
					}}
				>
					<input
						class="input"
						bind:value={chatInput}
						maxlength="2000"
						placeholder={t('type_message')}
						aria-label={t('type_message')}
					/>
					<button class="btn" type="submit" disabled={!chatInput.trim()}>{t('send')}</button>
				</form>
			</section>

			<button class="btn quiet block" style="margin-top:26px" onclick={endSession} disabled={busy}>
				{t('workspace_end')}
			</button>
		{/if}
	{:else if !error}
		<div class="skeleton" style="height:280px; margin-top:24px"></div>
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
	.row-between {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}
	.facts {
		margin-top: 10px;
		border-top: 1px solid var(--line);
	}
	.fact {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		padding: 8px 0;
		border-bottom: 1px solid var(--line);
	}
	.docs {
		margin: 10px 0 0;
		padding-left: 18px;
	}
	.draft {
		margin-top: 10px;
	}
	.chat {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		max-height: 320px;
		overflow-y: auto;
	}
	.msg {
		align-self: flex-start;
		max-width: 85%;
		background: var(--surface-2);
		border-radius: var(--r-md);
		padding: 10px 14px;
	}
	.msg.mine {
		align-self: flex-end;
		background: var(--accent-soft);
	}
	.msg span {
		display: block;
		margin-top: 4px;
		text-align: right;
	}
	.chatbox {
		display: flex;
		gap: 8px;
		margin-top: 12px;
	}
</style>
