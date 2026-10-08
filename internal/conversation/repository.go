package conversation

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status values for a conversation.
const (
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusAbandoned = "abandoned"
)

// Conversation is an intake session.
type Conversation struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	ProfileID *string        `json:"profile_id,omitempty"`
	Language  string         `json:"language"`
	Status    string         `json:"status"`
	Context   map[string]any `json:"context"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Messages  []Message      `json:"messages,omitempty"`
}

// Message is a single conversation turn.
type Message struct {
	ID             string         `json:"id"`
	ConversationID string         `json:"conversation_id"`
	Role           string         `json:"role"`
	ContentText    string         `json:"content_text"`
	ContentLang    string         `json:"content_lang,omitempty"`
	FactsExtracted map[string]any `json:"facts_extracted,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// Repository provides persistence for conversations.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a conversation Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a conversation does not exist.
var ErrNotFound = errors.New("not found")

// Create inserts a new conversation.
func (r *Repository) Create(ctx context.Context, userID string, profileID *string, lang string) (*Conversation, error) {
	var c Conversation
	err := r.pool.QueryRow(ctx, `
		INSERT INTO conversations (user_id, profile_id, language)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, profile_id, language, status, context, created_at, updated_at`,
		userID, profileID, lang).
		Scan(&c.ID, &c.UserID, &c.ProfileID, &c.Language, &c.Status, &c.Context, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetActive returns the user's active conversation, if any.
func (r *Repository) GetActive(ctx context.Context, userID string) (*Conversation, error) {
	var c Conversation
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, profile_id, language, status, context, created_at, updated_at
		FROM conversations WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC LIMIT 1`, userID).
		Scan(&c.ID, &c.UserID, &c.ProfileID, &c.Language, &c.Status, &c.Context, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Get returns a conversation owned by the user, with its messages.
func (r *Repository) Get(ctx context.Context, id, userID string) (*Conversation, error) {
	var c Conversation
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, profile_id, language, status, context, created_at, updated_at
		FROM conversations WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&c.ID, &c.UserID, &c.ProfileID, &c.Language, &c.Status, &c.Context, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	msgs, err := r.ListMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Messages = msgs
	return &c, nil
}

// ListMessages returns messages for a conversation in order.
func (r *Repository) ListMessages(ctx context.Context, convID string) ([]Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, role, content_text, COALESCE(content_lang, ''), facts_extracted, created_at
		FROM conversation_messages WHERE conversation_id = $1 ORDER BY created_at`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.ContentText, &m.ContentLang, &m.FactsExtracted, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// AddMessage appends a message.
func (r *Repository) AddMessage(ctx context.Context, convID, role, text, lang string, facts map[string]any) (*Message, error) {
	var m Message
	var factsArg any
	if len(facts) > 0 {
		factsArg = facts
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO conversation_messages (conversation_id, role, content_text, content_lang, facts_extracted)
		VALUES ($1, $2, $3, NULLIF($4,''), $5)
		RETURNING id, conversation_id, role, content_text, COALESCE(content_lang, ''), facts_extracted, created_at`,
		convID, role, text, lang, factsArg).
		Scan(&m.ID, &m.ConversationID, &m.Role, &m.ContentText, &m.ContentLang, &m.FactsExtracted, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateContext stores the conversation context and optionally the language.
func (r *Repository) UpdateContext(ctx context.Context, id string, convCtx map[string]any, lang string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE conversations SET context = $2, language = COALESCE(NULLIF($3,''), language), updated_at = NOW()
		WHERE id = $1`, id, convCtx, lang)
	return err
}

// SetStatus updates the conversation status.
func (r *Repository) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE conversations SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

// SetProfile links a profile to the conversation.
func (r *Repository) SetProfile(ctx context.Context, id, profileID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE conversations SET profile_id = $2, updated_at = NOW() WHERE id = $1`, id, profileID)
	return err
}
