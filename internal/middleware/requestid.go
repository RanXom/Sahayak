package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/sahayak/sahayak/internal/domain"
)

// RequestID ensures every request has an X-Request-Id and stores it in context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := domain.WithRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestIDFrom returns the request ID from a context.
func requestIDFrom(ctx context.Context) string {
	return domain.RequestIDFromContext(ctx)
}
