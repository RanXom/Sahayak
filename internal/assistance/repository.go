package assistance

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a state transition is no longer valid
// (e.g. a request already accepted by someone else).
var ErrConflict = errors.New("conflict")

// Repository persists marketplace rows.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// --- requests ---

// CreateRequest inserts a pending assistance request.
func (r *Repository) CreateRequest(ctx context.Context, citizenID string, applicationID, schemeID *string, language, category, details string, shareFacts, shareDocs bool) (*Request, error) {
	var out Request
	err := r.pool.QueryRow(ctx, `
		INSERT INTO assistance_requests
			(citizen_user_id, application_id, scheme_id, language, category, details, share_profile_facts, share_documents)
		VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8)
		RETURNING id, citizen_user_id, application_id::text, scheme_id::text, language, category,
			details, share_profile_facts, share_documents, status, matched_helper_user_id::text,
			created_at, matched_at, closed_at, expires_at`,
		citizenID, strOrEmpty(applicationID), strOrEmpty(schemeID), language, category, details, shareFacts, shareDocs,
	).Scan(&out.ID, &out.CitizenUserID, &out.ApplicationID, &out.SchemeID, &out.Language, &out.Category,
		&out.Details, &out.ShareProfileFacts, &out.ShareDocuments, &out.Status, &out.MatchedHelperID,
		&out.CreatedAt, &out.MatchedAt, &out.ClosedAt, &out.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRequest returns one request with the matched helper's display name.
func (r *Repository) GetRequest(ctx context.Context, id string) (*Request, error) {
	var out Request
	err := r.pool.QueryRow(ctx, `
		SELECT q.id, q.citizen_user_id, q.application_id::text, q.scheme_id::text, q.language,
			q.category, q.details, q.share_profile_facts, q.share_documents, q.status,
			q.matched_helper_user_id::text, q.created_at, q.matched_at, q.closed_at, q.expires_at,
			h.display_name
		FROM assistance_requests q
		LEFT JOIN helper_profiles h ON h.user_id = q.matched_helper_user_id
		WHERE q.id = $1`, id,
	).Scan(&out.ID, &out.CitizenUserID, &out.ApplicationID, &out.SchemeID, &out.Language,
		&out.Category, &out.Details, &out.ShareProfileFacts, &out.ShareDocuments, &out.Status,
		&out.MatchedHelperID, &out.CreatedAt, &out.MatchedAt, &out.ClosedAt, &out.ExpiresAt,
		&out.HelperDisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &out, err
}

// ListCitizenRequests returns a citizen's requests, newest first.
func (r *Repository) ListCitizenRequests(ctx context.Context, citizenID string) ([]Request, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT q.id, q.citizen_user_id, q.application_id::text, q.scheme_id::text, q.language,
			q.category, q.details, q.share_profile_facts, q.share_documents, q.status,
			q.matched_helper_user_id::text, q.created_at, q.matched_at, q.closed_at, q.expires_at,
			h.display_name
		FROM assistance_requests q
		LEFT JOIN helper_profiles h ON h.user_id = q.matched_helper_user_id
		WHERE q.citizen_user_id = $1 ORDER BY q.created_at DESC LIMIT 50`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

// ListPending returns all currently matchable requests (oldest first).
func (r *Repository) ListPending(ctx context.Context) ([]Request, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, citizen_user_id, application_id::text, scheme_id::text, language,
			category, details, share_profile_facts, share_documents, status,
			matched_helper_user_id::text, created_at, matched_at, closed_at, expires_at, NULL
		FROM assistance_requests
		WHERE status = 'pending' AND expires_at > NOW()
		ORDER BY created_at ASC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}

// AcceptRequest atomically assigns a pending request. Zero rows means it is
// gone (accepted/cancelled/expired) — the caller reports a conflict.
func (r *Repository) AcceptRequest(ctx context.Context, id, helperID string) (*Request, error) {
	var out Request
	err := r.pool.QueryRow(ctx, `
		UPDATE assistance_requests
		SET status = 'accepted', matched_helper_user_id = $2, matched_at = NOW()
		WHERE id = $1 AND status = 'pending' AND expires_at > NOW()
		RETURNING id, citizen_user_id, application_id::text, scheme_id::text, language,
			category, details, share_profile_facts, share_documents, status,
			matched_helper_user_id::text, created_at, matched_at, closed_at, expires_at`,
		id, helperID,
	).Scan(&out.ID, &out.CitizenUserID, &out.ApplicationID, &out.SchemeID, &out.Language,
		&out.Category, &out.Details, &out.ShareProfileFacts, &out.ShareDocuments, &out.Status,
		&out.MatchedHelperID, &out.CreatedAt, &out.MatchedAt, &out.ClosedAt, &out.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConflict
	}
	return &out, err
}

// SetRequestStatus moves a request between states only from an allowed prior
// state; zero rows means the transition is invalid.
func (r *Repository) SetRequestStatus(ctx context.Context, id, to string, from []string) (*Request, error) {
	var out Request
	err := r.pool.QueryRow(ctx, `
		UPDATE assistance_requests
		SET status = $2, closed_at = CASE WHEN $2 IN ('completed','cancelled','expired') THEN NOW() ELSE closed_at END
		WHERE id = $1 AND status = ANY($3)
		RETURNING id, citizen_user_id, application_id::text, scheme_id::text, language,
			category, details, share_profile_facts, share_documents, status,
			matched_helper_user_id::text, created_at, matched_at, closed_at, expires_at`,
		id, to, from,
	).Scan(&out.ID, &out.CitizenUserID, &out.ApplicationID, &out.SchemeID, &out.Language,
		&out.Category, &out.Details, &out.ShareProfileFacts, &out.ShareDocuments, &out.Status,
		&out.MatchedHelperID, &out.CreatedAt, &out.MatchedAt, &out.ClosedAt, &out.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConflict
	}
	return &out, err
}

// ExpireOverdue flips lapsed pending requests to expired. Returns the count.
func (r *Repository) ExpireOverdue(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE assistance_requests SET status = 'expired', closed_at = NOW()
		WHERE status = 'pending' AND expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func scanRequests(rows pgx.Rows) ([]Request, error) {
	var out []Request
	for rows.Next() {
		var q Request
		if err := rows.Scan(&q.ID, &q.CitizenUserID, &q.ApplicationID, &q.SchemeID, &q.Language,
			&q.Category, &q.Details, &q.ShareProfileFacts, &q.ShareDocuments, &q.Status,
			&q.MatchedHelperID, &q.CreatedAt, &q.MatchedAt, &q.ClosedAt, &q.ExpiresAt,
			&q.HelperDisplayName); err != nil {
			return nil, err
		}
		q.WaitingSeconds = int64(time.Since(q.CreatedAt).Seconds())
		out = append(out, q)
	}
	if out == nil {
		out = []Request{}
	}
	return out, rows.Err()
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// --- helper profile reads (matching inputs) ---

// HelperMatchProfile carries what matching needs.
type HelperMatchProfile struct {
	Languages   []string
	Categories  []string
	IsAvailable bool
	Verified    bool
	Active      bool
}

// GetHelperMatchProfile loads a helper's matching inputs.
func (r *Repository) GetHelperMatchProfile(ctx context.Context, userID string) (*HelperMatchProfile, error) {
	var p HelperMatchProfile
	var langs, cats []string
	err := r.pool.QueryRow(ctx, `
		SELECT languages, categories, is_available, verified, active
		FROM helper_profiles WHERE user_id = $1`, userID,
	).Scan(&langs, &cats, &p.IsAvailable, &p.Verified, &p.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Languages = langs
	p.Categories = cats
	return &p, nil
}

// HelperHasActiveSession reports whether the helper already holds an
// unfinished assignment (rule: one active session at a time).
func (r *Repository) HelperHasActiveSession(ctx context.Context, helperID string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM helper_sessions
		WHERE helper_user_id = $1 AND status = 'active'`, helperID).Scan(&n)
	return n > 0, err
}

// --- sessions (stored in helper_sessions) ---

// Session is the assigned assistance session.
type Session struct {
	ID            string     `json:"id"`
	RequestID     *string    `json:"request_id,omitempty"`
	HelperUserID  string     `json:"helper_user_id"`
	CitizenUserID *string    `json:"citizen_user_id,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// CreateSession opens the session for an accepted request.
func (r *Repository) CreateSession(ctx context.Context, requestID, helperID, citizenID, purpose string) (*Session, error) {
	var s Session
	id := uuid.NewString()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO helper_sessions (id, helper_user_id, citizen_user_id, purpose, status, request_id)
		VALUES ($1, $2, $3, $4, 'active', $5)
		RETURNING id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at`,
		id, helperID, citizenID, purpose, requestID,
	).Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSession returns a session by id.
func (r *Repository) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		SELECT id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at
		FROM helper_sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

// GetSessionByRequest returns the session for a request, if any.
func (r *Repository) GetSessionByRequest(ctx context.Context, requestID string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		SELECT id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at
		FROM helper_sessions WHERE request_id = $1 ORDER BY created_at DESC LIMIT 1`, requestID,
	).Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

// CitizenSessions lists a citizen's sessions, newest first.
func (r *Repository) CitizenSessions(ctx context.Context, citizenID string) ([]Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at
		FROM helper_sessions WHERE citizen_user_id = $1 ORDER BY created_at DESC LIMIT 50`, citizenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []Session{}
	}
	return out, rows.Err()
}

// EndSession closes a session from active.
func (r *Repository) EndSession(ctx context.Context, id, endedBy, status string) (*Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, `
		UPDATE helper_sessions SET status = $2, completed_at = NOW(), ended_by = $3
		WHERE id = $1 AND status = 'active'
		RETURNING id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at`,
		id, status, endedBy,
	).Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConflict
	}
	return &s, err
}

// --- messages ---

// Message is one chat line in a session.
type Message struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	SenderID   string    `json:"-"`
	SenderRole string    `json:"sender_role"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	Mine       bool      `json:"mine,omitempty"`
}

// AddMessage stores a chat message; duplicate client_msg_id is a no-op that
// returns the original row (idempotent sends).
func (r *Repository) AddMessage(ctx context.Context, sessionID, senderID, senderRole, clientMsgID, body string) (*Message, error) {
	var m Message
	clientMsgID = trimTo(clientMsgID, 64)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO session_messages (session_id, sender_user_id, sender_role, client_msg_id, body)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		ON CONFLICT (session_id, client_msg_id) WHERE client_msg_id IS NOT NULL
		DO UPDATE SET body = EXCLUDED.body
		RETURNING id, session_id, sender_role, body, created_at`,
		sessionID, senderID, senderRole, clientMsgID, body,
	).Scan(&m.ID, &m.SessionID, &m.SenderRole, &m.Body, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMessages returns messages after an optional timestamp cursor.
func (r *Repository) ListMessages(ctx context.Context, sessionID string, since time.Time, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, session_id, sender_user_id::text, sender_role, body, created_at FROM session_messages
		WHERE session_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)
		ORDER BY created_at ASC, id ASC LIMIT $3`, sessionID, nullTime(since), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.SenderID, &m.SenderRole, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []Message{}
	}
	return out, rows.Err()
}

// --- signaling ---

// HelperSessions lists sessions assigned to a helper, newest first.
func (r *Repository) HelperSessions(ctx context.Context, helperID string) ([]Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id::text, helper_user_id, citizen_user_id::text, status, created_at, completed_at
		FROM helper_sessions WHERE helper_user_id = $1 ORDER BY created_at DESC LIMIT 50`, helperID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.RequestID, &s.HelperUserID, &s.CitizenUserID, &s.Status, &s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []Session{}
	}
	return out, rows.Err()
}

// Signal is one WebRTC signaling payload.
type Signal struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	SenderID  string         `json:"-"`
	Kind      string         `json:"kind"`
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
	Mine      bool           `json:"mine,omitempty"`
}

// AddSignal stores a signaling message (offer/answer/ice/hangup/join/leave).
func (r *Repository) AddSignal(ctx context.Context, sessionID, senderID, kind string, payload map[string]any) (*Signal, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var s Signal
	err = r.pool.QueryRow(ctx, `
		INSERT INTO session_signals (session_id, sender_user_id, kind, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, session_id, kind, payload, created_at`,
		sessionID, senderID, kind, raw,
	).Scan(&s.ID, &s.SessionID, &s.Kind, &s.Payload, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSignals returns signaling messages after an optional cursor.
func (r *Repository) ListSignals(ctx context.Context, sessionID string, since time.Time, limit int) ([]Signal, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, session_id, sender_user_id::text, kind, payload, created_at FROM session_signals
		WHERE session_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)
		ORDER BY created_at ASC, id ASC LIMIT $3`, sessionID, nullTime(since), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Signal
	for rows.Next() {
		var s Signal
		if err := rows.Scan(&s.ID, &s.SessionID, &s.SenderID, &s.Kind, &s.Payload, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []Signal{}
	}
	return out, rows.Err()
}

// --- drafts ---

// DraftValue is one shared field value with its provenance.
type DraftValue struct {
	FieldName  string    `json:"field_name"`
	Value      any       `json:"value"`
	ProposedBy string    `json:"proposed_by"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UpsertDraft proposes (or directly sets) a draft value. Last write wins;
// author and timestamp are always visible so nothing is silently lost.
func (r *Repository) UpsertDraft(ctx context.Context, sessionID, field string, value any, by, status string) (*DraftValue, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var d DraftValue
	err = r.pool.QueryRow(ctx, `
		INSERT INTO session_draft_values (session_id, field_name, value, proposed_by, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (session_id, field_name)
		DO UPDATE SET value = EXCLUDED.value, proposed_by = EXCLUDED.proposed_by,
			status = EXCLUDED.status, updated_at = NOW()
		RETURNING field_name, value, proposed_by::text, status, updated_at`,
		sessionID, field, raw, by, status,
	).Scan(&d.FieldName, &d.Value, &d.ProposedBy, &d.Status, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDrafts returns all draft values for a session.
func (r *Repository) ListDrafts(ctx context.Context, sessionID string) ([]DraftValue, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT field_name, value, proposed_by::text, status, updated_at
		FROM session_draft_values WHERE session_id = $1 ORDER BY field_name`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DraftValue
	for rows.Next() {
		var d DraftValue
		if err := rows.Scan(&d.FieldName, &d.Value, &d.ProposedBy, &d.Status, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []DraftValue{}
	}
	return out, rows.Err()
}

// --- consent ---

// Consent records a grant with its scope.
type Consent struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Scope     map[string]any `json:"scope"`
	Valid     bool           `json:"valid"`
	CreatedAt time.Time      `json:"created_at"`
}

// GrantConsent records a consent grant.
func (r *Repository) GrantConsent(ctx context.Context, sessionID, requestID, grantedBy string, scope map[string]any, statement, ip, ua string) (*Consent, error) {
	raw, err := json.Marshal(scope)
	if err != nil {
		return nil, err
	}
	var c Consent
	err = r.pool.QueryRow(ctx, `
		INSERT INTO consent_records (session_id, request_id, granted_by, scope, statement, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, session_id, scope, valid, created_at`,
		sessionID, requestID, grantedBy, raw, statement, ip, ua,
	).Scan(&c.ID, &c.SessionID, &c.Scope, &c.Valid, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ValidConsent reports whether a live consent grant exists for the session.
func (r *Repository) ValidConsent(ctx context.Context, sessionID string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM consent_records WHERE session_id = $1 AND valid = TRUE`, sessionID).Scan(&n)
	return n > 0, err
}

// RevokeConsents invalidates all grants for a session.
func (r *Repository) RevokeConsents(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE consent_records SET valid = FALSE, revoked_at = NOW()
		WHERE session_id = $1 AND valid = TRUE`, sessionID)
	return err
}

func trimTo(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
