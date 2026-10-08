package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the auth module.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an auth Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// OTPChallenge is a stored OTP challenge.
type OTPChallenge struct {
	ID             string
	PhoneHash      string
	PhoneLast4     string
	PhoneEncrypted string
	OTPHash        string
	Purpose        string
	Attempts       int
	MaxAttempts    int
	ExpiresAt      time.Time
	ConsumedAt     *time.Time
}

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// FindUserByPhoneHash returns a user by hashed phone, or ErrNotFound.
func (r *Repository) FindUserByPhoneHash(ctx context.Context, phoneHash string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, phone_last4, role, preferred_lang, created_at, updated_at, last_login_at
		FROM users WHERE phone_hash = $1`, phoneHash)
	return scanUser(row)
}

// FindUserByID returns a user by ID.
func (r *Repository) FindUserByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, phone_last4, role, preferred_lang, created_at, updated_at, last_login_at
		FROM users WHERE id = $1`, id)
	return scanUser(row)
}

// CreateUser inserts a new user.
func (r *Repository) CreateUser(ctx context.Context, phoneHash, phoneLast4, phoneEncrypted, lang, role string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (phone_hash, phone_last4, phone_encrypted, preferred_lang, role)
		VALUES ($1, $2, NULLIF($3,''), $4, $5)
		RETURNING id, phone_last4, role, preferred_lang, created_at, updated_at, last_login_at`,
		phoneHash, phoneLast4, phoneEncrypted, lang, role)
	return scanUser(row)
}

// EncryptedPhone returns the encrypted phone number for a user.
func (r *Repository) EncryptedPhone(ctx context.Context, userID string) (string, error) {
	var enc *string
	err := r.pool.QueryRow(ctx, `SELECT phone_encrypted FROM users WHERE id = $1`, userID).Scan(&enc)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if enc == nil {
		return "", nil
	}
	return *enc, nil
}

// TouchLastLogin updates last_login_at.
func (r *Repository) TouchLastLogin(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`, userID)
	return err
}

// UpdatePreferredLang updates the user's preferred language.
func (r *Repository) UpdatePreferredLang(ctx context.Context, userID, lang string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET preferred_lang = $2, updated_at = NOW() WHERE id = $1`, userID, lang)
	return err
}

// CreateOTPChallenge stores a new OTP challenge.
func (r *Repository) CreateOTPChallenge(ctx context.Context, phoneHash, phoneLast4, phoneEncrypted, otpHash, purpose string, expiresAt time.Time) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO otp_challenges (phone_hash, phone_last4, phone_encrypted, otp_hash, purpose, expires_at)
		VALUES ($1, $2, NULLIF($3,''), $4, $5, $6) RETURNING id`,
		phoneHash, phoneLast4, phoneEncrypted, otpHash, purpose, expiresAt).Scan(&id)
	return id, err
}

// GetOTPChallenge fetches an OTP challenge by ID.
func (r *Repository) GetOTPChallenge(ctx context.Context, id string) (*OTPChallenge, error) {
	var c OTPChallenge
	var enc *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, phone_hash, phone_last4, phone_encrypted, otp_hash, purpose, attempts, max_attempts, expires_at, consumed_at
		FROM otp_challenges WHERE id = $1`, id).
		Scan(&c.ID, &c.PhoneHash, &c.PhoneLast4, &enc, &c.OTPHash, &c.Purpose, &c.Attempts, &c.MaxAttempts, &c.ExpiresAt, &c.ConsumedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if enc != nil {
		c.PhoneEncrypted = *enc
	}
	return &c, nil
}

// IncrementOTPAttempts increments the attempt counter.
func (r *Repository) IncrementOTPAttempts(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE otp_challenges SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// ConsumeOTPChallenge marks a challenge consumed.
func (r *Repository) ConsumeOTPChallenge(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE otp_challenges SET consumed_at = NOW() WHERE id = $1`, id)
	return err
}

// CountRecentOTPRequests counts OTP requests for a phone since the given time.
func (r *Repository) CountRecentOTPRequests(ctx context.Context, phoneHash string, since time.Time) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM otp_challenges WHERE phone_hash = $1 AND created_at >= $2`,
		phoneHash, since).Scan(&n)
	return n, err
}

// StoreRefreshToken persists a hashed refresh token.
func (r *Repository) StoreRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return err
}

// GetRefreshToken returns the user ID and metadata for a refresh token hash.
func (r *Repository) GetRefreshToken(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, revokedAt *time.Time, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&userID, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, nil, ErrNotFound
	}
	return userID, expiresAt, revokedAt, err
}

// RevokeRefreshToken revokes a single refresh token.
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`, tokenHash)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.PhoneLast4, &u.Role, &u.PreferredLang, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
