package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/auth"
	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/db"
	"github.com/sahayak/sahayak/internal/jobs"
	"github.com/sahayak/sahayak/internal/obs"
	"github.com/sahayak/sahayak/migrations"
)

func main() {
	adminPhone := flag.String("admin-phone", "", "phone number to create/promote as admin (or set ADMIN_PHONE)")
	flag.Parse()

	phone := *adminPhone
	if phone == "" {
		phone = os.Getenv("ADMIN_PHONE")
	}

	if err := run(phone); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(adminPhone string) error {
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

	if adminPhone != "" {
		if err := upsertAdmin(ctx, pool, cfg, adminPhone); err != nil {
			return err
		}
		logger.Info("admin_upserted", "phone_last4", last4(adminPhone))
	}
	return nil
}

func upsertAdmin(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, phone string) error {
	normalized, err := auth.NormalizePhone(phone)
	if err != nil {
		return err
	}
	hash := auth.HashPhone(normalized, cfg.PhoneHashPepper)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (phone_hash, phone_last4, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (phone_hash) DO UPDATE SET role = 'admin', updated_at = NOW()`,
		hash, auth.Last4(normalized))
	return err
}

func last4(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	return phone[len(phone)-4:]
}
