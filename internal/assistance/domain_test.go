package assistance

import (
	"testing"
	"time"
)

func TestMatchScore_LanguageGate(t *testing.T) {
	now := time.Now()
	ok, _ := MatchScore("mr", "application_help", now.Add(-time.Hour), []string{"hi", "en"}, []string{"application_help"}, now)
	if ok {
		t.Fatal("request must be filtered out when helper lacks the language")
	}
	ok, score := MatchScore("mr", "application_help", now.Add(-time.Hour), []string{"mr"}, []string{"scheme_discovery"}, now)
	if !ok || score <= 1000 {
		t.Fatalf("language match must pass with base score, got ok=%v score=%d", ok, score)
	}
}

func TestMatchScore_CategoryAndWaitingOrder(t *testing.T) {
	now := time.Now()
	_, catScore := MatchScore("hi", "application_help", now.Add(-time.Minute), []string{"hi"}, []string{"application_help"}, now)
	_, noCatScore := MatchScore("hi", "application_help", now.Add(-time.Minute), []string{"hi"}, []string{"scheme_discovery"}, now)
	if catScore <= noCatScore {
		t.Fatalf("category match must outrank mismatch: %d vs %d", catScore, noCatScore)
	}
	_, oldScore := MatchScore("hi", "scheme_discovery", now.Add(-2*time.Hour), []string{"hi"}, []string{"scheme_discovery"}, now)
	_, newScore := MatchScore("hi", "scheme_discovery", now.Add(-time.Minute), []string{"hi"}, []string{"scheme_discovery"}, now)
	if oldScore <= newScore {
		t.Fatalf("older request must outrank newer: %d vs %d", oldScore, newScore)
	}
	// Language+category must dominate any waiting-time difference.
	_, best := MatchScore("hi", "application_help", now.Add(-time.Minute), []string{"hi"}, []string{"application_help"}, now)
	_, worst := MatchScore("hi", "scheme_discovery", now.Add(-10*time.Hour), []string{"hi"}, []string{"document_help"}, now)
	if best <= worst {
		t.Fatalf("category must dominate waiting time: %d vs %d", best, worst)
	}
}

func TestSortCandidates_OldestBreaksTies(t *testing.T) {
	now := time.Now()
	a := Candidate{Request: Request{ID: "a", CreatedAt: now.Add(-time.Hour)}, Score: 1100}
	b := Candidate{Request: Request{ID: "b", CreatedAt: now.Add(-2 * time.Hour)}, Score: 1100}
	in := []Candidate{a, b}
	SortCandidates(in)
	if in[0].ID != "b" {
		t.Fatalf("tie must break oldest-first, got %s first", in[0].ID)
	}
}

func TestValidLanguage_CoversSarvamSet(t *testing.T) {
	for _, l := range []string{"hi", "en", "ta", "mr", "bn", "te", "kn", "ml", "gu", "pa", "or", "ur"} {
		if !ValidLanguage(l) {
			t.Fatalf("expected %s to be valid", l)
		}
	}
	if ValidLanguage("xx") {
		t.Fatal("xx must be invalid")
	}
}
