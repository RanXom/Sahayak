package location

import (
	"context"
	"errors"

	"github.com/sahayak/sahayak/internal/domain"
)

// Service implements location use cases.
type Service struct {
	repo *Repository
}

// NewService constructs the location service.
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// ListStates returns all states with coverage.
func (s *Service) ListStates(ctx context.Context) ([]domain.StateWithCoverage, error) {
	states, err := s.repo.ListStates(ctx)
	if err != nil {
		return nil, domain.Internal("Could not load states", err)
	}
	for i := range states {
		if states[i].Coverage.Level == domain.CoverageNone && states[i].Coverage.Note == "" {
			states[i].Coverage.Note = "State scheme coverage is being developed. Central schemes are still available."
		}
	}
	return states, nil
}

// ListDistricts returns districts for a state.
func (s *Service) ListDistricts(ctx context.Context, stateID int) ([]domain.District, error) {
	if _, err := s.repo.GetState(ctx, stateID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("State not found")
		}
		return nil, domain.Internal("Could not load state", err)
	}
	districts, err := s.repo.ListDistricts(ctx, stateID)
	if err != nil {
		return nil, domain.Internal("Could not load districts", err)
	}
	return districts, nil
}
