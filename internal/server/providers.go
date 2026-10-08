package server

import (
	"context"
	"os"

	"github.com/sahayak/sahayak/internal/auth"
	"github.com/sahayak/sahayak/internal/platform/llm"
	"github.com/sahayak/sahayak/internal/platform/sms"
	"github.com/sahayak/sahayak/internal/platform/translate"
	"github.com/sahayak/sahayak/internal/platform/voice"
	"github.com/sahayak/sahayak/internal/platform/whatsapp"
)

// smsProvider selects the configured SMS provider, falling back to a console
// provider when no credentials are configured (development).
func (s *Server) smsProvider() sms.Provider {
	switch s.cfg.SMSProvider {
	case "msg91":
		if s.cfg.MSG91AuthKey != "" {
			return sms.NewMSG91Provider(s.cfg.MSG91AuthKey, s.cfg.MSG91Sender)
		}
	}
	return sms.NewConsoleProvider(s.logger)
}

// geminiProvider returns a Gemini provider or nil when no key is set.
func (s *Server) geminiProvider() llm.Provider {
	if s.cfg.Gemini.APIKey == "" {
		return nil
	}
	return llm.NewGeminiProvider(s.cfg.Gemini.APIKey, s.cfg.Gemini.Model, s.cfg.Gemini.BaseURL, s.cfg.LLMTimeout)
}

// openAIProvider returns an OpenAI-compatible provider or nil when no key is set.
func (s *Server) openAIProvider() llm.Provider {
	if s.cfg.OpenAI.APIKey == "" {
		return nil
	}
	return llm.NewOpenAIProvider(s.cfg.OpenAI.APIKey, s.cfg.OpenAI.Model, s.cfg.OpenAI.BaseURL, s.cfg.LLMTimeout)
}

// llmProvider builds a failover chain: the configured provider first, the other
// paid provider second, and a deterministic fallback last. A single provider
// outage therefore degrades rather than breaks.
func (s *Server) llmProvider() llm.Provider {
	gemini := s.geminiProvider()
	openai := s.openAIProvider()

	type named struct {
		name string
		p    llm.Provider
	}
	var ordered []named
	if s.cfg.LLMProvider == "openai" {
		ordered = []named{{"openai", openai}, {"gemini", gemini}}
	} else {
		ordered = []named{{"gemini", gemini}, {"openai", openai}}
	}

	var chain []llm.Provider
	for _, n := range ordered {
		if n.p != nil {
			chain = append(chain, llm.Named{Name: n.name, Provider: n.p})
		}
	}
	chain = append(chain, llm.Named{Name: "fallback", Provider: llm.NewFallbackProvider()})

	if len(chain) == 1 {
		return chain[0]
	}
	return llm.NewChainWithLogger(s.logger, chain...)
}

// llmLive reports whether at least one real LLM provider is configured.
func (s *Server) llmLive() bool {
	return s.cfg.Gemini.APIKey != "" || s.cfg.OpenAI.APIKey != ""
}

// voiceProviders builds STT/TTS failover chains from the configured providers.
func (s *Server) voiceProviders() (voice.STTProvider, voice.TTSProvider) {
	type bundle struct {
		stt voice.STTProvider
		tts voice.TTSProvider
	}
	sarvam := bundle{}
	if s.cfg.SarvamAPIKey != "" {
		p := voice.NewSarvamProvider(s.cfg.SarvamAPIKey, s.cfg.Sarvam.STTModel, s.cfg.Sarvam.TTSModel, s.cfg.Sarvam.Speaker)
		sarvam = bundle{stt: p, tts: p}
	}
	bhashini := bundle{}
	if s.cfg.Bhashini.APIKey != "" {
		p := voice.NewBhashiniProvider(s.cfg.Bhashini.APIKey, s.cfg.Bhashini.UserID, s.cfg.Bhashini.BaseURL)
		bhashini = bundle{stt: p, tts: p}
	}

	ordered := []bundle{}
	first, second := sarvam, bhashini
	if s.cfg.VoiceProvider == "bhashini" {
		first, second = bhashini, sarvam
	}
	if first.stt != nil {
		ordered = append(ordered, first)
	}
	if second.stt != nil {
		ordered = append(ordered, second)
	}
	if len(ordered) == 0 {
		s.logger.Warn("voice_provider_disabled", "reason", "no voice API key configured", "provider", s.cfg.VoiceProvider)
		return nil, nil
	}
	if len(ordered) == 1 {
		return ordered[0].stt, ordered[0].tts
	}

	stts := make([]voice.STTProvider, 0, len(ordered))
	ttss := make([]voice.TTSProvider, 0, len(ordered))
	for _, b := range ordered {
		stts = append(stts, b.stt)
		ttss = append(ttss, b.tts)
	}
	return voice.NewChainSTT(stts...), voice.NewChainTTS(ttss...)
}

// voiceLive reports whether at least one voice provider is configured.
func (s *Server) voiceLive() bool {
	return s.cfg.SarvamAPIKey != "" || s.cfg.Bhashini.APIKey != ""
}

// firebaseVerifier builds the Firebase ID-token verifier, or nil when
// Firebase phone auth is not configured.
func (s *Server) firebaseVerifier() auth.TokenVerifier {
	if s.cfg.FirebaseProjectID == "" {
		return nil
	}
	creds := []byte(s.cfg.FirebaseCredentials)
	if len(creds) == 0 && s.cfg.FirebaseCredentialsFile != "" {
		b, err := os.ReadFile(s.cfg.FirebaseCredentialsFile)
		if err != nil {
			s.logger.Error("firebase_credentials_read_failed", "error", err.Error())
			return nil
		}
		creds = b
	}
	v, err := auth.NewFirebaseVerifier(context.Background(), s.cfg.FirebaseProjectID, creds)
	if err != nil {
		s.logger.Error("firebase_init_failed", "error", err.Error())
		return nil
	}
	return v
}

// firebaseLive reports whether Firebase phone auth is configured.
func (s *Server) firebaseLive() bool {
	return s.cfg.FirebaseProjectID != ""
}

// translateProvider returns the configured translator, defaulting to passthrough.
func (s *Server) translateProvider() translate.Provider {
	if s.cfg.VoiceProvider == "bhashini" && s.cfg.Bhashini.APIKey != "" {
		return translate.NewBhashiniNMT(s.cfg.Bhashini.APIKey, s.cfg.Bhashini.UserID, s.cfg.Bhashini.BaseURL, "")
	}
	return translate.NewFallback()
}

// whatsappProvider selects the configured WhatsApp provider, falling back to a
// console provider in development.
func (s *Server) whatsappProvider() whatsapp.Provider {
	switch s.cfg.WhatsAppProvider {
	case "gupshup":
		if s.cfg.WhatsAppAPIKey != "" {
			return whatsapp.NewGupshupProvider(s.cfg.WhatsAppAPIKey)
		}
	}
	return whatsapp.NewConsoleProvider(s.logger)
}
