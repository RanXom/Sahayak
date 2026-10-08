package location

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/domain"
)

// Repository provides persistence for the location module.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a location Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ErrNotFound is returned when a state or district does not exist.
var ErrNotFound = errors.New("not found")

// ListStates returns all active states with their coverage status.
func (r *Repository) ListStates(ctx context.Context) ([]domain.StateWithCoverage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.code, s.name_en, s.name_hi, COALESCE(s.name_local, ''), s.is_ut,
		       COALESCE(c.level, 'none'), COALESCE(c.scheme_count, 0),
		       COALESCE(c.last_updated, NOW()), COALESCE(c.notes, '')
		FROM states s
		LEFT JOIN coverage_status c ON c.state_id = s.id
		WHERE s.active = TRUE
		ORDER BY s.name_en`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.StateWithCoverage
	for rows.Next() {
		var sc domain.StateWithCoverage
		if err := rows.Scan(&sc.ID, &sc.Code, &sc.NameEN, &sc.NameHI, &sc.NameLocal, &sc.IsUT,
			&sc.Coverage.Level, &sc.Coverage.SchemeCount, &sc.Coverage.LastUpdated, &sc.Coverage.Note); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// GetState returns a single state.
func (r *Repository) GetState(ctx context.Context, id int) (*domain.State, error) {
	var s domain.State
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name_en, name_hi, COALESCE(name_local, ''), is_ut
		FROM states WHERE id = $1 AND active = TRUE`, id).
		Scan(&s.ID, &s.Code, &s.NameEN, &s.NameHI, &s.NameLocal, &s.IsUT)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// StateCode returns the code for a state ID, or "" if unknown.
func (r *Repository) StateCode(ctx context.Context, id int) (string, error) {
	var code string
	err := r.pool.QueryRow(ctx, `SELECT code FROM states WHERE id = $1`, id).Scan(&code)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return code, err
}

// FindStateByName resolves a state by its English or Hindi name (case-insensitive).
func (r *Repository) FindStateByName(ctx context.Context, name string) (*domain.State, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNotFound
	}
	var s domain.State
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name_en, name_hi, COALESCE(name_local, ''), is_ut
		FROM states
		WHERE active = TRUE AND (LOWER(name_en) = LOWER($1) OR name_hi = $1 OR LOWER(COALESCE(name_local,'')) = LOWER($1))
		LIMIT 1`, name).
		Scan(&s.ID, &s.Code, &s.NameEN, &s.NameHI, &s.NameLocal, &s.IsUT)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetCoverage returns coverage status for a state.
func (r *Repository) GetCoverage(ctx context.Context, stateID int) (*domain.Coverage, error) {
	var c domain.Coverage
	err := r.pool.QueryRow(ctx, `
		SELECT level, scheme_count, last_updated, COALESCE(notes, '')
		FROM coverage_status WHERE state_id = $1`, stateID).
		Scan(&c.Level, &c.SchemeCount, &c.LastUpdated, &c.Note)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetDistrict returns a single district.
func (r *Repository) GetDistrict(ctx context.Context, id int) (*domain.District, error) {
	var d domain.District
	err := r.pool.QueryRow(ctx, `
		SELECT id, state_id, name_en, name_hi, COALESCE(name_local, ''), COALESCE(census_code, '')
		FROM districts WHERE id = $1 AND active = TRUE`, id).
		Scan(&d.ID, &d.StateID, &d.NameEN, &d.NameHI, &d.NameLocal, &d.CensusCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDistricts returns active districts for a state.
func (r *Repository) ListDistricts(ctx context.Context, stateID int) ([]domain.District, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, state_id, name_en, name_hi, COALESCE(name_local, ''), COALESCE(census_code, '')
		FROM districts WHERE state_id = $1 AND active = TRUE
		ORDER BY name_en`, stateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.District
	for rows.Next() {
		var d domain.District
		if err := rows.Scan(&d.ID, &d.StateID, &d.NameEN, &d.NameHI, &d.NameLocal, &d.CensusCode); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
