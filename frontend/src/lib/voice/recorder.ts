import { voiceConfig, type StopReason } from './config';

export interface RecorderEvents {
	onLevel?: (level: number) => void;
	onElapsed?: (ms: number) => void;
	onAutoStop?: (reason: StopReason, blob: Blob) => void;
}

export function isRecordingSupported(): boolean {
	return (
		typeof window !== 'undefined' &&
		typeof navigator !== 'undefined' &&
		!!navigator.mediaDevices?.getUserMedia &&
		typeof MediaRecorder !== 'undefined'
	);
}

function pickMimeType(): string {
	const candidates = ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/ogg'];
	for (const c of candidates) {
		if (typeof MediaRecorder !== 'undefined' && MediaRecorder.isTypeSupported?.(c)) return c;
	}
	return '';
}

/**
 * VoiceRecorder captures microphone audio, exposes a smoothed amplitude level
 * for visualisation, and enforces the recording safety limits (hard duration
 * cap + silence auto-stop) defined in `voice/config.ts`.
 */
export class VoiceRecorder {
	private stream: MediaStream | null = null;
	private recorder: MediaRecorder | null = null;
	private chunks: Blob[] = [];
	private ctx: AudioContext | null = null;
	private analyser: AnalyserNode | null = null;
	private timeData: Uint8Array<ArrayBuffer> | null = null;

	private timer: number | null = null;
	private startedAt = 0;
	private stopped = false;
	private events: RecorderEvents = {};

	// Silence detection state
	private calibrating = true;
	private calibration: number[] = [];
	private noiseFloor = 0;
	private threshold = voiceConfig().silenceThreshold;
	private speechStartedAt = 0;
	private armed = false;
	private lastVoiceAt = 0;
	private level = 0;

	async start(events: RecorderEvents = {}): Promise<void> {
		if (!isRecordingSupported()) throw new Error('Recording is not supported on this device');
		this.events = events;
		this.chunks = [];
		this.stopped = false;
		this.calibrating = true;
		this.calibration = [];
		this.armed = false;
		this.speechStartedAt = 0;
		this.level = 0;

		this.stream = await navigator.mediaDevices.getUserMedia({
			audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true }
		});

		const AudioCtx =
			window.AudioContext ?? (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
		this.ctx = new AudioCtx();
		if (this.ctx.state === 'suspended') await this.ctx.resume();
		const source = this.ctx.createMediaStreamSource(this.stream);
		this.analyser = this.ctx.createAnalyser();
		this.analyser.fftSize = 1024;
		this.analyser.smoothingTimeConstant = 0.75;
		source.connect(this.analyser);
		this.timeData = new Uint8Array(this.analyser.fftSize);

		const mimeType = pickMimeType();
		this.recorder = mimeType
			? new MediaRecorder(this.stream, { mimeType })
			: new MediaRecorder(this.stream);
		this.recorder.ondataavailable = (e) => {
			if (e.data.size > 0) this.chunks.push(e.data);
		};
		this.recorder.start(250);

		this.startedAt = performance.now();
		const cfg = voiceConfig();
		this.timer = window.setInterval(() => this.tick(), cfg.sampleIntervalMs);
	}

	private tick(): void {
		if (this.stopped || !this.analyser || !this.timeData) return;
		const cfg = voiceConfig();
		const now = performance.now();
		const elapsed = now - this.startedAt;

		const rms = this.readRms();
		// Attack/release smoothing so the level (and silence test) is not jittery.
		const attack = 0.5;
		const release = 0.12;
		this.level = rms > this.level ? this.level + (rms - this.level) * attack : this.level + (rms - this.level) * release;
		this.events.onLevel?.(Math.min(1, this.level * 4));
		this.events.onElapsed?.(elapsed);

		// Calibrate the noise floor over the first moments of the recording.
		if (this.calibrating) {
			this.calibration.push(rms);
			if (elapsed >= cfg.calibrationMs) {
				this.calibrating = false;
				this.calibration.sort((a, b) => a - b);
				this.noiseFloor = this.calibration[Math.floor(this.calibration.length / 2)] ?? 0;
				this.threshold = Math.max(this.noiseFloor * cfg.noiseFloorMultiplier, cfg.silenceThreshold);
			}
			if (elapsed >= cfg.maxDurationMs) this.autoStop('max-duration');
			return;
		}

		const isSpeech = this.level > Math.max(this.threshold, cfg.minSpeechRms);
		if (!this.armed) {
			if (isSpeech) {
				if (this.speechStartedAt === 0) this.speechStartedAt = now;
				if (now - this.speechStartedAt >= cfg.minSpeechMs) {
					this.armed = true;
					this.lastVoiceAt = now;
				}
			} else {
				this.speechStartedAt = 0;
			}
		} else {
			if (isSpeech) this.lastVoiceAt = now;
			else if (now - this.lastVoiceAt >= cfg.silenceDurationMs) {
				this.autoStop('silence');
				return;
			}
		}

		if (elapsed >= cfg.maxDurationMs) this.autoStop('max-duration');
	}

	private readRms(): number {
		if (!this.analyser || !this.timeData) return 0;
		this.analyser.getByteTimeDomainData(this.timeData);
		let sum = 0;
		for (let i = 0; i < this.timeData.length; i++) {
			const v = (this.timeData[i] - 128) / 128;
			sum += v * v;
		}
		return Math.sqrt(sum / this.timeData.length);
	}

	private autoStop(reason: StopReason): void {
		if (this.stopped) return;
		this.stop()
			.then((blob) => this.events.onAutoStop?.(reason, blob))
			.catch(() => this.events.onAutoStop?.(reason, new Blob()));
	}

	stop(): Promise<Blob> {
		return new Promise((resolve) => {
			if (this.stopped || !this.recorder) {
				resolve(new Blob(this.chunks, { type: 'audio/webm' }));
				return;
			}
			this.stopped = true;
			const finish = () => {
				const blob = new Blob(this.chunks, { type: this.recorder?.mimeType || 'audio/webm' });
				this.cleanup();
				resolve(blob);
			};
			this.recorder.onstop = finish;
			if (this.recorder.state !== 'inactive') this.recorder.stop();
			else finish();
		});
	}

	cancel(): void {
		this.stopped = true;
		try {
			if (this.recorder && this.recorder.state !== 'inactive') this.recorder.stop();
		} catch {
			/* noop */
		}
		this.cleanup();
	}

	levelValue(): number {
		return Math.min(1, this.level * 4);
	}

	getAnalyser(): AnalyserNode | null {
		return this.analyser;
	}

	private cleanup(): void {
		if (this.timer !== null) {
			clearInterval(this.timer);
			this.timer = null;
		}
		this.stream?.getTracks().forEach((t) => t.stop());
		this.stream = null;
		this.analyser = null;
		this.recorder = null;
		this.ctx?.close().catch(() => {});
		this.ctx = null;
	}
}
