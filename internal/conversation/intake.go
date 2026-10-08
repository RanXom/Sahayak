package conversation

// QuestionPriority is the order in which intake questions are asked.
var QuestionPriority = []string{
	"state",
	"gender",
	"age",
	"marital_status",
	"income_annual",
	"occupation",
	"residence_type",
	"caste_category",
	"ration_card_type",
	"land_acres",
	"has_bank_account",
	"has_aadhaar",
	"family_members",
}

// requiredForReady are the facts considered sufficient to run a first evaluation.
var requiredForReady = map[string]bool{
	"state": true, "gender": true, "age": true,
	"marital_status": true, "income_annual": true,
	"occupation": true, "residence_type": true,
}

// MaxTurns is the maximum number of assistant follow-up questions per conversation.
const MaxTurns = 15

// NextField returns the next fact to ask about, or "" when intake is complete.
// A fact is skipped if it is already present or was already asked.
func NextField(facts map[string]any, asked []string) string {
	askedSet := make(map[string]bool, len(asked))
	for _, a := range asked {
		askedSet[a] = true
	}
	for _, f := range QuestionPriority {
		if _, present := facts[f]; present {
			continue
		}
		if f == "state" {
			// State is captured as state_id on the profile, not a free fact.
			continue
		}
		if askedSet[f] {
			continue
		}
		return f
	}
	return ""
}

// IsReady reports whether enough core facts exist to run eligibility.
func IsReady(facts map[string]any) bool {
	for f := range requiredForReady {
		if f == "state" {
			continue
		}
		if _, ok := facts[f]; !ok {
			return false
		}
	}
	return true
}

// MissingCore returns the core facts still missing (excluding state).
func MissingCore(facts map[string]any) []string {
	var out []string
	for _, f := range QuestionPriority {
		if f == "state" {
			continue
		}
		if !requiredForReady[f] {
			continue
		}
		if _, ok := facts[f]; !ok {
			out = append(out, f)
		}
	}
	return out
}
