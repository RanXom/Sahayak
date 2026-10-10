// Package assistance implements the client-helper marketplace: assistance
// requests, deterministic language-first matching, sessions with consent,
// chat, WebRTC signaling and shared application drafts.
package assistance

import (
	"sort"
	"time"
)

// Request lifecycle states.
const (
	StatusPending   = "pending"
	StatusAccepted  = "accepted"
	StatusInSession = "in_session"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
	StatusExpired   = "expired"
)

// Assistance categories a helper can claim and a request can ask for.
var Categories = []string{
	"scheme_discovery",
	"application_help",
	"document_help",
}

// Request is a citizen's ask for remote help.
type Request struct {
	ID                string     `json:"id"`
	CitizenUserID     string     `json:"citizen_user_id,omitempty"`
	ApplicationID     *string    `json:"application_id,omitempty"`
	SchemeID          *string    `json:"scheme_id,omitempty"`
	Language          string     `json:"language"`
	Category          string     `json:"category"`
	Details           string     `json:"details,omitempty"`
	ShareProfileFacts bool       `json:"share_profile_facts"`
	ShareDocuments    bool       `json:"share_documents"`
	Status            string     `json:"status"`
	MatchedHelperID   *string    `json:"matched_helper_user_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	MatchedAt         *time.Time `json:"matched_at,omitempty"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	WaitingSeconds    int64      `json:"waiting_seconds,omitempty"`
	HelperDisplayName *string    `json:"helper_display_name,omitempty"`
}

// Candidate is a pending request with its match score for one helper.
type Candidate struct {
	Request
	Score int `json:"match_score"`
}

// MatchScore ranks a request for a helper. Language compatibility is a hard
// requirement (ok=false filters the request out); category match, then waiting
// time (older first) order the rest. Deterministic, no LLM involved.
func MatchScore(language, category string, createdAt time.Time, helperLangs, helperCats []string, now time.Time) (ok bool, score int) {
	langOK := false
	for _, l := range helperLangs {
		if l == language {
			langOK = true
			break
		}
	}
	if !langOK {
		return false, 0
	}
	score = 1000 // language gate passed
	for _, c := range helperCats {
		if c == category {
			score += 100
			break
		}
	}
	// Older requests first: +1 per waiting minute, capped so language and
	// category always dominate recency.
	waiting := int(now.Sub(createdAt).Minutes())
	if waiting < 0 {
		waiting = 0
	}
	// Capped below the category bonus so language, then category, always
	// dominate recency.
	if waiting > 99 {
		waiting = 99
	}
	score += waiting
	return true, score
}

// SortCandidates orders by score desc, then oldest first (stable, testable).
func SortCandidates(in []Candidate) {
	sort.SliceStable(in, func(a, b int) bool {
		if in[a].Score != in[b].Score {
			return in[a].Score > in[b].Score
		}
		return in[a].CreatedAt.Before(in[b].CreatedAt)
	})
}

// ValidCategory reports whether c is a known assistance category.
func ValidCategory(c string) bool {
	for _, k := range Categories {
		if k == c {
			return true
		}
	}
	return false
}

// ValidLanguage reports whether l is a supported assistance language code.
func ValidLanguage(l string) bool {
	switch l {
	case "hi", "en", "ta", "mr", "bn", "te", "kn", "ml", "gu", "pa", "or", "ur":
		return true
	}
	return false
}
