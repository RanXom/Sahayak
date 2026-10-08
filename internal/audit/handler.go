package audit

import (
	"net/http"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes audit query endpoints (admin only).
type Handler struct {
	repo *Repository
}

// NewHandler constructs an audit Handler.
func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

// Query handles GET /admin/audit.
func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := httpx.PageParams(r)
	entries, total, err := h.repo.Query(r.Context(), r.URL.Query().Get("action"), perPage, offset)
	if err != nil {
		httpx.WriteError(w, domain.Internal("Could not query audit log", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"entries":    entries,
		"pagination": httpx.Pagination{Page: page, PerPage: perPage, Total: total},
	})
}
