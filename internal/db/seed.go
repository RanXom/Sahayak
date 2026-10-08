package db

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
)

// seedOrder defines the execution order for idempotent seed files. Order
// matters: states must exist before districts and schemes reference them.
var seedOrder = []string{
	"seed/states.sql",
	"seed/districts.sql",
	"seed/document_types.sql",
	"seed/initial_schemes.sql",
}

// Seed applies all seed files in order. Every seed file is idempotent.
func Seed(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	for _, name := range seedOrder {
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("read seed %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("apply seed %s: %w", name, err)
		}
	}
	return nil
}
