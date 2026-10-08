package helper

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes helper mode HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a helper Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func userID(r *http.Request) (string, bool) {
	c, ok := domain.UserFromContext(r.Context())
	if !ok {
		return "", false
	}
	return c.UserID, true
}

type registerBody struct {
	HelperType   string `json:"helper_type"`
	Organization string `json:"organization"`
	IDNumber     string `json:"id_number"`
}

// Register handles POST /helper/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body registerBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	profile, err := h.svc.Register(r.Context(), id, body.HelperType, body.Organization, body.IDNumber)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, profile)
}

// Dashboard handles GET /helper/dashboard.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	data, err := h.svc.Dashboard(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, data)
}

type startBody struct {
	Purpose string `json:"purpose"`
}

// StartSession handles POST /helper/sessions.
func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body startBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	sess, err := h.svc.StartSession(r.Context(), id, body.Purpose)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, sess)
}

// ListSessions handles GET /helper/sessions.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	sessions, err := h.svc.ListToday(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

// GetSession handles GET /helper/sessions/{id}.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	sess, err := h.svc.GetSession(r.Context(), id, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

type consentBody struct {
	ConsentText string `json:"consent_text"`
}

// RecordConsent handles POST /helper/sessions/{id}/consent.
func (h *Handler) RecordConsent(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body consentBody
	if err := httpx.DecodeJSONOptional(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	sess, err := h.svc.RecordConsent(r.Context(), id, chi.URLParam(r, "id"), body.ConsentText)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

// EndSession handles POST /helper/sessions/{id}/end.
func (h *Handler) EndSession(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	sess, err := h.svc.EndSession(r.Context(), id, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

// AdminList handles GET /admin/helpers.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	verifiedOnly := r.URL.Query().Get("verified") == "true"
	helpers, err := h.svc.List(r.Context(), verifiedOnly)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"helpers": helpers})
}

// AdminVerify handles POST /admin/helpers/{userId}/verify.
func (h *Handler) AdminVerify(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	if err := h.svc.Verify(r.Context(), claims.UserID, chi.URLParam(r, "userId")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}
