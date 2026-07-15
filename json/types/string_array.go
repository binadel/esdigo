package types

import "github.com/binadel/esdigo/json"

// StringArray is a JSON array of strings, stored unboxed in a []string. Its
// elements are decoded copies (unlike the scalar String, they do not alias the
// input buffer), so they stay valid after the buffer is reused. It carries the
// usual tri-state: Present, Defined, and Valid.
type StringArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []string
}

// IsPresent reports whether the field appeared in the input.
func (a *StringArray) IsPresent() bool {
	return a.Present
}

// IsDefined reports whether the field was present and non-null.
func (a *StringArray) IsDefined() bool {
	return a.Defined
}

// IsValid reports whether the array was well-formed and every element was a
// string (no element was dropped).
func (a *StringArray) IsValid() bool {
	return a.Valid
}

// Elements returns the decoded element slice, for a generic array validator.
func (a *StringArray) Elements() []string {
	return a.Value
}

// Set assigns value and marks the field present, defined, and valid.
func (a *StringArray) Set(value []string) {
	*a = StringArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (a *StringArray) SetNull() {
	*a = StringArray{
		Present: true,
	}
}

// WriteJSON writes the array, or null when the field is not defined. It returns
// false only when the field is defined but invalid.
func (a *StringArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				w.WriteString(v)
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

// ReadJSON reads a JSON array of strings (or null) into a. A non-string element
// is dropped and marks the array Valid=false (the string elements are still
// kept). Only a malformed array — an unskippable element or a missing separator —
// stops the reader.
func (a *StringArray) ReadJSON(r *json.Reader) bool {
	*a = StringArray{
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
			if value, ok := r.ReadString(); ok {
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
