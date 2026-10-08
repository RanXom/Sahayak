package eligibility

import (
	"testing"

	"github.com/sahayak/sahayak/internal/domain"
)

func ptr[T any](v T) *T { return &v }

func TestEvaluateRule_ThreeStates(t *testing.T) {
	expr := `age >= 40 && gender == "female" && marital_status == "widowed" && income_annual <= 200000`

	tests := []struct {
		name string
		env  map[string]any
		want RuleStatus
	}{
		{
			name: "eligible widow",
			env:  map[string]any{"age": 45, "gender": "female", "marital_status": "widowed", "income_annual": 120000},
			want: StatusPass,
		},
		{
			name: "too young",
			env:  map[string]any{"age": 30, "gender": "female", "marital_status": "widowed", "income_annual": 120000},
			want: StatusFail,
		},
		{
			name: "missing income is unknown not fail",
			env:  map[string]any{"age": 45, "gender": "female", "marital_status": "widowed"},
			want: StatusUnknown,
		},
		{
			name: "missing everything is unknown",
			env:  map[string]any{},
			want: StatusUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateRule(expr, tc.env)
			if got.Status != tc.want {
				t.Fatalf("status = %s, want %s (missing=%v err=%s)", got.Status, tc.want, got.MissingFields, got.Error)
			}
		})
	}
}

func TestEvaluateRule_BoundaryValues(t *testing.T) {
	expr := `age >= 60`
	if got := EvaluateRule(expr, map[string]any{"age": 60}); got.Status != StatusPass {
		t.Fatalf("age 60 should pass, got %s", got.Status)
	}
	if got := EvaluateRule(expr, map[string]any{"age": 59}); got.Status != StatusFail {
		t.Fatalf("age 59 should fail, got %s", got.Status)
	}
}

func TestEvaluateRule_InOperator(t *testing.T) {
	expr := `caste_category in ["sc", "st"]`
	if got := EvaluateRule(expr, map[string]any{"caste_category": "sc"}); got.Status != StatusPass {
		t.Fatalf("sc should pass, got %s", got.Status)
	}
	if got := EvaluateRule(expr, map[string]any{"caste_category": "general"}); got.Status != StatusFail {
		t.Fatalf("general should fail, got %s", got.Status)
	}
}

func TestFlattenProfile_DerivedFields(t *testing.T) {
	facts := map[string]any{
		"marital_status": "widowed",
		"family_members": []any{
			map[string]any{"relation": "son", "is_student": true, "class": 11},
			map[string]any{"relation": "daughter", "gender": "female", "age": 7},
		},
	}
	env := FlattenProfile(facts, "RJ")
	if env["is_widow"] != true {
		t.Errorf("is_widow = %v, want true", env["is_widow"])
	}
	if env["has_student_child"] != true {
		t.Errorf("has_student_child = %v, want true", env["has_student_child"])
	}
	if env["student_child_class"] != 11 {
		t.Errorf("student_child_class = %v, want 11", env["student_child_class"])
	}
	if env["has_girl_child"] != true {
		t.Errorf("has_girl_child = %v, want true", env["has_girl_child"])
	}
	if env["girl_child_age"] != 7 {
		t.Errorf("girl_child_age = %v, want 7", env["girl_child_age"])
	}
	if env["state_code"] != "RJ" {
		t.Errorf("state_code = %v, want RJ", env["state_code"])
	}
}

func TestEvaluateScheme_Classification(t *testing.T) {
	scheme := &domain.Scheme{
		NameEN: "Widow Pension",
		Rules: []domain.EligibilityRule{
			{RuleName: "Age >= 40", Expression: "age >= 40", IsMandatory: true},
			{RuleName: "Widow", Expression: `marital_status == "widowed"`, IsMandatory: true},
			{RuleName: "Income", Expression: "income_annual <= 200000", IsMandatory: true},
		},
	}

	t.Run("eligible", func(t *testing.T) {
		env := map[string]any{"age": 45, "marital_status": "widowed", "income_annual": 100000}
		got := EvaluateScheme(scheme, env)
		if got.Status != MatchEligible {
			t.Fatalf("status = %s, want eligible", got.Status)
		}
		if got.PassedRules != 3 {
			t.Fatalf("passed = %d, want 3", got.PassedRules)
		}
	})

	t.Run("not eligible on fail", func(t *testing.T) {
		env := map[string]any{"age": 30, "marital_status": "widowed", "income_annual": 100000}
		got := EvaluateScheme(scheme, env)
		if got.Status != MatchNotEligible {
			t.Fatalf("status = %s, want not_eligible", got.Status)
		}
	})

	t.Run("potentially eligible on missing", func(t *testing.T) {
		env := map[string]any{"age": 45, "marital_status": "widowed"}
		got := EvaluateScheme(scheme, env)
		if got.Status != MatchPotential {
			t.Fatalf("status = %s, want potentially_eligible", got.Status)
		}
		if len(got.MissingFields) != 1 || got.MissingFields[0] != "income_annual" {
			t.Fatalf("missing = %v, want [income_annual]", got.MissingFields)
		}
	})
}

func TestRank_EligibleBeforePotential(t *testing.T) {
	results := []SchemeResult{
		{Status: MatchNotEligible, Scheme: &domain.Scheme{NameEN: "C"}},
		{Status: MatchPotential, Scheme: &domain.Scheme{NameEN: "B"}},
		{Status: MatchEligible, Scheme: &domain.Scheme{NameEN: "A"}},
	}
	Rank(results)
	if results[0].Status != MatchEligible || results[1].Status != MatchPotential || results[2].Status != MatchNotEligible {
		t.Fatalf("unexpected order: %s, %s, %s", results[0].Status, results[1].Status, results[2].Status)
	}
}
