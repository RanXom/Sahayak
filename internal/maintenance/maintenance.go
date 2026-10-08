// Package maintenance implements scheduled housekeeping jobs.
package maintenance

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sahayak/sahayak/internal/scheme"
)

// Service performs link checks and data retention.
type Service struct {
	pool     *pgxpool.Pool
	schemes  *scheme.Repository
	logger   *slog.Logger
	audioDir string
	pdfDir   string
	client   *http.Client
}

// New constructs a maintenance service.
func New(pool *pgxpool.Pool, schemes *scheme.Repository, logger *slog.Logger, audioDir, pdfDir string) *Service {
	return &Service{
		pool:     pool,
		schemes:  schemes,
		logger:   logger,
		audioDir: audioDir,
		pdfDir:   pdfDir,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// LinkCheck verifies published scheme application URLs and logs broken ones.
func (s *Service) LinkCheck(ctx context.Context) error {
	urls, err := s.schemes.ListApplicationURLs(ctx)
	if err != nil {
		return err
	}
	for _, u := range urls {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, u.URL, nil)
		if err != nil {
			continue
		}
		resp, err := s.client.Do(req)
		if err != nil {
			s.logger.Warn("link_check_failed", "scheme_id", u.ID, "error", err.Error())
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			s.logger.Warn("link_check_broken", "scheme_id", u.ID, "status", resp.StatusCode)
		}
	}
	return nil
}

// DataRetention deletes expired records and temporary files per policy.
func (s *Service) DataRetention(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM otp_challenges WHERE created_at < NOW() - INTERVAL '24 hours'`); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked_at < NOW() - INTERVAL '30 days'`); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM conversation_messages WHERE created_at < NOW() - INTERVAL '90 days'`); err != nil {
		return err
	}
	s.cleanupDir(s.audioDir, 24*time.Hour, nil)
	s.cleanupDir(s.pdfDir, 90*24*time.Hour, []string{".pdf"})
	return nil
}

func (s *Service) cleanupDir(dir string, maxAge time.Duration, exts []string) {
	if dir == "" {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if len(exts) > 0 {
			matched := false
			ext := strings.ToLower(filepath.Ext(path))
			for _, e := range exts {
				if ext == e {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}
		if info, err := d.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
		return nil
	})
}
