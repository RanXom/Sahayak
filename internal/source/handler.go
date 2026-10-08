package source

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes source registry HTTP endpoints (admin only).
type Handler struct {
	svc *Service
}

// NewHandler constructs a source Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListSources handles GET /admin/sources.
func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListSources(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sources": out})
}

// CreateSource handles POST /admin/sources.
func (h *Handler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var in domain.Source
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.CreateSource(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

// ListCandidates handles GET /admin/candidates.
func (h *Handler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := httpx.PageParams(r)
	out, total, err := h.svc.ListCandidates(r.Context(), r.URL.Query().Get("status"), perPage, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"candidates": out,
		"pagination": httpx.Pagination{Page: page, PerPage: perPage, Total: total},
	})
}

type reviewBody struct {
	Notes string `json:"notes"`
}

func reviewerID(r *http.Request) string {
	if c, ok := domain.UserFromContext(r.Context()); ok {
		return c.UserID
	}
	return ""
}

// AcceptCandidate handles POST /admin/candidates/{id}/accept.
func (h *Handler) AcceptCandidate(w http.ResponseWriter, r *http.Request) {
	var body reviewBody
	if err := httpx.DecodeJSONOptional(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.Accept(r.Context(), chi.URLParam(r, "id"), reviewerID(r), body.Notes)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// RejectCandidate handles POST /admin/candidates/{id}/reject.
func (h *Handler) RejectCandidate(w http.ResponseWriter, r *http.Request) {
	var body reviewBody
	if err := httpx.DecodeJSONOptional(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.Reject(r.Context(), chi.URLParam(r, "id"), reviewerID(r), body.Notes)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}
