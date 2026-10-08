package profile

import (
	"net/http"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes profile HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a profile Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func userID(r *http.Request) (string, bool) {
	c, ok := domain.UserFromContext(r.Context())
	if !ok {
		return "", false
	}
	return c.UserID, true
}

// Get handles GET /profile.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	view, err := h.svc.View(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

// Update handles PUT /profile.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var in UpdateInput
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.svc.Update(r.Context(), id, in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	view, err := h.svc.View(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

type mergeFactsBody struct {
	Facts      map[string]any `json:"facts"`
	StateID    *int           `json:"state_id"`
	DistrictID *int           `json:"district_id"`
}

// MergeFacts handles PATCH /profile/facts.
func (h *Handler) MergeFacts(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body mergeFactsBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := h.svc.MergeFacts(r.Context(), id, body.Facts, body.StateID, body.DistrictID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	view, err := h.svc.View(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

// GetDocuments handles GET /profile/documents.
func (h *Handler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	docs, err := h.svc.GetDocuments(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"documents": docs})
}

type updateDocumentsBody struct {
	Documents []DocumentUpdate `json:"documents"`
}

// UpdateDocuments handles PUT /profile/documents.
func (h *Handler) UpdateDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body updateDocumentsBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	docs, err := h.svc.UpdateDocuments(r.Context(), id, body.Documents)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"documents": docs})
}
