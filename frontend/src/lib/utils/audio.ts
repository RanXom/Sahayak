// Audio capture helpers built on the MediaRecorder API.

export function isRecordingSupported(): boolean {
	return (
		typeof window !== 'undefined' &&
		typeof navigator !== 'undefined' &&
		!!navigator.mediaDevices &&
		typeof MediaRecorder !== 'undefined'
	);
}

export class Recorder {
	private recorder: MediaRecorder | null = null;
	private chunks: Blob[] = [];
	private stream: MediaStream | null = null;

	async start(): Promise<void> {
		if (!isRecordingSupported()) {
			throw new Error('Recording is not supported on this device');
		}
		this.stream = await navigator.mediaDevices.getUserMedia({ audio: true });
		const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
			? 'audio/webm;codecs=opus'
			: 'audio/webm';
		this.recorder = new MediaRecorder(this.stream, { mimeType });
		this.chunks = [];
		this.recorder.ondataavailable = (e) => {
			if (e.data.size > 0) this.chunks.push(e.data);
		};
		this.recorder.start();
	}

	stop(): Promise<Blob> {
		return new Promise((resolve, reject) => {
			if (!this.recorder) {
				reject(new Error('Recorder was not started'));
				return;
			}
			this.recorder.onstop = () => {
				const blob = new Blob(this.chunks, { type: 'audio/webm' });
				this.cleanup();
				resolve(blob);
			};
			this.recorder.stop();
		});
	}

	cancel(): void {
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		this.cleanup();
	}

	private cleanup(): void {
		this.stream?.getTracks().forEach((t) => t.stop());
		this.stream = null;
		this.recorder = null;
	}
}

export function playAudio(url: string): Promise<void> {
	return new Promise((resolve) => {
		const audio = new Audio(url);
		audio.onended = () => resolve();
		audio.onerror = () => resolve();
		audio.play().catch(() => resolve());
	});
}
