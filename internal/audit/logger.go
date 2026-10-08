// Package audit records security-relevant actions. It never stores PII.
package audit

import (
	"context"
	"log/slog"
	"time"
)

// Entry is a single audit log record.
type Entry struct {
	ActorID      *string
	ActorRole    string
	Action       string
	ResourceType string
	ResourceID   string
	Details      map[string]any
	IPAddress    string
	UserAgent    string
}

// Logger writes audit entries.
type Logger interface {
	Log(ctx context.Context, e Entry)
}

// DBLogger writes audit entries to PostgreSQL.
type DBLogger struct {
	repo   *Repository
	logger *slog.Logger
}

// NewDBLogger constructs a DB-backed audit logger.
func NewDBLogger(repo *Repository, logger *slog.Logger) *DBLogger {
	return &DBLogger{repo: repo, logger: logger}
}

// Log persists an audit entry. Failures are logged but never propagated.
func (l *DBLogger) Log(ctx context.Context, e Entry) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := l.repo.Insert(ctx, e); err != nil {
		l.logger.Error("audit_write_failed", "action", e.Action, "error", err)
	}
}

// NopLogger discards audit entries (useful in tests).
type NopLogger struct{}

// Log implements Logger.
func (NopLogger) Log(context.Context, Entry) {}
