package domain

import "context"

type contextKey string

const (
	userContextKey      contextKey = "sahayak.user"
	requestIDContextKey contextKey = "sahayak.request_id"
	langContextKey      contextKey = "sahayak.lang"
)

// Claims are the authenticated principal's JWT claims.
type Claims struct {
	UserID string
	Role   string
	Lang   string
}

// WithUser attaches claims to the context.
func WithUser(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, userContextKey, c)
}

// UserFromContext returns the authenticated claims, if any.
func UserFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(userContextKey).(*Claims)
	return c, ok
}

// WithRequestID attaches a request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, id)
}

// RequestIDFromContext returns the request ID, if any.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

// WithLanguage attaches the resolved request language.
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, langContextKey, lang)
}

// LanguageFromContext returns the resolved request language.
func LanguageFromContext(ctx context.Context) string {
	lang, _ := ctx.Value(langContextKey).(string)
	return lang
}
