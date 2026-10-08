<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api/client';
	import { login, loginWithFirebase } from '$lib/stores/auth';
	import { confirmFirebaseOtp, firebaseEnabled, sendFirebaseOtp } from '$lib/firebase';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	let phone = $state('');
	let otp = $state('');
	let challengeId = $state('');
	let step = $state<'phone' | 'otp'>('phone');
	let error = $state('');
	let busy = $state(false);
	let demoOtp = $state('');
	// Login method: Firebase SMS when both sides support it, else server OTP.
	let method = $state<'firebase' | 'server'>('server');
	let firebaseOffered = $state(false);

	onMount(async () => {
		try {
			const cfg = await api.getConfig();
			firebaseOffered = !!cfg.firebase_auth && firebaseEnabled;
			if (firebaseOffered) method = 'firebase';
		} catch {
			firebaseOffered = false;
		}
	});

	function switchMethod(m: 'firebase' | 'server') {
		method = m;
		step = 'phone';
		otp = '';
		challengeId = '';
		demoOtp = '';
		error = '';
	}

	async function requestOtp() {
		error = '';
		busy = true;
		try {
			if (method === 'firebase') {
				await sendFirebaseOtp(phone.trim(), 'recaptcha-container');
			} else {
				const res = await api.requestOtp(phone.trim());
				challengeId = res.challenge_id;
				if (res.demo_otp) {
					demoOtp = res.demo_otp;
					otp = res.demo_otp;
				}
			}
			step = 'otp';
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Could not send code';
		} finally {
			busy = false;
		}
	}

	async function verify() {
		error = '';
		busy = true;
		try {
			if (method === 'firebase') {
				const idToken = await confirmFirebaseOtp(otp.trim());
				await loginWithFirebase(idToken);
			} else {
				await login(challengeId, otp.trim());
			}
			goto('/');
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Verification failed';
		} finally {
			busy = false;
		}
	}
</script>

<section class="shell read login">
	<a class="back" href="/" aria-label={t('back')}><Icon name="back" size={18} /></a>
	<p class="eyebrow">{t('app_name')}</p>
	<h1 style="margin-top:10px">{t('login')}</h1>
	<p class="muted" style="margin-top:8px">{t('tagline')}</p>

	{#if error}<div class="banner" style="margin-top:20px">{error}</div>{/if}

	{#if firebaseOffered}
		<div class="methods" role="tablist">
			<button
				type="button"
				class="method"
				class:active={method === 'firebase'}
				onclick={() => switchMethod('firebase')}
			>
				{t('login_firebase')}
			</button>
			<button
				type="button"
				class="method"
				class:active={method === 'server'}
				onclick={() => switchMethod('server')}
			>
				{t('login_server_code')}
			</button>
		</div>
	{/if}
	<div id="recaptcha-container"></div>

	{#if step === 'phone'}
		<form onsubmit={(e) => (e.preventDefault(), requestOtp())}>
			<div class="field" style="margin-top:24px">
				<label for="phone">{t('phone_number')}</label>
				<input
					id="phone"
					class="input"
					type="tel"
					inputmode="tel"
					placeholder="+91XXXXXXXXXX"
					bind:value={phone}
					autocomplete="tel"
				/>
			</div>
			<button class="btn block" type="submit" disabled={busy || phone.replace(/\D/g, '').length < 10}>
				{t('send_otp')}
			</button>
		</form>
	{:else}
		{#if demoOtp}
			<div class="demo">
				<span class="eyebrow">Demo</span>
				<span>Code <strong class="mono">{demoOtp}</strong> — pre-filled</span>
			</div>
		{/if}
		<form onsubmit={(e) => (e.preventDefault(), verify())}>
			<div class="field" style="margin-top:20px">
				<label for="otp">{t('enter_otp')}</label>
				<input
					id="otp"
					class="input otp"
					type="text"
					inputmode="numeric"
					maxlength="6"
					bind:value={otp}
					autocomplete="one-time-code"
				/>
			</div>
			<button class="btn block" type="submit" disabled={busy || otp.length < 6}>{t('verify')}</button>
			<button class="btn quiet block" type="button" style="margin-top:8px" onclick={() => (step = 'phone')}>
				{t('phone_number')}
			</button>
		</form>
	{/if}
</section>

<style>
	.login {
		padding-top: 6vh;
		max-width: 420px;
	}
	.back {
		display: inline-grid;
		place-items: center;
		width: 40px;
		height: 40px;
		border-radius: var(--r-pill);
		border: 1px solid var(--line);
		color: var(--ink-2);
		margin-bottom: 28px;
	}
	.back:hover {
		background: var(--surface);
		color: var(--ink);
	}
	.demo {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 20px;
		padding: 11px 14px;
		border-radius: var(--r-md);
		background: var(--accent-soft);
		color: var(--accent-strong);
		font-size: var(--step--1);
		font-weight: 500;
	}
	.otp {
		letter-spacing: 0.4em;
		font-size: var(--step-2);
		text-align: center;
	}
	.methods {
		display: flex;
		gap: 8px;
		margin-top: 20px;
	}
	.method {
		flex: 1;
		padding: 10px;
		border-radius: var(--r-md);
		border: 1px solid var(--line);
		background: transparent;
		color: var(--ink-2);
		font-size: var(--step--1);
		font-weight: 600;
	}
	.method.active {
		background: var(--accent-soft);
		color: var(--accent-strong);
		border-color: transparent;
	}
</style>
