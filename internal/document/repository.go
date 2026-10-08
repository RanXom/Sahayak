package document

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the document module.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a document Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a document type does not exist.
var ErrNotFound = errors.New("not found")

const docTypeColumns = `
	id, name_en, name_hi, description_en, description_hi, category,
	COALESCE(issuing_authority_en, ''), COALESCE(issuing_authority_hi, ''),
	COALESCE(acquisition_url, ''), COALESCE(acquisition_steps_en, ''), COALESCE(acquisition_steps_hi, ''),
	COALESCE(typical_time, ''), COALESCE(typical_fee, ''), last_verified, COALESCE(source_url, '')`

func scanDocumentType(row pgx.Row) (*domain.DocumentType, error) {
	var d domain.DocumentType
	err := row.Scan(&d.ID, &d.NameEN, &d.NameHI, &d.DescriptionEN, &d.DescriptionHI, &d.Category,
		&d.IssuingAuthorityEN, &d.IssuingAuthorityHI, &d.AcquisitionURL,
		&d.AcquisitionStepsEN, &d.AcquisitionStepsHI, &d.TypicalTime, &d.TypicalFee,
		&d.LastVerified, &d.SourceURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDocumentTypes returns all document types.
func (r *Repository) ListDocumentTypes(ctx context.Context) ([]domain.DocumentType, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+docTypeColumns+` FROM document_types ORDER BY category, name_en`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DocumentType
	for rows.Next() {
		d, err := scanDocumentType(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, rows.Err()
}

// GetDocumentType returns a single document type.
func (r *Repository) GetDocumentType(ctx context.Context, id string) (*domain.DocumentType, error) {
	return scanDocumentType(r.pool.QueryRow(ctx, `SELECT `+docTypeColumns+` FROM document_types WHERE id = $1`, id))
}

// ServiceLocation is a physical office where documents can be obtained.
type ServiceLocation struct {
	ID           string   `json:"id"`
	LocationType string   `json:"location_type"`
	Name         string   `json:"name"`
	Address      string   `json:"address,omitempty"`
	DistrictID   *int     `json:"district_id,omitempty"`
	Latitude     *float64 `json:"lat,omitempty"`
	Longitude    *float64 `json:"lng,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Services     []string `json:"services,omitempty"`
}

// ListServiceLocations returns service locations for a district offering a service.
func (r *Repository) ListServiceLocations(ctx context.Context, districtID int, service string) ([]ServiceLocation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, location_type, name, COALESCE(address, ''), district_id, latitude, longitude,
		       COALESCE(phone, ''), services
		FROM service_locations
		WHERE active = TRUE AND district_id = $1
		  AND ($2 = '' OR $2 = ANY(services))
		ORDER BY name
		LIMIT 10`, districtID, service)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ServiceLocation
	for rows.Next() {
		var l ServiceLocation
		if err := rows.Scan(&l.ID, &l.LocationType, &l.Name, &l.Address, &l.DistrictID,
			&l.Latitude, &l.Longitude, &l.Phone, &l.Services); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
