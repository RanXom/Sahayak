package helper

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HelperProfile is a welfare worker's verified profile.
type HelperProfile struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	HelperType   string     `json:"helper_type"`
	Organization string     `json:"organization,omitempty"`
	Verified     bool       `json:"verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	Active       bool       `json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Session is a helper-assisted citizen session.
type Session struct {
	ID             string     `json:"id"`
	HelperUserID   string     `json:"helper_user_id"`
	CitizenUserID  *string    `json:"citizen_user_id,omitempty"`
	ConsentGiven   bool       `json:"consent_given"`
	ConsentText    string     `json:"consent_text,omitempty"`
	ConsentAt      *time.Time `json:"consent_at,omitempty"`
	Purpose        string     `json:"purpose"`
	Status         string     `json:"status"`
	ConversationID *string    `json:"conversation_id,omitempty"`
	ProfileID      *string    `json:"profile_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// Repository provides persistence for helper mode.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a helper Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// UpsertHelperProfile creates or updates a helper profile.
func (r *Repository) UpsertHelperProfile(ctx context.Context, userID, helperType, organization, idNumber string) (*HelperProfile, error) {
	var h HelperProfile
	err := r.pool.QueryRow(ctx, `
		INSERT INTO helper_profiles (user_id, helper_type, organization, id_number)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''))
		ON CONFLICT (user_id) DO UPDATE SET
			helper_type = EXCLUDED.helper_type,
			organization = EXCLUDED.organization,
			id_number = EXCLUDED.id_number
		RETURNING id, user_id, helper_type, COALESCE(organization,''), verified, verified_at, active, created_at`,
		userID, helperType, organization, idNumber).
		Scan(&h.ID, &h.UserID, &h.HelperType, &h.Organization, &h.Verified, &h.VerifiedAt, &h.Active, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// GetHelperProfile returns a helper profile by user ID.
func (r *Repository) GetHelperProfile(ctx context.Context, userID string) (*HelperProfile, error) {
	var h HelperProfile
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, helper_type, COALESCE(organization,''), verified, verified_at, active, created_at
		FROM helper_profiles WHERE user_id = $1`, userID).
		Scan(&h.ID, &h.UserID, &h.HelperType, &h.Organization, &h.Verified, &h.VerifiedAt, &h.Active, &h.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// ListHelpers returns helper profiles, optionally filtered by verification.
func (r *Repository) ListHelpers(ctx context.Context, verifiedOnly bool) ([]HelperProfile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, helper_type, COALESCE(organization,''), verified, verified_at, active, created_at
		FROM helper_profiles
		WHERE ($1 = FALSE OR verified = TRUE)
		ORDER BY created_at DESC`, verifiedOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HelperProfile
	for rows.Next() {
		var h HelperProfile
		if err := rows.Scan(&h.ID, &h.UserID, &h.HelperType, &h.Organization, &h.Verified, &h.VerifiedAt, &h.Active, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// Verify marks a helper profile verified.
func (r *Repository) Verify(ctx context.Context, helperUserID, adminID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE helper_profiles SET verified = TRUE, verified_at = NOW(), verified_by = $2 WHERE user_id = $1`,
		helperUserID, adminID)
	return err
}

// CreateSession starts a new helper session.
func (r *Repository) CreateSession(ctx context.Context, helperUserID, purpose string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		INSERT INTO helper_sessions (helper_user_id, purpose)
		VALUES ($1, $2)
		RETURNING id, helper_user_id, citizen_user_id, consent_given, COALESCE(consent_text,''), consent_at,
		          purpose, status, conversation_id, profile_id, created_at, completed_at`,
		helperUserID, purpose).
		Scan(&s.ID, &s.HelperUserID, &s.CitizenUserID, &s.ConsentGiven, &s.ConsentText, &s.ConsentAt,
			&s.Purpose, &s.Status, &s.ConversationID, &s.ProfileID, &s.CreatedAt, &s.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSession returns a session owned by the helper.
func (r *Repository) GetSession(ctx context.Context, id, helperUserID string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		SELECT id, helper_user_id, citizen_user_id, consent_given, COALESCE(consent_text,''), consent_at,
		       purpose, status, conversation_id, profile_id, created_at, completed_at
		FROM helper_sessions WHERE id = $1 AND helper_user_id = $2`, id, helperUserID).
		Scan(&s.ID, &s.HelperUserID, &s.CitizenUserID, &s.ConsentGiven, &s.ConsentText, &s.ConsentAt,
			&s.Purpose, &s.Status, &s.ConversationID, &s.ProfileID, &s.CreatedAt, &s.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// RecordConsent stores the consent statement and timestamp.
func (r *Repository) RecordConsent(ctx context.Context, id, consentText string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE helper_sessions SET consent_given = TRUE, consent_text = $2, consent_at = NOW() WHERE id = $1`,
		id, consentText)
	return err
}

// LinkSession attaches a conversation and profile to a session.
func (r *Repository) LinkSession(ctx context.Context, id string, conversationID, profileID *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE helper_sessions SET conversation_id = $2, profile_id = $3 WHERE id = $1`, id, conversationID, profileID)
	return err
}

// EndSession completes a session.
func (r *Repository) EndSession(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE helper_sessions SET status = 'completed', completed_at = NOW() WHERE id = $1`, id)
	return err
}

// ListToday returns today's sessions for a helper.
func (r *Repository) ListToday(ctx context.Context, helperUserID string) ([]Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, helper_user_id, citizen_user_id, consent_given, COALESCE(consent_text,''), consent_at,
		       purpose, status, conversation_id, profile_id, created_at, completed_at
		FROM helper_sessions
		WHERE helper_user_id = $1 AND created_at >= date_trunc('day', NOW())
		ORDER BY created_at DESC`, helperUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.HelperUserID, &s.CitizenUserID, &s.ConsentGiven, &s.ConsentText, &s.ConsentAt,
			&s.Purpose, &s.Status, &s.ConversationID, &s.ProfileID, &s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CountToday counts a helper's sessions started today.
func (r *Repository) CountToday(ctx context.Context, helperUserID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM helper_sessions
		WHERE helper_user_id = $1 AND created_at >= date_trunc('day', NOW())`, helperUserID).Scan(&n)
	return n, err
}
