// Package voice defines the speech-to-text and text-to-speech provider
// abstractions. Providers are swappable via configuration.
package voice

import "context"

// TranscribeResult is the output of speech recognition.
type TranscribeResult struct {
	Text       string
	Language   string
	Confidence float64
}

// SynthesizeResult is the output of speech synthesis.
type SynthesizeResult struct {
	Audio      []byte
	Format     string
	SampleRate int
}

// STTProvider converts audio to text.
type STTProvider interface {
	Transcribe(ctx context.Context, audio []byte, lang, format string) (*TranscribeResult, error)
}

// TTSProvider converts text to audio.
type TTSProvider interface {
	Synthesize(ctx context.Context, text, lang string) (*SynthesizeResult, error)
}
