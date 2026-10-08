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

// OpenAIProvider implements Provider against OpenAI-compatible chat APIs.
type OpenAIProvider struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

// NewOpenAIProvider constructs an OpenAI-compatible provider.
func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *OpenAIProvider) generate(ctx context.Context, system, user string, jsonMode bool) (string, error) {
	msgs := []chatMessage{}
	if system != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: system})
	}
	msgs = append(msgs, chatMessage{Role: "user", Content: user})

	req := chatRequest{Model: p.Model, Messages: msgs, Temperature: 0.1}
	if jsonMode {
		req.ResponseFormat = map[string]any{"type": "json_object"}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, string(respBody))
	}
	var cr chatResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if cr.Error != nil {
		return "", fmt.Errorf("openai error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	return cr.Choices[0].Message.Content, nil
}

// ExtractFacts extracts structured facts from user text.
func (p *OpenAIProvider) ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error) {
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
func (p *OpenAIProvider) GenerateFollowUp(ctx context.Context, req FollowUpRequest) (*FollowUpResponse, error) {
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
func (p *OpenAIProvider) GenerateExplanation(ctx context.Context, req ExplanationRequest) (*ExplanationResponse, error) {
	system := "Explain welfare scheme eligibility using ONLY the provided rule outcomes. Never invent criteria. Write in " + languageName(req.Language) + "."
	user := fmt.Sprintf("Scheme: %s\nStatus: %s\nPassed: %s\nFailed: %s\nUnconfirmed: %s",
		req.SchemeName, req.Status, strings.Join(req.PassedRules, "; "), strings.Join(req.FailedRules, "; "), strings.Join(req.MissingRules, "; "))
	out, err := p.generate(ctx, system, user, false)
	if err != nil || strings.TrimSpace(out) == "" {
		return &ExplanationResponse{Text: FallbackExplanation(req)}, nil
	}
	return &ExplanationResponse{Text: strings.TrimSpace(out)}, nil
}

// DetectLanguage returns an ISO 639-1 code for the given text.
func (p *OpenAIProvider) DetectLanguage(ctx context.Context, text string) (string, error) {
	out, err := p.generate(ctx, "Detect the dominant language. Reply with only the ISO 639-1 code.", text, false)
	if err != nil {
		return "", err
	}
	code := strings.TrimSpace(strings.ToLower(out))
	if i := strings.IndexAny(code, " \n"); i > 0 {
		code = code[:i]
	}
	if len(code) > 3 || code == "" {
		return "hi", nil
	}
	return code, nil
}
