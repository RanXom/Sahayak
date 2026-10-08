package application

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status values for an application.
const (
	StatusDraft     = "draft"
	StatusReady     = "ready"
	StatusSubmitted = "submitted"
	StatusTracking  = "tracking"
	StatusCompleted = "completed"
	StatusRejected  = "rejected"
)

// Application is a citizen's application for a scheme.
type Application struct {
	ID             string         `json:"id"`
	UserID         string         `json:"user_id"`
	ProfileID      string         `json:"profile_id"`
	SchemeID       string         `json:"scheme_id"`
	Status         string         `json:"status"`
	FormData       map[string]any `json:"form_data"`
	PDFPath        *string        `json:"-"`
	HasPDF         bool           `json:"has_pdf"`
	SubmissionRef  string         `json:"submission_ref,omitempty"`
	SubmittedAt    *time.Time     `json:"submitted_at,omitempty"`
	ExternalStatus string         `json:"external_status,omitempty"`
	Notes          string         `json:"notes,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Repository provides persistence for applications.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an application Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when an application does not exist.
var ErrNotFound = errors.New("not found")

func scanApp(row pgx.Row) (*Application, error) {
	var a Application
	var pdfPath *string
	err := row.Scan(&a.ID, &a.UserID, &a.ProfileID, &a.SchemeID, &a.Status, &a.FormData, &pdfPath,
		&a.SubmissionRef, &a.SubmittedAt, &a.ExternalStatus, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.PDFPath = pdfPath
	a.HasPDF = pdfPath != nil && *pdfPath != ""
	return &a, nil
}

const appColumns = `id, user_id, profile_id, scheme_id, status, form_data, pdf_path,
	COALESCE(submission_ref, ''), submitted_at, COALESCE(external_status, ''), COALESCE(notes, ''), created_at, updated_at`

// Create inserts a new application.
func (r *Repository) Create(ctx context.Context, userID, profileID, schemeID string, formData map[string]any) (*Application, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO applications (user_id, profile_id, scheme_id, form_data)
		VALUES ($1, $2, $3, $4)
		RETURNING `+appColumns, userID, profileID, schemeID, formData)
	return scanApp(row)
}

// Get returns an application owned by the user.
func (r *Repository) Get(ctx context.Context, id, userID string) (*Application, error) {
	return scanApp(r.pool.QueryRow(ctx, `SELECT `+appColumns+` FROM applications WHERE id = $1 AND user_id = $2`, id, userID))
}

// GetByID returns an application regardless of owner (used internally).
func (r *Repository) GetByID(ctx context.Context, id string) (*Application, error) {
	return scanApp(r.pool.QueryRow(ctx, `SELECT `+appColumns+` FROM applications WHERE id = $1`, id))
}

// ListByUser returns a user's applications.
func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*Application, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+appColumns+` FROM applications WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Application
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// UpdateFormData updates form data and status.
func (r *Repository) UpdateFormData(ctx context.Context, id string, formData map[string]any, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE applications SET form_data = $2, status = $3, updated_at = NOW() WHERE id = $1`,
		id, formData, status)
	return err
}

// SetPDFPath stores the generated PDF path and marks the application ready.
func (r *Repository) SetPDFPath(ctx context.Context, id, path string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE applications SET pdf_path = $2, status = 'ready', updated_at = NOW() WHERE id = $1`, id, path)
	return err
}

// SetSubmission records a submission reference and status.
func (r *Repository) SetSubmission(ctx context.Context, id, ref, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE applications
		SET submission_ref = NULLIF($2,''), status = $3, submitted_at = NOW(), updated_at = NOW()
		WHERE id = $1`, id, ref, status)
	return err
}

// SetExternalStatus updates the tracked external status.
func (r *Repository) SetExternalStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE applications SET external_status = $2, last_status_check = NOW(), updated_at = NOW() WHERE id = $1`, id, status)
	return err
}
