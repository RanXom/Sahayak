<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { refreshSession } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	const LANGS = ['hi', 'en', 'ta', 'mr', 'bn', 'te', 'kn', 'ml', 'gu', 'pa', 'or', 'ur'];
	const TYPES = ['volunteer', 'asha', 'csc_operator', 'field_worker', 'ngo_worker'];
	const CATS = ['scheme_discovery', 'application_help', 'document_help'];

	let helperType = $state('volunteer');
	let displayName = $state('');
	let organization = $state('');
	let languages = $state<string[]>(['hi']);
	let categories = $state<string[]>(['application_help']);
	let experience = $state('');
	let agreed = $state(false);
	let error = $state('');
	let busy = $state(false);

	function toggle(arr: string[], v: string): string[] {
		return arr.includes(v) ? arr.filter((x) => x !== v) : [...arr, v];
	}

	function langName(code: string): string {
		const names: Record<string, Record<string, string>> = {
			hi: { hi: 'हिन्दी', en: 'Hindi', ta: 'இந்தி' },
			en: { hi: 'English', en: 'English', ta: 'ஆங்கிலம்' },
			ta: { hi: 'तमिल', en: 'Tamil', ta: 'தமிழ்' },
			mr: { hi: 'मराठी', en: 'Marathi', ta: 'மராத்தி' }
		};
		return names[code]?.[lang] ?? code;
	}

	async function submit() {
		error = '';
		busy = true;
		try {
			await api.registerHelper({
				helper_type: helperType,
				organization: organization.trim() || undefined,
				display_name: displayName.trim() || undefined,
				languages,
				categories,
				experience: experience.trim() || undefined,
				terms_accepted: agreed
			});
			// Role changed citizen -> helper: rotate tokens so the JWT carries it.
			await refreshSession();
			goto('/helper');
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('error_state');
		} finally {
			busy = false;
		}
	}
</script>

<section class="shell read page">
	<a class="back" href="/login" aria-label={t('back')}><Icon name="back" size={18} /></a>
	<p class="eyebrow" style="margin-top:14px">{t('helper_register_title')}</p>
	<h1 style="margin-top:8px">{t('helper_register_heading')}</h1>
	<p class="muted sm" style="margin-top:10px">{t('helper_register_note')}</p>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}

	<form
		onsubmit={(e) => {
			e.preventDefault();
			submit();
		}}
	>
		<div class="field" style="margin-top:22px">
			<label for="htype">{t('helper_type')}</label>
			<select id="htype" class="select" bind:value={helperType}>
				{#each TYPES as ty}
					<option value={ty}>{t('helper_type_' + ty)}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<label for="dname">{t('helper_display_name')}</label>
			<input
				id="dname"
				class="input"
				type="text"
				bind:value={displayName}
				maxlength={80}
			/>
		</div>

		<div class="field">
			<label for="org">{t('helper_org')}</label>
			<input
				id="org"
				class="input"
				type="text"
				bind:value={organization}
				maxlength={120}
			/>
		</div>

		<fieldset class="checkgroup">
			<legend>{t('helper_languages')}</legend>
			<p class="muted sm">{t('helper_languages_hint')}</p>
			<div class="checks">
				{#each LANGS as code}
					<label class="checkline">
						<span>{langName(code)}</span>
						<input
							type="checkbox"
							checked={languages.includes(code)}
							onchange={() => (languages = toggle(languages, code))}
						/>
					</label>
				{/each}
			</div>
		</fieldset>

		<fieldset class="checkgroup">
			<legend>{t('helper_categories')}</legend>
			<div class="checks">
				{#each CATS as cat}
					<label class="checkline">
						<span>{t('helper_cat_' + cat)}</span>
						<input
							type="checkbox"
							checked={categories.includes(cat)}
							onchange={() => (categories = toggle(categories, cat))}
						/>
					</label>
				{/each}
			</div>
		</fieldset>

		<div class="field">
			<label for="exp">{t('helper_experience')}</label>
			<textarea
				id="exp"
				class="textarea"
				rows="3"
				bind:value={experience}
				maxlength={500}
			></textarea>
		</div>

		<label class="checkline agree">
			<span class="sm">{t('helper_terms')}</span>
			<input
				type="checkbox"
				checked={agreed}
				onchange={(e) => (agreed = e.currentTarget.checked)}
			/>
		</label>

		<button class="btn block" type="submit" disabled={busy || !agreed}>
			{t('helper_register_submit')}
		</button>
	</form>
</section>

<style>
	.back {
		display: inline-flex;
		align-items: center;
		color: var(--ink-2);
		font-weight: 600;
		font-size: var(--step--1);
		padding: 8px 0;
	}
	.checkgroup {
		border: 0;
		padding: 0;
		margin: 0 0 16px;
	}
	.checkgroup legend {
		font-size: var(--step--1);
		font-weight: 600;
		color: var(--ink-2);
		padding: 0;
		margin-bottom: 4px;
	}
	.checks {
		margin-top: 6px;
		border-top: 1px solid var(--line);
	}
	.agree {
		margin: 4px 0 18px;
		gap: 10px;
	}
	.agree input {
		width: 22px;
		height: 22px;
		flex: none;
	}
	.sm {
		font-size: var(--step--1);
	}
</style>
