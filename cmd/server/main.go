package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sahayak/sahayak/internal/config"
	"github.com/sahayak/sahayak/internal/db"
	"github.com/sahayak/sahayak/internal/obs"
	"github.com/sahayak/sahayak/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := server.New(cfg, logger, pool)

	if cfg.ServerMode == "all" || cfg.ServerMode == "worker" {
		if err := srv.StartWorkers(ctx); err != nil {
			logger.Error("workers_start_failed", "error", err)
		} else {
			logger.Info("workers_started")
		}
	}

	httpServer := &http.Server{
		Addr:              ":" + itoa(cfg.Port),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server_started", "port", cfg.Port, "mode", cfg.ServerMode, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown_initiated")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.StopWorkers(shutdownCtx); err != nil {
		logger.Error("workers_stop_error", "error", err)
	}
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown_error", "error", err)
	}
	logger.Info("shutdown_complete")
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
