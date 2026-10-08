package voice

import "context"

// ChainSTT tries multiple speech-to-text providers in order.
type ChainSTT struct {
	providers []STTProvider
}

// NewChainSTT builds an STT failover chain.
func NewChainSTT(providers ...STTProvider) *ChainSTT {
	return &ChainSTT{providers: providers}
}

// Transcribe tries each provider until one succeeds.
func (c *ChainSTT) Transcribe(ctx context.Context, audio []byte, lang, format string) (*TranscribeResult, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := p.Transcribe(ctx, audio, lang, format)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// ChainTTS tries multiple text-to-speech providers in order.
type ChainTTS struct {
	providers []TTSProvider
}

// NewChainTTS builds a TTS failover chain.
func NewChainTTS(providers ...TTSProvider) *ChainTTS {
	return &ChainTTS{providers: providers}
}

// Synthesize tries each provider until one succeeds.
func (c *ChainTTS) Synthesize(ctx context.Context, text, lang string) (*SynthesizeResult, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := p.Synthesize(ctx, text, lang)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
