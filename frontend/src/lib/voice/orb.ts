// The Sahayak voice orb.
//
// A canvas-2D "liquid intelligence" sphere: a deformed circle whose silhouette
// is driven by summed sine harmonics, filled with moving radial gradients, lit
// by internal light blobs, finished with a rim light and a clipped grain
// texture. It reacts to real audio by reading an AnalyserNode (microphone
// while listening, response audio while speaking).

export type OrbState = 'idle' | 'listening' | 'processing' | 'speaking' | 'error';

interface Params {
	energy: number; // 0..1 overall activity
	flow: number; // 0..1 internal flow speed
	deform: number; // 0..1 silhouette deformation
	glow: number; // 0..1 outer glow
	scale: number; // relative size
	sat: number; // 0..1 saturation multiplier
}

const TARGETS: Record<OrbState, Params> = {
	idle: { energy: 0.08, flow: 0.16, deform: 0.34, glow: 0.22, scale: 1.0, sat: 0.9 },
	listening: { energy: 0.5, flow: 0.55, deform: 1.0, glow: 0.5, scale: 1.04, sat: 1.0 },
	processing: { energy: 0.34, flow: 0.92, deform: 0.5, glow: 0.42, scale: 0.99, sat: 0.95 },
	speaking: { energy: 0.5, flow: 0.62, deform: 0.92, glow: 0.56, scale: 1.03, sat: 1.0 },
	error: { energy: 0.06, flow: 0.1, deform: 0.18, glow: 0.08, scale: 0.96, sat: 0.28 }
};

const TAU = Math.PI * 2;
const POINTS = 160;

function lerp(a: number, b: number, t: number): number {
	return a + (b - a) * t;
}
function clamp(v: number, lo: number, hi: number): number {
	return Math.min(hi, Math.max(lo, v));
}

export class Orb {
	private canvas: HTMLCanvasElement;
	private ctx: CanvasRenderingContext2D;
	private raf = 0;
	private time = 0;
	private last = 0;
	private w = 0;
	private h = 0;
	private dpr = 1;
	private state: OrbState = 'idle';
	private params: Params = { ...TARGETS.idle };
	private analyser: AnalyserNode | null = null;
	private freq: Uint8Array<ArrayBuffer> | null = null;
	private timeData: Uint8Array<ArrayBuffer> | null = null;
	private energy = 0;
	private bands = { low: 0, mid: 0, high: 0 };
	private noise: HTMLCanvasElement | null = null;
	private reduced = false;
	private ro: ResizeObserver | null = null;
	private visible = true;

	constructor(canvas: HTMLCanvasElement) {
		this.canvas = canvas;
		const ctx = canvas.getContext('2d');
		if (!ctx) throw new Error('canvas 2d unavailable');
		this.ctx = ctx;
		this.reduced =
			typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches;
		this.noise = this.makeNoise();
		this.resize();
		this.ro = new ResizeObserver(() => this.resize());
		this.ro.observe(canvas.parentElement ?? canvas);
		this.onVisibility = this.onVisibility.bind(this);
		document.addEventListener('visibilitychange', this.onVisibility);
	}

	private onVisibility(): void {
		this.visible = !document.hidden;
	}

	setState(state: OrbState): void {
		this.state = state;
	}

	setAnalyser(node: AnalyserNode | null): void {
		this.analyser = node;
		if (node) {
			this.freq = new Uint8Array(node.frequencyBinCount);
			this.timeData = new Uint8Array(node.fftSize);
		} else {
			this.freq = null;
			this.timeData = null;
		}
	}

	private resize(): void {
		const rect = (this.canvas.parentElement ?? this.canvas).getBoundingClientRect();
		const size = Math.max(1, Math.min(rect.width, rect.height));
		this.dpr = clamp(window.devicePixelRatio || 1, 1, 2);
		this.w = size;
		this.h = size;
		this.canvas.width = Math.round(size * this.dpr);
		this.canvas.height = Math.round(size * this.dpr);
		this.canvas.style.width = `${size}px`;
		this.canvas.style.height = `${size}px`;
	}

	start(): void {
		this.last = performance.now();
		const loop = (now: number) => {
			this.raf = requestAnimationFrame(loop);
			const dt = Math.min(0.05, (now - this.last) / 1000);
			this.last = now;
			if (!this.visible) return;
			this.frame(dt);
		};
		this.raf = requestAnimationFrame(loop);
	}

	destroy(): void {
		cancelAnimationFrame(this.raf);
		this.ro?.disconnect();
		document.removeEventListener('visibilitychange', this.onVisibility);
	}

	private readAudio(): void {
		if (!this.analyser || !this.freq || !this.timeData) {
			this.energy = lerp(this.energy, 0, 0.1);
			this.bands = { low: 0, mid: 0, high: 0 };
			return;
		}
		try {
			this.analyser.getByteFrequencyData(this.freq);
			this.analyser.getByteTimeDomainData(this.timeData);
		} catch {
			// The source's AudioContext may have closed; fade out gracefully.
			this.energy = lerp(this.energy, 0, 0.1);
			return;
		}

		let sum = 0;
		for (let i = 0; i < this.timeData.length; i++) {
			const v = (this.timeData[i] - 128) / 128;
			sum += v * v;
		}
		const rms = Math.sqrt(sum / this.timeData.length);
		const target = clamp(rms * 3.4, 0, 1);
		const k = target > this.energy ? 0.4 : 0.08; // attack faster than release
		this.energy = lerp(this.energy, target, k);

		const n = this.freq.length;
		const avg = (a: number, b: number) => {
			let s = 0;
			for (let i = a; i < b; i++) s += this.freq![i];
			return s / (b - a) / 255;
		};
		this.bands.low = lerp(this.bands.low, avg(1, Math.floor(n * 0.06)), 0.3);
		this.bands.mid = lerp(this.bands.mid, avg(Math.floor(n * 0.06), Math.floor(n * 0.28)), 0.3);
		this.bands.high = lerp(this.bands.high, avg(Math.floor(n * 0.28), Math.floor(n * 0.6)), 0.3);
	}

	private frame(dt: number): void {
		const speed = this.reduced ? 0.12 : 1;
		this.time += dt * speed;
		this.readAudio();

		const target = TARGETS[this.state];
		const p = this.params;
		const ease = this.reduced ? 0.03 : 0.06;
		// Audio boosts energy/deform only in reactive states.
		const reactive = this.state === 'listening' || this.state === 'speaking';
		const audio = reactive ? this.energy : 0;

		p.energy = lerp(p.energy, target.energy + audio * 0.55, ease);
		p.flow = lerp(p.flow, target.flow + audio * 0.4, ease);
		p.deform = lerp(p.deform, target.deform + audio * 0.5, ease);
		p.glow = lerp(p.glow, target.glow + audio * 0.5, ease);
		p.scale = lerp(p.scale, target.scale + audio * 0.06, ease);
		p.sat = lerp(p.sat, target.sat, ease);

		this.draw(p, audio);
	}

	private draw(p: Params, audio: number): void {
		const ctx = this.ctx;
		const { w, h } = this;
		ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
		ctx.clearRect(0, 0, w, h);

		const cx = w / 2;
		const cy = h / 2;
		const R = Math.min(w, h) * 0.34 * p.scale;
		const t = this.time;
		const hue = 158 + p.flow * 10 - audio * 6;

		// --- silhouette (deformed circle) ---
		const pts: Array<[number, number]> = [];
		for (let i = 0; i < POINTS; i++) {
			const a = (i / POINTS) * TAU;
			const wave =
				Math.sin(3 * a + t * 0.7) * 0.5 +
				Math.sin(5 * a - t * 0.9) * 0.28 +
				Math.sin(7 * a + t * 1.3) * 0.16 +
				Math.sin(2 * a - t * 0.4) * 0.2;
			const r = R * (1 + wave * (0.05 + p.deform * 0.16));
			pts.push([cx + Math.cos(a) * r, cy + Math.sin(a) * r]);
		}
		const path = new Path2D();
		path.moveTo(pts[0][0], pts[0][1]);
		for (let i = 1; i < pts.length; i++) path.lineTo(pts[i][0], pts[i][1]);
		path.closePath();

		// --- outer atmosphere ---
		const glowR = R * (1.7 + p.glow * 0.5);
		const glow = ctx.createRadialGradient(cx, cy, R * 0.7, cx, cy, glowR);
		glow.addColorStop(0, `hsla(${hue}, ${55 * p.sat}%, 42%, ${0.28 * p.glow})`);
		glow.addColorStop(1, 'hsla(160, 40%, 30%, 0)');
		ctx.fillStyle = glow;
		ctx.beginPath();
		ctx.arc(cx, cy, glowR, 0, TAU);
		ctx.fill();

		// --- base body ---
		const ox = Math.cos(t * 0.5) * R * 0.22;
		const oy = Math.sin(t * 0.4) * R * 0.22;
		const body = ctx.createRadialGradient(
			cx - R * 0.3 + ox,
			cy - R * 0.34 + oy,
			R * 0.1,
			cx,
			cy,
			R * 1.05
		);
		body.addColorStop(0, `hsl(${hue + 6}, ${70 * p.sat}%, ${62 + p.energy * 8}%)`);
		body.addColorStop(0.45, `hsl(${hue}, ${62 * p.sat}%, ${38 + p.energy * 6}%)`);
		body.addColorStop(0.8, `hsl(${hue - 4}, ${55 * p.sat}%, ${20 + p.energy * 4}%)`);
		body.addColorStop(1, `hsl(${hue - 8}, ${50 * p.sat}%, ${12}%)`);
		ctx.save();
		ctx.clip(path);
		ctx.fillStyle = body;
		ctx.fillRect(cx - R * 1.3, cy - R * 1.3, R * 2.6, R * 2.6);

		// --- internal light blobs (flowing) ---
		ctx.globalCompositeOperation = 'lighter';
		for (let b = 0; b < 3; b++) {
			const ph = t * (0.5 + p.flow * 0.8) + b * 2.1;
			const bx = cx + Math.cos(ph) * R * (0.34 + 0.1 * b) + ox * 0.4;
			const by = cy + Math.sin(ph * 1.2) * R * (0.3 + 0.08 * b) + oy * 0.4;
			const br = R * (0.5 - b * 0.08);
			const blob = ctx.createRadialGradient(bx, by, 0, bx, by, br);
			const l = 66 - b * 6 + audio * 12;
			blob.addColorStop(0, `hsla(${hue + 10 + b * 8}, ${72 * p.sat}%, ${l}%, ${0.22 + audio * 0.2})`);
			blob.addColorStop(1, 'hsla(160, 60%, 40%, 0)');
			ctx.fillStyle = blob;
			ctx.beginPath();
			ctx.arc(bx, by, br, 0, TAU);
			ctx.fill();
		}
		ctx.globalCompositeOperation = 'source-over';

		// --- grain texture, clipped to the body ---
		if (this.noise) {
			ctx.globalCompositeOperation = 'overlay';
			ctx.globalAlpha = 0.05;
			ctx.fillStyle = ctx.createPattern(this.noise, 'repeat')!;
			ctx.fillRect(cx - R * 1.3, cy - R * 1.3, R * 2.6, R * 2.6);
			ctx.globalAlpha = 1;
			ctx.globalCompositeOperation = 'source-over';
		}

		// --- inner vignette for depth ---
		const vig = ctx.createRadialGradient(cx, cy, R * 0.35, cx, cy, R);
		vig.addColorStop(0, 'rgba(0,0,0,0)');
		vig.addColorStop(1, 'rgba(0,0,0,0.28)');
		ctx.fillStyle = vig;
		ctx.fillRect(cx - R * 1.3, cy - R * 1.3, R * 2.6, R * 2.6);
		ctx.restore();

		// --- rim light ---
		const rim = ctx.createLinearGradient(cx - R, cy - R, cx + R, cy + R);
		rim.addColorStop(0, `hsla(${hue + 20}, 90%, ${72}%, ${0.55 * p.sat})`);
		rim.addColorStop(0.5, 'rgba(255,255,255,0.06)');
		rim.addColorStop(1, 'rgba(255,255,255,0.02)');
		ctx.strokeStyle = rim;
		ctx.lineWidth = 1.25;
		ctx.stroke(path);
	}

	private makeNoise(): HTMLCanvasElement {
		const c = document.createElement('canvas');
		c.width = c.height = 96;
		const ctx = c.getContext('2d');
		if (!ctx) return c;
		const img = ctx.createImageData(96, 96);
		for (let i = 0; i < img.data.length; i += 4) {
			const v = 200 + Math.random() * 55;
			img.data[i] = img.data[i + 1] = img.data[i + 2] = v;
			img.data[i + 3] = 255;
		}
		ctx.putImageData(img, 0, 0);
		return c;
	}
}
