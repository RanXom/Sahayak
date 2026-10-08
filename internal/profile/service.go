package profile

import (
	"context"
	"errors"
	"time"

	"github.com/sahayak/sahayak/internal/domain"
)

// LocationResolver resolves state and district metadata.
type LocationResolver interface {
	GetState(ctx context.Context, id int) (*domain.State, error)
	GetDistrict(ctx context.Context, id int) (*domain.District, error)
}

// Service implements profile use cases.
type Service struct {
	repo     *Repository
	location LocationResolver
}

// NewService constructs the profile service.
func NewService(repo *Repository, location LocationResolver) *Service {
	return &Service{repo: repo, location: location}
}

// View is the API representation of a profile with resolved location.
type View struct {
	ID        string                   `json:"id"`
	State     *domain.State            `json:"state,omitempty"`
	District  *domain.District         `json:"district,omitempty"`
	Facts     map[string]any           `json:"facts"`
	Documents []domain.ProfileDocument `json:"documents"`
	Completed bool                     `json:"completed"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// View returns the enriched profile representation.
func (s *Service) View(ctx context.Context, userID string) (*View, error) {
	p, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	v := &View{
		ID:        p.ID,
		Facts:     p.Facts,
		Documents: p.Documents,
		Completed: p.Completed,
		UpdatedAt: p.UpdatedAt,
	}
	if v.Facts == nil {
		v.Facts = map[string]any{}
	}
	if v.Documents == nil {
		v.Documents = []domain.ProfileDocument{}
	}
	if p.StateID != nil && s.location != nil {
		if st, err := s.location.GetState(ctx, *p.StateID); err == nil {
			v.State = st
		}
	}
	if p.DistrictID != nil && s.location != nil {
		if d, err := s.location.GetDistrict(ctx, *p.DistrictID); err == nil {
			v.District = d
		}
	}
	return v, nil
}

// UpdateInput is the payload for updating a profile.
type UpdateInput struct {
	StateID    *int           `json:"state_id"`
	DistrictID *int           `json:"district_id"`
	Facts      map[string]any `json:"facts"`
}

// GetOrCreate returns the user's profile, creating one if absent.
func (s *Service) GetOrCreate(ctx context.Context, userID string) (*domain.Profile, error) {
	p, err := s.repo.GetByUserID(ctx, userID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, domain.Internal("Could not load profile", err)
	}
	p, err = s.repo.Create(ctx, userID, nil, nil)
	if err != nil {
		return nil, domain.Internal("Could not create profile", err)
	}
	return p, nil
}

// Get returns the user's profile.
func (s *Service) Get(ctx context.Context, userID string) (*domain.Profile, error) {
	return s.GetOrCreate(ctx, userID)
}

// Update replaces the profile facts and location.
func (s *Service) Update(ctx context.Context, userID string, in UpdateInput) (*domain.Profile, error) {
	p, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	facts := FilterKnownFacts(in.Facts)
	if err := ValidateFacts(facts); err != nil {
		return nil, domain.InvalidInput(err.Error())
	}
	if err := s.repo.ReplaceFacts(ctx, p.ID, facts, in.StateID, in.DistrictID, computeCompleted(in.StateID, facts)); err != nil {
		return nil, domain.Internal("Could not update profile", err)
	}
	s.syncDocumentsFromFacts(ctx, p.ID, facts)
	return s.Get(ctx, userID)
}

// MergeFacts merges validated facts into the profile.
func (s *Service) MergeFacts(ctx context.Context, userID string, facts map[string]any, stateID, districtID *int) (*domain.Profile, error) {
	p, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	filtered := FilterKnownFacts(facts)
	if err := ValidateFacts(filtered); err != nil {
		return nil, domain.InvalidInput(err.Error())
	}
	merged := make(map[string]any, len(p.Facts)+len(filtered))
	for k, v := range p.Facts {
		merged[k] = v
	}
	for k, v := range filtered {
		merged[k] = v
	}
	if stateID == nil {
		stateID = p.StateID
	}
	if districtID == nil {
		districtID = p.DistrictID
	}
	if err := s.repo.ReplaceFacts(ctx, p.ID, merged, stateID, districtID, computeCompleted(stateID, merged)); err != nil {
		return nil, domain.Internal("Could not update profile", err)
	}
	s.syncDocumentsFromFacts(ctx, p.ID, merged)
	return s.Get(ctx, userID)
}

// GetDocuments returns the document inventory.
func (s *Service) GetDocuments(ctx context.Context, userID string) ([]domain.ProfileDocument, error) {
	p, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.Documents, nil
}

// DocumentUpdate is a single document inventory entry.
type DocumentUpdate struct {
	DocumentType string `json:"document_type"`
	HasDocument  bool   `json:"has_document"`
	Notes        string `json:"notes"`
}

// UpdateDocuments updates one or more document inventory entries.
func (s *Service) UpdateDocuments(ctx context.Context, userID string, updates []DocumentUpdate) ([]domain.ProfileDocument, error) {
	p, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, u := range updates {
		if err := ValidateDocumentType(u.DocumentType); err != nil {
			return nil, err
		}
		if err := s.repo.UpsertDocument(ctx, p.ID, u.DocumentType, u.HasDocument, u.Notes); err != nil {
			return nil, domain.Internal("Could not update documents", err)
		}
	}
	// Mirror document possession back into facts so rules that read
	// has_aadhaar / has_bank_account stay consistent with the checklist.
	facts := p.Facts
	if facts == nil {
		facts = map[string]any{}
	}
	changed := false
	for _, u := range updates {
		switch u.DocumentType {
		case "aadhaar":
			facts["has_aadhaar"] = u.HasDocument
			changed = true
		case "bank_account":
			facts["has_bank_account"] = u.HasDocument
			changed = true
		}
	}
	if changed {
		if err := s.repo.ReplaceFacts(ctx, p.ID, facts, p.StateID, p.DistrictID, p.Completed); err != nil {
			return nil, domain.Internal("Could not update documents", err)
		}
	}
	return s.repo.GetDocuments(ctx, p.ID)
}

// syncDocumentsFromFacts keeps the document inventory consistent with facts
// extracted during the conversation (e.g. "I have an Aadhaar card").
func (s *Service) syncDocumentsFromFacts(ctx context.Context, profileID string, facts map[string]any) {
	if v, ok := facts["has_aadhaar"].(bool); ok {
		_ = s.repo.UpsertDocument(ctx, profileID, "aadhaar", v, "")
	}
	if v, ok := facts["has_bank_account"].(bool); ok {
		_ = s.repo.UpsertDocument(ctx, profileID, "bank_account", v, "")
	}
	if v, ok := facts["ration_card_type"].(string); ok && v != "" && v != "none" {
		_ = s.repo.UpsertDocument(ctx, profileID, "ration_card", true, "")
	}
	if v, ok := toFloat(facts["land_acres"]); ok && v > 0 {
		_ = s.repo.UpsertDocument(ctx, profileID, "land_record", true, "")
	}
}

// computeCompleted determines whether enough core facts exist for matching.
func computeCompleted(stateID *int, facts map[string]any) bool {
	if stateID == nil {
		if _, ok := facts["state"]; !ok {
			return false
		}
	}
	_, hasAge := facts["age"]
	_, hasGender := facts["gender"]
	return hasAge && hasGender
}
