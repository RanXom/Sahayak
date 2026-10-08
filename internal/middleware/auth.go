package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// TokenVerifier validates an access token and returns its claims.
type TokenVerifier interface {
	VerifyAccessToken(token string) (*domain.Claims, error)
}

// Authenticate validates the bearer token and stores claims in the context.
func Authenticate(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				httpx.WriteError(w, domain.Unauthorized("Missing or malformed Authorization header"))
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			claims, err := verifier.VerifyAccessToken(token)
			if err != nil {
				httpx.WriteError(w, domain.Unauthorized("Invalid or expired token"))
				return
			}
			ctx := domain.WithUser(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// VerifiedHelperChecker reports whether a user is a verified helper.
type VerifiedHelperChecker interface {
	IsVerified(ctx context.Context, userID string) bool
}

// RequireVerifiedHelper restricts a route to verified helpers.
func RequireVerifiedHelper(checker VerifiedHelperChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := domain.UserFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, domain.Unauthorized("Authentication required"))
				return
			}
			if claims.Role == domain.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}
			if !checker.IsVerified(r.Context(), claims.UserID) {
				httpx.WriteError(w, domain.Forbidden("Verified helper account required"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole restricts a route to the given roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := domain.UserFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, domain.Unauthorized("Authentication required"))
				return
			}
			if !allowed[claims.Role] {
				httpx.WriteError(w, domain.Forbidden("You do not have permission to access this resource"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
