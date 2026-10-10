package assistance

import (
	"context"
	"strings"
	"time"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/domain"
)

// SendMessage stores one chat line. Only participants of an active session
// may write; client_msg_id makes retries idempotent.
func (s *Service) SendMessage(ctx context.Context, userID, sessionID, body, clientMsgID string) (*Message, error) {
	_, role, err := s.assertActiveParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, domain.InvalidInput("Message text is required")
	}
	m, err := s.repo.AddMessage(ctx, sessionID, userID, role, clientMsgID, body)
	if err != nil {
		return nil, domain.Internal("Could not send message", err)
	}
	return m, nil
}

// ListMessages returns chat history after a cursor for polling clients.
func (s *Service) ListMessages(ctx context.Context, userID, sessionID string, since time.Time, limit int) ([]Message, error) {
	sess, _, err := s.assertParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	out, err := s.repo.ListMessages(ctx, sess.ID, since, limit)
	if err != nil {
		return nil, domain.Internal("Could not load messages", err)
	}
	for i := range out {
		out[i].Mine = out[i].SenderID == userID
		out[i].SenderID = ""
	}
	return out, nil
}

// ValidSignalKinds are the accepted WebRTC signaling payloads.
var ValidSignalKinds = []string{"offer", "answer", "ice", "hangup", "join", "leave"}

// SendSignal stores one signaling message for the other participant to poll.
func (s *Service) SendSignal(ctx context.Context, userID, sessionID, kind string, payload map[string]any) (*Signal, error) {
	_, _, err := s.assertActiveParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	ok := false
	for _, k := range ValidSignalKinds {
		if k == kind {
			ok = true
			break
		}
	}
	if !ok {
		return nil, domain.InvalidInput("Unknown signal kind")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	sig, err := s.repo.AddSignal(ctx, sessionID, userID, kind, payload)
	if err != nil {
		return nil, domain.Internal("Could not send signal", err)
	}
	return sig, nil
}

// ListSignals returns signaling messages after a cursor for polling clients.
func (s *Service) ListSignals(ctx context.Context, userID, sessionID string, since time.Time, limit int) ([]Signal, error) {
	sess, _, err := s.assertParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	out, err := s.repo.ListSignals(ctx, sess.ID, since, limit)
	if err != nil {
		return nil, domain.Internal("Could not load signals", err)
	}
	for i := range out {
		out[i].Mine = out[i].SenderID == userID
		out[i].SenderID = ""
	}
	return out, nil
}

// ProposeDraft writes a session draft value. Helpers create proposals;
// citizens writing directly confirm immediately.
func (s *Service) ProposeDraft(ctx context.Context, userID, role, sessionID, field string, value any) (*DraftValue, error) {
	if _, _, err := s.assertActiveParticipant(ctx, userID, sessionID); err != nil {
		return nil, err
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return nil, domain.InvalidInput("Field name is required")
	}
	status := "proposed"
	if role == domain.RoleCitizen {
		status = "confirmed"
	}
	d, err := s.repo.UpsertDraft(ctx, sessionID, field, value, userID, status)
	if err != nil {
		return nil, domain.Internal("Could not save draft value", err)
	}
	return d, nil
}

// ConfirmDraft marks a helper's proposal confirmed. Citizens only.
func (s *Service) ConfirmDraft(ctx context.Context, userID, sessionID, field string) (*DraftValue, error) {
	sess, role, err := s.assertActiveParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleCitizen {
		return nil, domain.InvalidInput("Only the citizen can confirm values")
	}
	drafts, err := s.repo.ListDrafts(ctx, sess.ID)
	if err != nil {
		return nil, domain.Internal("Could not load drafts", err)
	}
	var cur *DraftValue
	for i := range drafts {
		if drafts[i].FieldName == field {
			cur = &drafts[i]
			break
		}
	}
	if cur == nil {
		return nil, domain.NotFound("Draft value not found")
	}
	d, err := s.repo.UpsertDraft(ctx, sess.ID, field, cur.Value, userID, "confirmed")
	if err != nil {
		return nil, domain.Internal("Could not confirm value", err)
	}
	return d, nil
}

// ApplyDrafts writes confirmed values into the application form. Requires a
// live consent grant and an active session.
func (s *Service) ApplyDrafts(ctx context.Context, userID, sessionID, ip, ua string) (map[string]any, error) {
	sess, role, err := s.assertActiveParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleCitizen {
		return nil, domain.InvalidInput("Only the citizen can apply values")
	}
	ok, err := s.repo.ValidConsent(ctx, sessionID)
	if err != nil {
		return nil, domain.Internal("Could not check consent", err)
	}
	if !ok {
		return nil, domain.InvalidInput("Consent is required before applying values")
	}
	if sess.RequestID == nil {
		return nil, domain.InvalidInput("This session has no linked request")
	}
	req, err := s.repo.GetRequest(ctx, *sess.RequestID)
	if err != nil {
		return nil, domain.Internal("Could not load request", err)
	}
	if req.ApplicationID == nil || *req.ApplicationID == "" {
		return nil, domain.InvalidInput("This session has no linked application")
	}
	drafts, err := s.repo.ListDrafts(ctx, sess.ID)
	if err != nil {
		return nil, domain.Internal("Could not load drafts", err)
	}
	form := map[string]any{}
	for _, d := range drafts {
		if d.Status == "confirmed" {
			form[d.FieldName] = d.Value
		}
	}
	if len(form) == 0 {
		return nil, domain.InvalidInput("No confirmed values to apply")
	}
	app, err := s.apps.Get(ctx, *req.ApplicationID, userID)
	if err != nil {
		return nil, domain.Internal("Could not load application", err)
	}
	merged := map[string]any{}
	for k, v := range app.FormData {
		merged[k] = v
	}
	for k, v := range form {
		merged[k] = v
	}
	if err := s.apps.UpdateFormData(ctx, app.ID, merged, app.Status); err != nil {
		return nil, domain.Internal("Could not update application", err)
	}
	for field := range form {
		_, _ = s.repo.UpsertDraft(ctx, sess.ID, field, form[field], userID, "applied")
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &userID, ActorRole: role,
		Action: "assist.draft.apply", ResourceType: "session", ResourceID: sessionID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"fields": len(form)}})
	return form, nil
}

// SharedContext returns what the helper is allowed to see: the linked
// application, profile facts and documents strictly within the live consent
// scope, plus the draft board.
func (s *Service) SharedContext(ctx context.Context, userID, sessionID string) (map[string]any, error) {
	sess, role, err := s.assertActiveParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.RequestID == nil {
		return nil, domain.NotFound("Session has no linked request")
	}
	req, err := s.repo.GetRequest(ctx, *sess.RequestID)
	if err != nil {
		return nil, domain.Internal("Could not load request", err)
	}
	out := map[string]any{"request": req}
	if role == domain.RoleCitizen {
		drafts, err := s.repo.ListDrafts(ctx, sess.ID)
		if err != nil {
			return nil, domain.Internal("Could not load drafts", err)
		}
		out["drafts"] = drafts
		return out, nil
	}
	// Helper view: consent-gated.
	ok, err := s.repo.ValidConsent(ctx, sessionID)
	if err != nil {
		return nil, domain.Internal("Could not check consent", err)
	}
	if !ok {
		return nil, domain.InvalidInput("The citizen has not granted access yet")
	}
	consents, err := s.repo.ListDrafts(ctx, sess.ID)
	if err != nil {
		return nil, domain.Internal("Could not load drafts", err)
	}
	out["drafts"] = consents
	// Scope comes from the request the citizen filed.
	if req.ShareProfileFacts {
		if p, err := s.profiles.GetByUserID(ctx, req.CitizenUserID); err == nil {
			out["profile_facts"] = p.Facts
		}
	}
	if req.ShareDocuments {
		if docs, err := s.profiles.GetDocuments(ctx, req.CitizenUserID); err == nil {
			names := make([]map[string]any, 0, len(docs))
			for _, d := range docs {
				names = append(names, map[string]any{
					"document_type": d.DocumentType, "has_document": d.HasDocument,
				})
			}
			out["documents"] = names
		}
	}
	if req.ApplicationID != nil && *req.ApplicationID != "" {
		if app, err := s.apps.GetByID(ctx, *req.ApplicationID); err == nil {
			out["application"] = map[string]any{
				"id": app.ID, "scheme_id": app.SchemeID, "status": app.Status,
				"form_data": app.FormData,
			}
		}
	}
	return out, nil
}
