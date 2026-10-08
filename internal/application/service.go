package application

import (
	"context"
	"errors"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/profile"
	"github.com/sahayak/sahayak/internal/scheme"
)

// Service implements application use cases.
type Service struct {
	repo     *Repository
	schemes  *scheme.Repository
	profiles *profile.Repository
	pdf      *PDFGenerator
	notify   func(ctx context.Context, userID, appID, status string)
}

// NewService constructs the application service.
func NewService(repo *Repository, schemes *scheme.Repository, profiles *profile.Repository, pdf *PDFGenerator) *Service {
	return &Service{repo: repo, schemes: schemes, profiles: profiles, pdf: pdf}
}

// SetStatusNotifier registers a callback invoked when application status changes.
func (s *Service) SetStatusNotifier(fn func(ctx context.Context, userID, appID, status string)) {
	s.notify = fn
}

// FormFieldView is a rendered application field.
type FormFieldView struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	Value      any    `json:"value,omitempty"`
	AutoFilled bool   `json:"auto_filled"`
	Options    any    `json:"options,omitempty"`
}

// FormView is the application form for review.
type FormView struct {
	Application   *Application    `json:"application"`
	Scheme        *domain.Scheme  `json:"scheme"`
	Fields        []FormFieldView `json:"form_fields"`
	AutoFilled    map[string]any  `json:"auto_filled"`
	MissingFields []string        `json:"missing_fields"`
}

// CreateDraft creates an application draft with auto-filled fields.
func (s *Service) CreateDraft(ctx context.Context, userID, schemeID, lang string) (*FormView, error) {
	sch, err := s.schemes.GetPublishedByID(ctx, schemeID)
	if err != nil {
		if errors.Is(err, scheme.ErrNotFound) {
			return nil, domain.NotFound("Scheme not found")
		}
		return nil, domain.Internal("Could not load scheme", err)
	}
	p, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, profile.ErrNotFound) {
			return nil, domain.InvalidInput("Please complete your profile first")
		}
		return nil, domain.Internal("Could not load profile", err)
	}

	auto := AutoFill(sch.ApplicationFields, p.Facts)
	app, err := s.repo.Create(ctx, userID, p.ID, schemeID, auto)
	if err != nil {
		return nil, domain.Internal("Could not create application", err)
	}
	return s.buildFormView(app, sch, lang), nil
}

// Get returns an application form view.
func (s *Service) Get(ctx context.Context, userID, appID, lang string) (*FormView, error) {
	app, err := s.repo.Get(ctx, appID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Application not found")
		}
		return nil, domain.Internal("Could not load application", err)
	}
	sch, err := s.schemes.GetPublishedByID(ctx, app.SchemeID)
	if err != nil {
		return nil, domain.Internal("Could not load scheme", err)
	}
	return s.buildFormView(app, sch, lang), nil
}

// List returns a user's applications.
func (s *Service) List(ctx context.Context, userID string) ([]*Application, error) {
	out, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, domain.Internal("Could not list applications", err)
	}
	return out, nil
}

// UpdateForm stores edited form data and validates required fields.
func (s *Service) UpdateForm(ctx context.Context, userID, appID string, formData map[string]any) (*Application, error) {
	app, err := s.repo.Get(ctx, appID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Application not found")
		}
		return nil, domain.Internal("Could not load application", err)
	}
	sch, err := s.schemes.GetPublishedByID(ctx, app.SchemeID)
	if err != nil {
		return nil, domain.Internal("Could not load scheme", err)
	}
	merged := make(map[string]any, len(app.FormData)+len(formData))
	for k, v := range app.FormData {
		merged[k] = v
	}
	for k, v := range formData {
		merged[k] = v
	}
	missing := missingRequired(sch.ApplicationFields, merged)
	status := StatusDraft
	if len(missing) == 0 {
		status = StatusReady
	}
	if err := s.repo.UpdateFormData(ctx, app.ID, merged, status); err != nil {
		return nil, domain.Internal("Could not update application", err)
	}
	return s.repo.Get(ctx, appID, userID)
}

// GeneratePDF renders and stores the application PDF.
func (s *Service) GeneratePDF(ctx context.Context, userID, appID, lang string) (*Application, error) {
	app, err := s.repo.Get(ctx, appID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Application not found")
		}
		return nil, domain.Internal("Could not load application", err)
	}
	sch, err := s.schemes.GetPublishedByID(ctx, app.SchemeID)
	if err != nil {
		return nil, domain.Internal("Could not load scheme", err)
	}
	path, err := s.pdf.Generate(userID, app.ID, sch, app.FormData, lang)
	if err != nil {
		return nil, domain.Internal("Could not generate PDF", err)
	}
	if err := s.repo.SetPDFPath(ctx, app.ID, path); err != nil {
		return nil, domain.Internal("Could not save PDF reference", err)
	}
	return s.repo.Get(ctx, appID, userID)
}

// PDFPath returns the filesystem path of a generated PDF.
func (s *Service) PDFPath(ctx context.Context, userID, appID string) (string, error) {
	app, err := s.repo.Get(ctx, appID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", domain.NotFound("Application not found")
		}
		return "", domain.Internal("Could not load application", err)
	}
	if app.PDFPath == nil || *app.PDFPath == "" {
		return "", domain.NotFound("PDF has not been generated yet")
	}
	return *app.PDFPath, nil
}

// RecordSubmission records a user-reported submission reference.
func (s *Service) RecordSubmission(ctx context.Context, userID, appID, ref string) (*Application, error) {
	if _, err := s.repo.Get(ctx, appID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Application not found")
		}
		return nil, domain.Internal("Could not load application", err)
	}
	if err := s.repo.SetSubmission(ctx, appID, ref, StatusTracking); err != nil {
		return nil, domain.Internal("Could not record submission", err)
	}
	if s.notify != nil {
		s.notify(ctx, userID, appID, "submitted")
	}
	return s.repo.Get(ctx, appID, userID)
}

// UpdateStatus records a user-reported external status.
func (s *Service) UpdateStatus(ctx context.Context, userID, appID, status string) (*Application, error) {
	if status == "" {
		return nil, domain.InvalidInput("status is required")
	}
	if _, err := s.repo.Get(ctx, appID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("Application not found")
		}
		return nil, domain.Internal("Could not load application", err)
	}
	if err := s.repo.SetExternalStatus(ctx, appID, status); err != nil {
		return nil, domain.Internal("Could not update status", err)
	}
	if s.notify != nil {
		s.notify(ctx, userID, appID, status)
	}
	return s.repo.Get(ctx, appID, userID)
}

func (s *Service) buildFormView(app *Application, sch *domain.Scheme, lang string) *FormView {
	auto := AutoFill(sch.ApplicationFields, app.FormData)
	view := &FormView{
		Application: app,
		Scheme:      sch,
		Fields:      []FormFieldView{},
		AutoFilled:  auto,
	}
	for _, f := range sch.ApplicationFields {
		value, hasValue := app.FormData[f.FieldName]
		if !hasValue {
			if v, ok := auto[f.FieldName]; ok {
				value = v
				hasValue = true
			}
		}
		view.Fields = append(view.Fields, FormFieldView{
			Name:       f.FieldName,
			Label:      f.Label(lang),
			Type:       f.FieldType,
			Required:   f.IsRequired,
			Value:      value,
			AutoFilled: hasValue && f.ProfileMapping != "",
			Options:    f.Options,
		})
	}
	view.MissingFields = missingRequired(sch.ApplicationFields, app.FormData)
	return view
}

func missingRequired(fields []domain.ApplicationField, formData map[string]any) []string {
	var missing []string
	for _, f := range fields {
		if !f.IsRequired {
			continue
		}
		v, ok := formData[f.FieldName]
		if !ok || v == nil || v == "" {
			missing = append(missing, f.FieldName)
		}
	}
	return missing
}
