package domain

import "time"

// Role constants.
const (
	RoleCitizen = "citizen"
	RoleHelper  = "helper"
	RoleAdmin   = "admin"
)

// User is the authenticated principal.
type User struct {
	ID            string     `json:"id"`
	PhoneLast4    string     `json:"phone_last4"`
	Role          string     `json:"role"`
	PreferredLang string     `json:"preferred_lang"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// State is an Indian state or union territory.
type State struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	NameEN    string `json:"name_en"`
	NameHI    string `json:"name_hi"`
	NameLocal string `json:"name_local,omitempty"`
	IsUT      bool   `json:"is_ut"`
}

// District is a district within a state.
type District struct {
	ID         int    `json:"id"`
	StateID    int    `json:"state_id"`
	NameEN     string `json:"name_en"`
	NameHI     string `json:"name_hi"`
	NameLocal  string `json:"name_local,omitempty"`
	CensusCode string `json:"census_code,omitempty"`
}

// CoverageLevel enumerates scheme coverage levels.
const (
	CoverageFull    = "full"
	CoveragePartial = "partial"
	CoverageNone    = "none"
)

// Coverage describes scheme coverage for a state.
type Coverage struct {
	Level       string    `json:"level"`
	SchemeCount int       `json:"scheme_count"`
	LastUpdated time.Time `json:"last_updated"`
	Note        string    `json:"note,omitempty"`
}

// StateWithCoverage combines a state with its coverage status.
type StateWithCoverage struct {
	State
	Coverage Coverage `json:"coverage"`
}

// Profile is a citizen's welfare profile.
type Profile struct {
	ID         string            `json:"id"`
	UserID     string            `json:"user_id"`
	StateID    *int              `json:"state_id,omitempty"`
	DistrictID *int              `json:"district_id,omitempty"`
	Facts      map[string]any    `json:"facts"`
	Documents  []ProfileDocument `json:"documents,omitempty"`
	Completed  bool              `json:"completed"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ProfileDocument records whether a user possesses a document type.
type ProfileDocument struct {
	DocumentType string `json:"document_type"`
	HasDocument  bool   `json:"has_document"`
	Notes        string `json:"notes,omitempty"`
}

// DocumentType is reference data describing a document and how to acquire it.
type DocumentType struct {
	ID                 string     `json:"id"`
	NameEN             string     `json:"name_en"`
	NameHI             string     `json:"name_hi"`
	DescriptionEN      string     `json:"description_en"`
	DescriptionHI      string     `json:"description_hi"`
	Category           string     `json:"category"`
	IssuingAuthorityEN string     `json:"issuing_authority_en,omitempty"`
	IssuingAuthorityHI string     `json:"issuing_authority_hi,omitempty"`
	AcquisitionURL     string     `json:"acquisition_url,omitempty"`
	AcquisitionStepsEN string     `json:"acquisition_steps_en,omitempty"`
	AcquisitionStepsHI string     `json:"acquisition_steps_hi,omitempty"`
	TypicalTime        string     `json:"typical_time,omitempty"`
	TypicalFee         string     `json:"typical_fee,omitempty"`
	LastVerified       *time.Time `json:"last_verified,omitempty"`
	SourceURL          string     `json:"source_url,omitempty"`
}
