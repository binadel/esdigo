package validation

import (
	"testing"

	"github.com/binadel/esdigo/json"
)

// TestPropertiesValidate exercises the min/max boundaries and that the count is
// echoed on the result.
func TestPropertiesValidate(t *testing.T) {
	cases := []struct {
		name  string
		v     *Properties
		count int
		valid bool
	}{
		{"min met", NewProperties().Min(1), 1, true},
		{"min unmet", NewProperties().Min(2), 1, false},
		{"max met", NewProperties().Max(3), 3, true},
		{"max exceeded", NewProperties().Max(2), 3, false},
		{"in range", NewProperties().Min(1).Max(3), 2, true},
		{"no bounds", NewProperties(), 0, true},
	}
	for _, tc := range cases {
		r := tc.v.Validate(tc.count)
		if r.IsValid() != tc.valid {
			t.Errorf("%s: IsValid() = %v, want %v", tc.name, r.IsValid(), tc.valid)
		}
		if r.Value != tc.count {
			t.Errorf("%s: Value = %d, want %d", tc.name, r.Value, tc.count)
		}
	}
}

// TestPropertiesErrorJSON checks the failure serializes at the object's path with
// the code and the bound parameter.
func TestPropertiesErrorJSON(t *testing.T) {
	r := NewProperties("obj").Min(2).Validate(1)
	w := json.NewWriter(64)
	r.WriteJSON(w)
	b, _ := w.Build()
	want := `{"path":["obj"],"errors":[{"code":"MIN_PROPERTIES","message":"object must have at least the minimum number of properties","minProperties":2}]}`
	if got := string(b); got != want {
		t.Errorf("JSON = %s, want %s", got, want)
	}
}
