package types

import "github.com/binadel/esdigo/json"

// Boolean is a JSON boolean field with tri-state tracking: Present (the field
// appeared at all), Defined (it was present and non-null), and Valid (it held a
// real boolean). This lets a decoder tell apart absent, null, and malformed
// without a separate error per field. See json.OptionalValue.
type Boolean struct {
	Present bool
	Defined bool
	Valid   bool
	Value   bool
}

// IsPresent reports whether the field appeared in the input.
func (b *Boolean) IsPresent() bool {
	return b.Present
}

// IsDefined reports whether the field was present and non-null.
func (b *Boolean) IsDefined() bool {
	return b.Defined
}

// IsValid reports whether a usable boolean was read.
func (b *Boolean) IsValid() bool {
	return b.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (b *Boolean) Set(value bool) {
	*b = Boolean{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (b *Boolean) SetNull() {
	*b = Boolean{
		Present: true,
	}
}

// WriteJSON writes the boolean, or null when the field is not defined. It returns
// false only when the field is defined but invalid, i.e. there is no value to
// serialize.
func (b *Boolean) WriteJSON(w *json.Writer) bool {
	if b.Defined {
		if b.Valid {
			w.WriteBoolean(b.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

// ReadJSON reads a JSON boolean (or null) into b. Per the json.ValueReader
// contract the returned bool means "the reader can continue", NOT "the value is
// valid": a non-boolean is skipped and leaves Valid=false so the enclosing object
// or array keeps parsing.
func (b *Boolean) ReadJSON(r *json.Reader) bool {
	*b = Boolean{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	b.Defined = true

	if b.Value, b.Valid = r.ReadBoolean(); b.Valid {
		r.SkipWhitespace()
		return true
	}

	return r.SkipValue()
}
