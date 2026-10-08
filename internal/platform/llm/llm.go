// Package llm defines the LLM provider abstraction used strictly for fact
// extraction, follow-up question generation and grounded explanations.
//
// The LLM is never permitted to decide eligibility.
package llm

import (
	"context"
	"fmt"
)

// Message is a single conversation message supplied as context.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExtractFactsRequest asks the LLM to extract structured facts.
type ExtractFactsRequest struct {
	UserText            string
	CurrentProfile      map[string]any
	Language            string
	ConversationHistory []Message
}

// ExtractFactsResponse is the structured result of fact extraction.
type ExtractFactsResponse struct {
	Facts         map[string]any `json:"facts"`
	Confidence    float64        `json:"confidence"`
	Ambiguities   []string       `json:"ambiguities"`
	NeedsFollowUp []string       `json:"needs_follow_up"`
}

// FollowUpRequest asks the LLM to phrase the next intake question. It also
// carries the user's latest message, known profile facts and recent history so
// the model can answer questions about Sahayak before asking the next question.
type FollowUpRequest struct {
	UserText      string
	ProfileFacts  map[string]any
	History       []Message
	MissingFields []string
	Language      string
	Context       map[string]any
}

// FollowUpResponse is the generated follow-up question.
type FollowUpResponse struct {
	Question      string `json:"question"`
	QuestionAbout string `json:"question_about"`
	IsLast        bool   `json:"is_last"`
}

// ExplanationRequest asks the LLM to phrase a grounded explanation.
type ExplanationRequest struct {
	SchemeName   string
	Status       string
	PassedRules  []string
	FailedRules  []string
	MissingRules []string
	Language     string
}

// ExplanationResponse is the generated explanation.
type ExplanationResponse struct {
	Text string `json:"text"`
}

// Provider is the LLM abstraction. Implementations must be safe for concurrent use.
type Provider interface {
	ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error)
	GenerateFollowUp(ctx context.Context, req FollowUpRequest) (*FollowUpResponse, error)
	GenerateExplanation(ctx context.Context, req ExplanationRequest) (*ExplanationResponse, error)
	DetectLanguage(ctx context.Context, text string) (string, error)
}

// Named wraps a Provider with a human-readable name so the failover chain can
// log which provider answered or failed.
type Named struct {
	Name string
	Provider
}

// ProviderName reports the wrapped provider's name.
func (n Named) ProviderName() string { return n.Name }

// providerName resolves a provider's log name, falling back to its Go type.
func providerName(p Provider) string {
	if n, ok := p.(interface{ ProviderName() string }); ok {
		return n.ProviderName()
	}
	return fmt.Sprintf("%T", p)
}
