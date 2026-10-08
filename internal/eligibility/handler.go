package eligibility

import (
	"net/http"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes eligibility HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs an eligibility Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Evaluate handles POST /eligibility/evaluate.
func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	lang := domain.LanguageFromContext(r.Context())
	res, err := h.svc.Evaluate(r.Context(), claims.UserID, lang)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// Results handles GET /eligibility/results.
func (h *Handler) Results(w http.ResponseWriter, r *http.Request) {
	h.Evaluate(w, r)
}
