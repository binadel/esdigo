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

// FieldPathRepresentation is a global config for path representation.
var FieldPathRepresentation = PathRepresentationArray

// FieldPath represent the path to the field.
type FieldPath struct {
	segments []string
	json     []byte
}

// Field creates a new field path.
func Field(path []string) FieldPath {
	var jsonStr string
	switch FieldPathRepresentation {
	case PathRepresentationArray:
		jsonStr = `["` + strings.Join(path, `","`) + `"]`
	case PathRepresentationDotted:
		jsonStr = `"` + strings.Join(path, ".") + `"`
	case PathRepresentationSlashed:
		jsonStr = `"` + strings.Join(path, "/") + `"`
	default:
		panic("invalid FieldPath representation type")
	}
	return FieldPath{
		segments: path,
		json:     []byte(jsonStr),
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
