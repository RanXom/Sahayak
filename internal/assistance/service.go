package assistance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sahayak/sahayak/internal/application"
	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/profile"
	"github.com/sahayak/sahayak/internal/scheme"
)

// Service implements the marketplace use cases.
type Service struct {
	repo     *Repository
	apps     *application.Repository
	profiles *profile.Repository
	schemes  *scheme.Repository
	audit    audit.Logger
}

// Deps wires the Service.
type Deps struct {
	Repo     *Repository
	Apps     *application.Repository
	Profiles *profile.Repository
	Schemes  *scheme.Repository
	Audit    audit.Logger
}

// NewService constructs a Service.
func NewService(d Deps) *Service {
	if d.Audit == nil {
		d.Audit = audit.NopLogger{}
	}
	return &Service{repo: d.Repo, apps: d.Apps, profiles: d.Profiles, schemes: d.Schemes, audit: d.Audit}
}

// CreateRequestInput is the citizen's help request.
type CreateRequestInput struct {
	ApplicationID     *string
	SchemeID          *string
	Language          string
	Category          string
	Details           string
	ShareProfileFacts bool
	ShareDocuments    bool
}

// CreateRequest files a pending assistance request.
func (s *Service) CreateRequest(ctx context.Context, citizenID string, in CreateRequestInput, ip, ua string) (*Request, error) {
	in.Language = strings.ToLower(strings.TrimSpace(in.Language))
	in.Category = strings.TrimSpace(in.Category)
	if !ValidLanguage(in.Language) {
		return nil, domain.InvalidInput("Unsupported assistance language")
	}
	if !ValidCategory(in.Category) {
		return nil, domain.InvalidInput("Unknown assistance category")
	}
	if in.ApplicationID != nil && *in.ApplicationID != "" {
		if _, err := s.apps.Get(ctx, *in.ApplicationID, citizenID); err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return nil, domain.NotFound("Application not found")
			}
			return nil, domain.Internal("Could not load application", err)
		}
	}
	if in.SchemeID != nil && *in.SchemeID != "" {
		if _, err := s.schemes.GetPublishedByID(ctx, *in.SchemeID); err != nil {
			if errors.Is(err, scheme.ErrNotFound) {
				return nil, domain.NotFound("Scheme not found")
			}
			return nil, domain.Internal("Could not load scheme", err)
		}
	}
	q, err := s.repo.CreateRequest(ctx, citizenID, in.ApplicationID, in.SchemeID, in.Language, in.Category, strings.TrimSpace(in.Details), in.ShareProfileFacts, in.ShareDocuments)
	if err != nil {
		return nil, domain.Internal("Could not create request", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &citizenID, ActorRole: domain.RoleCitizen,
		Action: "assist.request.create", ResourceType: "request", ResourceID: q.ID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"language": q.Language, "category": q.Category}})
	return q, nil
}

// Queue returns pending requests ranked for one helper. Expired rows are
// swept first so helpers never see stale work.
func (s *Service) Queue(ctx context.Context, helperID string) ([]Candidate, bool, error) {
	if _, err := s.repo.ExpireOverdue(ctx); err != nil {
		return nil, false, domain.Internal("Could not sweep requests", err)
	}
	prof, err := s.repo.GetHelperMatchProfile(ctx, helperID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, false, domain.InvalidInput("Complete your helper profile first")
		}
		return nil, false, domain.Internal("Could not load helper profile", err)
	}
	if !prof.IsAvailable {
		return []Candidate{}, false, nil
	}
	pending, err := s.repo.ListPending(ctx)
	if err != nil {
		return nil, false, domain.Internal("Could not list requests", err)
	}
	now := time.Now()
	var out []Candidate
	for _, q := range pending {
		ok, score := MatchScore(q.Language, q.Category, q.CreatedAt, prof.Languages, prof.Categories, now)
		if !ok {
			continue
		}
		out = append(out, Candidate{Request: q, Score: score})
	}
	SortCandidates(out)
	busy, err := s.repo.HelperHasActiveSession(ctx, helperID)
	if err != nil {
		return nil, false, domain.Internal("Could not check assignments", err)
	}
	return out, busy, nil
}

// Accept atomically assigns a pending request to a helper and opens the
// session. Exactly one concurrent accepter wins; the other gets a conflict.
func (s *Service) Accept(ctx context.Context, helperID, requestID, ip, ua string) (*Session, error) {
	prof, err := s.repo.GetHelperMatchProfile(ctx, helperID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.InvalidInput("Complete your helper profile first")
		}
		return nil, domain.Internal("Could not load helper profile", err)
	}
	if !prof.IsAvailable {
		return nil, domain.InvalidInput("Mark yourself available before accepting requests")
	}
	if busy, err := s.repo.HelperHasActiveSession(ctx, helperID); err != nil {
		return nil, domain.Internal("Could not check assignments", err)
	} else if busy {
		return nil, domain.InvalidInput("Finish your active session before accepting another")
	}
	q, err := s.repo.AcceptRequest(ctx, requestID, helperID)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, domain.InvalidInput("This request has already been assigned")
		}
		return nil, domain.Internal("Could not accept request", err)
	}
	if !languageAllowed(q.Language, prof.Languages) {
		// Belt and suspenders: the atomic accept above cannot check languages.
		// Roll the request back to pending rather than mis-assigning it.
		_, _ = s.repo.SetRequestStatus(ctx, requestID, StatusPending, []string{StatusAccepted})
		return nil, domain.InvalidInput("This request needs a different assistance language")
	}
	sess, err := s.repo.CreateSession(ctx, q.ID, helperID, q.CitizenUserID, q.Category)
	if err != nil {
		return nil, domain.Internal("Could not open session", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &helperID, ActorRole: domain.RoleHelper,
		Action: "assist.request.accept", ResourceType: "request", ResourceID: q.ID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"session_id": sess.ID}})
	return sess, nil
}

// JoinRequest lets the citizen confirm the assigned helper, granting consent
// scoped to the request. This moves pending/accepted work into session.
func (s *Service) JoinRequest(ctx context.Context, citizenID, requestID string, shareFacts, shareDocs bool, ip, ua string) (*Session, error) {
	q, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Request not found")
		}
		return nil, domain.Internal("Could not load request", err)
	}
	if q.CitizenUserID != citizenID {
		return nil, domain.NotFound("Request not found")
	}
	if q.Status != StatusAccepted || q.MatchedHelperID == nil {
		return nil, domain.InvalidInput("No helper is assigned to this request yet")
	}
	sess, err := s.repo.GetSessionByRequest(ctx, requestID)
	if err != nil {
		return nil, domain.Internal("Could not load session", err)
	}
	if _, err := s.repo.SetRequestStatus(ctx, requestID, StatusInSession, []string{StatusAccepted}); err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, domain.InvalidInput("This request is no longer awaiting you")
		}
		return nil, domain.Internal("Could not start session", err)
	}
	scope := map[string]any{
		"application_id":      q.ApplicationID,
		"scheme_id":           q.SchemeID,
		"share_profile_facts": shareFacts,
		"share_documents":     shareDocs,
	}
	statement := fmt.Sprintf("I allow the assigned helper to view %s for request %s to help prepare my application.",
		scopeSummary(shareFacts, shareDocs), requestID)
	if _, err := s.repo.GrantConsent(ctx, sess.ID, requestID, citizenID, scope, statement, ip, ua); err != nil {
		return nil, domain.Internal("Could not record consent", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &citizenID, ActorRole: domain.RoleCitizen,
		Action: "assist.consent.grant", ResourceType: "session", ResourceID: sess.ID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"scope": scope}})
	return sess, nil
}

// CancelRequest withdraws a request. Active work is closed and consent revoked.
func (s *Service) CancelRequest(ctx context.Context, citizenID, requestID, ip, ua string) (*Request, error) {
	q, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Request not found")
		}
		return nil, domain.Internal("Could not load request", err)
	}
	if q.CitizenUserID != citizenID {
		return nil, domain.NotFound("Request not found")
	}
	switch q.Status {
	case StatusPending:
		// No session exists yet.
	case StatusAccepted, StatusInSession:
		if sess, err := s.repo.GetSessionByRequest(ctx, requestID); err == nil && sess.Status == "active" {
			if _, err := s.repo.EndSession(ctx, sess.ID, "citizen", "cancelled"); err != nil {
				return nil, domain.Internal("Could not close session", err)
			}
			_ = s.repo.RevokeConsents(ctx, sess.ID)
		}
	default:
		return nil, domain.InvalidInput("This request can no longer be cancelled")
	}
	out, err := s.repo.SetRequestStatus(ctx, requestID, StatusCancelled, []string{StatusPending, StatusAccepted, StatusInSession})
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, domain.InvalidInput("This request can no longer be cancelled")
		}
		return nil, domain.Internal("Could not cancel request", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &citizenID, ActorRole: domain.RoleCitizen,
		Action: "assist.request.cancel", ResourceType: "request", ResourceID: requestID,
		IPAddress: ip, UserAgent: ua})
	return out, nil
}

// EndSession closes an active session, completes the request and revokes
// helper access immediately.
func (s *Service) EndSession(ctx context.Context, userID, role, sessionID, ip, ua string) (*Session, error) {
	sess, roleOf, err := s.assertParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	_ = roleOf
	out, err := s.repo.EndSession(ctx, sessionID, role, "completed")
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, domain.InvalidInput("This session is already closed")
		}
		return nil, domain.Internal("Could not close session", err)
	}
	_ = s.repo.RevokeConsents(ctx, sessionID)
	if sess.RequestID != nil {
		_, _ = s.repo.SetRequestStatus(ctx, *sess.RequestID, StatusCompleted,
			[]string{StatusAccepted, StatusInSession})
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &userID, ActorRole: role,
		Action: "assist.session.end", ResourceType: "session", ResourceID: sessionID,
		IPAddress: ip, UserAgent: ua})
	return out, nil
}

// assertParticipant loads a session only for its assigned citizen or helper.
// Outsiders get NotFound so session existence does not leak.
func (s *Service) assertParticipant(ctx context.Context, userID, sessionID string) (*Session, string, error) {
	sess, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, "", domain.NotFound("Session not found")
		}
		return nil, "", domain.Internal("Could not load session", err)
	}
	if sess.HelperUserID == userID {
		return sess, domain.RoleHelper, nil
	}
	if sess.CitizenUserID != nil && *sess.CitizenUserID == userID {
		return sess, domain.RoleCitizen, nil
	}
	return nil, "", domain.NotFound("Session not found")
}

// assertActiveParticipant additionally requires an unfinished session.
func (s *Service) assertActiveParticipant(ctx context.Context, userID, sessionID string) (*Session, string, error) {
	sess, role, err := s.assertParticipant(ctx, userID, sessionID)
	if err != nil {
		return nil, "", err
	}
	if sess.Status != "active" {
		return nil, "", domain.InvalidInput("This session is closed")
	}
	return sess, role, nil
}

func languageAllowed(lang string, langs []string) bool {
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}

func scopeSummary(facts, docs bool) string {
	switch {
	case facts && docs:
		return "profile details and documents"
	case facts:
		return "profile details"
	case docs:
		return "documents"
	default:
		return "application context only"
	}
}
