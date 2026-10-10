package helper

import (
	"context"
	"errors"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/domain"
)

// ConsentStatement is the exact text shown to a citizen before a helper session.
const ConsentStatement = "I consent to Sahayak assisting me with welfare scheme discovery and application. My information will be used only for this purpose and will not be shared without my permission."

// Accounts promotes citizen accounts to the helper role.
type Accounts interface {
	PromoteToHelper(ctx context.Context, userID string) (bool, error)
}

// Service implements helper mode use cases.
type Service struct {
	repo      *Repository
	accounts  Accounts
	audit     audit.Logger
	maxPerDay int
}

// NewService constructs the helper service.
func NewService(repo *Repository, accounts Accounts, auditLogger audit.Logger, maxPerDay int) *Service {
	if maxPerDay <= 0 {
		maxPerDay = 50
	}
	return &Service{repo: repo, accounts: accounts, audit: auditLogger, maxPerDay: maxPerDay}
}

// ValidHelperTypes are the accepted helper kinds. Registration never implies
// government certification; verified status comes only from admin review.
func ValidHelperTypes() []string {
	return []string{"asha", "csc_operator", "field_worker", "volunteer", "ngo_worker"}
}

// Register creates or updates a helper profile and promotes the account to
// the helper role (citizens only; admins untouched). Helpers must still pass
// admin verification to earn the verified badge.
func (s *Service) Register(ctx context.Context, userID string, in ProfileInput, ip, ua string) (*HelperProfile, error) {
	if err := ValidateProfileInput(in); err != nil {
		return nil, err
	}
	h, err := s.repo.UpsertHelperProfile(ctx, userID, in)
	if err != nil {
		return nil, domain.Internal("Could not register helper", err)
	}
	promoted, err := s.accounts.PromoteToHelper(ctx, userID)
	if err != nil {
		return nil, domain.Internal("Could not update account role", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &userID, ActorRole: domain.RoleHelper,
		Action: "helper.register", ResourceType: "helper", ResourceID: userID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"promoted": promoted}})
	return h, nil
}

// UpdateProfile edits the caller's helper profile.
func (s *Service) UpdateProfile(ctx context.Context, userID string, in ProfileInput, ip, ua string) (*HelperProfile, error) {
	if err := ValidateProfileInput(in); err != nil {
		return nil, err
	}
	h, err := s.repo.UpsertHelperProfile(ctx, userID, in)
	if err != nil {
		return nil, domain.Internal("Could not update helper profile", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &userID, ActorRole: domain.RoleHelper,
		Action: "helper.profile.update", ResourceType: "helper", ResourceID: userID,
		IPAddress: ip, UserAgent: ua})
	return h, nil
}

// SetAvailability flips queue visibility.
func (s *Service) SetAvailability(ctx context.Context, userID string, available bool, ip, ua string) error {
	if _, err := s.repo.GetHelperProfile(ctx, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.InvalidInput("Complete your helper profile first")
		}
		return domain.Internal("Could not load helper profile", err)
	}
	if err := s.repo.SetAvailability(ctx, userID, available); err != nil {
		return domain.Internal("Could not update availability", err)
	}
	s.audit.Log(ctx, audit.Entry{ActorID: &userID, ActorRole: domain.RoleHelper,
		Action: "helper.availability", ResourceType: "helper", ResourceID: userID,
		IPAddress: ip, UserAgent: ua, Details: map[string]any{"is_available": available}})
	return nil
}

// ValidateProfileInput enforces the registration contract.
func ValidateProfileInput(in ProfileInput) error {
	ok := false
	for _, t := range ValidHelperTypes() {
		if t == in.HelperType {
			ok = true
			break
		}
	}
	if !ok {
		return domain.InvalidInput("Invalid helper_type")
	}
	if len(in.Languages) == 0 {
		return domain.InvalidInput("Select at least one assistance language")
	}
	for _, l := range in.Languages {
		if !validLang(l) {
			return domain.InvalidInput("Unsupported language: " + l)
		}
	}
	for _, c := range in.Categories {
		if !validCategory(c) {
			return domain.InvalidInput("Unknown assistance category: " + c)
		}
	}
	if !in.TermsAccepted {
		return domain.InvalidInput("Helper terms must be accepted")
	}
	return nil
}

func validLang(l string) bool {
	switch l {
	case "hi", "en", "ta", "mr", "bn", "te", "kn", "ml", "gu", "pa", "or", "ur":
		return true
	}
	return false
}

func validCategory(c string) bool {
	switch c {
	case "scheme_discovery", "application_help", "document_help":
		return true
	}
	return false
}

// Verify marks a helper verified (admin action) and ensures the role.
func (s *Service) Verify(ctx context.Context, adminID, helperUserID string) error {
	if err := s.repo.Verify(ctx, helperUserID, adminID); err != nil {
		return domain.Internal("Could not verify helper", err)
	}
	_, _ = s.accounts.PromoteToHelper(ctx, helperUserID)
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &adminID,
		ActorRole:    domain.RoleAdmin,
		Action:       "helper.verify",
		ResourceType: "helper",
		ResourceID:   helperUserID,
	})
	return nil
}

// Profile returns the caller's helper profile.
func (s *Service) Profile(ctx context.Context, userID string) (*HelperProfile, error) {
	h, err := s.repo.GetHelperProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Helper profile not found")
		}
		return nil, domain.Internal("Could not load helper profile", err)
	}
	return h, nil
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
