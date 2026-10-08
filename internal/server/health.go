package server

import (
	"context"
	"net/http"
	"time"

	"github.com/sahayak/sahayak/internal/httpx"
)

// healthResponse is the payload for GET /healthz.
type healthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Version string            `json:"version"`
	Uptime  int64             `json:"uptime_seconds"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	status := "ok"
	if err := s.pool.Ping(ctx); err != nil {
		checks["database"] = "error"
		status = "degraded"
	} else {
		checks["database"] = "ok"
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}
	httpx.WriteJSON(w, code, healthResponse{
		Status:  status,
		Checks:  checks,
		Version: s.version,
		Uptime:  int64(time.Since(s.startedAt).Seconds()),
	})
}
