// Package server wires the HTTP router, middleware and module handlers.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/auth"
	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/crypto"
	"github.com/sahayak/sahayak/internal/location"
	"github.com/sahayak/sahayak/internal/middleware"
)

// Version is the application version reported by /healthz.
const Version = "0.1.0"

// Server owns the HTTP router and shared dependencies.
type Server struct {
	cfg     *config.Config
	logger  *slog.Logger
	pool    *pgxpool.Pool
	version string

	startedAt time.Time

	tokens      *auth.TokenManager
	authService *auth.Service
	authHandler *auth.Handler

	locationHandler *location.Handler

	// Rate limiters
	anonLimiter  *middleware.RateLimiter
	userLimiter  *middleware.RateLimiter
	voiceLimiter *middleware.RateLimiter

	router chi.Router
}

// New constructs the server and builds its routes.
func New(cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) *Server {
	s := &Server{
		cfg:       cfg,
		logger:    logger,
		pool:      pool,
		version:   Version,
		startedAt: time.Now(),
	}
	s.tokens = auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	s.initRateLimiters()
	s.initModules()
	s.buildRouter()
	return s
}

// Handler returns the configured HTTP handler.
func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) initRateLimiters() {
	rl := s.cfg.RateLimit
	s.anonLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: rl.AnonPerMinute, PerHour: rl.AnonPerHour})
	s.userLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: rl.CitizenPerMinute, PerHour: rl.CitizenPerHour})
	s.voiceLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: rl.VoicePerMinute, PerHour: rl.VoicePerHour})
}

func (s *Server) initModules() {
	auditRepo := audit.NewRepository(s.pool)
	auditLogger := audit.NewDBLogger(auditRepo, s.logger)

	cipher, err := crypto.New(s.cfg.PhoneHashPepper)
	if err != nil {
		s.logger.Error("crypto_init_failed", "error", err)
	}

	authRepo := auth.NewRepository(s.pool)
	s.authService = auth.NewService(auth.Deps{
		Repo:        authRepo,
		Tokens:      s.tokens,
		SMS:         s.smsProvider(),
		Audit:       auditLogger,
		Cipher:      cipher,
		Pepper:      s.cfg.PhoneHashPepper,
		OTPPer10Min: s.cfg.RateLimit.OTPPerPhone10Min,
		OTPPerDay:   s.cfg.RateLimit.OTPPerPhoneDay,
	})
	s.authHandler = auth.NewHandler(s.authService)

	locationRepo := location.NewRepository(s.pool)
	s.locationHandler = location.NewHandler(location.NewService(locationRepo))
}

func (s *Server) buildRouter() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(s.logger))
	r.Use(middleware.Logger(s.logger))
	r.Use(middleware.CORS(s.corsOrigins()))
	r.Use(middleware.Language("hi"))

	r.Get("/healthz", s.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Limit(s.anonLimiter, middleware.IPKey))
			r.Post("/auth/otp/request", s.authHandler.RequestOTP)
			r.Post("/auth/otp/verify", s.authHandler.VerifyOTP)
			r.Post("/auth/refresh", s.authHandler.Refresh)
			r.Get("/locations/states", s.locationHandler.ListStates)
			r.Get("/locations/states/{id}/districts", s.locationHandler.ListDistricts)
		})

		// Authenticated routes (citizen, helper, admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(s.tokens))
			r.Use(middleware.Limit(s.userLimiter, middleware.UserKey))

			r.Get("/auth/me", s.authHandler.Me)
			r.Post("/auth/logout", s.authHandler.Logout)
		})
	})

	s.router = r
}

func (s *Server) corsOrigins() []string {
	if len(s.cfg.AllowedOrigins) > 0 {
		return s.cfg.AllowedOrigins
	}
	if s.cfg.IsProduction() {
		// Same-origin deployments need no CORS. Set CORS_ALLOWED_ORIGINS for
		// cross-origin (e.g. separate api. and app. subdomains).
		return []string{}
	}
	return []string{"*"}
}
