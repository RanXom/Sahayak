package assistance

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes the marketplace HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func userID(r *http.Request) (string, string, bool) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		return "", "", false
	}
	return claims.UserID, claims.Role, true
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

type createRequestBody struct {
	ApplicationID     *string `json:"application_id"`
	SchemeID          *string `json:"scheme_id"`
	Language          string  `json:"language"`
	Category          string  `json:"category"`
	Details           string  `json:"details"`
	ShareProfileFacts bool    `json:"share_profile_facts"`
	ShareDocuments    bool    `json:"share_documents"`
}

// CreateRequest handles POST /assistance/requests.
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body createRequestBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.CreateRequest(r.Context(), uid, CreateRequestInput{
		ApplicationID: body.ApplicationID, SchemeID: body.SchemeID,
		Language: body.Language, Category: body.Category, Details: body.Details,
		ShareProfileFacts: body.ShareProfileFacts, ShareDocuments: body.ShareDocuments,
	}, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

// MyRequests handles GET /assistance/requests.
func (h *Handler) MyRequests(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, err := h.svc.repo.ListCitizenRequests(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, domain.Internal("Could not list requests", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"requests": out})
}

// GetRequest handles GET /assistance/requests/{id} for owner or assignee.
func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	q, err := h.svc.repo.GetRequest(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, domain.NotFound("Request not found"))
		return
	}
	if q.CitizenUserID != uid && (q.MatchedHelperID == nil || *q.MatchedHelperID != uid) {
		httpx.WriteError(w, domain.NotFound("Request not found"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, q)
}

// Queue handles GET /assistance/queue (helpers).
func (h *Handler) Queue(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, busy, err := h.svc.Queue(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"requests": out, "has_active_session": busy})
}

// Accept handles POST /assistance/requests/{id}/accept (helpers).
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	sess, err := h.svc.Accept(r.Context(), uid, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

// Join handles POST /assistance/requests/{id}/join (citizen confirms + consents).
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body struct {
		ShareProfileFacts bool `json:"share_profile_facts"`
		ShareDocuments    bool `json:"share_documents"`
	}
	_ = httpx.DecodeJSON(w, r, &body)
	sess, err := h.svc.JoinRequest(r.Context(), uid, chi.URLParam(r, "id"), body.ShareProfileFacts, body.ShareDocuments, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

// Cancel handles POST /assistance/requests/{id}/cancel (citizen).
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, err := h.svc.CancelRequest(r.Context(), uid, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// MySessions handles GET /assistance/sessions.
func (h *Handler) MySessions(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var out []Session
	var err error
	if role == domain.RoleHelper {
		out, err = h.svc.repo.HelperSessions(r.Context(), uid)
	} else {
		out, err = h.svc.repo.CitizenSessions(r.Context(), uid)
	}
	if err != nil {
		httpx.WriteError(w, domain.Internal("Could not list sessions", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

// GetSession handles GET /assistance/sessions/{id} (participants only).
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	sess, _, err := h.svc.assertParticipant(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sess)
}

// SharedContext handles GET /assistance/sessions/{id}/context.
func (h *Handler) SharedContext(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, err := h.svc.SharedContext(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// EndSession handles POST /assistance/sessions/{id}/end.
func (h *Handler) EndSession(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, err := h.svc.EndSession(r.Context(), uid, role, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

type messageBody struct {
	Body        string `json:"body"`
	ClientMsgID string `json:"client_msg_id"`
}

// SendMessage handles POST /assistance/sessions/{id}/messages.
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body messageBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.SendMessage(r.Context(), uid, chi.URLParam(r, "id"), body.Body, body.ClientMsgID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

// ListMessages handles GET /assistance/sessions/{id}/messages?since=&limit=.
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	since, limit := parseCursor(r)
	out, err := h.svc.ListMessages(r.Context(), uid, chi.URLParam(r, "id"), since, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"messages": out})
}

type signalBody struct {
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload"`
}

// SendSignal handles POST /assistance/sessions/{id}/signals.
func (h *Handler) SendSignal(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body signalBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.SendSignal(r.Context(), uid, chi.URLParam(r, "id"), body.Kind, body.Payload)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

// ListSignals handles GET /assistance/sessions/{id}/signals?since=&limit=.
func (h *Handler) ListSignals(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	since, limit := parseCursor(r)
	out, err := h.svc.ListSignals(r.Context(), uid, chi.URLParam(r, "id"), since, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"signals": out})
}

type draftBody struct {
	FieldName string `json:"field_name"`
	Value     any    `json:"value"`
}

// ProposeDraft handles PUT /assistance/sessions/{id}/drafts.
func (h *Handler) ProposeDraft(w http.ResponseWriter, r *http.Request) {
	uid, role, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body draftBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.ProposeDraft(r.Context(), uid, role, chi.URLParam(r, "id"), body.FieldName, body.Value)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// ConfirmDraft handles POST /assistance/sessions/{id}/drafts/confirm.
func (h *Handler) ConfirmDraft(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body struct {
		FieldName string `json:"field_name"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	out, err := h.svc.ConfirmDraft(r.Context(), uid, chi.URLParam(r, "id"), body.FieldName)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// ApplyDrafts handles POST /assistance/sessions/{id}/drafts/apply.
func (h *Handler) ApplyDrafts(w http.ResponseWriter, r *http.Request) {
	uid, _, ok := userID(r)
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	out, err := h.svc.ApplyDrafts(r.Context(), uid, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"applied": out})
}

func parseCursor(r *http.Request) (time.Time, int) {
	var since time.Time
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}
	limit := 50
	return since, limit
}
