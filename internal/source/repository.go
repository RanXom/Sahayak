package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the source registry.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a source Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// ListSources returns all sources.
func (r *Repository) ListSources(ctx context.Context) ([]domain.Source, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, source_type, COALESCE(url, ''), COALESCE(description, ''), is_active,
		       fetch_config, last_fetched_at, COALESCE(last_fetch_status, ''), created_at
		FROM sources ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Source
	for rows.Next() {
		var s domain.Source
		if err := rows.Scan(&s.ID, &s.Name, &s.SourceType, &s.URL, &s.Description, &s.IsActive,
			&s.FetchConfig, &s.LastFetchedAt, &s.LastFetchStatus, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CreateSource inserts a source.
func (r *Repository) CreateSource(ctx context.Context, s domain.Source) (*domain.Source, error) {
	var out domain.Source
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sources (name, source_type, url, description, fetch_config, is_active)
		VALUES ($1,$2,$3,$4,$5,TRUE)
		RETURNING id, name, source_type, COALESCE(url, ''), COALESCE(description, ''), is_active,
		          fetch_config, last_fetched_at, COALESCE(last_fetch_status, ''), created_at`,
		s.Name, s.SourceType, s.URL, s.Description, s.FetchConfig).
		Scan(&out.ID, &out.Name, &out.SourceType, &out.URL, &out.Description, &out.IsActive,
			&out.FetchConfig, &out.LastFetchedAt, &out.LastFetchStatus, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListCandidates returns candidate schemes by status.
func (r *Repository) ListCandidates(ctx context.Context, status string, limit, offset int) ([]domain.CandidateScheme, int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source_id, COALESCE(external_id, ''), raw_data, name, status, content_hash,
		       fetched_at, reviewed_at, reviewed_by, COALESCE(reviewer_notes, ''), promoted_to, created_at
		FROM candidate_schemes
		WHERE ($1 = '' OR status = $1)
		ORDER BY fetched_at DESC
		LIMIT $2 OFFSET $3`, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.CandidateScheme
	for rows.Next() {
		var c domain.CandidateScheme
		if err := rows.Scan(&c.ID, &c.SourceID, &c.ExternalID, &c.RawData, &c.Name, &c.Status, &c.ContentHash,
			&c.FetchedAt, &c.ReviewedAt, &c.ReviewedBy, &c.ReviewerNotes, &c.PromotedTo, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidate_schemes WHERE ($1 = '' OR status = $1)`, status).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// CreateCandidate stores a candidate scheme, deduplicating by content hash.
func (r *Repository) CreateCandidate(ctx context.Context, c domain.CandidateScheme) (*domain.CandidateScheme, error) {
	raw, err := json.Marshal(c.RawData)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])

	var out domain.CandidateScheme
	err = r.pool.QueryRow(ctx, `
		INSERT INTO candidate_schemes (source_id, external_id, raw_data, name, status, content_hash)
		VALUES ($1,$2,$3,$4,'pending',$5)
		ON CONFLICT (source_id, content_hash) DO UPDATE SET fetched_at = NOW()
		RETURNING id, source_id, COALESCE(external_id, ''), raw_data, name, status, content_hash,
		          fetched_at, reviewed_at, reviewed_by, COALESCE(reviewer_notes, ''), promoted_to, created_at`,
		c.SourceID, nullIfEmpty(c.ExternalID), raw, c.Name, hash).
		Scan(&out.ID, &out.SourceID, &out.ExternalID, &out.RawData, &out.Name, &out.Status, &out.ContentHash,
			&out.FetchedAt, &out.ReviewedAt, &out.ReviewedBy, &out.ReviewerNotes, &out.PromotedTo, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCandidate returns a candidate by ID.
func (r *Repository) GetCandidate(ctx context.Context, id string) (*domain.CandidateScheme, error) {
	var c domain.CandidateScheme
	err := r.pool.QueryRow(ctx, `
		SELECT id, source_id, COALESCE(external_id, ''), raw_data, name, status, content_hash,
		       fetched_at, reviewed_at, reviewed_by, COALESCE(reviewer_notes, ''), promoted_to, created_at
		FROM candidate_schemes WHERE id = $1`, id).
		Scan(&c.ID, &c.SourceID, &c.ExternalID, &c.RawData, &c.Name, &c.Status, &c.ContentHash,
			&c.FetchedAt, &c.ReviewedAt, &c.ReviewedBy, &c.ReviewerNotes, &c.PromotedTo, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Review marks a candidate accepted or rejected.
func (r *Repository) Review(ctx context.Context, id, status, reviewerID, notes string, promotedTo *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE candidate_schemes
		SET status = $2, reviewer_notes = NULLIF($3,''), reviewed_at = NOW(), reviewed_by = $4, promoted_to = $5
		WHERE id = $1`, id, status, notes, reviewerID, promotedTo)
	return err
}

// CreateDraftSchemeFromCandidate creates a minimal draft scheme linked to a candidate's source.
func (r *Repository) CreateDraftSchemeFromCandidate(ctx context.Context, name, slug, sourceID string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO schemes (slug, status, name_en, name_hi, benefit_type, scope, source_id)
		VALUES ($1, 'draft', $2, $2, 'other', 'central', $3)
		ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
		RETURNING id`, slug, name, sourceID).Scan(&id)
	return id, err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
