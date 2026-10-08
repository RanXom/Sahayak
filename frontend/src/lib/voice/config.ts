// Central voice-capture configuration. Every limit lives here so it can be
// tuned without touching the recorder or UI. Values can be overridden at
// runtime via `configureVoice({...})`.

export interface VoiceConfig {
	/** Absolute hard cap on a single recording. */
	maxDurationMs: number;
	/** Continuous silence (after speech) that auto-stops the recording. */
	silenceDurationMs: number;
	/** Absolute RMS floor considered silence (used if calibration is inconclusive). */
	silenceThreshold: number;
	/** Multiplier applied to the measured noise floor to derive the speech threshold. */
	noiseFloorMultiplier: number;
	/** How long to sample the room before trusting the noise floor. */
	calibrationMs: number;
	/** Minimum speech before silence detection is armed (avoids stopping on a cough). */
	minSpeechMs: number;
	/** Sampling cadence for the level monitor. */
	sampleIntervalMs: number;
	/** Minimum RMS to count as "speech" for arming. */
	minSpeechRms: number;
}

export const defaultVoiceConfig: VoiceConfig = {
	maxDurationMs: 120_000,
	silenceDurationMs: 12_000,
	silenceThreshold: 0.02,
	noiseFloorMultiplier: 3.2,
	calibrationMs: 600,
	minSpeechMs: 500,
	sampleIntervalMs: 100,
	minSpeechRms: 0.025
};

let config: VoiceConfig = { ...defaultVoiceConfig };

export function voiceConfig(): VoiceConfig {
	return config;
}

export function configureVoice(partial: Partial<VoiceConfig>): void {
	config = { ...config, ...partial };
}

export type StopReason = 'manual' | 'silence' | 'max-duration' | 'error';
