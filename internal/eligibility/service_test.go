package eligibility

import (
	"strings"
	"testing"

	"github.com/sahayak/sahayak/internal/document"
	"github.com/sahayak/sahayak/internal/domain"
)

func TestBuildVoiceSummary_ListsMatchesDocsAndMissingFacts(t *testing.T) {
	res := &EvaluateResponse{
		Results: []ResultView{
			{
				Scheme:      &domain.Scheme{NameEN: "Widow Pension", NameHI: "विधवा पेंशन"},
				Status:      MatchEligible,
				RuleResults: []RuleEvaluation{{Status: StatusPass}, {Status: StatusPass}},
				Documents: &document.CheckResult{
					MissingCount: 2,
					Documents: []document.DocumentStatus{
						{DocumentName: "Income Certificate", IsMandatory: true, UserHas: false},
						{DocumentName: "Death Certificate", IsMandatory: true, UserHas: false},
						{DocumentName: "Aadhaar Card", IsMandatory: true, UserHas: true},
					},
				},
			},
			{
				Scheme:        &domain.Scheme{NameEN: "PMAY-G", NameHI: "पीएमआवास-ग्रामीण"},
				Status:        MatchPotential,
				MissingFields: []string{"has_pucca_house"},
				RuleResults:   []RuleEvaluation{{Status: StatusPass}, {Status: StatusUnknown}},
			},
		},
	}

	en := BuildVoiceSummary(res, "en")
	for _, want := range []string{"Widow Pension", "100%", "Income Certificate", "pucca house"} {
		if !strings.Contains(en, want) {
			t.Errorf("English summary missing %q:\n%s", want, en)
		}
	}

	hi := BuildVoiceSummary(res, "hi")
	for _, want := range []string{"विधवा पेंशन", "100%", "पक्का घर"} {
		if !strings.Contains(hi, want) {
			t.Errorf("Hindi summary missing %q:\n%s", want, hi)
		}
	}
}

func TestBuildMissingSummary(t *testing.T) {
	res := &EvaluateResponse{
		Results: []ResultView{
			{
				Scheme:        &domain.Scheme{NameEN: "PMAY-G", NameHI: "पीएमआवास-ग्रामीण"},
				Status:        MatchPotential,
				MissingFields: []string{"has_pucca_house", "income_annual"},
			},
			{
				Scheme: &domain.Scheme{NameEN: "Widow Pension"},
				Status: MatchEligible,
				Documents: &document.CheckResult{
					Documents: []document.DocumentStatus{
						{DocumentName: "Income Certificate", IsMandatory: true, UserHas: false},
					},
				},
			},
		},
	}
	en := BuildMissingSummary(res, "en")
	for _, want := range []string{"pucca house", "annual family income", "Income Certificate"} {
		if !strings.Contains(en, want) {
			t.Errorf("missing summary missing %q:\n%s", want, en)
		}
	}
}

func TestMatchPercent(t *testing.T) {
	eligible := ResultView{Status: MatchEligible}
	if got := matchPercent(eligible); got != 100 {
		t.Errorf("eligible = %d, want 100", got)
	}
	potential := ResultView{
		Status:      MatchPotential,
		RuleResults: []RuleEvaluation{{Status: StatusPass}, {Status: StatusPass}, {Status: StatusUnknown}},
	}
	if got := matchPercent(potential); got != 66 {
		t.Errorf("potential = %d, want 66", got)
	}
}
