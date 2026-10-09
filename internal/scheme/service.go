package scheme

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/rules"
)

// Service implements scheme use cases.
type Service struct {
	repo *Repository
}

// NewService constructs the scheme service.
func NewService(repo *Repository) *Service { return &Service{repo: repo} }

// Input is the admin payload for creating or updating a scheme.
type Input struct {
	Slug                 string                          `json:"slug"`
	NameEN               string                          `json:"name_en"`
	NameHI               string                          `json:"name_hi"`
	DescriptionEN        string                          `json:"description_en"`
	DescriptionHI        string                          `json:"description_hi"`
	BenefitType          string                          `json:"benefit_type"`
	BenefitAmount        string                          `json:"benefit_amount"`
	BenefitDescriptionEN string                          `json:"benefit_description_en"`
	BenefitDescriptionHI string                          `json:"benefit_description_hi"`
	Scope                string                          `json:"scope"`
	StateID              *int                            `json:"state_id"`
	Ministry             string                          `json:"ministry"`
	Department           string                          `json:"department"`
	ApplicationURL       string                          `json:"application_url"`
	OfficialURL          string                          `json:"official_url"`
	SourceID             *string                         `json:"source_id"`
	IsActive             *bool                           `json:"is_active"`
	Rules                []domain.EligibilityRule        `json:"rules"`
	RequiredDocuments    []domain.SchemeRequiredDocument `json:"required_documents"`
	ApplicationFields    []domain.ApplicationField       `json:"application_fields"`
}

// ListPublished returns published, active schemes (central + optional state)
// without evaluation internals, for public guest browsing.
func (s *Service) ListPublished(ctx context.Context, stateID *int) ([]*domain.Scheme, error) {
	out, err := s.repo.ListPublishedBrief(ctx, stateID)
	if err != nil {
		return nil, domain.Internal("Could not list schemes", err)
	}
	return out, nil
}

// GetPublished returns a published scheme by ID.
func (s *Service) GetPublished(ctx context.Context, id string) (*domain.Scheme, error) {
	sch, err := s.repo.GetPublishedByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Scheme not found")
		}
		return nil, domain.Internal("Could not load scheme", err)
	}
	return sch, nil
}

// ListAdmin returns schemes of any status.
func (s *Service) ListAdmin(ctx context.Context, status string, limit, offset int) ([]*domain.Scheme, int64, error) {
	out, total, err := s.repo.ListAdmin(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, domain.Internal("Could not list schemes", err)
	}
	return out, total, nil
}

// ListForEvaluation returns published schemes scoped to a state, with rules.
func (s *Service) ListForEvaluation(ctx context.Context, stateID *int) ([]*domain.Scheme, error) {
	out, err := s.repo.ListPublished(ctx, stateID)
	if err != nil {
		return nil, domain.Internal("Could not load schemes", err)
	}
	return out, nil
}

// Create creates a scheme with its children.
func (s *Service) Create(ctx context.Context, in Input) (*domain.Scheme, error) {
	if err := validateInput(&in, true); err != nil {
		return nil, err
	}
	sch := inputToScheme(in)
	if sch.Slug == "" {
		sch.Slug = slugify(in.NameEN)
	}
	created, err := s.repo.Create(ctx, sch)
	if err != nil {
		return nil, domain.Internal("Could not create scheme", err)
	}
	if err := s.applyChildren(ctx, created.ID, in); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, created.ID)
}

// Update updates a scheme and replaces its children.
func (s *Service) Update(ctx context.Context, id string, in Input) (*domain.Scheme, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Scheme not found")
		}
		return nil, domain.Internal("Could not load scheme", err)
	}
	if err := validateInput(&in, false); err != nil {
		return nil, err
	}
	sch := inputToScheme(in)
	sch.ID = id
	if _, err := s.repo.Update(ctx, sch); err != nil {
		return nil, domain.Internal("Could not update scheme", err)
	}
	if err := s.applyChildren(ctx, id, in); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Publish publishes a scheme.
func (s *Service) Publish(ctx context.Context, id string) error {
	sch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.NotFound("Scheme not found")
		}
		return domain.Internal("Could not load scheme", err)
	}
	if len(sch.Rules) == 0 {
		return domain.InvalidInput("A scheme must have at least one eligibility rule before publishing")
	}
	if err := s.repo.SetStatus(ctx, id, domain.SchemePublished); err != nil {
		return domain.Internal("Could not publish scheme", err)
	}
	return nil
}

// Archive archives a scheme.
func (s *Service) Archive(ctx context.Context, id string) error {
	if err := s.repo.SetStatus(ctx, id, domain.SchemeArchived); err != nil {
		return domain.Internal("Could not archive scheme", err)
	}
	return nil
}

// SetStatus sets an arbitrary lifecycle status.
func (s *Service) SetStatus(ctx context.Context, id, status string) error {
	switch status {
	case domain.SchemeDraft, domain.SchemeVerified, domain.SchemePublished, domain.SchemeArchived:
	default:
		return domain.InvalidInput("Invalid scheme status")
	}
	if err := s.repo.SetStatus(ctx, id, status); err != nil {
		return domain.Internal("Could not update scheme status", err)
	}
	return nil
}

func (s *Service) applyChildren(ctx context.Context, id string, in Input) error {
	if in.Rules != nil {
		if err := s.repo.ReplaceRules(ctx, id, in.Rules); err != nil {
			return domain.Internal("Could not save eligibility rules", err)
		}
	}
	if in.RequiredDocuments != nil {
		if err := s.repo.ReplaceRequiredDocuments(ctx, id, in.RequiredDocuments); err != nil {
			return domain.Internal("Could not save required documents", err)
		}
	}
	if in.ApplicationFields != nil {
		if err := s.repo.ReplaceApplicationFields(ctx, id, in.ApplicationFields); err != nil {
			return domain.Internal("Could not save application fields", err)
		}
	}
	return nil
}

func validateInput(in *Input, requireName bool) error {
	if requireName && strings.TrimSpace(in.NameEN) == "" {
		return domain.InvalidInput("name_en is required")
	}
	if in.NameHI == "" {
		in.NameHI = in.NameEN
	}
	switch in.BenefitType {
	case "pension", "scholarship", "insurance", "housing", "health", "agriculture", "employment", "financial", "women_child", "other":
	default:
		return domain.InvalidInput("Invalid benefit_type")
	}
	switch in.Scope {
	case "":
		in.Scope = domain.ScopeCentral
	case domain.ScopeCentral, domain.ScopeState:
	default:
		return domain.InvalidInput("Invalid scope")
	}
	if in.Scope == domain.ScopeState && in.StateID == nil {
		return domain.InvalidInput("state_id is required for state-scoped schemes")
	}
	for i := range in.Rules {
		if strings.TrimSpace(in.Rules[i].RuleName) == "" {
			return domain.InvalidInput("Each eligibility rule requires a rule_name")
		}
		if err := rules.Validate(in.Rules[i].Expression); err != nil {
			return domain.InvalidInput("Rule \"" + in.Rules[i].RuleName + "\": " + err.Error())
		}
		if err := rules.ValidateVariables(in.Rules[i].Expression); err != nil {
			return domain.InvalidInput("Rule \"" + in.Rules[i].RuleName + "\": " + err.Error())
		}
	}
	return nil
}

func inputToScheme(in Input) *domain.Scheme {
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	return &domain.Scheme{
		Slug:                 in.Slug,
		NameEN:               in.NameEN,
		NameHI:               in.NameHI,
		DescriptionEN:        in.DescriptionEN,
		DescriptionHI:        in.DescriptionHI,
		BenefitType:          in.BenefitType,
		BenefitAmount:        in.BenefitAmount,
		BenefitDescriptionEN: in.BenefitDescriptionEN,
		BenefitDescriptionHI: in.BenefitDescriptionHI,
		Scope:                in.Scope,
		StateID:              in.StateID,
		Ministry:             in.Ministry,
		Department:           in.Department,
		ApplicationURL:       in.ApplicationURL,
		OfficialURL:          in.OfficialURL,
		SourceID:             in.SourceID,
		IsActive:             isActive,
	}
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlug.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
