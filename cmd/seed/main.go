package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/db"
	"github.com/sahayak/sahayak/internal/jobs"
	"github.com/sahayak/sahayak/internal/obs"
	"github.com/sahayak/sahayak/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
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

	if err := db.Migrate(ctx, pool, migrations.FS, migrations.Dir); err != nil {
		return err
	}
	if err := jobs.Migrate(ctx, pool); err != nil {
		return err
	}
	if err := db.Seed(ctx, pool, migrations.FS); err != nil {
		return err
	}
	logger.Info("seed_complete")
	return nil
}
