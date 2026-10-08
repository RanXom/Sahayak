package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// SarvamProvider implements STT and TTS using Sarvam AI (Saarika/Saaras and Bulbul).
type SarvamProvider struct {
	APIKey   string
	BaseURL  string
	STTModel string
	TTSModel string
	Speaker  string
	Client   *http.Client
}

// NewSarvamProvider constructs a Sarvam provider.
func NewSarvamProvider(apiKey, sttModel, ttsModel, speaker string) *SarvamProvider {
	if sttModel == "" {
		sttModel = "saaras:v4"
	}
	if ttsModel == "" {
		ttsModel = "bulbul:v3"
	}
	if speaker == "" {
		speaker = "priya"
	}
	return &SarvamProvider{
		APIKey:   apiKey,
		BaseURL:  "https://api.sarvam.ai",
		STTModel: sttModel,
		TTSModel: ttsModel,
		Speaker:  speaker,
		Client:   &http.Client{Timeout: 45 * time.Second},
	}
}

// Transcribe sends audio to Sarvam's speech-to-text endpoint.
func (p *SarvamProvider) Transcribe(ctx context.Context, audio []byte, lang, format string) (*TranscribeResult, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", "audio."+normalizeFormat(format))
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(audio); err != nil {
		return nil, err
	}
	if lang == "" {
		lang = "unknown"
	}
	_ = mw.WriteField("language_code", sarvamLang(lang))
	_ = mw.WriteField("model", p.STTModel)
	// "mode" is only applicable to saaras:v3; saaras:v4 selects the mode
	// implicitly and rejects unsupported combinations.
	if strings.HasPrefix(p.STTModel, "saaras:v3") {
		_ = mw.WriteField("mode", "transcribe")
	}
	mw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/speech-to-text", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("api-subscription-key", p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sarvam stt status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Transcript          string  `json:"transcript"`
		LanguageCode        string  `json:"language_code"`
		LanguageProbability float64 `json:"language_probability"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode sarvam stt: %w", err)
	}
	return &TranscribeResult{Text: out.Transcript, Language: out.LanguageCode, Confidence: out.LanguageProbability}, nil
}

// Synthesize sends text to Sarvam's text-to-speech endpoint.
func (p *SarvamProvider) Synthesize(ctx context.Context, text, lang string) (*SynthesizeResult, error) {
	if lang == "" {
		lang = "hi-IN"
	}
	payload := map[string]any{
		"text":          text,
		"language_code": sarvamLang(lang),
		"speaker":       p.Speaker,
		"model":         p.TTSModel,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/text-to-speech", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-subscription-key", p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<24))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sarvam tts status %d: %s", resp.StatusCode, string(respBody))
	}
	var out struct {
		Audios []string `json:"audios"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode sarvam tts: %w", err)
	}
	if len(out.Audios) == 0 {
		return nil, fmt.Errorf("sarvam tts returned no audio")
	}
	audio, err := base64.StdEncoding.DecodeString(out.Audios[0])
	if err != nil {
		return nil, fmt.Errorf("decode sarvam audio: %w", err)
	}
	return &SynthesizeResult{Audio: audio, Format: "wav", SampleRate: 22050}, nil
}

func sarvamLang(code string) string {
	switch code {
	case "", "unknown":
		return "unknown"
	case "hi", "hi-IN":
		return "hi-IN"
	case "en", "en-IN":
		return "en-IN"
	case "ta":
		return "ta-IN"
	case "te":
		return "te-IN"
	case "kn":
		return "kn-IN"
	case "bn":
		return "bn-IN"
	case "mr":
		return "mr-IN"
	case "gu":
		return "gu-IN"
	case "pa":
		return "pa-IN"
	default:
		return "hi-IN"
	}
}

func normalizeFormat(format string) string {
	switch format {
	case "webm", "ogg", "wav", "mp3", "m4a":
		return format
	default:
		return "wav"
	}
}
