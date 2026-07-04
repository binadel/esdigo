package types

import "github.com/binadel/esdigo/json"

// Array is a JSON array of elements V. PV is the pointer type *V (constrained to
// read and write itself as JSON), which lets ReadJSON allocate an element with
// `new(V)` and decode into it in place — no reflection and no factory method. Use
// it for arrays of objects (Array[Foo, *Foo]) or nested arrays; the specialized
// scalar arrays (NumberArray, StringArray, BooleanArray) are leaner for scalars.
// It carries the usual tri-state: Present, Defined, and Valid.
type Array[V any, PV json.ValueReadWriter[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []PV
}

// IsPresent reports whether the field appeared in the input.
func (a *Array[V, PV]) IsPresent() bool {
	return a.Present
}

// IsDefined reports whether the field was present and non-null.
func (a *Array[V, PV]) IsDefined() bool {
	return a.Defined
}

// IsValid reports whether the array was well-formed with no element skipped. It
// does not assert each element is individually valid — an element that decoded
// but is itself invalid stays in Value; check the elements for that.
func (a *Array[V, PV]) IsValid() bool {
	return a.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (a *Array[V, PV]) Set(value []PV) {
	*a = Array[V, PV]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (a *Array[V, PV]) SetNull() {
	*a = Array[V, PV]{
		Present: true,
	}
}

// WriteJSON writes the array, or null when the field is not defined. It returns
// false if the field is defined-but-invalid or if any element fails to write.
func (a *Array[V, PV]) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				if ok := v.WriteJSON(w); !ok {
					return false
				}
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

// ReadJSON reads a JSON array (or null) into a. Each element that reads
// successfully is appended; an element the reader cannot decode but can skip is
// dropped and marks the whole array Valid=false (kept elements are still
// available). Only a malformed array — a broken element or a missing separator —
// stops the reader and returns false.
func (a *Array[V, PV]) ReadJSON(r *json.Reader) bool {
	*a = Array[V, PV]{
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
			item := PV(new(V))
			if item.ReadJSON(r) {
				a.Value = append(a.Value, item)
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
