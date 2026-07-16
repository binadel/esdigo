package types

import "github.com/binadel/esdigo/json"

// Any is a field of arbitrary JSON shape, backed by the untyped DOM (json.Value).
// Use it for schema-less values (e.g. JSON Schema additionalProperties: true). It
// carries the usual tri-state: Present, Defined, and Valid. See json.OptionalValue.
type Any struct {
	Present bool
	Defined bool
	Valid   bool
	Value   json.Value
}

// IsPresent reports whether the field appeared in the input.
func (a *Any) IsPresent() bool {
	return a.Present
}

// IsDefined reports whether the field was present and non-null.
func (a *Any) IsDefined() bool {
	return a.Defined
}

// IsValid reports whether a usable value was read.
func (a *Any) IsValid() bool {
	return a.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (a *Any) Set(value json.Value) {
	*a = Any{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (a *Any) SetNull() {
	*a = Any{
		Present: true,
	}
}

// WriteJSON writes the value, or null when the field is not defined. It returns
// false if the field is defined-but-invalid, or if the DOM value cannot be
// serialized.
func (a *Any) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			if ok := w.WriteValue(a.Value); !ok {
				return false
			}
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

// ReadJSON reads any JSON value (or null) into a. Unlike the scalar wrappers there
// is no "wrong type" to recover from — Any accepts every type — so a read that
// fails is a genuine syntax error and returns false (the reader cannot continue).
func (a *Any) ReadJSON(r *json.Reader) bool {
	*a = Any{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	a.Defined = true

	if a.Value, a.Valid = r.ReadValue(); a.Valid {
		r.SkipWhitespace()
		return true
	}

	return false
}
