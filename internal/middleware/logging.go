package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sahayak/sahayak/internal/domain"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Logger logs one structured line per request. It never logs query values.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			userID := ""
			if c, ok := domain.UserFromContext(r.Context()); ok {
				userID = c.UserID
			}
			level := slog.LevelInfo
			if rec.status >= 500 {
				level = slog.LevelError
			} else if rec.status >= 400 {
				level = slog.LevelWarn
			}

			logger.Log(r.Context(), level, "http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestIDFrom(r.Context()),
				"user_id", userID,
			)
		})
	}
}

// Language resolves the request language from Accept-Language or the lang query param.
func Language(defaultLang string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := r.URL.Query().Get("lang")
			if lang == "" {
				lang = r.Header.Get("Accept-Language")
				if i := strings.IndexAny(lang, ",;"); i >= 0 {
					lang = lang[:i]
				}
				lang = strings.TrimSpace(lang)
			}
			lang = normalizeLang(lang, defaultLang)
			ctx := domain.WithLanguage(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func normalizeLang(lang, fallback string) string {
	switch lang {
	case "en", "hi", "ta", "te", "kn", "bn", "mr", "gu", "pa", "or", "as", "ur", "bho", "raj":
		return lang
	default:
		return fallback
	}
}
