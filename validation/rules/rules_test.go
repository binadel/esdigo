package rules

import "testing"

// allCodes is every exported rule code, used to assert they stay distinct and
// non-empty (a duplicated or empty code would make failure reports ambiguous).
var allCodes = []Code{
	Required, NotNull, Boolean, Number, Integer, OutOfRange, String, Array, Object,
	Length, MinLength, MaxLength, Pattern, Regex, Date, Time, DateTime, Duration,
	Email, IP, IPv4, IPv6, Uri, Uuid, UuidVersion, ExactItems, MinItems, MaxItems,
	Minimum, Maximum, ExclusiveMinimum, ExclusiveMaximum, MultipleOf, Enum, Const,
	UniqueItems, Hostname, UriReference, JsonPointer,
}

// TestCodeValues spot-checks the wire strings a few codes serialize to.
func TestCodeValues(t *testing.T) {
	cases := map[Code]string{
		Required:    "REQUIRED",
		NotNull:     "NOT_NULL",
		IPv4:        "IPv4",
		MultipleOf:  "MULTIPLE_OF",
		JsonPointer: "JSON_POINTER",
	}
	for code, want := range cases {
		if string(code) != want {
			t.Errorf("code = %q, want %q", string(code), want)
		}
	}
}

// TestCodesUniqueAndNonEmpty guards against a copy-paste duplicate or an empty code.
func TestCodesUniqueAndNonEmpty(t *testing.T) {
	seen := make(map[Code]bool, len(allCodes))
	for _, c := range allCodes {
		if c == "" {
			t.Errorf("found an empty code")
		}
		if seen[c] {
			t.Errorf("duplicate code %q", c)
		}
		seen[c] = true
	}
}
