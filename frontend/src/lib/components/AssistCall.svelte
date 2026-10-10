<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, ApiError, type AssistSignal } from '$lib/api/client';
	import { language, translate, type Lang } from '$lib/i18n';

	let { sessionId = '' }: { sessionId?: string } = $props();

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));
	function t(k: string) {
		return translate(lang, k);
	}

	type CallState = 'idle' | 'calling' | 'in-call' | 'incoming' | 'ended' | 'error';
	let callState = $state<CallState>('idle');
	let error = $state('');
	let muted = $state(false);
	let mediaNotice = $state('');
	let videoOn = $state(true);
	let lastSeen = $state('');
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	let localVideo = $state<HTMLVideoElement | null>(null);
	let remoteVideo = $state<HTMLVideoElement | null>(null);
	let remoteStream = $state<MediaStream | null>(null);

	// Streams must (re)attach whenever the video elements mount: the answerer
	// acquires media while the call UI is still hidden, so a one-shot
	// assignment in ensureMedia/ontrack lands on null and is lost forever.
	$effect(() => {
		if (localVideo && localStream && localVideo.srcObject !== localStream) {
			localVideo.srcObject = localStream;
		}
	});
	$effect(() => {
		if (remoteVideo && remoteStream && remoteVideo.srcObject !== remoteStream) {
			remoteVideo.srcObject = remoteStream;
		}
	});
	let pc: RTCPeerConnection | null = null;
	let localStream: MediaStream | null = null;
	let remoteOffer: AssistSignal | null = null;

	const ICE_SERVERS: RTCConfiguration = {
		iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
	};

	async function send(kind: string, payload: Record<string, unknown> = {}) {
		await api.sendAssistSignal(sessionId, kind, payload);
	}

	function attachHandlers() {
		if (!pc) return;
		pc.onicecandidate = (e) => {
			if (e.candidate) {
				send('ice', {
					candidate: e.candidate.candidate,
					sdpMid: e.candidate.sdpMid,
					sdpMLineIndex: e.candidate.sdpMLineIndex
				}).catch(() => {});
			}
		};
		pc.ontrack = (e) => {
			if (e.streams[0]) remoteStream = e.streams[0];
		};
		pc.onconnectionstatechange = () => {
			if (!pc) return;
			if (pc.connectionState === 'connected') callState = 'in-call';
			if (pc.connectionState === 'failed') {
				error = t('call_failed');
				callState = 'error';
			}
		};
	}

	async function ensureMedia(withVideo: boolean) {
		mediaNotice = '';
		try {
			localStream = await navigator.mediaDevices.getUserMedia({
				audio: true,
				video: withVideo
			});
		} catch (e) {
			// Typical on same-device testing (camera already held by the other
			// tab) or camera-less phones: fall back to voice-only instead of
			// failing the whole call.
			if (!withVideo) throw e;
			const name = e instanceof DOMException ? e.name : '';
			if (name !== 'OverconstrainedError' && name !== 'NotReadableError' && name !== 'NotFoundError' && name !== 'NotAllowedError') {
				throw e;
			}
			localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
			mediaNotice = t('call_camera_fallback');
		}
		if (localVideo) {
			localVideo.srcObject = localStream;
			localVideo.muted = true;
		}
		videoOn = localStream.getVideoTracks().length > 0;
	}

	async function startCall(withVideo: boolean) {
		error = '';
		callState = 'calling';
		try {
			await ensureMedia(withVideo);
			pc = new RTCPeerConnection(ICE_SERVERS);
			localStream?.getTracks().forEach((tr) => pc?.addTrack(tr, localStream as MediaStream));
			attachHandlers();
			await send('join', { video: withVideo });
			const offer = await pc.createOffer();
			await pc.setLocalDescription(offer);
			await send('offer', { sdp: offer.sdp, type: offer.type });
		} catch (e) {
			error = e instanceof ApiError ? e.message : t('call_media_denied');
			callState = 'error';
			cleanup();
		}
	}

	async function answerCall() {
		if (!remoteOffer) return;
		error = '';
		try {
			await ensureMedia(true);
			pc = new RTCPeerConnection(ICE_SERVERS);
			localStream?.getTracks().forEach((tr) => pc?.addTrack(tr, localStream as MediaStream));
			attachHandlers();
			const sdp = remoteOffer.payload['sdp'] as string | undefined;
			const type = (remoteOffer.payload['type'] as RTCSdpType | undefined) ?? 'offer';
			if (!sdp) throw new Error('bad offer');
			await pc.setRemoteDescription({ sdp, type });
			const answer = await pc.createAnswer();
			await pc.setLocalDescription(answer);
			await send('answer', { sdp: answer.sdp, type: answer.type });
			remoteOffer = null;
			callState = 'calling';
		} catch {
			error = t('call_failed');
			callState = 'error';
			cleanup();
		}
	}

	async function declineCall() {
		remoteOffer = null;
		callState = 'idle';
		await send('hangup', { reason: 'declined' }).catch(() => {});
	}

	async function poll() {
		let signals: AssistSignal[] = [];
		try {
			const res = await api.listAssistSignals(sessionId, lastSeen || undefined);
			signals = res.signals;
		} catch {
			return;
		}
		for (const sig of signals) {
			lastSeen = (sig as AssistSignal & { created_at: string }).created_at > lastSeen
				? (sig as AssistSignal & { created_at: string }).created_at
				: lastSeen;
			if (sig.mine) continue;
			if (sig.kind === 'offer' && callState === 'idle') {
				remoteOffer = sig;
				callState = 'incoming';
			} else if (sig.kind === 'answer' && pc && callState === 'calling') {
				const sdp = sig.payload['sdp'] as string | undefined;
				if (sdp) {
					try {
						await pc.setRemoteDescription({ sdp, type: 'answer' });
					} catch {
						/* late/duplicate answer */
					}
				}
			} else if (sig.kind === 'ice' && pc) {
				const c = sig.payload['candidate'] as string | undefined;
				if (c) {
					try {
						await pc.addIceCandidate({
							candidate: c,
							sdpMid: (sig.payload['sdpMid'] as string | undefined) ?? null,
							sdpMLineIndex: (sig.payload['sdpMLineIndex'] as number | undefined) ?? null
						});
					} catch {
						/* out-of-order candidate */
					}
				}
			} else if (sig.kind === 'hangup') {
				endCall(false);
			}
		}
	}

	function toggleMute() {
		muted = !muted;
		localStream?.getAudioTracks().forEach((tr) => (tr.enabled = !muted));
	}

	function toggleVideo() {
		videoOn = !videoOn;
		localStream?.getVideoTracks().forEach((tr) => (tr.enabled = videoOn));
	}

	function cleanup() {
		try {
			pc?.close();
		} catch {
			/* ignore */
		}
		pc = null;
		localStream?.getTracks().forEach((tr) => tr.stop());
		localStream = null;
		remoteStream = null;
		if (remoteVideo) remoteVideo.srcObject = null;
	}

	async function endCall(notify = true) {
		if (notify) await send('hangup', {}).catch(() => {});
		cleanup();
		remoteOffer = null;
		if (callState !== 'error') callState = 'ended';
	}

	onMount(() => {
		pollTimer = setInterval(poll, 2500);
		poll();
		return () => {
			if (pollTimer) clearInterval(pollTimer);
		};
	});

	onDestroy(() => {
		if (pollTimer) clearInterval(pollTimer);
		cleanup();
	});
</script>

<div class="callbox panel soft pad">
	<p class="eyebrow">{t('call_title')}</p>
	{#if error}<div class="banner" style="margin-top:10px">{error}</div>{/if}
	{#if mediaNotice}<p class="muted sm" style="margin-top:10px">{mediaNotice}</p>{/if}

	{#if callState === 'idle' || callState === 'ended'}
		<div class="callbtns">
			<button class="btn sm" onclick={() => startCall(true)}>{t('call_video')}</button>
			<button class="btn ghost sm" onclick={() => startCall(false)}>{t('call_audio')}</button>
		</div>
		<p class="muted sm" style="margin-top:8px">{t('call_stun_note')}</p>
	{:else if callState === 'incoming'}
		<p style="margin-top:10px; font-weight:700">{t('call_incoming')}</p>
		<div class="callbtns">
			<button class="btn sm" onclick={answerCall}>{t('call_answer')}</button>
			<button class="btn ghost sm" onclick={declineCall}>{t('call_decline')}</button>
		</div>
	{:else}
		<div class="videos">
			<video bind:this={remoteVideo} autoplay playsinline class="remote"></video>
			<video bind:this={localVideo} autoplay playsinline muted class="local"></video>
		</div>
		<p class="muted sm" style="margin-top:8px">
			{callState === 'calling' ? t('call_connecting') : t('call_connected')}
		</p>
		<div class="callbtns">
			<button class="btn quiet sm" onclick={toggleMute} aria-pressed={muted}>
				{muted ? t('call_unmute') : t('call_mute')}
			</button>
			<button class="btn quiet sm" onclick={toggleVideo} aria-pressed={videoOn}>
				{videoOn ? t('call_video_off') : t('call_video_on')}
			</button>
			<button class="btn ghost sm" onclick={() => endCall()}>{t('call_end')}</button>
		</div>
	{/if}
</div>

<style>
	.callbox {
		margin-top: 18px;
	}
	.callbtns {
		display: flex;
		gap: 8px;
		margin-top: 12px;
		flex-wrap: wrap;
	}
	.videos {
		position: relative;
		margin-top: 12px;
		border-radius: var(--r-md);
		overflow: hidden;
		background: #000;
		aspect-ratio: 4 / 3;
	}
	.remote {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.local {
		position: absolute;
		right: 10px;
		bottom: 10px;
		width: 96px;
		aspect-ratio: 3 / 4;
		object-fit: cover;
		border-radius: var(--r-sm);
		background: #222;
	}
	.sm {
		font-size: var(--step--1);
	}
</style>
