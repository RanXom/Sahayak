package profile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
)

// numericFacts are fact keys that must be numeric.
var numericFacts = map[string]bool{
	"age": true, "income_annual": true, "land_acres": true,
	"family_size": true, "disability_percentage": true,
	"girl_child_age": true, "student_child_class": true,
}

// boolFacts are fact keys that must be boolean.
var boolFacts = map[string]bool{
	"has_bank_account": true, "has_aadhaar": true, "minority_status": true,
	"has_pucca_house": true, "has_electricity": true,
	"is_pregnant_or_lactating": true, "has_girl_child": true,
	"has_student_child": true, "is_widow": true,
}

// stringFacts are fact keys that must be strings (free text or identifiers).
var stringFacts = map[string]bool{
	"gender": true, "marital_status": true, "occupation": true,
	"caste_category": true, "residence_type": true, "ration_card_type": true,
	"education_level": true, "disability_status": true, "employment_status": true,
	"full_name": true, "address": true, "bank_account_number": true,
	"aadhaar_number": true, "ration_card_number": true, "state": true,
}

// enumFacts constrains string facts to a known set of values.
var enumFacts = map[string][]string{
	"gender":            {"male", "female", "other"},
	"marital_status":    {"single", "married", "widowed", "divorced", "separated"},
	"caste_category":    {"general", "obc", "sc", "st", "ews"},
	"residence_type":    {"rural", "urban"},
	"ration_card_type":  {"apl", "bpl", "aay", "none"},
	"education_level":   {"illiterate", "primary", "secondary", "higher_secondary", "graduate", "post_graduate"},
	"disability_status": {"none", "moderate", "severe", "multiple"},
	"employment_status": {"employed", "unemployed", "self_employed", "student", "retired", "homemaker"},
}

// AllowedFactKeys returns the set of fact keys accepted from clients/LLMs.
func AllowedFactKeys() map[string]bool {
	out := make(map[string]bool, len(numericFacts)+len(boolFacts)+len(stringFacts))
	for k := range numericFacts {
		out[k] = true
	}
	for k := range boolFacts {
		out[k] = true
	}
	for k := range stringFacts {
		out[k] = true
	}
	out["family_members"] = true
	return out
}

// ValidateFacts rejects unknown keys and type/enum violations. Unknown keys are
// dropped by the caller (LLM hallucination mitigation); here we return an error
// listing them so the API can respond with a clear message.
func ValidateFacts(facts map[string]any) error {
	for k, v := range facts {
		if k == "family_members" {
			if err := validateFamilyMembers(v); err != nil {
				return err
			}
			continue
		}
		switch {
		case numericFacts[k]:
			if _, ok := toFloat(v); !ok {
				return fmt.Errorf("fact %q must be numeric", k)
			}
		case boolFacts[k]:
			if _, ok := v.(bool); !ok {
				return fmt.Errorf("fact %q must be true or false", k)
			}
		case stringFacts[k]:
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("fact %q must be text", k)
			}
			if allowed, ok := enumFacts[k]; ok && !contains(allowed, s) {
				return fmt.Errorf("fact %q has an invalid value", k)
			}
		default:
			return fmt.Errorf("unknown fact %q", k)
		}
	}
	return nil
}

// FilterKnownFacts drops unknown keys and normalizes numeric values.
func FilterKnownFacts(facts map[string]any) map[string]any {
	allowed := AllowedFactKeys()
	out := make(map[string]any, len(facts))
	for k, v := range facts {
		if !allowed[k] {
			continue
		}
		if numericFacts[k] {
			if f, ok := toFloat(v); ok {
				out[k] = f
				continue
			}
		}
		out[k] = v
	}
	return out
}

func validateFamilyMembers(v any) error {
	list, ok := v.([]any)
	if !ok {
		return fmt.Errorf("fact \"family_members\" must be a list")
	}
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("each family member must be an object")
		}
		for k := range m {
			switch k {
			case "relation", "age", "education", "gender", "is_student", "class":
			default:
				return fmt.Errorf("unknown family member field %q", k)
			}
		}
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// keyAliases maps common LLM-produced keys to the controlled vocabulary.
var keyAliases = map[string]string{
	"state_of_residence": "state",
	"state_name":         "state",
	"residence_state":    "state",
	"annual_income":      "income_annual",
	"family_income":      "income_annual",
	"household_income":   "income_annual",
	"income":             "income_annual",
	"land_holding_acres": "land_acres",
	"land_holding":       "land_acres",
	"land_size":          "land_acres",
	"land_area":          "land_acres",
	"caste":              "caste_category",
	"social_category":    "caste_category",
	"category":           "caste_category",
	"residential_area":   "residence_type",
	"area_type":          "residence_type",
	"has_bank":           "has_bank_account",
	"bank_account":       "has_bank_account",
	"has_aadhaar_card":   "has_aadhaar",
	"aadhaar":            "has_aadhaar",
	"education":          "education_level",
	"disability":         "disability_status",
	"is_minority":        "minority_status",
	"ration_card":        "ration_card_type",
	"bpl_card":           "ration_card_type",
}

// valueAliases normalizes enum-like string values to their canonical form.
var valueAliases = map[string]map[string]string{
	"gender": {
		"woman": "female", "women": "female", "female": "female", "girl": "female",
		"man": "male", "men": "male", "male": "male", "boy": "male",
		"other": "other", "transgender": "other",
	},
	"marital_status": {
		"widow": "widowed", "widowed": "widowed", "widowhood": "widowed",
		"unmarried": "single", "single": "single", "never married": "single",
		"married": "married", "wife": "married", "husband": "married",
		"divorced": "divorced", "separated": "separated",
	},
	"caste_category": {
		"general": "general", "gen": "general", "ur": "general", "unreserved": "general",
		"obc": "obc", "other backward class": "obc", "other backward caste": "obc",
		"sc": "sc", "scheduled caste": "sc", "scheduled_caste": "sc",
		"st": "st", "scheduled tribe": "st", "scheduled_tribe": "st", "tribal": "st",
		"ews": "ews",
	},
	"residence_type": {
		"rural": "rural", "village": "rural", "village area": "rural",
		"urban": "urban", "city": "urban", "town": "urban",
	},
	"ration_card_type": {
		"bpl": "bpl", "below poverty line": "bpl", "apl": "apl", "above poverty line": "apl",
		"aay": "aay", "antyodaya": "aay", "antyodaya anna yojana": "aay",
		"none": "none", "no": "none",
	},
	"disability_status": {
		"none": "none", "no": "none", "moderate": "moderate", "severe": "severe", "multiple": "multiple",
	},
}

var occupationAliases = map[string]string{
	"farming": "farmer", "agriculture": "farmer", "agriculturalist": "farmer", "farmer": "farmer",
	"agri": "farmer", "cultivator": "farmer",
	"street vendor": "street_vendor", "street_vendor": "street_vendor", "vendor": "street_vendor",
	"hawker": "street_vendor", "street hawker": "street_vendor",
	"fisherman": "fisher", "fisher": "fisher", "fish farmer": "fish_farmer", "fish_farmer": "fish_farmer",
	"daily wage": "labourer", "daily wage worker": "labourer", "labourer": "labourer",
	"laborer": "labourer", "worker": "labourer", "farm labourer": "labourer",
	"student": "student", "homemaker": "homemaker", "housewife": "homemaker",
	"unemployed": "unemployed", "self employed": "self_employed", "self_employed": "self_employed",
}

// SanitizeExtractedFacts normalizes untrusted (LLM) facts into the controlled
// vocabulary and drops anything that cannot be mapped. It never errors, so an
// imperfect extraction cannot break the conversation.
func SanitizeExtractedFacts(raw map[string]any) map[string]any {
	out := make(map[string]any, len(raw))
	for k, v := range raw {
		key := strings.ToLower(strings.TrimSpace(k))

		// Boolean BPL flags map to a ration card type.
		if key == "has_bpl_card" || key == "bpl_card" || key == "bpl" {
			if b, ok := v.(bool); ok && b {
				out["ration_card_type"] = "bpl"
			}
			continue
		}

		if alias, ok := keyAliases[key]; ok {
			key = alias
		}
		if !AllowedFactKeys()[key] {
			continue
		}

		if key == "occupation" {
			if s, ok := v.(string); ok {
				if canon, ok := occupationAliases[strings.ToLower(strings.TrimSpace(s))]; ok {
					out[key] = canon
				} else {
					out[key] = strings.ToLower(strings.TrimSpace(s))
				}
			}
			continue
		}

		if nv, ok := normalizeValue(key, v); ok {
			out[key] = nv
		}
	}
	return out
}

func normalizeValue(key string, v any) (any, bool) {
	if numericFacts[key] {
		if f, ok := toFloat(v); ok {
			return f, true
		}
		if s, ok := v.(string); ok {
			if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
				return f, true
			}
		}
		return nil, false
	}
	if boolFacts[key] {
		if b, ok := v.(bool); ok {
			return b, true
		}
		if s, ok := v.(string); ok {
			switch strings.ToLower(strings.TrimSpace(s)) {
			case "yes", "true", "haan", "हाँ", "हां":
				return true, true
			case "no", "false", "nahi", "नहीं":
				return false, true
			}
		}
		return nil, false
	}
	if key == "family_members" {
		if _, ok := v.([]any); ok {
			return v, true
		}
		return nil, false
	}

	s, ok := v.(string)
	if !ok {
		return nil, false
	}
	s = strings.TrimSpace(s)
	if allowed, isEnum := enumFacts[key]; isEnum {
		lower := strings.ToLower(s)
		if m, ok := valueAliases[key]; ok {
			if canon, ok := m[lower]; ok {
				return canon, true
			}
		}
		for _, a := range allowed {
			if lower == a {
				return a, true
			}
		}
		return nil, false
	}
	return s, true
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case nil:
		return 0, false
	default:
		return 0, false
	}
}

// DocumentTypes is the set of document types the API accepts.
var DocumentTypes = map[string]bool{
	"aadhaar": true, "income_cert": true, "domicile_cert": true, "caste_cert": true,
	"bank_account": true, "ration_card": true, "voter_id": true, "birth_cert": true,
	"pan_card": true, "disability_cert": true, "land_record": true, "student_cert": true,
	"death_cert": true, "marriage_cert": true,
}

// ValidateDocumentType reports whether a document type is supported.
func ValidateDocumentType(t string) error {
	if !DocumentTypes[t] {
		return domain.InvalidInput("Unsupported document type: " + t)
	}
	return nil
}
