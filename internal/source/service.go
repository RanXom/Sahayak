package source

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
)

// Service implements source registry use cases.
type Service struct {
	repo *Repository
}

// NewService constructs the source service.
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// ListSources returns all sources.
func (s *Service) ListSources(ctx context.Context) ([]domain.Source, error) {
	out, err := s.repo.ListSources(ctx)
	if err != nil {
		return nil, domain.Internal("Could not list sources", err)
	}
	return out, nil
}

// CreateSource creates a source.
func (s *Service) CreateSource(ctx context.Context, in domain.Source) (*domain.Source, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.SourceType) == "" {
		return nil, domain.InvalidInput("name and source_type are required")
	}
	switch in.SourceType {
	case "official_api", "gov_dataset", "gov_portal", "gazette", "manual":
	default:
		return nil, domain.InvalidInput("Invalid source_type")
	}
	out, err := s.repo.CreateSource(ctx, in)
	if err != nil {
		return nil, domain.Internal("Could not create source", err)
	}
	return out, nil
}

// ListCandidates lists candidate schemes.
func (s *Service) ListCandidates(ctx context.Context, status string, limit, offset int) ([]domain.CandidateScheme, int64, error) {
	out, total, err := s.repo.ListCandidates(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, domain.Internal("Could not list candidates", err)
	}
	return out, total, nil
}

// CreateCandidate stores a candidate scheme.
func (s *Service) CreateCandidate(ctx context.Context, in domain.CandidateScheme) (*domain.CandidateScheme, error) {
	if in.SourceID == "" || strings.TrimSpace(in.Name) == "" {
		return nil, domain.InvalidInput("source_id and name are required")
	}
	out, err := s.repo.CreateCandidate(ctx, in)
	if err != nil {
		return nil, domain.Internal("Could not store candidate", err)
	}
	return out, nil
}

// Accept promotes a candidate to a draft scheme.
func (s *Service) Accept(ctx context.Context, id, reviewerID, notes string) (*domain.CandidateScheme, error) {
	c, err := s.repo.GetCandidate(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Candidate not found")
		}
		return nil, domain.Internal("Could not load candidate", err)
	}
	slug := slugify(c.Name)
	schemeID, err := s.repo.CreateDraftSchemeFromCandidate(ctx, c.Name, slug, c.SourceID)
	if err != nil {
		return nil, domain.Internal("Could not create draft scheme", err)
	}
	if err := s.repo.Review(ctx, id, domain.CandidateAccepted, reviewerID, notes, &schemeID); err != nil {
		return nil, domain.Internal("Could not update candidate", err)
	}
	return s.repo.GetCandidate(ctx, id)
}

// Reject rejects a candidate.
func (s *Service) Reject(ctx context.Context, id, reviewerID, notes string) (*domain.CandidateScheme, error) {
	if err := s.repo.Review(ctx, id, domain.CandidateRejected, reviewerID, notes, nil); err != nil {
		return nil, domain.Internal("Could not update candidate", err)
	}
	return s.repo.GetCandidate(ctx, id)
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlug.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
