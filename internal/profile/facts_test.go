package profile

import "testing"

func TestSanitizeExtractedFacts_NormalizesLLMOutput(t *testing.T) {
	// This mirrors real output from Gemini for the sentence:
	// "I am a 45 year old widow from Rajasthan, I farm 2 acres and have a BPL card"
	raw := map[string]any{
		"age":                45,
		"marital_status":     "widow",
		"state_of_residence": "Rajasthan",
		"occupation":         "farming",
		"land_holding_acres": 2.0,
		"has_bpl_card":       true,
	}

	got := SanitizeExtractedFacts(raw)

	want := map[string]any{
		"age":              float64(45),
		"marital_status":   "widowed",
		"state":            "Rajasthan",
		"occupation":       "farmer",
		"land_acres":       float64(2),
		"ration_card_type": "bpl",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d facts %v, want %d %v", len(got), got, len(want), want)
	}
	for k, wv := range want {
		if got[k] != wv {
			t.Errorf("fact %q = %v, want %v", k, got[k], wv)
		}
	}
}

func TestSanitizeExtractedFacts_DropsUnknownAndInvalid(t *testing.T) {
	raw := map[string]any{
		"favourite_colour": "blue",     // unknown key
		"gender":           "woman",    // alias -> female
		"caste_category":   "nonsense", // invalid enum -> dropped
		"has_aadhaar":      "yes",      // string bool -> true
		"income_annual":    "120000",   // numeric string -> parsed
		"residence_type":   "village",  // alias -> rural
	}
	got := SanitizeExtractedFacts(raw)

	if _, ok := got["favourite_colour"]; ok {
		t.Error("unknown key should be dropped")
	}
	if _, ok := got["caste_category"]; ok {
		t.Error("invalid enum value should be dropped")
	}
	if got["gender"] != "female" {
		t.Errorf("gender = %v, want female", got["gender"])
	}
	if got["residence_type"] != "rural" {
		t.Errorf("residence_type = %v, want rural", got["residence_type"])
	}
	if got["has_aadhaar"] != true {
		t.Errorf("has_aadhaar = %v, want true", got["has_aadhaar"])
	}
	if got["income_annual"] != float64(120000) {
		t.Errorf("income_annual = %v, want 120000", got["income_annual"])
	}
}

func TestValidateFacts_RejectsUnknown(t *testing.T) {
	if err := ValidateFacts(map[string]any{"made_up": 1}); err == nil {
		t.Fatal("expected error for unknown fact")
	}
	if err := ValidateFacts(map[string]any{"gender": "female", "age": 30}); err != nil {
		t.Fatalf("valid facts rejected: %v", err)
	}
}
