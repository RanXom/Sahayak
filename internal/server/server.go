// Package server wires the HTTP router, middleware and module handlers.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/application"
	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/auth"
	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/conversation"
	"github.com/sahayak/sahayak/internal/crypto"
	"github.com/sahayak/sahayak/internal/document"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/eligibility"
	"github.com/sahayak/sahayak/internal/helper"
	"github.com/sahayak/sahayak/internal/jobs"
	"github.com/sahayak/sahayak/internal/location"
	"github.com/sahayak/sahayak/internal/maintenance"
	"github.com/sahayak/sahayak/internal/middleware"
	"github.com/sahayak/sahayak/internal/notification"
	"github.com/sahayak/sahayak/internal/profile"
	"github.com/sahayak/sahayak/internal/scheme"
	"github.com/sahayak/sahayak/internal/source"
)

// Version is the application version reported by /healthz.
const Version = "0.1.0"

// Server owns the HTTP router and shared dependencies.
type Server struct {
	cfg     *config.Config
	logger  *slog.Logger
	pool    *pgxpool.Pool
	version string

	startedAt    time.Time
	riverStarted atomic.Bool

	tokens      *auth.TokenManager
	authService *auth.Service
	authHandler *auth.Handler

	locationHandler     *location.Handler
	profileHandler      *profile.Handler
	schemeHandler       *scheme.Handler
	sourceHandler       *source.Handler
	documentHandler     *document.Handler
	eligibilityHandler  *eligibility.Handler
	auditHandler        *audit.Handler
	conversationHandler *conversation.Handler
	applicationHandler  *application.Handler
	notificationHandler *notification.Handler
	notificationSvc     *notification.Service
	helperHandler       *helper.Handler
	helperSvc           *helper.Service
	maintenance         *maintenance.Service
	jobsClient          *jobs.Client
	audioStore          *conversation.FileAudioStore

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
	s.initJobs()
	s.buildRouter()
	return s
}

// Handler returns the configured HTTP handler.
func (s *Server) Handler() http.Handler { return s.router }

// Tokens exposes the token manager (used by the worker/other components).
func (s *Server) Tokens() *auth.TokenManager { return s.tokens }

func (s *Server) initRateLimiters() {
	rl := s.cfg.RateLimit
	anonMin, anonHour := rl.AnonPerMinute, rl.AnonPerHour
	userMin, userHour := rl.CitizenPerMinute, rl.CitizenPerHour
	if s.cfg.DemoOTP != "" {
		// A shared venue network can look like one client. Relax limits in demo mode.
		anonMin, anonHour = anonMin*20, anonHour*20
		userMin, userHour = userMin*20, userHour*20
		s.logger.Warn("demo_mode_enabled", "detail", "fixed OTP login and relaxed rate limits are active")
	}
	s.anonLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: anonMin, PerHour: anonHour})
	s.userLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: userMin, PerHour: userHour})
	s.voiceLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{PerMinute: rl.VoicePerMinute, PerHour: rl.VoicePerHour})
}

func (s *Server) initModules() {
	auditRepo := audit.NewRepository(s.pool)
	auditLogger := audit.NewDBLogger(auditRepo, s.logger)
	s.auditHandler = audit.NewHandler(auditRepo)

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
		DemoOTP:     s.cfg.DemoOTP,
		OTPPer10Min: s.cfg.RateLimit.OTPPerPhone10Min,
		OTPPerDay:   s.cfg.RateLimit.OTPPerPhoneDay,
	})
	s.authHandler = auth.NewHandler(s.authService)

	locationRepo := location.NewRepository(s.pool)
	s.locationHandler = location.NewHandler(location.NewService(locationRepo))

	profileRepo := profile.NewRepository(s.pool)
	profileSvc := profile.NewService(profileRepo, locationRepo)
	s.profileHandler = profile.NewHandler(profileSvc)

	schemeRepo := scheme.NewRepository(s.pool)
	schemeSvc := scheme.NewService(schemeRepo)
	s.schemeHandler = scheme.NewHandler(schemeSvc)

	sourceRepo := source.NewRepository(s.pool)
	s.sourceHandler = source.NewHandler(source.NewService(sourceRepo))

	documentRepo := document.NewRepository(s.pool)
	documentSvc := document.NewService(documentRepo, schemeRepo, profileRepo)
	s.documentHandler = document.NewHandler(documentSvc)

	eligibilitySvc := eligibility.NewService(profileRepo, schemeSvc, locationRepo, documentSvc)
	s.eligibilityHandler = eligibility.NewHandler(eligibilitySvc)

	audioStore, err := conversation.NewFileAudioStore(s.cfg.DataDir + "/audio")
	if err != nil {
		s.logger.Error("audio_store_init_failed", "error", err, "fallback", os.TempDir())
		audioStore, err = conversation.NewFileAudioStore(os.TempDir() + "/sahayak-audio")
		if err != nil {
			s.logger.Error("audio_store_fallback_failed", "error", err)
		}
	}
	s.audioStore = audioStore
	stt, tts := s.voiceProviders()
	convSvc := conversation.NewService(conversation.Deps{
		Repo:        conversation.NewRepository(s.pool),
		Profiles:    profileSvc,
		LLM:         s.llmProvider(),
		STT:         stt,
		TTS:         tts,
		Audio:       audioStore,
		Location:    locationRepo,
		Audit:       auditLogger,
		Logger:      s.logger,
		Eligibility: eligibilitySvc,
	})
	s.conversationHandler = conversation.NewHandler(convSvc, audioStore, s.cfg.MaxAudioMB)

	pdfGen, err := application.NewPDFGenerator(s.cfg.PDFDir, s.cfg.FontDir)
	if err != nil {
		s.logger.Error("pdf_generator_init_failed", "error", err)
		pdfGen, _ = application.NewPDFGenerator(os.TempDir()+"/sahayak-pdfs", s.cfg.FontDir)
	}
	appRepo := application.NewRepository(s.pool)
	appSvc := application.NewService(appRepo, schemeRepo, profileRepo, pdfGen)
	s.applicationHandler = application.NewHandler(appSvc)

	notifRepo := notification.NewRepository(s.pool)
	s.notificationSvc = notification.NewService(notification.Deps{
		Repo:     notifRepo,
		Phones:   authRepo,
		Cipher:   cipher,
		WhatsApp: s.whatsappProvider(),
		SMS:      s.smsProvider(),
		Audit:    auditLogger,
	})
	s.notificationHandler = notification.NewHandler(s.notificationSvc)

	appSvc.SetStatusNotifier(func(ctx context.Context, userID, appID, status string) {
		_, _ = s.notificationSvc.Notify(ctx, notification.NotifyRequest{
			UserID:        userID,
			ApplicationID: &appID,
			Channel:       notification.ChannelWhatsApp,
			TemplateKey:   notification.TemplateStatusUpdate,
			Language:      "hi",
			Params:        map[string]string{"status": status},
		})
	})

	helperRepo := helper.NewRepository(s.pool)
	s.helperSvc = helper.NewService(helperRepo, auditLogger, s.cfg.RateLimit.HelperSessionsPerDay)
	s.helperHandler = helper.NewHandler(s.helperSvc)

	s.maintenance = maintenance.New(s.pool, schemeRepo, s.logger, s.cfg.DataDir+"/audio", s.cfg.PDFDir)
}

// initJobs wires the River client when running in worker or all mode.
func (s *Server) initJobs() {
	if s.cfg.ServerMode == "api" {
		return
	}
	handlers := jobs.Handlers{
		Notification: func(ctx context.Context, args jobs.NotificationArgs) error {
			return s.notificationSvc.Dispatch(ctx, args.NotificationID)
		},
		StatusCheck: func(context.Context, jobs.StatusCheckArgs) error { return nil },
		LinkCheck: func(ctx context.Context, _ jobs.LinkCheckArgs) error {
			return s.maintenance.LinkCheck(ctx)
		},
		DataRetention: func(ctx context.Context, _ jobs.DataRetentionArgs) error {
			return s.maintenance.DataRetention(ctx)
		},
	}
	client, err := jobs.NewClient(s.pool, handlers)
	if err != nil {
		s.logger.Error("jobs_client_init_failed", "error", err)
		return
	}
	s.jobsClient = client
	s.notificationSvc.SetEnqueuer(client)
}

// StartWorkers starts the background job workers.
func (s *Server) StartWorkers(ctx context.Context) error {
	if s.jobsClient == nil {
		return nil
	}
	if err := s.jobsClient.Start(ctx); err != nil {
		return err
	}
	s.riverStarted.Store(true)
	return nil
}

// StopWorkers gracefully stops the background job workers.
func (s *Server) StopWorkers(ctx context.Context) error {
	if s.jobsClient == nil {
		return nil
	}
	return s.jobsClient.Stop(ctx)
}

// NotificationService exposes the notification service (used by other modules).
func (s *Server) NotificationService() *notification.Service { return s.notificationSvc }

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
			r.Get("/config", s.handleConfig)
			r.Get("/locations/states", s.locationHandler.ListStates)
			r.Get("/locations/states/{id}/districts", s.locationHandler.ListDistricts)
			r.Get("/documents/types", s.documentHandler.ListTypes)
			r.Get("/audio/temp/{id}", s.conversationHandler.TempAudio)
		})

		// Authenticated routes (citizen, helper, admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(s.tokens))
			r.Use(middleware.Limit(s.userLimiter, middleware.UserKey))

			r.Get("/auth/me", s.authHandler.Me)
			r.Post("/auth/logout", s.authHandler.Logout)

			r.Get("/profile", s.profileHandler.Get)
			r.Put("/profile", s.profileHandler.Update)
			r.Patch("/profile/facts", s.profileHandler.MergeFacts)
			r.Get("/profile/documents", s.profileHandler.GetDocuments)
			r.Put("/profile/documents", s.profileHandler.UpdateDocuments)

			r.Get("/schemes/{id}", s.schemeHandler.Get)
			r.Get("/schemes/{id}/documents", s.documentHandler.SchemeDocuments)

			r.Post("/eligibility/evaluate", s.eligibilityHandler.Evaluate)
			r.Get("/eligibility/results", s.eligibilityHandler.Results)

			r.Post("/conversations", s.conversationHandler.Start)
			r.Get("/conversations/{id}", s.conversationHandler.Get)
			r.Post("/conversations/{id}/message", s.conversationHandler.Message)
			r.Post("/conversations/{id}/voice", s.conversationHandler.Voice)

			r.Post("/applications", s.applicationHandler.Create)
			r.Get("/applications", s.applicationHandler.List)
			r.Get("/applications/{id}", s.applicationHandler.Get)
			r.Put("/applications/{id}/form", s.applicationHandler.UpdateForm)
			r.Post("/applications/{id}/pdf", s.applicationHandler.GeneratePDF)
			r.Get("/applications/{id}/pdf", s.applicationHandler.DownloadPDF)
			r.Post("/applications/{id}/submit", s.applicationHandler.Submit)
			r.Put("/applications/{id}/status", s.applicationHandler.UpdateStatus)

			r.Get("/notifications", s.notificationHandler.List)
			r.Get("/notifications/preferences", s.notificationHandler.GetPreferences)
			r.Put("/notifications/preferences", s.notificationHandler.UpdatePreferences)

			r.Post("/helper/register", s.helperHandler.Register)
		})

		// Helper routes (verified helpers and admins)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(s.tokens))
			r.Use(middleware.RequireRole(domain.RoleHelper, domain.RoleAdmin))
			r.Use(middleware.RequireVerifiedHelper(s.helperSvc))
			r.Use(middleware.Limit(s.userLimiter, middleware.UserKey))

			r.Get("/helper/dashboard", s.helperHandler.Dashboard)
			r.Post("/helper/sessions", s.helperHandler.StartSession)
			r.Get("/helper/sessions", s.helperHandler.ListSessions)
			r.Get("/helper/sessions/{id}", s.helperHandler.GetSession)
			r.Post("/helper/sessions/{id}/consent", s.helperHandler.RecordConsent)
			r.Post("/helper/sessions/{id}/end", s.helperHandler.EndSession)
		})

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(s.tokens))
			r.Use(middleware.RequireRole(domain.RoleAdmin))
			r.Use(middleware.Limit(s.userLimiter, middleware.UserKey))

			r.Get("/admin/schemes", s.schemeHandler.AdminList)
			r.Post("/admin/schemes", s.schemeHandler.AdminCreate)
			r.Put("/admin/schemes/{id}", s.schemeHandler.AdminUpdate)
			r.Post("/admin/schemes/{id}/publish", s.schemeHandler.AdminPublish)
			r.Post("/admin/schemes/{id}/archive", s.schemeHandler.AdminArchive)

			r.Get("/admin/sources", s.sourceHandler.ListSources)
			r.Post("/admin/sources", s.sourceHandler.CreateSource)
			r.Get("/admin/candidates", s.sourceHandler.ListCandidates)
			r.Post("/admin/candidates/{id}/accept", s.sourceHandler.AcceptCandidate)
			r.Post("/admin/candidates/{id}/reject", s.sourceHandler.RejectCandidate)

			r.Get("/admin/audit", s.auditHandler.Query)

			r.Get("/admin/helpers", s.helperHandler.AdminList)
			r.Post("/admin/helpers/{userId}/verify", s.helperHandler.AdminVerify)
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
