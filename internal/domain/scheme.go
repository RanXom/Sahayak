package domain

import "time"

// Scheme lifecycle statuses.
const (
	SchemeDraft     = "draft"
	SchemeVerified  = "verified"
	SchemePublished = "published"
	SchemeArchived  = "archived"
)

// Scheme scopes.
const (
	ScopeCentral = "central"
	ScopeState   = "state"
)

// Scheme is a government welfare scheme.
type Scheme struct {
	ID                   string     `json:"id"`
	Slug                 string     `json:"slug"`
	Status               string     `json:"status"`
	NameEN               string     `json:"name_en"`
	NameHI               string     `json:"name_hi"`
	DescriptionEN        string     `json:"description_en,omitempty"`
	DescriptionHI        string     `json:"description_hi,omitempty"`
	BenefitType          string     `json:"benefit_type"`
	BenefitAmount        string     `json:"benefit_amount,omitempty"`
	BenefitDescriptionEN string     `json:"benefit_description_en,omitempty"`
	BenefitDescriptionHI string     `json:"benefit_description_hi,omitempty"`
	Scope                string     `json:"scope"`
	StateID              *int       `json:"state_id,omitempty"`
	Ministry             string     `json:"ministry,omitempty"`
	Department           string     `json:"department,omitempty"`
	ApplicationURL       string     `json:"application_url,omitempty"`
	OfficialURL          string     `json:"official_url,omitempty"`
	Deadline             *time.Time `json:"deadline,omitempty"`
	IsActive             bool       `json:"is_active"`
	SourceID             *string    `json:"source_id,omitempty"`
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`
	VerifiedBy           *string    `json:"verified_by,omitempty"`
	PublishedAt          *time.Time `json:"published_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	Rules             []EligibilityRule        `json:"rules,omitempty"`
	RequiredDocuments []SchemeRequiredDocument `json:"required_documents,omitempty"`
	ApplicationFields []ApplicationField       `json:"application_fields,omitempty"`
}

// Name returns the localized scheme name.
func (s *Scheme) Name(lang string) string {
	if lang == "hi" && s.NameHI != "" {
		return s.NameHI
	}
	return s.NameEN
}

// Description returns the localized scheme description.
func (s *Scheme) Description(lang string) string {
	if lang == "hi" && s.DescriptionHI != "" {
		return s.DescriptionHI
	}
	return s.DescriptionEN
}

// EligibilityRule is a single expr-lang expression belonging to a scheme.
type EligibilityRule struct {
	ID            string `json:"id"`
	SchemeID      string `json:"scheme_id"`
	RuleName      string `json:"rule_name"`
	Expression    string `json:"expression"`
	DescriptionEN string `json:"description_en,omitempty"`
	DescriptionHI string `json:"description_hi,omitempty"`
	IsMandatory   bool   `json:"is_mandatory"`
	SortOrder     int    `json:"sort_order"`
}

// SchemeRequiredDocument links a scheme to a required document type.
type SchemeRequiredDocument struct {
	ID             string `json:"id"`
	SchemeID       string `json:"scheme_id"`
	DocumentType   string `json:"document_type"`
	DocumentNameEN string `json:"document_name_en"`
	DocumentNameHI string `json:"document_name_hi"`
	IsMandatory    bool   `json:"is_mandatory"`
	SortOrder      int    `json:"sort_order"`
}

// ApplicationField describes a single field on a scheme application form.
type ApplicationField struct {
	ID             string `json:"id"`
	SchemeID       string `json:"scheme_id"`
	FieldName      string `json:"field_name"`
	FieldLabelEN   string `json:"field_label_en"`
	FieldLabelHI   string `json:"field_label_hi"`
	FieldType      string `json:"field_type"`
	ProfileMapping string `json:"profile_mapping,omitempty"`
	IsRequired     bool   `json:"is_required"`
	SortOrder      int    `json:"sort_order"`
	Options        any    `json:"options,omitempty"`
}

// Label returns the localized field label.
func (f *ApplicationField) Label(lang string) string {
	if lang == "hi" && f.FieldLabelHI != "" {
		return f.FieldLabelHI
	}
	return f.FieldLabelEN
}

// Source is an official data source tracked for provenance.
type Source struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	SourceType      string     `json:"source_type"`
	URL             string     `json:"url,omitempty"`
	Description     string     `json:"description,omitempty"`
	IsActive        bool       `json:"is_active"`
	FetchConfig     any        `json:"fetch_config,omitempty"`
	LastFetchedAt   *time.Time `json:"last_fetched_at,omitempty"`
	LastFetchStatus string     `json:"last_fetch_status,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Candidate scheme review statuses.
const (
	CandidatePending   = "pending"
	CandidateReviewing = "reviewing"
	CandidateAccepted  = "accepted"
	CandidateRejected  = "rejected"
)

// CandidateScheme is a discovered but unverified scheme.
type CandidateScheme struct {
	ID            string     `json:"id"`
	SourceID      string     `json:"source_id"`
	ExternalID    string     `json:"external_id,omitempty"`
	RawData       any        `json:"raw_data"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	ContentHash   string     `json:"content_hash"`
	FetchedAt     time.Time  `json:"fetched_at"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy    *string    `json:"reviewed_by,omitempty"`
	ReviewerNotes string     `json:"reviewer_notes,omitempty"`
	PromotedTo    *string    `json:"promoted_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
