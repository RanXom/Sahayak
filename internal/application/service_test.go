package application

import (
	"encoding/json"
	"testing"

	"github.com/sahayak/sahayak/internal/domain"
)

// Regression test: missing_fields must encode as [] (not null) when nothing
// is missing — the frontend reads .length on it and crashes on null.
func TestMissingRequired_EmptyEncodesAsArray(t *testing.T) {
	fields := []domain.ApplicationField{
		{FieldName: "full_name", IsRequired: true},
	}
	got := missingRequired(fields, map[string]any{"full_name": "Test User"})
	if got == nil {
		t.Fatal("missingRequired returned nil, want empty slice")
	}
	b, err := json.Marshal(map[string]any{"missing_fields": got})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"missing_fields":[]}` {
		t.Fatalf("got %s, want empty JSON array", b)
	}
}
