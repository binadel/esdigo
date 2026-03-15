package validation

import "github.com/binadel/esdigo/json"

const (
	keyPath   = `"path":`
	keyErrors = `,"errors":`
)

// Result represents the validation result for the field.
type Result[T any] struct {
	Path    FieldPath
	Errors  []Error
	Present bool
	Defined bool
	Value   T
}

// IsValid returns whether the result is valid.
func (r *Result[T]) IsValid() bool {
	return len(r.Errors) == 0
}

// WriteJSON writes JSON form of the result.
func (r *Result[T]) WriteJSON(w *json.Writer) bool {
	w.BeginObject()

	w.WriteRawString(keyPath)
	r.Path.WriteJSON(w)

	w.WriteRawString(keyErrors)
	w.BeginArray()
	for i, err := range r.Errors {
		if i > 0 {
			w.ValueSeparator()
		}
		err.WriteJSON(w)
	}
	w.EndArray()

	w.EndObject()
	return true
}
