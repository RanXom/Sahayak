package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GeminiProvider implements Provider against the Google Gemini API.
type GeminiProvider struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

// NewGeminiProvider constructs a Gemini provider.
func NewGeminiProvider(apiKey, model, baseURL string) *GeminiProvider {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &GeminiProvider{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  map[string]any  `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *GeminiProvider) generate(ctx context.Context, system, user string, jsonMode bool) (string, error) {
	req := geminiRequest{
		Contents: []geminiContent{{Role: "user", Parts: []geminiPart{{Text: user}}}},
	}
	if system != "" {
		req.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: system}}}
	}
	if jsonMode {
		req.GenerationConfig = map[string]any{
			"responseMimeType": "application/json",
			"temperature":      0.1,
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, p.Model, p.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini status %d: %s", resp.StatusCode, string(respBody))
	}

	var gr geminiResponse
	if err := json.Unmarshal(respBody, &gr); err != nil {
		return "", fmt.Errorf("decode gemini response: %w", err)
	}
	if gr.Error != nil {
		return "", fmt.Errorf("gemini error: %s", gr.Error.Message)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}
	return gr.Candidates[0].Content.Parts[0].Text, nil
}

// ExtractFacts extracts structured facts from user text.
func (p *GeminiProvider) ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error) {
	out, err := p.generate(ctx, ExtractionSystemPrompt(), BuildExtractionUserPrompt(req), true)
	if err != nil {
		return nil, err
	}
	var res ExtractFactsResponse
	if err := json.Unmarshal([]byte(stripCodeFence(out)), &res); err != nil {
		return nil, fmt.Errorf("parse extraction JSON: %w", err)
	}
	if res.Facts == nil {
		res.Facts = map[string]any{}
	}
	return &res, nil
}

// GenerateFollowUp asks the LLM to phrase the next question.
func (p *GeminiProvider) GenerateFollowUp(ctx context.Context, req FollowUpRequest) (*FollowUpResponse, error) {
	field := ""
	if len(req.MissingFields) > 0 {
		field = req.MissingFields[0]
	}
	out, err := p.generate(ctx, FollowUpSystemPrompt(req.Language), BuildFollowUpUserPrompt(req), true)
	if err != nil {
		return &FollowUpResponse{Question: QuestionForField(field, req.Language), QuestionAbout: field}, nil
	}
	var res FollowUpResponse
	if err := json.Unmarshal([]byte(stripCodeFence(out)), &res); err != nil || res.Question == "" {
		return &FollowUpResponse{Question: QuestionForField(field, req.Language), QuestionAbout: field}, nil
	}
	if res.QuestionAbout == "" {
		res.QuestionAbout = field
	}
	return &res, nil
}

// GenerateExplanation produces a grounded explanation from rule outcomes.
func (p *GeminiProvider) GenerateExplanation(ctx context.Context, req ExplanationRequest) (*ExplanationResponse, error) {
	langName := languageName(req.Language)
	system := "You explain welfare scheme eligibility results. Use ONLY the rule outcomes provided. Never invent criteria, schemes, documents or URLs. Be concise and respectful. Write in " + langName + "."
	user := fmt.Sprintf("Scheme: %s\nStatus: %s\nPassed criteria: %s\nFailed criteria: %s\nUnconfirmed criteria: %s\nWrite a 1-2 sentence explanation for the citizen.",
		req.SchemeName, req.Status, strings.Join(req.PassedRules, "; "), strings.Join(req.FailedRules, "; "), strings.Join(req.MissingRules, "; "))

	out, err := p.generate(ctx, system, user, false)
	if err != nil {
		return &ExplanationResponse{Text: FallbackExplanation(req)}, nil
	}
	text := strings.TrimSpace(out)
	if text == "" {
		text = FallbackExplanation(req)
	}
	return &ExplanationResponse{Text: text}, nil
}

// DetectLanguage returns an ISO 639-1 code for the given text.
func (p *GeminiProvider) DetectLanguage(ctx context.Context, text string) (string, error) {
	system := "Detect the dominant language of the user's message. Respond with only the ISO 639-1 code (e.g. hi, en, ta, te, bn, mr)."
	out, err := p.generate(ctx, system, text, false)
	if err != nil {
		return "", err
	}
	code := strings.TrimSpace(strings.ToLower(out))
	code = strings.Trim(code, ".\"' \n")
	if i := strings.IndexAny(code, " \n"); i > 0 {
		code = code[:i]
	}
	if len(code) > 3 {
		return "hi", nil
	}
	return code, nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func languageName(code string) string {
	switch code {
	case "hi":
		return "Hindi"
	case "en":
		return "English"
	case "ta":
		return "Tamil"
	case "te":
		return "Telugu"
	case "kn":
		return "Kannada"
	case "bn":
		return "Bengali"
	case "mr":
		return "Marathi"
	case "gu":
		return "Gujarati"
	case "pa":
		return "Punjabi"
	default:
		return "Hindi"
	}
}
