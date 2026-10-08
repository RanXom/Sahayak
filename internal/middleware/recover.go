package middleware

import (
	"log/slog"
	"net/http"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Recover catches panics and returns a 500 without leaking internals.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic_recovered",
						"panic", rec,
						"path", r.URL.Path,
						"request_id", requestIDFrom(r.Context()),
					)
					httpx.WriteError(w, domain.Internal("An unexpected error occurred", nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
