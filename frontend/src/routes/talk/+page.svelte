<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, ApiError, resolveMediaUrl, type TurnResult } from '$lib/api/client';
	import { language, translate, type Lang } from '$lib/i18n';
	import { authenticated, handleAuthError } from '$lib/stores/auth';
	import { VoiceRecorder, isRecordingSupported } from '$lib/voice/recorder';
	import { voiceConfig } from '$lib/voice/config';
	import Orb from '$lib/components/Orb.svelte';
	import Icon from '$lib/components/Icon.svelte';
	import type { OrbState } from '$lib/voice/orb';

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	type Turn = { user: string; assistant: string };
	let turns = $state<Turn[]>([]);
	let conversationId = $state('');
	let uiState = $state<OrbState>('idle');
	let analyser = $state<AnalyserNode | null>(null);
	let errorMsg = $state('');
	let elapsed = $state(0);
	let voiceAvailable = $state(false);
	let textMode = $state(false);
	let input = $state('');
	let showTranscript = $state(false);
	let ready = $state(false);

	let recorder: VoiceRecorder | null = null;
	let playbackCtx: AudioContext | null = null;
	let playbackAnalyser: AnalyserNode | null = null;
	let lastAudioUrl = $state('');
	let needsTapToPlay = $state(false);

	const maxMs = voiceConfig().maxDurationMs;
	let lastAssistant = $derived(turns.length ? turns[turns.length - 1].assistant : '');
	let lastUser = $derived(turns.length ? turns[turns.length - 1].user : '');

	let stateLabel = $derived(
		uiState === 'listening'
			? t('listening')
			: uiState === 'processing'
				? t('processing')
				: uiState === 'speaking'
					? t('speaking')
					: uiState === 'error'
						? t('error_state')
						: t('tap_to_speak')
	);

	let hint = $derived(
		uiState === 'listening' ? t('tap_to_stop') : uiState === 'idle' ? t('silence_hint') : ''
	);

	let mmss = $derived(
		`${Math.floor(elapsed / 60000)}:${String(Math.floor((elapsed % 60000) / 1000)).padStart(2, '0')}`
	);

	onMount(async () => {
		if (!$authenticated) {
			goto('/login');
			return;
		}
		try {
			const cfg = await api.getConfig();
			voiceAvailable = cfg.voice_enabled && isRecordingSupported();
			if (!voiceAvailable) textMode = true;
		} catch {
			textMode = true;
		}
		try {
			const conv = await api.startConversation(lang);
			conversationId = conv.id;
			turns = conv.messages
				.filter((m) => m.role === 'assistant')
				.map((m) => ({ user: '', assistant: m.content_text }));
		} catch (e) {
			errorMsg = e instanceof ApiError ? e.message : 'Could not start session';
		}
	});

	onDestroy(() => {
		recorder?.cancel();
		playbackCtx?.close().catch(() => {});
	});

	async function toggleMic() {
		if (uiState === 'listening') {
			await stopListening();
			return;
		}
		if (uiState === 'processing' || uiState === 'speaking') return;
		if (!voiceAvailable) {
			textMode = true;
			return;
		}
		errorMsg = '';
		uiState = 'idle';
		elapsed = 0;
		recorder = new VoiceRecorder();
		try {
			await recorder.start({
				onElapsed: (ms) => (elapsed = ms),
				onAutoStop: (_reason, blob) => handleBlob(blob)
			});
			analyser = recorder.getAnalyser();
			uiState = 'listening';
		} catch {
			errorMsg = t('voice_unavailable');
			uiState = 'error';
			textMode = true;
		}
	}

	async function stopListening() {
		if (!recorder) return;
		analyser = null;
		const blob = await recorder.stop();
		recorder = null;
		await handleBlob(blob);
	}

	async function handleBlob(blob: Blob) {
		recorder = null;
		analyser = null;
		if (blob.size < 800) {
			errorMsg = t('voice_unavailable');
			uiState = 'error';
			return;
		}
		uiState = 'processing';
		try {
			const res = await api.sendVoice(conversationId, blob, lang);
			await applyResult(res);
		} catch (e) {
			if (handleAuthError(e)) return;
			errorMsg = e instanceof ApiError ? e.message : t('error_state');
			uiState = 'error';
		}
	}

	async function sendText() {
		if (!input.trim() || uiState === 'processing') return;
		const text = input.trim();
		input = '';
		uiState = 'processing';
		errorMsg = '';
		try {
			const res = await api.sendMessage(conversationId, text, lang);
			await applyResult(res);
		} catch (e) {
			if (handleAuthError(e)) return;
			errorMsg = e instanceof ApiError ? e.message : t('error_state');
			uiState = 'error';
		}
	}

	async function applyResult(res: TurnResult) {
		turns = [...turns, { user: res.transcript, assistant: res.response_text }];
		ready = res.ready_for_review;
		if (res.response_audio_url) {
			await speak(resolveMediaUrl(res.response_audio_url));
		} else {
			uiState = 'idle';
		}
	}

	async function speak(url: string) {
		lastAudioUrl = url;
		needsTapToPlay = false;
		try {
			const Ctx =
				window.AudioContext ??
				(window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
			if (!playbackCtx) {
				playbackCtx = new Ctx();
				playbackAnalyser = playbackCtx.createAnalyser();
				playbackAnalyser.fftSize = 1024;
				playbackAnalyser.connect(playbackCtx.destination);
			}
			if (playbackCtx.state === 'suspended') await playbackCtx.resume();

			const audio = new Audio(url);
			// Required: the API is a different origin (local :8080, split
			// subdomains in prod). Without CORS mode, Web Audio zeroes the
			// output and playback is silently mute.
			audio.crossOrigin = 'anonymous';
			const src = playbackCtx.createMediaElementSource(audio);
			src.connect(playbackAnalyser!);
			analyser = playbackAnalyser;
			uiState = 'speaking';
			await new Promise<void>((resolve) => {
				audio.onended = () => resolve();
				audio.onerror = () => {
					console.warn('[talk] audio element error', audio.error);
					resolve();
				};
				audio.play().catch((err) => {
					// Autoplay policy: transient activation from the mic tap has
					// expired by the time the reply arrives. Fall back to an
					// explicit tap-to-play button (a real user gesture).
					console.warn('[talk] audio play blocked, needs tap', err);
					needsTapToPlay = true;
					resolve();
				});
			});
		} catch (err) {
			console.warn('[talk] speak failed', err);
			needsTapToPlay = lastAudioUrl !== '';
		}
		analyser = null;
		if (!needsTapToPlay) uiState = 'idle';
	}

	// replay plays the last reply; called from a button so the browser treats
	// it as a user gesture and allows audio output.
	async function replay() {
		if (!lastAudioUrl) return;
		needsTapToPlay = false;
		await speak(lastAudioUrl);
	}

	function reset() {
		recorder?.cancel();
		recorder = null;
		analyser = null;
		uiState = 'idle';
		errorMsg = '';
	}
</script>

<div class="stage talk">
	<header class="talk-top">
		<button class="iconbtn" onclick={() => goto('/')} aria-label={t('back')}>
			<Icon name="back" size={20} />
		</button>
		<svg
			class="talk-brand"
			width="30"
			height="33"
			viewBox="0 0 40 43"
			aria-hidden="true"
			focusable="false"
		>
			<g id="logogram" style="opacity: 1;"><path fill="#0E6E55" d="M20 28.4356C16.5306 28.4357 13.3035 30.2166 11.4521 33.1512L9.1543 36.7934L10.0371 37.3511C10.3132 37.5101 10.5937 37.6622 10.8779 37.8082L13.1436 34.2188C14.6285 31.8648 17.2171 30.437 20 30.4369C22.783 30.4369 25.3714 31.8647 26.8564 34.2188L29.1211 37.8082C29.4053 37.6623 29.6858 37.5101 29.9619 37.3511L30.8457 36.7934L28.5479 33.1512C26.6964 30.2165 23.4695 28.4356 20 28.4356ZM20.123 36.4428L20 36.4379C19.3855 36.4381 18.8239 36.774 18.5312 37.3072L18.4766 37.4166L17.3691 39.834C18.0662 39.9256 18.7744 39.9821 19.4922 40L20 38.8914L20.5068 40C21.2249 39.9821 21.9335 39.9256 22.6309 39.834L21.5244 37.4166L21.4697 37.3072C21.1966 36.8096 20.689 36.4842 20.123 36.4428ZM21 12.4063V0H19V12.4063L15.7891 0.422921L13.8574 0.940585L17.0674 12.923L10.8662 2.18005L9.13379 3.18021L15.335 13.9222L6.56543 5.15124L5.15039 6.56651L13.9199 15.3375L3.17969 9.13529L2.17969 10.868L12.9209 17.0702L0.94043 13.8597L0.422852 15.7917L12.4053 19.0031H0V21.0035H12.4053L0.422852 24.2149L0.94043 26.1469L12.9209 22.9354L2.17969 29.1386L3.17969 30.8713L13.9199 24.6681L5.15039 33.4401L6.56543 34.8554L13.1216 28.2964C13.684 27.7337 14 26.9706 14 26.175V20.435C14 17.1209 16.6865 14.4342 20 14.434C23.3137 14.434 26 17.1208 26 20.435V26.1759C26 26.9716 26.316 27.7347 26.8785 28.2974L33.4346 34.8554L34.8496 33.4401L26.0801 24.6691L36.8203 30.8713L37.8203 29.1386L27.0791 22.9354L39.0596 26.1469L39.5771 24.2149L27.5957 21.0035H40V19.0031H27.5947L39.5771 15.7917L39.0596 13.8597L27.0791 17.0692L37.8203 10.868L36.8203 9.13529L26.0781 15.3375L34.8496 6.56554L33.4355 5.15124L24.6641 13.9232L30.8662 3.18021L29.1338 2.18005L22.9316 12.923L26.1426 0.940585L24.2109 0.422921L21 12.4063ZM20 32.4372C18.0122 32.4374 16.1792 33.5112 15.207 35.2453L13.2021 38.8201C13.8407 39.051 14.4942 39.2501 15.1611 39.4159L16.9512 36.224C17.5694 35.1209 18.7357 34.4377 20 34.4376C21.2645 34.4376 22.4305 35.1208 23.0488 36.224L24.8379 39.4159C25.5048 39.2502 26.1583 39.0509 26.7969 38.8201L24.793 35.2453C23.8208 33.5111 21.9879 32.4372 20 32.4372Z"/></g>
		</svg>
		<select
			class="select lang"
			value={lang}
			onchange={(e) => language.set((e.currentTarget as HTMLSelectElement).value as Lang)}
			aria-label={t('language')}
		>
			<option value="hi">हिन्दी</option>
			<option value="en">English</option>
			<option value="ta">தமிழ்</option>
		</select>
	</header>

	<main class="stage-main">
		<button
			class="orb-btn"
			class:active={uiState === 'listening'}
			onclick={toggleMic}
			aria-label={stateLabel}
			disabled={uiState === 'processing' || uiState === 'speaking'}
		>
			<Orb mode={uiState} {analyser} size={320} label={stateLabel} />
		</button>

		<div class="status" aria-live="polite">
			<p class="state">{stateLabel}</p>
			{#if hint}<p class="hint dim">{hint}</p>{/if}
			{#if uiState === 'listening'}
				<p class="elapsed mono">{mmss} <span class="dim">/ {Math.floor(maxMs / 60000)}:00</span></p>
			{/if}
		</div>

		{#if errorMsg}
			<div class="err" role="alert">
				<span>{errorMsg}</span>
				<button class="btn ghost sm" onclick={reset}>{t('try_again')}</button>
			</div>
		{/if}

		{#if needsTapToPlay && lastAudioUrl}
			<div class="err" role="alert">
				<span>{t('tap_to_play')}</span>
				<button class="btn ghost sm" onclick={replay}>▶</button>
			</div>
		{/if}

		{#if lastAssistant && uiState !== 'listening' && uiState !== 'processing'}
			<div class="said">
				{#if lastUser}
					<p class="said-user dim">“{lastUser}”</p>
				{/if}
				<p class="said-assistant">{lastAssistant}</p>
				{#if ready}
					<a class="btn sm" href="/schemes" style="margin-top:16px">
						{t('check_eligibility')}
						<span class="disc"><Icon name="arrow" size={14} /></span>
					</a>
				{/if}
			</div>
		{/if}
	</main>

	<footer class="talk-bottom">
		{#if textMode}
			<form class="textline" onsubmit={(e) => (e.preventDefault(), sendText())}>
				<input
					class="input"
					placeholder={t('type_message')}
					bind:value={input}
					aria-label={t('type_message')}
				/>
				<button class="btn" type="submit" disabled={!input.trim() || uiState === 'processing'}>
					{t('send')}
				</button>
			</form>
		{:else}
			<div class="controls">
				<button class="btn quiet sm" onclick={() => (textMode = true)}>{t('type_instead')}</button>
				{#if turns.length}
					<button class="btn quiet sm" onclick={() => (showTranscript = !showTranscript)}>
						{showTranscript ? t('hide_transcript') : t('show_transcript')}
					</button>
				{/if}
			</div>
		{/if}

		{#if showTranscript && turns.length}
			<div class="transcript">
				{#each turns as turn}
					{#if turn.user}<p class="t-user"><span class="dim">{t('you_said')}</span> {turn.user}</p>{/if}
					<p class="t-assistant"><span class="dim">{t('sahayak_said')}</span> {turn.assistant}</p>
				{/each}
			</div>
		{/if}
	</footer>
</div>

<style>
	.talk {
		min-height: 100dvh;
		background: var(--canvas);
		color: var(--ink);
		display: flex;
		flex-direction: column;
		padding: 0 clamp(22px, 6vw, 44px) calc(24px + env(safe-area-inset-bottom));
	}
	.talk-top {
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		align-items: center;
		gap: 12px;
		padding: calc(14px + env(safe-area-inset-top)) 0 14px;
	}
	.talk-top .iconbtn {
		justify-self: start;
	}
	.talk-top .talk-brand {
		justify-self: center;
		display: block;
	}
	.talk-top .lang {
		justify-self: end;
	}
	.talk-brand {
		display: block;
	}
	.iconbtn {
		width: 42px;
		height: 42px;
		border-radius: var(--r-pill);
		border: 1px solid var(--line);
		background: transparent;
		color: var(--ink);
		display: grid;
		place-items: center;
		cursor: pointer;
		transition: background var(--dur-1) var(--ease);
	}
	.iconbtn:hover {
		background: var(--surface);
	}
	.lang {
		width: auto;
		min-height: 38px;
		padding: 4px 32px 4px 12px;
		font-size: var(--step--1);
		font-weight: 600;
		border-radius: var(--r-pill);
		background: transparent;
		color: var(--ink);
		border-color: var(--line);
		background-position: right 10px center;
	}
	.stage-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 20px;
		padding: 8px 0 20px;
		text-align: center;
	}
	.orb-btn {
		background: none;
		border: 0;
		padding: 0;
		cursor: pointer;
		-webkit-tap-highlight-color: transparent;
	}
	.orb-btn:disabled {
		cursor: default;
	}
	.status {
		min-height: 62px;
		display: grid;
		gap: 3px;
		place-items: center;
	}
	.state {
		font-size: var(--step-2);
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	.hint {
		font-size: var(--step--1);
	}
	.elapsed {
		font-size: var(--step--1);
		color: var(--ink-2);
	}
	.err {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		border-radius: var(--r-md);
		border: 1px solid var(--line);
		background: color-mix(in srgb, var(--surface) 70%, transparent);
		font-size: var(--step--1);
	}
	.said {
		max-width: 40ch;
		animation: fade var(--dur-3) var(--ease-out) both;
	}
	@keyframes fade {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	.said-user {
		font-size: var(--step--1);
		margin-bottom: 8px;
	}
	.said-assistant {
		font-size: var(--step-1);
		line-height: 1.55;
		color: var(--ink);
	}
	.talk-bottom {
		padding-top: 8px;
	}
	.textline {
		display: flex;
		gap: 8px;
		align-items: center;
	}
	.textline .input {
		background: var(--surface);
		border-color: var(--line);
		color: var(--ink);
	}
	.controls {
		display: flex;
		justify-content: center;
		gap: 6px;
	}
	.transcript {
		margin-top: 14px;
		max-height: 30vh;
		overflow-y: auto;
		display: grid;
		gap: 10px;
		text-align: left;
		font-size: var(--step--1);
		border-top: 1px solid var(--line);
		padding-top: 14px;
	}
	.t-user,
	.t-assistant {
		line-height: 1.5;
	}
	.t-user {
		color: var(--ink-2);
	}
	@media (min-width: 860px) {
		.said {
			max-width: 48ch;
		}
	}
</style>
