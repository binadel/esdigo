package validation

import (
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/binadel/esdigo/json"
)

func pathJSON(f FieldPath) string {
	w := json.NewWriter(32)
	f.WriteJSON(w)
	return string(w.Bytes())
}

func TestFieldPathRepresentations(t *testing.T) {
	orig := FieldPathRepresentation
	defer func() { FieldPathRepresentation = orig }()

	FieldPathRepresentation = PathRepresentationArray
	if got := pathJSON(Field([]string{"user", "age"})); got != `["user","age"]` {
		t.Errorf("array: %s", got)
	}
	FieldPathRepresentation = PathRepresentationDotted
	if got := pathJSON(Field([]string{"user", "age"})); got != `"user.age"` {
		t.Errorf("dotted: %s", got)
	}
	FieldPathRepresentation = PathRepresentationSlashed
	if got := pathJSON(Field([]string{"user", "age"})); got != `"user/age"` {
		t.Errorf("slashed: %s", got)
	}
}

func TestFieldPathSegments(t *testing.T) {
	segments := []string{"user", "address", "street"}
	if got := Field(segments).Segments(); !reflect.DeepEqual(got, segments) {
		t.Errorf("segments: got %v want %v", got, segments)
	}
}

// TestFieldPathEscaping is the regression guard: a segment containing a quote or
// backslash must be escaped so the emitted path stays valid JSON.
func TestFieldPathEscaping(t *testing.T) {
	orig := FieldPathRepresentation
	defer func() { FieldPathRepresentation = orig }()

	FieldPathRepresentation = PathRepresentationArray
	got := pathJSON(Field([]string{`a"b`, `c\d`}))
	if want := `["a\"b","c\\d"]`; got != want {
		t.Errorf("array escaping: got %s want %s", got, want)
	}
	if !stdjson.Valid([]byte(got)) {
		t.Errorf("array escaped path is not valid JSON: %s", got)
	}

	FieldPathRepresentation = PathRepresentationDotted
	got = pathJSON(Field([]string{`a"b`}))
	if !stdjson.Valid([]byte(got)) {
		t.Errorf("dotted escaped path is not valid JSON: %s", got)
	}
}
