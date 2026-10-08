package document

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes document HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a document Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListTypes handles GET /documents/types.
func (h *Handler) ListTypes(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListTypes(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"document_types": out})
}

// SchemeDocuments handles GET /schemes/{id}/documents.
func (h *Handler) SchemeDocuments(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	lang := domain.LanguageFromContext(r.Context())
	res, err := h.svc.CheckScheme(r.Context(), claims.UserID, chi.URLParam(r, "id"), lang)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}
