package llm

import "context"

// FallbackProvider is used when no LLM credentials are configured. It performs
// no extraction and falls back to deterministic questions and explanations so
// the application degrades to form-based input rather than failing.
type FallbackProvider struct{}

// NewFallbackProvider constructs a fallback provider.
func NewFallbackProvider() *FallbackProvider { return &FallbackProvider{} }

// ExtractFacts returns no extracted facts.
func (p *FallbackProvider) ExtractFacts(context.Context, ExtractFactsRequest) (*ExtractFactsResponse, error) {
	return &ExtractFactsResponse{Facts: map[string]any{}}, nil
}

// GenerateFollowUp returns a deterministic question.
func (p *FallbackProvider) GenerateFollowUp(_ context.Context, req FollowUpRequest) (*FollowUpResponse, error) {
	field := ""
	if len(req.MissingFields) > 0 {
		field = req.MissingFields[0]
	}
	return &FollowUpResponse{Question: QuestionForField(field, req.Language), QuestionAbout: field}, nil
}

// GenerateExplanation returns a deterministic explanation.
func (p *FallbackProvider) GenerateExplanation(_ context.Context, req ExplanationRequest) (*ExplanationResponse, error) {
	return &ExplanationResponse{Text: FallbackExplanation(req)}, nil
}

// DetectLanguage defaults to Hindi.
func (p *FallbackProvider) DetectLanguage(context.Context, string) (string, error) {
	return "hi", nil
}
