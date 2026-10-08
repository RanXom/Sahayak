// Package server wires the HTTP router, middleware and module handlers.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/config"
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
	s.buildRouter()
	return s
}

// Handler returns the configured HTTP handler.
func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) buildRouter() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(s.logger))
	r.Use(middleware.Logger(s.logger))
	r.Use(middleware.Language("hi"))

	r.Get("/healthz", s.handleHealth)

	s.router = r
}
