package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/db"
	"github.com/sahayak/sahayak/internal/obs"
	"github.com/sahayak/sahayak/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migrate failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := obs.NewLogger(os.Getenv("LOG_LEVEL"))
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		if err := db.Migrate(ctx, pool, migrations.FS, migrations.Dir); err != nil {
			return err
		}
		logger.Info("migrations_applied")
	case "down":
		if err := db.Rollback(ctx, pool, migrations.FS, migrations.Dir); err != nil {
			return err
		}
		logger.Info("migration_rolled_back")
	case "seed":
		if err := db.Seed(ctx, pool, migrations.FS); err != nil {
			return err
		}
		logger.Info("seed_applied")
	default:
		return fmt.Errorf("unknown command %q (use up|down|seed)", cmd)
	}
	return nil
}
