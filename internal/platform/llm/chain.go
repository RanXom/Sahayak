package llm

import "context"

// Chain tries multiple providers in order and returns the first successful
// result. It is used to fail over between LLM providers (e.g. Gemini ->
// OpenAI-compatible -> deterministic fallback) so a single provider outage does
// not break the demo.
type Chain struct {
	providers []Provider
}

// NewChain builds a failover chain. The last provider should be a
// non-networked fallback that never fails.
func NewChain(providers ...Provider) *Chain {
	return &Chain{providers: providers}
}

// ExtractFacts tries each provider until one succeeds.
func (c *Chain) ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := p.ExtractFacts(ctx, req)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// GenerateFollowUp tries each provider until one succeeds.
func (c *Chain) GenerateFollowUp(ctx context.Context, req FollowUpRequest) (*FollowUpResponse, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := p.GenerateFollowUp(ctx, req)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// GenerateExplanation tries each provider until one succeeds.
func (c *Chain) GenerateExplanation(ctx context.Context, req ExplanationRequest) (*ExplanationResponse, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := p.GenerateExplanation(ctx, req)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// DetectLanguage tries each provider until one succeeds.
func (c *Chain) DetectLanguage(ctx context.Context, text string) (string, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		lang, err := p.DetectLanguage(ctx, text)
		if err == nil {
			return lang, nil
		}
		lastErr = err
	}
	return "", lastErr
}
