package conversation

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/location"
	"github.com/sahayak/sahayak/internal/platform/llm"
	"github.com/sahayak/sahayak/internal/platform/voice"
	"github.com/sahayak/sahayak/internal/profile"
)

// EligibilitySummarizer produces spoken-language eligibility summaries.
type EligibilitySummarizer interface {
	VoiceSummary(ctx context.Context, userID, lang string) (string, error)
	MissingSummary(ctx context.Context, userID, lang string) (string, error)
}

// Deps are the dependencies for the conversation service.
type Deps struct {
	Repo        *Repository
	Profiles    *profile.Service
	LLM         llm.Provider
	STT         voice.STTProvider
	TTS         voice.TTSProvider
	Audio       AudioStore
	Location    *location.Repository
	Audit       audit.Logger
	Logger      *slog.Logger
	Eligibility EligibilitySummarizer
}

// Service orchestrates the intake conversation.
type Service struct {
	repo        *Repository
	profiles    *profile.Service
	llm         llm.Provider
	stt         voice.STTProvider
	tts         voice.TTSProvider
	audio       AudioStore
	location    *location.Repository
	audit       audit.Logger
	logger      *slog.Logger
	eligibility EligibilitySummarizer
}

// NewService constructs the conversation service.
func NewService(d Deps) *Service {
	logger := d.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:        d.Repo,
		profiles:    d.Profiles,
		llm:         d.LLM,
		stt:         d.STT,
		tts:         d.TTS,
		audio:       d.Audio,
		location:    d.Location,
		audit:       d.Audit,
		logger:      logger,
		eligibility: d.Eligibility,
	}
}

// FollowUp describes the next intake question.
type FollowUp struct {
	QuestionAbout  string `json:"question_about"`
	IsLastQuestion bool   `json:"is_last_question"`
}

// TurnResult is the outcome of processing one user turn.
type TurnResult struct {
	ConversationID   string         `json:"conversation_id"`
	Transcript       string         `json:"transcript"`
	DetectedLanguage string         `json:"detected_language"`
	ResponseText     string         `json:"response_text"`
	ResponseAudioURL string         `json:"response_audio_url,omitempty"`
	FactsExtracted   map[string]any `json:"facts_extracted"`
	ProfileSnapshot  map[string]any `json:"profile_snapshot"`
	FollowUp         *FollowUp      `json:"follow_up,omitempty"`
	ReadyForReview   bool           `json:"ready_for_review"`
	TurnCount        int            `json:"turn_count"`
}

// Start creates a new active conversation.
func (s *Service) Start(ctx context.Context, userID, lang string) (*Conversation, error) {
	if lang == "" {
		lang = "hi"
	}
	p, err := s.profiles.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	conv, err := s.repo.Create(ctx, userID, &p.ID, lang)
	if err != nil {
		return nil, domain.Internal("Could not start conversation", err)
	}
	// Greeting.
	greeting := greetingText(lang)
	if _, err := s.repo.AddMessage(ctx, conv.ID, "assistant", greeting, lang, nil); err != nil {
		return nil, domain.Internal("Could not start conversation", err)
	}
	conv.Messages = []Message{{Role: "assistant", ContentText: greeting, ContentLang: lang}}
	return conv, nil
}

// Get returns a conversation with messages.
func (s *Service) Get(ctx context.Context, userID, convID string) (*Conversation, error) {
	conv, err := s.repo.Get(ctx, convID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Conversation not found")
		}
		return nil, domain.Internal("Could not load conversation", err)
	}
	return conv, nil
}

// GetOrStart returns the active conversation or starts one.
func (s *Service) GetOrStart(ctx context.Context, userID, lang string) (*Conversation, error) {
	conv, err := s.repo.GetActive(ctx, userID)
	if err == nil {
		return s.Get(ctx, userID, conv.ID)
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, domain.Internal("Could not load conversation", err)
	}
	return s.Start(ctx, userID, lang)
}

// HandleVoice transcribes audio then processes it as text.
func (s *Service) HandleVoice(ctx context.Context, userID, convID string, audio []byte, format, lang string) (*TurnResult, error) {
	if s.stt == nil {
		return nil, domain.Unavailable("Voice input is temporarily unavailable. Please type your response instead.")
	}
	conv, err := s.Get(ctx, userID, convID)
	if err != nil {
		return nil, err
	}
	if lang == "" {
		lang = conv.Language
	}
	// Ask the STT provider to auto-detect the spoken language rather than
	// forcing the UI language, then mirror whatever was actually spoken.
	sttStart := time.Now()
	res, err := s.stt.Transcribe(ctx, audio, "", format)
	if err != nil {
		s.logger.Error("stt_failed", "provider_error", err.Error(), "language", lang, "duration_ms", time.Since(sttStart).Milliseconds())
		return nil, domain.Unavailable("Could not understand the audio. Please try again or type your response.")
	}
	s.logger.Info("stt_ok", "language", res.Language, "chars", len(res.Text), "duration_ms", time.Since(sttStart).Milliseconds())
	if strings.TrimSpace(res.Text) == "" {
		return nil, domain.InvalidInput("We could not hear anything. Please try again.")
	}
	detected := res.Language
	if len(detected) > 2 {
		detected = detected[:2]
	}
	if detected == "" {
		detected = lang
	}
	result, err := s.process(ctx, userID, conv, res.Text, detected)
	if err != nil {
		return nil, err
	}
	result.Transcript = res.Text
	return result, nil
}

// HandleText processes a typed message.
func (s *Service) HandleText(ctx context.Context, userID, convID, text, lang string) (*TurnResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, domain.InvalidInput("Message text is required")
	}
	conv, err := s.Get(ctx, userID, convID)
	if err != nil {
		return nil, err
	}
	if lang == "" {
		lang = conv.Language
	}
	return s.process(ctx, userID, conv, text, lang)
}

func (s *Service) process(ctx context.Context, userID string, conv *Conversation, text, lang string) (*TurnResult, error) {
	// Mirror the user's script when they write in an Indian language, so the
	// reply language matches what they actually used. Otherwise fall back to the
	// provided (UI) language, detecting only when it is unknown.
	lang = resolveReplyLang(text, lang)
	// Latin script carries no language signal (English vs Romanized Hindi /
	// Marathi / Tamil look identical), so detect concurrently with extraction
	// below — no added latency. Only non-Hindi results override: "hi" is also
	// the chain's last-resort fallback, so trusting it would flip English to
	// Hindi whenever the LLM is down.
	var detectCh chan detectResult
	if detectScriptLang(text) == "" {
		detectCh = make(chan detectResult, 1)
		go func() {
			code, err := s.llm.DetectLanguage(ctx, text)
			detectCh <- detectResult{code: code, err: err}
		}()
	}
	if lang == "" {
		if detected, err := s.llm.DetectLanguage(ctx, text); err == nil && detected != "" {
			lang = detected
		} else {
			lang = conv.Language
		}
	}
	convCtx := conv.Context
	if convCtx == nil {
		convCtx = map[string]any{}
	}
	turnCount := intFromAny(convCtx["turn_count"]) + 1
	asked := stringSlice(convCtx["asked_fields"])

	p, err := s.profiles.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Extract facts.
	extractStart := time.Now()
	extraction, err := s.llm.ExtractFacts(ctx, llm.ExtractFactsRequest{
		UserText:       text,
		CurrentProfile: p.Facts,
		Language:       lang,
	})
	if err != nil {
		// Degrade gracefully: no facts extracted, still record the message.
		s.logger.Warn("llm_extract_failed", "error", err.Error(), "language", lang, "duration_ms", time.Since(extractStart).Milliseconds())
		extraction = &llm.ExtractFactsResponse{Facts: map[string]any{}}
	} else {
		s.logger.Debug("llm_extract_ok", "language", lang, "facts", len(extraction.Facts), "duration_ms", time.Since(extractStart).Milliseconds())
	}
	// Pick up the concurrent language detection (started above for
	// Latin-script input, overlapping extraction so it costs no latency).
	if detectCh != nil {
		if res := <-detectCh; res.err == nil && supportedReplyLang(res.code) && res.code != "hi" {
			s.logger.Info("llm_detect_override", "from", lang, "to", res.code)
			lang = res.code
		}
	}
	facts := profile.SanitizeExtractedFacts(extraction.Facts)

	// 2. Resolve state name to state_id if present.
	var stateID *int
	if stateName, ok := facts["state"].(string); ok && stateName != "" {
		if st, err := s.location.FindStateByName(ctx, stateName); err == nil {
			stateID = &st.ID
		}
	}

	// 3. Merge into profile.
	updatedProfile, err := s.profiles.MergeFacts(ctx, userID, facts, stateID, nil)
	if err != nil {
		return nil, err
	}

	// 4. Record the user's message.
	if _, err := s.repo.AddMessage(ctx, conv.ID, "user", text, lang, facts); err != nil {
		return nil, domain.Internal("Could not save message", err)
	}

	// 5. Decide next question.
	nextField := NextField(updatedProfile.Facts, asked)
	ready := nextField == "" || turnCount >= MaxTurns || (IsReady(updatedProfile.Facts) && turnCount >= 5)

	result := &TurnResult{
		ConversationID:   conv.ID,
		Transcript:       text,
		DetectedLanguage: lang,
		FactsExtracted:   facts,
		ProfileSnapshot:  updatedProfile.Facts,
		TurnCount:        turnCount,
		ReadyForReview:   ready,
	}

	responseText := ""
	intent := intentOf(text)
	switch {
	case intent == intentMissing && s.eligibility != nil:
		// The user deliberately asked what is missing: list it.
		if txt := s.missingSummary(ctx, userID, lang); txt != "" {
			responseText = txt
		} else {
			responseText = readyText(lang)
		}
	case intent == intentSchemes && s.eligibility != nil:
		// The user asked about schemes: answer with matches, not a cross-question.
		if txt := s.eligibilitySummary(ctx, userID, lang); txt != "" {
			responseText = txt
		} else {
			responseText = readyText(lang)
		}
	case !ready:
		followUpStart := time.Now()
		fu, err := s.llm.GenerateFollowUp(ctx, llm.FollowUpRequest{
			UserText:      text,
			ProfileFacts:  updatedProfile.Facts,
			History:       llmHistory(conv.Messages),
			MissingFields: []string{nextField},
			Language:      lang,
			Context:       convCtx,
		})
		if err != nil || fu.Question == "" {
			if err != nil {
				s.logger.Warn("llm_followup_failed", "error", err.Error(), "language", lang, "field", nextField)
			}
			fu = &llm.FollowUpResponse{Question: llm.QuestionForField(nextField, lang), QuestionAbout: nextField}
		}
		s.logger.Info("llm_followup_generated",
			"language", lang, "field", nextField,
			"chars", len(fu.Question), "duration_ms", time.Since(followUpStart).Milliseconds())
		responseText = fu.Question
		result.FollowUp = &FollowUp{QuestionAbout: fu.QuestionAbout, IsLastQuestion: false}
		asked = append(asked, nextField)
	default:
		// Intake complete: tell the user, in the conversation, which schemes they
		// match and what is still missing, rather than sending them to a page.
		if txt := s.eligibilitySummary(ctx, userID, lang); txt != "" {
			responseText = txt
		} else {
			responseText = readyText(lang)
		}
	}

	// 6. Synthesize speech (best effort). The exact text above — whether the
	// LLM answer or the deterministic fallback — is what gets spoken.
	if s.tts != nil && responseText != "" {
		ttsStart := time.Now()
		if synth, err := s.tts.Synthesize(ctx, responseText, lang); err == nil && len(synth.Audio) > 0 {
			if id, err := s.audio.Save(synth.Audio, synth.Format); err == nil {
				result.ResponseAudioURL = "/api/v1/audio/temp/" + id
			}
			s.logger.Info("tts_synth_ok", "language", lang, "chars", len(responseText), "bytes", len(synth.Audio), "duration_ms", time.Since(ttsStart).Milliseconds())
		} else if err != nil {
			s.logger.Warn("tts_synth_failed", "error", err.Error(), "language", lang, "duration_ms", time.Since(ttsStart).Milliseconds())
		}
	}
	result.ResponseText = responseText

	if _, err := s.repo.AddMessage(ctx, conv.ID, "assistant", responseText, lang, nil); err != nil {
		return nil, domain.Internal("Could not save message", err)
	}

	// 7. Persist context.
	newCtx := map[string]any{
		"state":                  map[bool]string{true: "ready", false: "follow_up"}[ready],
		"facts_collected":        factKeys(updatedProfile.Facts),
		"current_question_about": nextField,
		"turn_count":             turnCount,
		"max_turns":              MaxTurns,
		"asked_fields":           asked,
	}
	if err := s.repo.UpdateContext(ctx, conv.ID, newCtx, lang); err != nil {
		return nil, domain.Internal("Could not update conversation", err)
	}
	if ready {
		_ = s.repo.SetStatus(ctx, conv.ID, StatusCompleted)
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &userID,
		Action:       "conversation.turn",
		ResourceType: "conversation",
		ResourceID:   conv.ID,
		Details:      map[string]any{"turn": turnCount, "ready": ready},
	})
	return result, nil
}

func greetingText(lang string) string {
	if lang == "hi" {
		return "नमस्ते! मैं सहायक हूँ। मुझे अपने बारे में बताइए, जैसे आपकी उम्र, आप कहाँ रहते हैं, और आप क्या काम करते हैं।"
	}
	if lang == "mr" {
		return "नमस्ते! मी सहायक आहे. मला तुमच्याबद्दल सांगा — उदाहरणार्थ तुमचे वय, तुम्ही कुठे राहता आणि तुम्ही काय काम करता."
	}
	return "Namaste! I am Sahayak. Tell me about yourself — for example your age, where you live, and what work you do."
}

func readyText(lang string) string {
	if lang == "hi" {
		return "धन्यवाद! आपकी प्रोफ़ाइल तैयार है। अब आप अपनी जानकारी की समीक्षा करें और पात्र योजनाएँ देखें।"
	}
	if lang == "mr" {
		return "धन्यवाद! तुमची प्रोफाइल तयार आहे. कृपया तुमची माहिती तपासा आणि तुम्ही पात्र असलेल्या योजना पाहा."
	}
	return "Thank you! Your profile is ready. Please review your details and see the schemes you may be eligible for."
}

func factKeys(facts map[string]any) []string {
	out := make([]string, 0, len(facts))
	for k := range facts {
		out = append(out, k)
	}
	return out
}

// Conversation intents that change the assistant's behaviour.
const (
	intentSchemes = "schemes"
	intentMissing = "missing"
)

// intentOf classifies a user message. Missing-info questions take precedence,
// then scheme/eligibility questions; everything else is treated as intake.
func intentOf(text string) string {
	t := strings.ToLower(text)
	missing := []string{
		"what is missing", "what's missing", "what do you need", "what else", "what more",
		"kya missing", "kya chahiye", "kya baki", "aur kya", "kya kami",
		"क्या कमी", "क्या चाहिए", "क्या बाकी", "और क्या",
	}
	for _, k := range missing {
		if strings.Contains(t, k) {
			return intentMissing
		}
	}
	schemes := []string{
		"scheme", "yojana", "योजना", "योजनाए", "eligible", "eligibility", "पात्र", "पात्रता",
		"benefit", "लाभ", "available", "उपलब्ध", "apply", "आवेदन", "form", "फॉर्म",
		"kaun si", "kaun-si", "konsi", "कौन सी", "कौनसी",
	}
	for _, k := range schemes {
		if strings.Contains(t, k) {
			return intentSchemes
		}
	}
	return ""
}

func (s *Service) eligibilitySummary(ctx context.Context, userID, lang string) string {
	if s.eligibility == nil {
		return ""
	}
	txt, err := s.eligibility.VoiceSummary(ctx, userID, lang)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(txt)
}

func (s *Service) missingSummary(ctx context.Context, userID, lang string) string {
	if s.eligibility == nil {
		return ""
	}
	txt, err := s.eligibility.MissingSummary(ctx, userID, lang)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(txt)
}

// detectResult carries an async language-detection outcome.
type detectResult struct {
	code string
	err  error
}

// supportedReplyLang reports whether code is a language we can actually reply
// and speak in (script coverage + TTS + prompts).
func supportedReplyLang(code string) bool {
	if code == "en" {
		return true
	}
	return scriptOf(code) != ""
}

// resolveReplyLang picks the reply language for a turn. A non-empty lang that
// is already written in the text's script (e.g. Marathi detected by STT for
// Devanagari input) is kept as-is; otherwise the script guess wins (e.g. Hindi
// input with an English UI falls back to Hindi). This matters because
// Devanagari is shared by Hindi, Marathi, Nepali and Sanskrit, and the script
// guess alone defaults all of it to Hindi.
func resolveReplyLang(text, lang string) string {
	script := detectScriptLang(text)
	if script == "" {
		return lang
	}
	if lang != "" && scriptOf(lang) == scriptOf(script) {
		return lang
	}
	return script
}

// scriptOf maps a language code to its primary script, using the same codes
// detectScriptLang returns so the two are directly comparable.
func scriptOf(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "hi", "mr", "ne", "sa", "bho":
		return "hi" // Devanagari
	case "bn", "as":
		return "bn" // Bengali-Assamese
	case "pa":
		return "pa" // Gurmukhi
	case "gu":
		return "gu" // Gujarati
	case "or":
		return "or" // Odia
	case "ta":
		return "ta" // Tamil
	case "te":
		return "te" // Telugu
	case "kn":
		return "kn" // Kannada
	case "ml":
		return "ml" // Malayalam
	case "ur":
		return "ur" // Arabic script
	default:
		return ""
	}
}

// detectScriptLang returns an ISO 639-1 code when the text is written in a
// non-Latin Indian script, or "" for Latin/unknown text.
func detectScriptLang(text string) string {
	for _, r := range text {
		switch {
		case r >= 0x0900 && r <= 0x097F:
			return "hi" // Devanagari (Hindi/Marathi)
		case r >= 0x0980 && r <= 0x09FF:
			return "bn" // Bengali/Assamese
		case r >= 0x0A00 && r <= 0x0A7F:
			return "pa" // Gurmukhi
		case r >= 0x0A80 && r <= 0x0AFF:
			return "gu" // Gujarati
		case r >= 0x0B00 && r <= 0x0B7F:
			return "or" // Odia
		case r >= 0x0B80 && r <= 0x0BFF:
			return "ta" // Tamil
		case r >= 0x0C00 && r <= 0x0C7F:
			return "te" // Telugu
		case r >= 0x0C80 && r <= 0x0CFF:
			return "kn" // Kannada
		case r >= 0x0D00 && r <= 0x0D7F:
			return "ml" // Malayalam
		case r >= 0x0600 && r <= 0x06FF:
			return "ur" // Arabic script (Urdu)
		}
	}
	return ""
}

// llmHistory converts stored conversation messages into LLM history, keeping
// only the most recent turns and skipping empty entries.
func llmHistory(messages []Message) []llm.Message {
	const maxHistory = 6
	start := 0
	if len(messages) > maxHistory {
		start = len(messages) - maxHistory
	}
	out := make([]llm.Message, 0, len(messages)-start)
	for _, m := range messages[start:] {
		content := strings.TrimSpace(m.ContentText)
		if content == "" {
			continue
		}
		out = append(out, llm.Message{Role: m.Role, Content: content})
	}
	return out
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return 0
}

func stringSlice(v any) []string {
	list, ok := v.([]any)
	if !ok {
		if ss, ok := v.([]string); ok {
			return ss
		}
		return nil
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
