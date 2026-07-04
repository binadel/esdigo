package validation

import (
	"strings"

	"github.com/binadel/esdigo/json"
)

// PathRepresentation determines how to represent a path, which can be:
// 1. in array form: ["user", "address", "street"]
// 2. joined by dots: "user.address.street"
// 3. joined by slashes: "user/address/street"
type PathRepresentation int

const (
	PathRepresentationArray PathRepresentation = iota
	PathRepresentationDotted
	PathRepresentationSlashed
)

// FieldPathRepresentation is a global config for path representation. It is read
// when Field builds a path (typically at validator construction), so set it once
// during initialization rather than mutating it concurrently with Field calls.
var FieldPathRepresentation = PathRepresentationArray

// FieldPath represent the path to the field.
type FieldPath struct {
	segments []string
	json     []byte
}

// Field creates a new field path, precomputing its JSON form per the current
// FieldPathRepresentation. Segments are written through the JSON writer, so any
// quote, backslash or control character in a segment is properly escaped.
func Field(path []string) FieldPath {
	w := json.NewWriter(32)
	switch FieldPathRepresentation {
	case PathRepresentationArray:
		w.BeginArray()
		for i, segment := range path {
			if i > 0 {
				w.ValueSeparator()
			}
			w.WriteString(segment)
		}
		w.EndArray()
	case PathRepresentationDotted:
		w.WriteString(strings.Join(path, "."))
	case PathRepresentationSlashed:
		w.WriteString(strings.Join(path, "/"))
	default:
		panic("invalid FieldPath representation type")
	}
	return FieldPath{
		segments: path,
		json:     append([]byte(nil), w.Bytes()...),
	}
}

// Segments return the path segments.
func (p FieldPath) Segments() []string {
	return p.segments
}

// WriteJSON writes JSON form of the field path.
func (p FieldPath) WriteJSON(w *json.Writer) bool {
	w.WriteRawBytes(p.json)
	return true
}
