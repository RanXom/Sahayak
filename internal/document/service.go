package document

import (
	"context"
	"errors"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/profile"
	"github.com/sahayak/sahayak/internal/scheme"
)

// Service implements document requirement matching.
type Service struct {
	repo     *Repository
	schemes  *scheme.Repository
	profiles *profile.Repository
}

// NewService constructs the document service.
func NewService(repo *Repository, schemes *scheme.Repository, profiles *profile.Repository) *Service {
	return &Service{repo: repo, schemes: schemes, profiles: profiles}
}

// AcquisitionInfo describes how to obtain a missing document.
type AcquisitionInfo struct {
	IssuingAuthority string            `json:"issuing_authority,omitempty"`
	Steps            string            `json:"steps,omitempty"`
	TypicalTime      string            `json:"typical_time,omitempty"`
	TypicalFee       string            `json:"typical_fee,omitempty"`
	SourceURL        string            `json:"source_url,omitempty"`
	AcquisitionURL   string            `json:"acquisition_url,omitempty"`
	NearestOffices   []ServiceLocation `json:"nearest_offices,omitempty"`
}

// DocumentStatus is the per-document result for a scheme.
type DocumentStatus struct {
	DocumentType string           `json:"type"`
	DocumentName string           `json:"name"`
	IsMandatory  bool             `json:"mandatory"`
	UserHas      bool             `json:"user_has"`
	Status       string           `json:"status"`
	Acquisition  *AcquisitionInfo `json:"acquisition,omitempty"`
}

// CheckResult is the document checklist for a scheme.
type CheckResult struct {
	SchemeID     string           `json:"scheme_id"`
	Documents    []DocumentStatus `json:"documents"`
	MissingCount int              `json:"missing_count"`
	AllPresent   bool             `json:"all_present"`
}

// CheckScheme computes the document checklist for a user and scheme.
func (s *Service) CheckScheme(ctx context.Context, userID, schemeID, lang string) (*CheckResult, error) {
	required, err := s.schemes.ListRequiredDocuments(ctx, schemeID)
	if err != nil {
		return nil, domain.Internal("Could not load scheme documents", err)
	}

	has := map[string]bool{}
	var districtID *int
	if p, err := s.profiles.GetByUserID(ctx, userID); err == nil {
		districtID = p.DistrictID
		for _, d := range p.Documents {
			has[d.DocumentType] = d.HasDocument
		}
	}

	result := &CheckResult{SchemeID: schemeID, Documents: []DocumentStatus{}}
	for _, req := range required {
		name := req.DocumentNameEN
		if lang == "hi" && req.DocumentNameHI != "" {
			name = req.DocumentNameHI
		}
		ds := DocumentStatus{
			DocumentType: req.DocumentType,
			DocumentName: name,
			IsMandatory:  req.IsMandatory,
			UserHas:      has[req.DocumentType],
		}
		if ds.UserHas {
			ds.Status = "available"
		} else {
			ds.Status = "missing"
			if req.IsMandatory {
				result.MissingCount++
			}
			if info := s.acquisition(ctx, req.DocumentType, lang, districtID); info != nil {
				ds.Acquisition = info
			}
		}
		result.Documents = append(result.Documents, ds)
	}
	result.AllPresent = result.MissingCount == 0
	return result, nil
}

// ListTypes returns all document types.
func (s *Service) ListTypes(ctx context.Context) ([]domain.DocumentType, error) {
	out, err := s.repo.ListDocumentTypes(ctx)
	if err != nil {
		return nil, domain.Internal("Could not load document types", err)
	}
	return out, nil
}

func (s *Service) acquisition(ctx context.Context, docType, lang string, districtID *int) *AcquisitionInfo {
	dt, err := s.repo.GetDocumentType(ctx, docType)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return nil
	}
	info := &AcquisitionInfo{
		TypicalTime:    dt.TypicalTime,
		TypicalFee:     dt.TypicalFee,
		SourceURL:      dt.SourceURL,
		AcquisitionURL: dt.AcquisitionURL,
	}
	if lang == "hi" {
		info.IssuingAuthority = firstNonEmpty(dt.IssuingAuthorityHI, dt.IssuingAuthorityEN)
		info.Steps = firstNonEmpty(dt.AcquisitionStepsHI, dt.AcquisitionStepsEN)
	} else {
		info.IssuingAuthority = firstNonEmpty(dt.IssuingAuthorityEN, dt.IssuingAuthorityHI)
		info.Steps = firstNonEmpty(dt.AcquisitionStepsEN, dt.AcquisitionStepsHI)
	}
	if districtID != nil {
		if offices, err := s.repo.ListServiceLocations(ctx, *districtID, docType); err == nil && len(offices) > 0 {
			info.NearestOffices = offices
		}
	}
	return info
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
