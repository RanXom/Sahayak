package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists audit entries.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an audit Repository.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Insert writes an audit entry.
func (r *Repository) Insert(ctx context.Context, e Entry) error {
	var details any
	if len(e.Details) > 0 {
		details = e.Details
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_log (actor_id, actor_role, action, resource_type, resource_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ActorID, e.ActorRole, e.Action, e.ResourceType, e.ResourceID, details, e.IPAddress, e.UserAgent)
	return err
}

// Query returns audit entries, most recent first.
func (r *Repository) Query(ctx context.Context, action string, limit, offset int) ([]Entry, int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT actor_id, actor_role, action, resource_type, resource_id, details, ip_address, user_agent
		FROM audit_log
		WHERE ($1 = '' OR action = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, action, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ActorID, &e.ActorRole, &e.Action, &e.ResourceType, &e.ResourceID, &e.Details, &e.IPAddress, &e.UserAgent); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_log WHERE ($1 = '' OR action = $1)`, action).Scan(&total); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}
