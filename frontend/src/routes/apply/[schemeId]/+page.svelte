<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, ApiError, apiBase, type ApplicationView } from '$lib/api/client';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let view = $state<ApplicationView | null>(null);
	let form = $state<Record<string, any>>({});
	let error = $state('');
	let status = $state('');
	let busy = $state(false);

	onMount(async () => {
		try {
			view = await api.createApplication(page.params.schemeId as string);
			form = { ...view.auto_filled };
			for (const f of view.form_fields) {
				if (form[f.name] === undefined && f.value !== undefined) form[f.name] = f.value;
			}
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not create application';
		}
	});

	async function save() {
		if (!view) return;
		busy = true;
		error = '';
		status = '';
		try {
			await api.updateApplicationForm(view.application.id, form);
			status = t('save');
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not save';
		} finally {
			busy = false;
		}
	}

	async function copyDetails() {
		if (!view) return;
		const lines = view.form_fields.map((f) => `${f.label}: ${form[f.name] ?? ''}`).join('\n');
		try {
			await navigator.clipboard.writeText(lines);
			status = 'Details copied — paste them into the official form';
		} catch {
			error = 'Could not copy to clipboard';
		}
	}

	async function download() {
		if (!view) return;
		busy = true;
		error = '';
		try {
			await api.updateApplicationForm(view.application.id, form);
			await api.generatePdf(view.application.id);
			const token = localStorage.getItem('sahayak_access');
			const res = await fetch(`${apiBase()}/applications/${view.application.id}/pdf`, {
				headers: { Authorization: `Bearer ${token}` }
			});
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'sahayak-application.pdf';
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not generate PDF';
		} finally {
			busy = false;
		}
	}
</script>

<section class="shell page read">
	<button class="back" onclick={() => goto('/schemes')} aria-label={t('back')}>
		<Icon name="back" size={18} /> {t('back')}
	</button>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	{#if view}
		<header style="margin-top:16px">
			<p class="eyebrow">{t('apply')}</p>
			<h1 style="margin-top:8px">{lang === 'hi' ? view.scheme.name_hi : view.scheme.name_en}</h1>
			<p class="muted sm" style="margin-top:8px">
				Review the pre-filled details. This is not an official submission.
			</p>
		</header>

		{#if status}<p class="saved" role="status">{status}</p>{/if}

		<div style="margin-top:22px">
			{#each view.form_fields as f}
				<div class="field">
					<label for={f.name}>
						{f.label}{#if f.required} *{/if}{#if f.auto_filled}<span class="dim"> · auto</span>{/if}
					</label>
					{#if f.type === 'select'}
						<select id={f.name} class="select" bind:value={form[f.name]}>
							<option value="">—</option>
							<option value="female">Female</option>
							<option value="male">Male</option>
							<option value="other">Other</option>
						</select>
					{:else if f.type === 'boolean'}
						<input id={f.name} type="checkbox" bind:checked={form[f.name]} />
					{:else if f.type === 'number'}
						<input id={f.name} class="input" type="number" bind:value={form[f.name]} />
					{:else if f.type === 'date'}
						<input id={f.name} class="input" type="date" bind:value={form[f.name]} />
					{:else}
						<input id={f.name} class="input" type="text" bind:value={form[f.name]} />
					{/if}
				</div>
			{/each}
		</div>

		{#if view.missing_fields.length > 0}
			<p class="muted sm">{t('missing_docs')}: {view.missing_fields.join(', ')}</p>
		{/if}

		<div class="actions">
			<button class="btn ghost block" onclick={copyDetails}>Copy details for the online form</button>
			<button class="btn ghost block" onclick={save} disabled={busy}>{t('save_changes')}</button>
			<button class="btn block" onclick={download} disabled={busy}>
				{t('download_pdf')}
				<span class="disc"><Icon name="arrow" size={15} /></span>
			</button>
		</div>
		<p class="dim sm" style="margin-top:14px">
			Official portals must be filled and submitted by you. Sahayak prepares the details and the PDF —
			it never submits on your behalf.
		</p>
	{:else if !error}
		<div class="skeleton" style="height:320px; margin-top:24px"></div>
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
