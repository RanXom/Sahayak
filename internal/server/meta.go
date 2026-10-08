package server

import (
	"net/http"

	"github.com/sahayak/sahayak/internal/httpx"
)

// handleConfig exposes non-sensitive runtime capabilities so the frontend can
// adapt (e.g. hide the mic when no voice provider is configured).
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"app_name":      "Sahayak",
		"version":       s.version,
		"voice_enabled": s.voiceLive(),
		"llm_enabled":   s.llmLive(),
		"languages":     []string{"hi", "en", "ta"},
	})
}
