package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// RateLimitConfig configures a single rate-limit tier.
type RateLimitConfig struct {
	PerMinute int
	PerHour   int
}

type bucket struct {
	minute   *rate.Limiter
	hour     *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a concurrency-safe in-memory limiter keyed by an arbitrary string.
// It is intentionally per-instance; horizontal scaling would require a shared store.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	cfg     RateLimitConfig
	cleanup time.Duration
	lastGC  time.Time
}

// NewRateLimiter creates a limiter for the given tier.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.PerMinute <= 0 {
		cfg.PerMinute = 60
	}
	if cfg.PerHour <= 0 {
		cfg.PerHour = 600
	}
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		cfg:     cfg,
		cleanup: 10 * time.Minute,
		lastGC:  time.Now(),
	}
}

func (l *RateLimiter) get(key string) *bucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.Sub(l.lastGC) > l.cleanup {
		for k, b := range l.buckets {
			if now.Sub(b.lastSeen) > l.cleanup {
				delete(l.buckets, k)
			}
		}
		l.lastGC = now
	}

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{
			minute: rate.NewLimiter(rate.Limit(float64(l.cfg.PerMinute)/60.0), l.cfg.PerMinute),
			hour:   rate.NewLimiter(rate.Limit(float64(l.cfg.PerHour)/3600.0), l.cfg.PerHour),
		}
		l.buckets[key] = b
	}
	b.lastSeen = now
	return b
}

// Allow reports whether a request for key may proceed.
func (l *RateLimiter) Allow(key string) bool {
	b := l.get(key)
	if !(b.minute.Allow() && b.hour.Allow()) {
		return false
	}
	return true
}

// KeyFunc derives the limiting key for a request.
type KeyFunc func(r *http.Request) string

// IPKey keys by remote IP.
func IPKey(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// UserKey keys by authenticated user ID, falling back to IP.
func UserKey(r *http.Request) string {
	if c, ok := domain.UserFromContext(r.Context()); ok {
		return "u:" + c.UserID
	}
	return IPKey(r)
}

// Limit enforces the limiter using the supplied key function.
func Limit(l *RateLimiter, keyFn KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(keyFn(r)) {
				httpx.WriteError(w, domain.RateLimited("Too many requests. Please slow down and try again shortly."))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
