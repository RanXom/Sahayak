package llm

import (
	"context"
	"log/slog"
	"time"
)

// Chain tries multiple providers in order and returns the first successful
// result. It is used to fail over between LLM providers (e.g. Gemini ->
// OpenAI-compatible -> deterministic fallback) so a single provider outage does
// not break the demo.
type Chain struct {
	providers []Provider
	logger    *slog.Logger
}

// NewChain builds a failover chain. The last provider should be a
// non-networked fallback that never fails.
func NewChain(providers ...Provider) *Chain {
	return &Chain{providers: providers, logger: slog.Default()}
}

// NewChainWithLogger builds a failover chain that logs provider attempts.
func NewChainWithLogger(logger *slog.Logger, providers ...Provider) *Chain {
	if logger == nil {
		logger = slog.Default()
	}
	return &Chain{providers: providers, logger: logger}
}

// attempt runs one provider call and logs its outcome. It returns the result
// and error unchanged so callers can fail over.
func attempt[T any](c *Chain, method string, p Provider, call func() (T, error)) (T, error) {
	start := time.Now()
	res, err := call()
	ms := time.Since(start).Milliseconds()
	if err != nil {
		c.logger.Warn("llm_provider_failed", "method", method, "provider", providerName(p), "duration_ms", ms, "error", err.Error())
		return res, err
	}
	c.logger.Debug("llm_provider_ok", "method", method, "provider", providerName(p), "duration_ms", ms)
	return res, nil
}

// ExtractFacts tries each provider until one succeeds.
func (c *Chain) ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error) {
	var lastErr error
	for _, p := range c.providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		res, err := attempt(c, "extract_facts", p, func() (*ExtractFactsResponse, error) {
			return p.ExtractFacts(ctx, req)
		})
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
		res, err := attempt(c, "generate_follow_up", p, func() (*FollowUpResponse, error) {
			return p.GenerateFollowUp(ctx, req)
		})
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
		res, err := attempt(c, "generate_explanation", p, func() (*ExplanationResponse, error) {
			return p.GenerateExplanation(ctx, req)
		})
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
		lang, err := attempt(c, "detect_language", p, func() (string, error) {
			return p.DetectLanguage(ctx, text)
		})
		if err == nil {
			return lang, nil
		}
		lastErr = err
	}
	return "", lastErr
}
