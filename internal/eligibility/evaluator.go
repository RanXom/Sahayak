package eligibility

import (
	"strconv"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/rules"
)

// RuleStatus is the three-state outcome of evaluating a rule.
type RuleStatus string

const (
	StatusPass    RuleStatus = "pass"
	StatusFail    RuleStatus = "fail"
	StatusUnknown RuleStatus = "unknown"
	StatusError   RuleStatus = "error"
)

// Scheme match statuses.
const (
	MatchEligible    = "eligible"
	MatchPotential   = "potentially_eligible"
	MatchNotEligible = "not_eligible"
)

// RuleEvaluation is the outcome of a single rule.
type RuleEvaluation struct {
	RuleName      string     `json:"rule"`
	Expression    string     `json:"-"`
	DescriptionEN string     `json:"description_en,omitempty"`
	DescriptionHI string     `json:"description_hi,omitempty"`
	Status        RuleStatus `json:"status"`
	MissingFields []string   `json:"missing_fields,omitempty"`
	Error         string     `json:"error,omitempty"`
}

// SchemeResult is the evaluation outcome for a scheme.
type SchemeResult struct {
	Scheme        *domain.Scheme   `json:"scheme"`
	Status        string           `json:"status"`
	MatchScore    float64          `json:"match_score"`
	PassedRules   int              `json:"passed_rules"`
	TotalRules    int              `json:"total_rules"`
	MissingFields []string         `json:"missing_fields,omitempty"`
	RuleResults   []RuleEvaluation `json:"rule_results"`
}

// FlattenProfile converts profile facts into the evaluation environment,
// adding derived fields and the state code.
func FlattenProfile(facts map[string]any, stateCode string) map[string]any {
	env := make(map[string]any, len(facts)+4)
	for k, v := range facts {
		env[k] = v
	}
	if stateCode != "" {
		env["state_code"] = stateCode
	}

	// Derived: widow.
	if ms, ok := env["marital_status"].(string); ok {
		env["is_widow"] = ms == "widowed"
	}

	// Derived: student child and girl child from family_members.
	if members, ok := env["family_members"].([]any); ok {
		hasStudent := false
		maxClass := 0
		hasGirl := false
		girlAge := 0
		for _, raw := range members {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			isStudent, _ := m["is_student"].(bool)
			if class, ok := toInt(m["class"]); ok && class > 0 {
				isStudent = true
				if class > maxClass {
					maxClass = class
				}
			}
			if isStudent {
				hasStudent = true
			}
			if gender, _ := m["gender"].(string); gender == "female" {
				if age, ok := toInt(m["age"]); ok && age < 10 {
					hasGirl = true
					girlAge = age
				}
			}
		}
		env["has_student_child"] = hasStudent
		env["student_child_class"] = maxClass
		env["has_girl_child"] = hasGirl
		if hasGirl {
			env["girl_child_age"] = girlAge
		}
	}
	return env
}

// EvaluateRule evaluates a single expression against the environment.
func EvaluateRule(expr string, env map[string]any) RuleEvaluation {
	res := RuleEvaluation{Expression: expr}

	fields, err := rules.ExtractFields(expr)
	if err != nil {
		res.Status = StatusError
		res.Error = err.Error()
		return res
	}
	var missing []string
	for _, f := range fields {
		v, ok := env[f]
		if !ok || v == nil {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		res.Status = StatusUnknown
		res.MissingFields = missing
		return res
	}

	ok, err := rules.Evaluate(expr, env)
	if err != nil {
		res.Status = StatusError
		res.Error = err.Error()
		return res
	}
	if ok {
		res.Status = StatusPass
	} else {
		res.Status = StatusFail
	}
	return res
}

// EvaluateScheme evaluates all rules of a scheme against the environment.
func EvaluateScheme(scheme *domain.Scheme, env map[string]any) SchemeResult {
	result := SchemeResult{Scheme: scheme, RuleResults: []RuleEvaluation{}}
	if len(scheme.Rules) == 0 {
		result.Status = MatchNotEligible
		return result
	}

	missingSet := map[string]bool{}
	var missing []string
	mandatoryFail := false
	mandatoryUnknown := false
	passed := 0

	for _, rule := range scheme.Rules {
		ev := EvaluateRule(rule.Expression, env)
		ev.RuleName = rule.RuleName
		ev.DescriptionEN = rule.DescriptionEN
		ev.DescriptionHI = rule.DescriptionHI
		result.RuleResults = append(result.RuleResults, ev)
		result.TotalRules++

		if ev.Status == StatusPass {
			passed++
		}
		if rule.IsMandatory {
			switch ev.Status {
			case StatusFail:
				mandatoryFail = true
			case StatusUnknown:
				mandatoryUnknown = true
				for _, f := range ev.MissingFields {
					if !missingSet[f] {
						missingSet[f] = true
						missing = append(missing, f)
					}
				}
			case StatusError:
				// A broken rule is treated as unknown rather than failing the user.
				mandatoryUnknown = true
			}
		}
	}

	result.PassedRules = passed
	switch {
	case mandatoryFail:
		result.Status = MatchNotEligible
	case mandatoryUnknown:
		result.Status = MatchPotential
		result.MissingFields = missing
	default:
		result.Status = MatchEligible
	}
	result.MatchScore = matchScore(result)
	return result
}

func matchScore(r SchemeResult) float64 {
	if r.TotalRules == 0 {
		return 0
	}
	base := float64(r.PassedRules) / float64(r.TotalRules)
	switch r.Status {
	case MatchEligible:
		return 0.7 + 0.3*base
	case MatchPotential:
		return 0.3 + 0.3*base
	default:
		return 0.3 * base
	}
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i, true
		}
	}
	return 0, false
}

// parseBenefitValue extracts an approximate numeric value from a benefit string
// such as "₹1,500/month" or "up to ₹2,00,000" for ranking purposes.
func parseBenefitValue(s string) float64 {
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
	if cleaned == "" {
		return 0
	}
	v, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return v
}
