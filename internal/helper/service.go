package helper

import (
	"context"
	"errors"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/domain"
)

// ConsentStatement is the exact text shown to a citizen before a helper session.
const ConsentStatement = "I consent to Sahayak assisting me with welfare scheme discovery and application. My information will be used only for this purpose and will not be shared without my permission."

// Service implements helper mode use cases.
type Service struct {
	repo      *Repository
	audit     audit.Logger
	maxPerDay int
}

// NewService constructs the helper service.
func NewService(repo *Repository, auditLogger audit.Logger, maxPerDay int) *Service {
	if maxPerDay <= 0 {
		maxPerDay = 50
	}
	return &Service{repo: repo, audit: auditLogger, maxPerDay: maxPerDay}
}

// Register creates or updates a helper profile.
func (s *Service) Register(ctx context.Context, userID, helperType, organization, idNumber string) (*HelperProfile, error) {
	switch helperType {
	case "asha", "csc_operator", "field_worker":
	default:
		return nil, domain.InvalidInput("Invalid helper_type")
	}
	h, err := s.repo.UpsertHelperProfile(ctx, userID, helperType, organization, idNumber)
	if err != nil {
		return nil, domain.Internal("Could not register helper", err)
	}
	return h, nil
}

// Verify marks a helper verified (admin action).
func (s *Service) Verify(ctx context.Context, adminID, helperUserID string) error {
	if err := s.repo.Verify(ctx, helperUserID, adminID); err != nil {
		return domain.Internal("Could not verify helper", err)
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &adminID,
		ActorRole:    domain.RoleAdmin,
		Action:       "helper.verify",
		ResourceType: "helper",
		ResourceID:   helperUserID,
	})
	return nil
}

// List returns helper profiles.
func (s *Service) List(ctx context.Context, verifiedOnly bool) ([]HelperProfile, error) {
	out, err := s.repo.ListHelpers(ctx, verifiedOnly)
	if err != nil {
		return nil, domain.Internal("Could not list helpers", err)
	}
	return out, nil
}

// IsVerified reports whether a user is a verified, active helper.
func (s *Service) IsVerified(ctx context.Context, userID string) bool {
	h, err := s.repo.GetHelperProfile(ctx, userID)
	return err == nil && h.Verified && h.Active
}

// StartSession begins a helper session (verified helpers only).
func (s *Service) StartSession(ctx context.Context, helperUserID, purpose string) (*Session, error) {
	switch purpose {
	case "scheme_discovery", "application_help", "document_help":
	default:
		return nil, domain.InvalidInput("Invalid purpose")
	}
	h, err := s.repo.GetHelperProfile(ctx, helperUserID)
	if err != nil || !h.Verified || !h.Active {
		return nil, domain.Forbidden("Your helper account is not verified")
	}
	count, err := s.repo.CountToday(ctx, helperUserID)
	if err != nil {
		return nil, domain.Internal("Could not start session", err)
	}
	if count >= s.maxPerDay {
		return nil, domain.RateLimited("Daily helper session limit reached")
	}
	sess, err := s.repo.CreateSession(ctx, helperUserID, purpose)
	if err != nil {
		return nil, domain.Internal("Could not start session", err)
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &helperUserID,
		ActorRole:    domain.RoleHelper,
		Action:       "helper.session.start",
		ResourceType: "helper_session",
		ResourceID:   sess.ID,
		Details:      map[string]any{"purpose": purpose, "consent": false},
	})
	return sess, nil
}

// RecordConsent records consent and returns the updated session.
func (s *Service) RecordConsent(ctx context.Context, helperUserID, sessionID, consentText string) (*Session, error) {
	sess, err := s.repo.GetSession(ctx, sessionID, helperUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Session not found")
		}
		return nil, domain.Internal("Could not load session", err)
	}
	if consentText == "" {
		consentText = ConsentStatement
	}
	if err := s.repo.RecordConsent(ctx, sess.ID, consentText); err != nil {
		return nil, domain.Internal("Could not record consent", err)
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &helperUserID,
		ActorRole:    domain.RoleHelper,
		Action:       "helper.session.consent",
		ResourceType: "helper_session",
		ResourceID:   sess.ID,
	})
	return s.repo.GetSession(ctx, sessionID, helperUserID)
}

// GetSession returns a session.
func (s *Service) GetSession(ctx context.Context, helperUserID, sessionID string) (*Session, error) {
	sess, err := s.repo.GetSession(ctx, sessionID, helperUserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Session not found")
		}
		return nil, domain.Internal("Could not load session", err)
	}
	return sess, nil
}

// EndSession completes a session and seals access to citizen data.
func (s *Service) EndSession(ctx context.Context, helperUserID, sessionID string) (*Session, error) {
	sess, err := s.GetSession(ctx, helperUserID, sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.EndSession(ctx, sess.ID); err != nil {
		return nil, domain.Internal("Could not end session", err)
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &helperUserID,
		ActorRole:    domain.RoleHelper,
		Action:       "helper.session.end",
		ResourceType: "helper_session",
		ResourceID:   sess.ID,
	})
	return s.repo.GetSession(ctx, helperUserID, sessionID)
}

// ListToday returns today's sessions for a helper.
func (s *Service) ListToday(ctx context.Context, helperUserID string) ([]Session, error) {
	out, err := s.repo.ListToday(ctx, helperUserID)
	if err != nil {
		return nil, domain.Internal("Could not list sessions", err)
	}
	return out, nil
}

// Dashboard summarizes a helper's day.
func (s *Service) Dashboard(ctx context.Context, helperUserID string) (map[string]any, error) {
	sessions, err := s.repo.ListToday(ctx, helperUserID)
	if err != nil {
		return nil, domain.Internal("Could not load dashboard", err)
	}
	active := 0
	for _, sess := range sessions {
		if sess.Status == "active" {
			active++
		}
	}
	return map[string]any{
		"sessions_today":  len(sessions),
		"active_sessions": active,
		"sessions":        sessions,
	}, nil
}
