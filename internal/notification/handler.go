package notification

import (
	"net/http"
	"strconv"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes notification HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a notification Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func userID(r *http.Request) (string, bool) {
	c, ok := domain.UserFromContext(r.Context())
	if !ok {
		return "", false
	}
	return c.UserID, true
}

// GetPreferences handles GET /notifications/preferences.
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	prefs, err := h.svc.GetPreferences(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"preferences": prefs})
}

type preferencesBody struct {
	Channel string `json:"channel"`
	Enabled bool   `json:"enabled"`
}

// UpdatePreferences handles PUT /notifications/preferences.
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body preferencesBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	prefs, err := h.svc.UpdatePreferences(r.Context(), id, body.Channel, body.Enabled)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"preferences": prefs})
}

// List handles GET /notifications.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	items, err := h.svc.List(r.Context(), id, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"notifications": items})
}
