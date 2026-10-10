<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, ApiError } from '$lib/api/client';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	const LANGS = ['hi', 'en', 'ta', 'mr', 'bn', 'te', 'kn', 'ml', 'gu', 'pa', 'or', 'ur'];
	const CATS = ['scheme_discovery', 'application_help', 'document_help'];

	let languageSel = $state<Lang>('hi');
	let category = $state('application_help');
	let details = $state('');
	let shareFacts = $state(true);
	let shareDocs = $state(false);
	let applicationId = $state('');
	let schemeId = $state('');
	let error = $state('');
	let busy = $state(false);

	onMount(() => {
		languageSel = lang;
		applicationId = page.url.searchParams.get('applicationId') ?? '';
		schemeId = page.url.searchParams.get('schemeId') ?? '';
	});

	async function submit() {
		error = '';
		busy = true;
		try {
			const res = await api.createAssistRequest({
				application_id: applicationId || undefined,
				scheme_id: schemeId || undefined,
				language: languageSel,
				category,
				details: details.trim(),
				share_profile_facts: shareFacts,
				share_documents: shareDocs
			});
			goto(`/help/${res.id}`);
		} catch (e) {
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
	<p class="eyebrow" style="margin-top:14px">{t('get_help_title')}</p>
	<h1 style="margin-top:8px">{t('help_new_heading')}</h1>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	<form
		onsubmit={(e) => {
			e.preventDefault();
			submit();
		}}
	>
		<div class="field" style="margin-top:22px">
			<label for="hlang">{t('help_language')}</label>
			<select id="hlang" class="select" bind:value={languageSel}>
				{#each LANGS as code}
					<option value={code}>{code}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<label for="hcat">{t('help_category')}</label>
			<select id="hcat" class="select" bind:value={category}>
				{#each CATS as cat}
					<option value={cat}>{t('helper_cat_' + cat)}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<label for="hdet">{t('help_details')}</label>
			<textarea
				id="hdet"
				class="textarea"
				rows="3"
				bind:value={details}
				maxlength="500"
				placeholder={t('help_details_hint')}
			></textarea>
		</div>

		<div class="panel soft pad" style="margin-bottom:16px">
			<p class="eyebrow">{t('help_share_title')}</p>
			<label class="checkline">
				<span class="sm">{t('help_share_facts')}</span>
				<input type="checkbox" bind:checked={shareFacts} />
			</label>
			<label class="checkline">
				<span class="sm">{t('help_share_docs')}</span>
				<input type="checkbox" bind:checked={shareDocs} />
			</label>
			<p class="muted sm" style="margin-top:8px">{t('help_share_note')}</p>
		</div>

		<button class="btn block" type="submit" disabled={busy}>{t('help_submit')}</button>
	</form>
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
