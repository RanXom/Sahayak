package scheme

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes scheme HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a scheme Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Get handles GET /schemes/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	sch, err := h.svc.GetPublished(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sch)
}

// AdminList handles GET /admin/schemes.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := httpx.PageParams(r)
	status := r.URL.Query().Get("status")
	schemes, total, err := h.svc.ListAdmin(r.Context(), status, perPage, offset)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"schemes":    schemes,
		"pagination": httpx.Pagination{Page: page, PerPage: perPage, Total: total},
	})
}

// AdminCreate handles POST /admin/schemes.
func (h *Handler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	var in Input
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	sch, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, sch)
}

// AdminUpdate handles PUT /admin/schemes/{id}.
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	var in Input
	if err := httpx.DecodeJSON(w, r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	sch, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sch)
}

// AdminPublish handles POST /admin/schemes/{id}/publish.
func (h *Handler) AdminPublish(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Publish(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": domain.SchemePublished})
}

// AdminArchive handles POST /admin/schemes/{id}/archive.
func (h *Handler) AdminArchive(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Archive(r.Context(), chi.URLParam(r, "id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": domain.SchemeArchived})
}
