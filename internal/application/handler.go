package application

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes application HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs an application Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func claimsFrom(r *http.Request) (*domain.Claims, bool) {
	return domain.UserFromContext(r.Context())
}

type createBody struct {
	SchemeID string `json:"scheme_id"`
}

// Create handles POST /applications.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body createBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if body.SchemeID == "" {
		httpx.WriteError(w, domain.InvalidInput("scheme_id is required"))
		return
	}
	view, err := h.svc.CreateDraft(r.Context(), claims.UserID, body.SchemeID, domain.LanguageFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, view)
}

// List handles GET /applications.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	apps, err := h.svc.List(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"applications": apps})
}

// Get handles GET /applications/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	view, err := h.svc.Get(r.Context(), claims.UserID, chi.URLParam(r, "id"), domain.LanguageFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, view)
}

type formBody struct {
	FormData map[string]any `json:"form_data"`
}

// UpdateForm handles PUT /applications/{id}/form.
func (h *Handler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body formBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	app, err := h.svc.UpdateForm(r.Context(), claims.UserID, chi.URLParam(r, "id"), body.FormData)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

// GeneratePDF handles POST /applications/{id}/pdf.
func (h *Handler) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	app, err := h.svc.GeneratePDF(r.Context(), claims.UserID, chi.URLParam(r, "id"), domain.LanguageFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

// DownloadPDF handles GET /applications/{id}/pdf.
func (h *Handler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	path, err := h.svc.PDFPath(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		httpx.WriteError(w, domain.NotFound("PDF not found"))
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(path))
	http.ServeFile(w, r, path)
}

type submitBody struct {
	SubmissionRef string `json:"submission_ref"`
}

// Submit handles POST /applications/{id}/submit (user-reported submission).
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body submitBody
	if err := httpx.DecodeJSONOptional(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	app, err := h.svc.RecordSubmission(r.Context(), claims.UserID, chi.URLParam(r, "id"), body.SubmissionRef)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}

type statusBody struct {
	Status string `json:"status"`
}

// UpdateStatus handles PUT /applications/{id}/status.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body statusBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	app, err := h.svc.UpdateStatus(r.Context(), claims.UserID, chi.URLParam(r, "id"), body.Status)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, app)
}
