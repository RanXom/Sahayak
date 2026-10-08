// Package migrations embeds the SQL migration and seed files so they can be
// applied by cmd/migrate without shipping files alongside the binary.
package migrations

import "embed"

// FS contains all .sql migration files and the seed/ directory.
//
//go:embed *.sql seed/*.sql
var FS embed.FS

// Dir is the root directory within FS containing migrations.
const Dir = "."
