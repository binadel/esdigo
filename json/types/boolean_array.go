package types

import "github.com/binadel/esdigo/json"

// BooleanArray is a JSON array of booleans, stored unboxed in a []bool. It
// carries the usual tri-state: Present, Defined, and Valid. See json.OptionalValue.
type BooleanArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []bool
}

// IsPresent reports whether the field appeared in the input.
func (a *BooleanArray) IsPresent() bool {
	return a.Present
}

// IsDefined reports whether the field was present and non-null.
func (a *BooleanArray) IsDefined() bool {
	return a.Defined
}

// IsValid reports whether the array was well-formed and every element was a
// boolean (no element was dropped).
func (a *BooleanArray) IsValid() bool {
	return a.Valid
}

// Elements returns the decoded element slice, for a generic array validator.
func (a *BooleanArray) Elements() []bool {
	return a.Value
}

// Set assigns value and marks the field present, defined, and valid.
func (a *BooleanArray) Set(value []bool) {
	*a = BooleanArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (a *BooleanArray) SetNull() {
	*a = BooleanArray{
		Present: true,
	}
}

// WriteJSON writes the array, or null when the field is not defined. It returns
// false only when the field is defined but invalid.
func (a *BooleanArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				w.WriteBoolean(v)
			}
			w.EndArray()
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

// ReadJSON reads a JSON array of booleans (or null) into a. A non-boolean element
// is dropped and marks the array Valid=false (the boolean elements are still
// kept). Only a malformed array — an unskippable element or a missing separator —
// stops the reader.
func (a *BooleanArray) ReadJSON(r *json.Reader) bool {
	*a = BooleanArray{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	a.Defined = true

	skipped := false
	if r.BeginArray() {
		r.SkipWhitespace()

		if r.EndArray() {
			r.SkipWhitespace()
			a.Valid = true
			return true
		}

		for {
			if value, ok := r.ReadBoolean(); ok {
				a.Value = append(a.Value, value)
			} else if r.SkipValue() {
				skipped = true
			} else {
				return false
			}

			r.SkipWhitespace()

			if r.EndArray() {
				a.Valid = !skipped
				return true
			}

			if !r.ValueSeparator() {
				r.SetSyntaxError("expected either end-array ']' or value-separator ','")
				return false
			}

			r.SkipWhitespace()
		}
	}

	return r.SkipValue()
}
