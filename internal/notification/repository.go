package notification

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Notification statuses.
const (
	StatusPending   = "pending"
	StatusSent      = "sent"
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
)

// Channels.
const (
	ChannelWhatsApp = "whatsapp"
	ChannelSMS      = "sms"
)

// Preference is a user's notification channel preference.
type Preference struct {
	Channel    string     `json:"channel"`
	Enabled    bool       `json:"enabled"`
	OptedInAt  *time.Time `json:"opted_in_at,omitempty"`
	OptedOutAt *time.Time `json:"opted_out_at,omitempty"`
}

// Notification is a queued or sent message.
type Notification struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	ApplicationID *string    `json:"application_id,omitempty"`
	Channel       string     `json:"channel"`
	TemplateKey   string     `json:"template_key"`
	Content       string     `json:"content"`
	Status        string     `json:"status"`
	ProviderRef   string     `json:"provider_ref,omitempty"`
	ScheduledFor  *time.Time `json:"scheduled_for,omitempty"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	Error         string     `json:"error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Repository provides persistence for notifications.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a notification Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// GetPreferences returns a user's notification preferences.
func (r *Repository) GetPreferences(ctx context.Context, userID string) ([]Preference, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT channel, enabled, opted_in_at, opted_out_at
		FROM notification_preferences WHERE user_id = $1 ORDER BY channel`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Preference
	for rows.Next() {
		var p Preference
		if err := rows.Scan(&p.Channel, &p.Enabled, &p.OptedInAt, &p.OptedOutAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertPreference sets a user's channel preference, recording the user's
// phone hash for reference.
func (r *Repository) UpsertPreference(ctx context.Context, userID, channel string, enabled bool) error {
	var optedIn, optedOut any
	if enabled {
		optedIn = time.Now()
	} else {
		optedOut = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notification_preferences (user_id, channel, enabled, phone_hash, opted_in_at, opted_out_at)
		SELECT $1, $2, $3, phone_hash, $4, $5 FROM users WHERE id = $1
		ON CONFLICT (user_id, channel) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			opted_in_at = COALESCE(EXCLUDED.opted_in_at, notification_preferences.opted_in_at),
			opted_out_at = EXCLUDED.opted_out_at`,
		userID, channel, enabled, optedIn, optedOut)
	return err
}

// IsEnabled reports whether a channel is enabled for the user.
func (r *Repository) IsEnabled(ctx context.Context, userID, channel string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `
		SELECT enabled FROM notification_preferences WHERE user_id = $1 AND channel = $2`, userID, channel).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return enabled, err
}

// Create inserts a notification and returns its ID.
func (r *Repository) Create(ctx context.Context, n Notification) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, application_id, channel, template_key, content, scheduled_for)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		n.UserID, n.ApplicationID, n.Channel, n.TemplateKey, n.Content, n.ScheduledFor).Scan(&id)
	return id, err
}

// Get returns a notification by ID.
func (r *Repository) Get(ctx context.Context, id string) (*Notification, error) {
	var n Notification
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, application_id, channel, template_key, content, status,
		       COALESCE(provider_ref, ''), scheduled_for, sent_at, COALESCE(error, ''), created_at
		FROM notifications WHERE id = $1`, id).
		Scan(&n.ID, &n.UserID, &n.ApplicationID, &n.Channel, &n.TemplateKey, &n.Content, &n.Status,
			&n.ProviderRef, &n.ScheduledFor, &n.SentAt, &n.Error, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// ListByUser returns a user's notifications, most recent first.
func (r *Repository) ListByUser(ctx context.Context, userID string, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, application_id, channel, template_key, content, status,
		       COALESCE(provider_ref, ''), scheduled_for, sent_at, COALESCE(error, ''), created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.ApplicationID, &n.Channel, &n.TemplateKey, &n.Content, &n.Status,
			&n.ProviderRef, &n.ScheduledFor, &n.SentAt, &n.Error, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// MarkSent records a successful send.
func (r *Repository) MarkSent(ctx context.Context, id, providerRef string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET status = 'sent', provider_ref = NULLIF($2,''), sent_at = NOW(), error = NULL
		WHERE id = $1`, id, providerRef)
	return err
}

// MarkFailed records a failed send.
func (r *Repository) MarkFailed(ctx context.Context, id, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET status = 'failed', error = $2 WHERE id = $1`, id, errMsg)
	return err
}
