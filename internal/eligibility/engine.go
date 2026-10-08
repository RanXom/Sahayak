package eligibility

import (
	"sort"

	"github.com/sahayak/sahayak/internal/domain"
)

// Engine evaluates a profile environment against a set of schemes.
type Engine struct{}

// NewEngine constructs an eligibility Engine.
func NewEngine() *Engine { return &Engine{} }

// Evaluate runs every scheme through the rule evaluator and returns ranked results.
func (e *Engine) Evaluate(env map[string]any, schemes []*domain.Scheme) []SchemeResult {
	results := make([]SchemeResult, 0, len(schemes))
	for _, s := range schemes {
		results = append(results, EvaluateScheme(s, env))
	}
	Rank(results)
	return results
}

// Rank orders results: eligible first, then potentially eligible, then
// not eligible; within a tier by benefit value and match score.
func Rank(results []SchemeResult) {
	statusRank := map[string]int{
		MatchEligible:    0,
		MatchPotential:   1,
		MatchNotEligible: 2,
	}
	sort.SliceStable(results, func(i, j int) bool {
		ri, rj := results[i], results[j]
		if statusRank[ri.Status] != statusRank[rj.Status] {
			return statusRank[ri.Status] < statusRank[rj.Status]
		}
		bi := parseBenefitValue(ri.Scheme.BenefitAmount)
		bj := parseBenefitValue(rj.Scheme.BenefitAmount)
		if bi != bj {
			return bi > bj
		}
		if ri.MatchScore != rj.MatchScore {
			return ri.MatchScore > rj.MatchScore
		}
		return ri.Scheme.NameEN < rj.Scheme.NameEN
	})
}

// Summarize returns counts by status.
func Summarize(results []SchemeResult) map[string]int {
	summary := map[string]int{
		MatchEligible:     0,
		MatchPotential:    0,
		MatchNotEligible:  0,
		"total_evaluated": len(results),
	}
	for _, r := range results {
		summary[r.Status]++
	}
	return summary
}
