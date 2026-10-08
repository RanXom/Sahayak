<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError, type StateInfo, type District, type ProfileView } from '$lib/api/client';
	import { authenticated } from '$lib/stores/auth';
	import { language, translate, type Lang } from '$lib/i18n';
	import { eligibility } from '$lib/stores/eligibility';
	import Ring from '$lib/components/Ring.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let states = $state<StateInfo[]>([]);
	let districts = $state<District[]>([]);
	let profile = $state<ProfileView | null>(null);
	let facts = $state<Record<string, any>>({});
	let stateId = $state<number | null>(null);
	let districtId = $state<number | null>(null);
	let docs = $state<Record<string, boolean>>({});
	let status = $state('');
	let error = $state('');
	let saving = $state(false);

	const docList: Array<[string, string, string]> = [
		['aadhaar', 'Aadhaar Card', 'आधार कार्ड'],
		['income_cert', 'Income Certificate', 'आय प्रमाण पत्र'],
		['domicile_cert', 'Domicile Certificate', 'मूल निवास प्रमाण पत्र'],
		['caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र'],
		['bank_account', 'Bank Account', 'बैंक खाता'],
		['ration_card', 'Ration Card', 'राशन कार्ड'],
		['voter_id', 'Voter ID', 'मतदाता पहचान पत्र'],
		['birth_cert', 'Birth Certificate', 'जन्म प्रमाण पत्र'],
		['pan_card', 'PAN Card', 'पैन कार्ड'],
		['disability_cert', 'Disability Certificate', 'विकलांगता प्रमाण पत्र'],
		['land_record', 'Land Record', 'भूमि रिकॉर्ड'],
		['student_cert', 'Student Certificate', 'छात्र प्रमाण पत्र'],
		['death_cert', 'Death Certificate', 'मृत्यु प्रमाण पत्र'],
		['marriage_cert', 'Marriage Certificate', 'विवाह प्रमाण पत्र']
	];

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
		const present = coreKeys.filter((k) => facts[k] !== undefined && facts[k] !== null && facts[k] !== '').length;
		return Math.min(1, (present + (stateId ? 1 : 0)) / (coreKeys.length + 1));
	});

	onMount(async () => {
		if (!$authenticated) {
			error = t('login');
			return;
		}
		try {
			const [p, s] = await Promise.all([api.getProfile(), api.states()]);
			profile = p;
			states = s.states;
			facts = { ...p.facts };
			stateId = p.state?.id ?? null;
			districtId = p.district?.id ?? null;
			for (const d of p.documents) docs[d.document_type] = d.has_document;
			if (stateId) await loadDistricts(stateId);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not load profile';
		}
	});

	async function loadDistricts(id: number) {
		try {
			districts = (await api.districts(id)).districts;
		} catch {
			districts = [];
		}
	}

	async function onStateChange(e: Event) {
		const id = Number((e.currentTarget as HTMLSelectElement).value) || null;
		stateId = id;
		districtId = null;
		if (id) await loadDistricts(id);
	}

	async function save() {
		saving = true;
		status = '';
		error = '';
		try {
			profile = await api.updateProfile({
				state_id: stateId ?? undefined,
				district_id: districtId ?? undefined,
				facts
			});
			await api.updateDocuments(
				Object.entries(docs).map(([document_type, has_document]) => ({ document_type, has_document }))
			);
			status = 'Saved';
			eligibility.set(null);
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not save profile';
		} finally {
			saving = false;
		}
	}
</script>

<section class="shell page read">
	<header class="head">
		<div>
			<p class="eyebrow">{t('profile')}</p>
			<h1 style="margin-top:8px">{t('understood')}</h1>
		</div>
		<Ring value={completeness} size={56} label={t('profile_completeness')} />
	</header>

	{#if error}<div class="banner" style="margin-top:16px">{error}</div>{/if}
	{#if status}<p class="saved" role="status">{status}</p>{/if}

	<details class="disclosure group" open>
		<summary><span>Basics</span><span class="plus" aria-hidden="true"></span></summary>
		<div class="body">
			<div class="field">
				<label for="state">{t('state')}</label>
				<select id="state" class="select" value={stateId ?? ''} onchange={onStateChange}>
					<option value="">—</option>
					{#each states as s}<option value={s.id}>{lang === 'hi' ? s.name_hi : s.name_en}</option>{/each}
				</select>
			</div>
			<div class="field">
				<label for="district">{t('district')}</label>
				<select id="district" class="select" bind:value={districtId}>
					<option value={null}>—</option>
					{#each districts as d}<option value={d.id}>{lang === 'hi' ? d.name_hi : d.name_en}</option>{/each}
				</select>
			</div>
			<div class="two">
				<div class="field">
					<label for="age">{t('age')}</label>
					<input id="age" class="input" type="number" bind:value={facts.age} />
				</div>
				<div class="field">
					<label for="gender">{t('gender')}</label>
					<select id="gender" class="select" bind:value={facts.gender}>
						<option value="female">Female · महिला</option>
						<option value="male">Male · पुरुष</option>
						<option value="other">Other</option>
					</select>
				</div>
			</div>
		</div>
	</details>

	<details class="disclosure group">
		<summary><span>Household &amp; income</span><span class="plus" aria-hidden="true"></span></summary>
		<div class="body">
			<div class="two">
				<div class="field">
					<label for="marital">Marital status</label>
					<select id="marital" class="select" bind:value={facts.marital_status}>
						<option value="single">Single</option>
						<option value="married">Married</option>
						<option value="widowed">Widowed · विधवा</option>
						<option value="divorced">Divorced</option>
					</select>
				</div>
				<div class="field">
					<label for="income">{t('income')}</label>
					<input id="income" class="input" type="number" bind:value={facts.income_annual} />
				</div>
				<div class="field">
					<label for="occupation">{t('occupation')}</label>
					<input id="occupation" class="input" bind:value={facts.occupation} />
				</div>
				<div class="field">
					<label for="caste">Category</label>
					<select id="caste" class="select" bind:value={facts.caste_category}>
						<option value="general">General</option>
						<option value="obc">OBC</option>
						<option value="sc">SC</option>
						<option value="st">ST</option>
						<option value="ews">EWS</option>
					</select>
				</div>
				<div class="field">
					<label for="residence">Residence</label>
					<select id="residence" class="select" bind:value={facts.residence_type}>
						<option value="rural">Rural · ग्रामीण</option>
						<option value="urban">Urban · शहरी</option>
					</select>
				</div>
				<div class="field">
					<label for="ration">Ration card</label>
					<select id="ration" class="select" bind:value={facts.ration_card_type}>
						<option value="bpl">BPL</option>
						<option value="aay">Antyodaya</option>
						<option value="apl">APL</option>
						<option value="none">None</option>
					</select>
				</div>
			</div>
		</div>
	</details>

	<details class="disclosure group">
		<summary><span>Land &amp; identity</span><span class="plus" aria-hidden="true"></span></summary>
		<div class="body">
			<div class="field">
				<label for="land">Land (acres)</label>
				<input id="land" class="input" type="number" step="0.1" bind:value={facts.land_acres} />
			</div>
			<div class="checkline">
				<span>Bank account</span>
				<input type="checkbox" bind:checked={facts.has_bank_account} />
			</div>
			<div class="checkline">
				<span>Aadhaar</span>
				<input type="checkbox" bind:checked={facts.has_aadhaar} />
			</div>
		</div>
	</details>

	<details class="disclosure group">
		<summary><span>{t('documents')}</span><span class="plus" aria-hidden="true"></span></summary>
		<div class="body">
			{#each docList as [id, en, hi]}
				<div class="checkline">
					<span>{lang === 'hi' ? hi : en}</span>
					<input type="checkbox" bind:checked={docs[id]} />
				</div>
			{/each}
		</div>
	</details>

	<div class="savebar">
		<button class="btn block" onclick={save} disabled={saving}>{saving ? '…' : t('save_changes')}</button>
	</div>
</section>

<style>
	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
	}
	.saved {
		margin-top: 14px;
		color: var(--accent-strong);
		font-weight: 600;
		font-size: var(--step--1);
	}
	.group {
		border-bottom: 1px solid var(--line);
	}
	.group:first-of-type {
		margin-top: 24px;
		border-top: 1px solid var(--line);
	}
	.body {
		padding: 4px 0 20px;
	}
	.two {
		display: grid;
		grid-template-columns: 1fr;
		gap: 0;
	}
	.savebar {
		position: sticky;
		bottom: calc(84px + env(safe-area-inset-bottom));
		margin-top: 24px;
		padding: 12px;
		border-radius: var(--r-lg);
		background: color-mix(in srgb, var(--surface) 86%, transparent);
		backdrop-filter: blur(12px);
		border: 1px solid var(--line);
	}
	@media (min-width: 560px) {
		.two {
			grid-template-columns: 1fr 1fr;
			gap: 0 16px;
		}
	}
	@media (min-width: 860px) {
		.savebar {
			bottom: 16px;
		}
	}
</style>
