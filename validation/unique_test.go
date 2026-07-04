package validation

import (
	"testing"
)

func TestArrayUniqueItems(t *testing.T) {
	if r := newIntArray().UniqueItems().Validate(readIntArray("[1,2,3]")); !r.IsValid() {
		t.Errorf("distinct items should pass: %s", resultJSON(r))
	}
	if r := newIntArray().UniqueItems().Validate(readIntArray("[]")); !r.IsValid() {
		t.Errorf("empty array should pass uniqueItems")
	}
	if r := newIntArray().UniqueItems().Validate(readIntArray("[1,2,2]")); r.IsValid() {
		t.Errorf("duplicate items should fail")
	} else {
		mustContain(t, "unique", resultJSON(r), `"UNIQUE_ITEMS"`)
	}
	// duplicates only flagged when requested
	if r := newIntArray().Validate(readIntArray("[1,1,1]")); !r.IsValid() {
		t.Errorf("without UniqueItems duplicates are fine: %s", resultJSON(r))
	}
}
