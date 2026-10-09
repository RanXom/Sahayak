package scheme

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the scheme module.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a scheme Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a scheme does not exist.
var ErrNotFound = errors.New("not found")

const schemeColumns = `
	id, slug, status, name_en, name_hi, COALESCE(description_en, ''), COALESCE(description_hi, ''),
	benefit_type, COALESCE(benefit_amount, ''), COALESCE(benefit_description_en, ''), COALESCE(benefit_description_hi, ''),
	scope, state_id, COALESCE(ministry, ''), COALESCE(department, ''),
	COALESCE(application_url, ''), COALESCE(official_url, ''), deadline, is_active, source_id,
	verified_at, verified_by, published_at, created_at, updated_at`

func scanScheme(row pgx.Row) (*domain.Scheme, error) {
	var s domain.Scheme
	err := row.Scan(
		&s.ID, &s.Slug, &s.Status, &s.NameEN, &s.NameHI, &s.DescriptionEN, &s.DescriptionHI,
		&s.BenefitType, &s.BenefitAmount, &s.BenefitDescriptionEN, &s.BenefitDescriptionHI,
		&s.Scope, &s.StateID, &s.Ministry, &s.Department,
		&s.ApplicationURL, &s.OfficialURL, &s.Deadline, &s.IsActive, &s.SourceID,
		&s.VerifiedAt, &s.VerifiedBy, &s.PublishedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByID returns a scheme with its rules, documents and fields.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Scheme, error) {
	s, err := scanScheme(r.pool.QueryRow(ctx, `SELECT `+schemeColumns+` FROM schemes WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return r.attach(ctx, s)
}

// GetPublishedByID returns a published scheme by ID.
func (r *Repository) GetPublishedByID(ctx context.Context, id string) (*domain.Scheme, error) {
	s, err := scanScheme(r.pool.QueryRow(ctx, `SELECT `+schemeColumns+` FROM schemes WHERE id = $1 AND status = 'published'`, id))
	if err != nil {
		return nil, err
	}
	return r.attach(ctx, s)
}

func (r *Repository) attach(ctx context.Context, s *domain.Scheme) (*domain.Scheme, error) {
	rules, err := r.ListRules(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	docs, err := r.ListRequiredDocuments(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	fields, err := r.ListApplicationFields(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.Rules, s.RequiredDocuments, s.ApplicationFields = rules, docs, fields
	return s, nil
}

// ListPublished returns published, active schemes, optionally scoped to a state.
// Central schemes are always included; state schemes are included when stateID matches.
func (r *Repository) ListPublished(ctx context.Context, stateID *int) ([]*domain.Scheme, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+schemeColumns+`
		FROM schemes
		WHERE status = 'published' AND is_active = TRUE
		  AND (scope = 'central' OR state_id = $1)
		ORDER BY name_en`, stateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Scheme
	for rows.Next() {
		s, err := scanScheme(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Attach rules for evaluation.
	for _, s := range out {
		if s.Rules, err = r.ListRules(ctx, s.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ListPublishedBrief returns published, active schemes without rules or other
// children. It backs the public guest-browsing endpoint: the payload must stay
// free of evaluation internals.
func (r *Repository) ListPublishedBrief(ctx context.Context, stateID *int) ([]*domain.Scheme, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+schemeColumns+`
		FROM schemes
		WHERE status = 'published' AND is_active = TRUE
		  AND (scope = 'central' OR state_id = $1)
		ORDER BY name_en`, stateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Scheme
	for rows.Next() {
		s, err := scanScheme(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAdmin returns schemes of any status with pagination.
func (r *Repository) ListAdmin(ctx context.Context, status string, limit, offset int) ([]*domain.Scheme, int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+schemeColumns+`
		FROM schemes
		WHERE ($1 = '' OR status = $1)
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.Scheme
	for rows.Next() {
		s, err := scanScheme(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM schemes WHERE ($1 = '' OR status = $1)`, status).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// Create inserts a new scheme and returns it.
func (r *Repository) Create(ctx context.Context, s *domain.Scheme) (*domain.Scheme, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO schemes (
			slug, status, name_en, name_hi, description_en, description_hi,
			benefit_type, benefit_amount, benefit_description_en, benefit_description_hi,
			scope, state_id, ministry, department, application_url, official_url, deadline,
			source_id, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,TRUE)
		RETURNING `+schemeColumns,
		s.Slug, defaultStatus(s.Status), s.NameEN, s.NameHI, s.DescriptionEN, s.DescriptionHI,
		s.BenefitType, s.BenefitAmount, s.BenefitDescriptionEN, s.BenefitDescriptionHI,
		s.Scope, s.StateID, s.Ministry, s.Department, s.ApplicationURL, s.OfficialURL, s.Deadline,
		s.SourceID)
	return scanScheme(row)
}

// Update modifies an existing scheme's editable fields.
func (r *Repository) Update(ctx context.Context, s *domain.Scheme) (*domain.Scheme, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE schemes SET
			name_en = $2, name_hi = $3, description_en = $4, description_hi = $5,
			benefit_type = $6, benefit_amount = $7, benefit_description_en = $8, benefit_description_hi = $9,
			scope = $10, state_id = $11, ministry = $12, department = $13,
			application_url = $14, official_url = $15, deadline = $16, source_id = $17,
			is_active = $18, updated_at = NOW()
		WHERE id = $1
		RETURNING `+schemeColumns,
		s.ID, s.NameEN, s.NameHI, s.DescriptionEN, s.DescriptionHI,
		s.BenefitType, s.BenefitAmount, s.BenefitDescriptionEN, s.BenefitDescriptionHI,
		s.Scope, s.StateID, s.Ministry, s.Department,
		s.ApplicationURL, s.OfficialURL, s.Deadline, s.SourceID, s.IsActive)
	return scanScheme(row)
}

// SetStatus transitions a scheme to a new status.
func (r *Repository) SetStatus(ctx context.Context, id, status string) error {
	switch status {
	case domain.SchemePublished:
		_, err := r.pool.Exec(ctx, `
			UPDATE schemes SET status = $2, published_at = NOW(), updated_at = NOW() WHERE id = $1`, id, status)
		return err
	case domain.SchemeVerified:
		_, err := r.pool.Exec(ctx, `
			UPDATE schemes SET status = $2, verified_at = NOW(), updated_at = NOW() WHERE id = $1`, id, status)
		return err
	default:
		_, err := r.pool.Exec(ctx, `UPDATE schemes SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
		return err
	}
}

// SchemeURL pairs a scheme ID with its application URL for link checking.
type SchemeURL struct {
	ID  string
	URL string
}

// ListApplicationURLs returns published schemes that have an application URL.
func (r *Repository) ListApplicationURLs(ctx context.Context) ([]SchemeURL, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, application_url FROM schemes
		WHERE status = 'published' AND is_active = TRUE
		  AND application_url IS NOT NULL AND application_url <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SchemeURL
	for rows.Next() {
		var u SchemeURL
		if err := rows.Scan(&u.ID, &u.URL); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// ListRules returns eligibility rules for a scheme.
func (r *Repository) ListRules(ctx context.Context, schemeID string) ([]domain.EligibilityRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, scheme_id, rule_name, expression, COALESCE(description_en, ''), COALESCE(description_hi, ''), is_mandatory, sort_order
		FROM scheme_eligibility_rules WHERE scheme_id = $1 ORDER BY sort_order, rule_name`, schemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []domain.EligibilityRule{}
	for rows.Next() {
		var rule domain.EligibilityRule
		if err := rows.Scan(&rule.ID, &rule.SchemeID, &rule.RuleName, &rule.Expression,
			&rule.DescriptionEN, &rule.DescriptionHI, &rule.IsMandatory, &rule.SortOrder); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// ListRequiredDocuments returns required documents for a scheme.
func (r *Repository) ListRequiredDocuments(ctx context.Context, schemeID string) ([]domain.SchemeRequiredDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, scheme_id, document_type, document_name_en, document_name_hi, is_mandatory, sort_order
		FROM scheme_required_documents WHERE scheme_id = $1 ORDER BY sort_order`, schemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []domain.SchemeRequiredDocument{}
	for rows.Next() {
		var d domain.SchemeRequiredDocument
		if err := rows.Scan(&d.ID, &d.SchemeID, &d.DocumentType, &d.DocumentNameEN, &d.DocumentNameHI, &d.IsMandatory, &d.SortOrder); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// ListApplicationFields returns application form fields for a scheme.
func (r *Repository) ListApplicationFields(ctx context.Context, schemeID string) ([]domain.ApplicationField, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, scheme_id, field_name, field_label_en, field_label_hi, field_type,
		       COALESCE(profile_mapping, ''), is_required, sort_order, options
		FROM scheme_application_fields WHERE scheme_id = $1 ORDER BY sort_order`, schemeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := []domain.ApplicationField{}
	for rows.Next() {
		var f domain.ApplicationField
		if err := rows.Scan(&f.ID, &f.SchemeID, &f.FieldName, &f.FieldLabelEN, &f.FieldLabelHI, &f.FieldType,
			&f.ProfileMapping, &f.IsRequired, &f.SortOrder, &f.Options); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

// ReplaceRules replaces all eligibility rules for a scheme in a transaction.
func (r *Repository) ReplaceRules(ctx context.Context, schemeID string, rules []domain.EligibilityRule) error {
	return r.replaceChildren(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM scheme_eligibility_rules WHERE scheme_id = $1`, schemeID); err != nil {
			return err
		}
		for i, rule := range rules {
			if _, err := tx.Exec(ctx, `
				INSERT INTO scheme_eligibility_rules (scheme_id, rule_name, expression, description_en, description_hi, is_mandatory, sort_order)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				schemeID, rule.RuleName, rule.Expression, rule.DescriptionEN, rule.DescriptionHI, rule.IsMandatory, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceRequiredDocuments replaces required documents for a scheme.
func (r *Repository) ReplaceRequiredDocuments(ctx context.Context, schemeID string, docs []domain.SchemeRequiredDocument) error {
	return r.replaceChildren(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM scheme_required_documents WHERE scheme_id = $1`, schemeID); err != nil {
			return err
		}
		for i, d := range docs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO scheme_required_documents (scheme_id, document_type, document_name_en, document_name_hi, is_mandatory, sort_order)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				schemeID, d.DocumentType, d.DocumentNameEN, d.DocumentNameHI, d.IsMandatory, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReplaceApplicationFields replaces application fields for a scheme.
func (r *Repository) ReplaceApplicationFields(ctx context.Context, schemeID string, fields []domain.ApplicationField) error {
	return r.replaceChildren(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM scheme_application_fields WHERE scheme_id = $1`, schemeID); err != nil {
			return err
		}
		for i, f := range fields {
			if _, err := tx.Exec(ctx, `
				INSERT INTO scheme_application_fields (scheme_id, field_name, field_label_en, field_label_hi, field_type, profile_mapping, is_required, sort_order, options)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
				schemeID, f.FieldName, f.FieldLabelEN, f.FieldLabelHI, f.FieldType, nullIfEmpty(f.ProfileMapping), f.IsRequired, i, f.Options); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) replaceChildren(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func defaultStatus(s string) string {
	if s == "" {
		return domain.SchemeDraft
	}
	return s
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
