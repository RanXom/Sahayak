package rules

import "testing"

func TestValidateAndExtract(t *testing.T) {
	fields, err := ExtractFields(`age >= 40 && gender == "female" && state_code == "RJ"`)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	want := map[string]bool{"age": true, "gender": true, "state_code": true}
	if len(fields) != len(want) {
		t.Fatalf("fields = %v, want %v", fields, want)
	}
	for _, f := range fields {
		if !want[f] {
			t.Errorf("unexpected field %q", f)
		}
	}
}

func TestValidateRejectsUnknownField(t *testing.T) {
	if err := ValidateVariables(`age >= 40 && made_up_field == true`); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestValidateRejectsSyntaxError(t *testing.T) {
	if err := Validate(`age >=`); err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestEvaluate(t *testing.T) {
	ok, err := Evaluate(`age >= 18`, map[string]any{"age": 20})
	if err != nil || !ok {
		t.Fatalf("expected true, got %v err=%v", ok, err)
	}
	ok, err = Evaluate(`caste_category in ["sc","st"]`, map[string]any{"caste_category": "general"})
	if err != nil || ok {
		t.Fatalf("expected false, got %v err=%v", ok, err)
	}
}
