// Package translate defines the machine translation provider abstraction.
package translate

import "context"

// Provider translates text between languages.
type Provider interface {
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}

// Fallback returns the source text unchanged. Used when no provider is configured.
type Fallback struct{}

// NewFallback constructs a fallback translator.
func NewFallback() *Fallback { return &Fallback{} }

// Translate returns the source text unchanged.
func (f *Fallback) Translate(_ context.Context, text, _, _ string) (string, error) {
	return text, nil
}
