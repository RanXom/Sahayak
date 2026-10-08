package profile

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the profile module.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a profile Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a profile does not exist.
var ErrNotFound = errors.New("not found")

// GetByUserID returns a profile with its documents, or ErrNotFound.
func (r *Repository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	var p domain.Profile
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, state_id, district_id, facts, completed, created_at, updated_at
		FROM profiles WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`, userID).
		Scan(&p.ID, &p.UserID, &p.StateID, &p.DistrictID, &p.Facts, &p.Completed, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	docs, err := r.GetDocuments(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Documents = docs
	return &p, nil
}

// GetByID returns a profile by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Profile, error) {
	var p domain.Profile
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, state_id, district_id, facts, completed, created_at, updated_at
		FROM profiles WHERE id = $1`, id).
		Scan(&p.ID, &p.UserID, &p.StateID, &p.DistrictID, &p.Facts, &p.Completed, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	docs, err := r.GetDocuments(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Documents = docs
	return &p, nil
}

// Create inserts a new profile for a user.
func (r *Repository) Create(ctx context.Context, userID string, stateID, districtID *int) (*domain.Profile, error) {
	var p domain.Profile
	err := r.pool.QueryRow(ctx, `
		INSERT INTO profiles (user_id, state_id, district_id)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, state_id, district_id, facts, completed, created_at, updated_at`,
		userID, stateID, districtID).
		Scan(&p.ID, &p.UserID, &p.StateID, &p.DistrictID, &p.Facts, &p.Completed, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Documents = []domain.ProfileDocument{}
	return &p, nil
}

// UpdateFacts merges new facts into the profile and updates location/completion.
func (r *Repository) UpdateFacts(ctx context.Context, profileID string, facts map[string]any, stateID, districtID *int, completed bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profiles
		SET facts = facts || $2::jsonb,
		    state_id = COALESCE($3, state_id),
		    district_id = COALESCE($4, district_id),
		    completed = $5,
		    updated_at = NOW()
		WHERE id = $1`, profileID, facts, stateID, districtID, completed)
	return err
}

// ReplaceFacts replaces the entire facts object.
func (r *Repository) ReplaceFacts(ctx context.Context, profileID string, facts map[string]any, stateID, districtID *int, completed bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE profiles
		SET facts = $2,
		    state_id = $3,
		    district_id = $4,
		    completed = $5,
		    updated_at = NOW()
		WHERE id = $1`, profileID, facts, stateID, districtID, completed)
	return err
}

// GetDocuments returns the document inventory for a profile.
func (r *Repository) GetDocuments(ctx context.Context, profileID string) ([]domain.ProfileDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT document_type, has_document, COALESCE(notes, '')
		FROM profile_documents WHERE profile_id = $1 ORDER BY document_type`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []domain.ProfileDocument{}
	for rows.Next() {
		var d domain.ProfileDocument
		if err := rows.Scan(&d.DocumentType, &d.HasDocument, &d.Notes); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// UpsertDocument inserts or updates a document inventory entry.
func (r *Repository) UpsertDocument(ctx context.Context, profileID, docType string, has bool, notes string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO profile_documents (profile_id, document_type, has_document, notes, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), NOW())
		ON CONFLICT (profile_id, document_type) DO UPDATE SET
			has_document = EXCLUDED.has_document,
			notes = EXCLUDED.notes,
			updated_at = NOW()`, profileID, docType, has, notes)
	return err
}
