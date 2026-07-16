package types

import "github.com/binadel/esdigo/json"

// Object is a nullable wrapper around a single nested value V — typically a
// code-generated struct. It adds the tri-state (Present, Defined, Valid) and
// null handling on top of V, then delegates the actual object parsing to V's own
// ReadJSON/WriteJSON. PV is the pointer type *V, used to allocate and populate a
// V in place. Use it for an optional/nullable object field.
type Object[V any, PV json.ValueReadWriter[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   PV
}

// IsPresent reports whether the field appeared in the input.
func (o *Object[V, PV]) IsPresent() bool {
	return o.Present
}

// IsDefined reports whether the field was present and non-null.
func (o *Object[V, PV]) IsDefined() bool {
	return o.Defined
}

// IsValid reports whether the nested value was read successfully.
func (o *Object[V, PV]) IsValid() bool {
	return o.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (o *Object[V, PV]) Set(value PV) {
	*o = Object[V, PV]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (o *Object[V, PV]) SetNull() {
	*o = Object[V, PV]{
		Present: true,
	}
}

// WriteJSON writes the nested value, or null when the field is not defined. It
// returns false if the field is defined-but-invalid or if the value fails to write.
func (o *Object[V, PV]) WriteJSON(w *json.Writer) bool {
	if o.Defined {
		if o.Valid {
			if ok := o.Value.WriteJSON(w); !ok {
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

// ReadJSON reads the nested value (or null) into o, delegating to V.ReadJSON.
// Valid is set from whether that read reported success; if V could not be read
// but the input can be skipped, the value is skipped and left invalid so the
// enclosing object or array keeps parsing.
func (o *Object[V, PV]) ReadJSON(r *json.Reader) bool {
	*o = Object[V, PV]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	o.Defined = true

	o.Value = PV(new(V))
	if o.Valid = o.Value.ReadJSON(r); o.Valid {
		r.SkipWhitespace()
		return true
	}

	return r.SkipValue()
}
