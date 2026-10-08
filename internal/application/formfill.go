package application

import (
	"strconv"
	"strings"

	"github.com/sahayak/sahayak/internal/domain"
)

// AutoFill maps a scheme's application fields to profile facts.
func AutoFill(fields []domain.ApplicationField, facts map[string]any) map[string]any {
	out := make(map[string]any)
	for _, f := range fields {
		if f.ProfileMapping == "" {
			continue
		}
		if v, ok := extractFromProfile(facts, f.ProfileMapping); ok {
			out[f.FieldName] = v
		}
	}
	return out
}

// extractFromProfile resolves a mapping path such as "age" or
// "family_members[0].age" against the facts object.
func extractFromProfile(facts map[string]any, path string) (any, bool) {
	if path == "" {
		return nil, false
	}
	segments := strings.Split(path, ".")
	var current any = facts
	for _, seg := range segments {
		name, index, hasIndex := parseSegment(seg)
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[name]
		if !ok {
			return nil, false
		}
		if hasIndex {
			list, ok := v.([]any)
			if !ok || index < 0 || index >= len(list) {
				return nil, false
			}
			v = list[index]
		}
		current = v
	}
	return current, current != nil
}

func parseSegment(seg string) (name string, index int, hasIndex bool) {
	open := strings.IndexByte(seg, '[')
	if open < 0 || !strings.HasSuffix(seg, "]") {
		return seg, 0, false
	}
	name = seg[:open]
	idxStr := seg[open+1 : len(seg)-1]
	i, err := strconv.Atoi(idxStr)
	if err != nil {
		return seg, 0, false
	}
	return name, i, true
}
