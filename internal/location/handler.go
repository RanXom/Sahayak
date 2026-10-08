package location

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes location HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a location Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListStates handles GET /locations/states.
func (h *Handler) ListStates(w http.ResponseWriter, r *http.Request) {
	states, err := h.svc.ListStates(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"states": states})
}

// ListDistricts handles GET /locations/states/{id}/districts.
func (h *Handler) ListDistricts(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, domain.InvalidInput("Invalid state id"))
		return
	}
	districts, err := h.svc.ListDistricts(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"districts": districts})
}
