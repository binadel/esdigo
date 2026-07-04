package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

// The exported aliases pair NumberArray with a codec for each backing type — the
// array counterparts of the Number aliases (Int64Array, Float64Array, ...).
type (
	IntArray    = NumberArray[int, scalarInt[int]]
	Int8Array   = NumberArray[int8, scalarInt[int8]]
	Int16Array  = NumberArray[int16, scalarInt[int16]]
	Int32Array  = NumberArray[int32, scalarInt[int32]]
	Int64Array  = NumberArray[int64, scalarInt[int64]]
	UIntArray   = NumberArray[uint, scalarInt[uint]]
	UInt8Array  = NumberArray[uint8, scalarInt[uint8]]
	UInt16Array = NumberArray[uint16, scalarInt[uint16]]
	UInt32Array = NumberArray[uint32, scalarInt[uint32]]
	UInt64Array = NumberArray[uint64, scalarInt[uint64]]

	Float32Array = NumberArray[float32, scalarFloat[float32]]
	Float64Array = NumberArray[float64, scalarFloat[float64]]

	BigIntArray   = NumberArray[*big.Int, bigIntCodec]
	BigFloatArray = NumberArray[*big.Float, bigFloatCodec]

	RawNumberArray = NumberArray[[]byte, rawCodec]
)

// NumberArray is a JSON array of numbers decoded by codec C. Unlike the generic
// Array it stores its elements unboxed in a []V and decodes each with the codec
// directly, so it is leaner for scalar numbers. It carries the usual tri-state:
// Present, Defined, and Valid. Use the aliases above.
type NumberArray[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []V
}

// IsPresent reports whether the field appeared in the input.
func (a *NumberArray[V, C]) IsPresent() bool {
	return a.Present
}

// IsDefined reports whether the field was present and non-null.
func (a *NumberArray[V, C]) IsDefined() bool {
	return a.Defined
}

// IsValid reports whether the array was well-formed and every element was a
// representable number (no element was dropped).
func (a *NumberArray[V, C]) IsValid() bool {
	return a.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (a *NumberArray[V, C]) Set(value []V) {
	*a = NumberArray[V, C]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (a *NumberArray[V, C]) SetNull() {
	*a = NumberArray[V, C]{
		Present: true,
	}
}

// WriteJSON writes the array, or null when the field is not defined. It returns
// false only when the field is defined but invalid.
func (a *NumberArray[V, C]) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			var codec C
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				codec.write(w, v)
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

// ReadJSON reads a JSON array of numbers (or null) into a. An element that is not
// a number, or is a number that does not fit V, is dropped and marks the array
// Valid=false (representable elements are still kept in Value). Only a malformed
// array — an unskippable element or a missing separator — stops the reader.
func (a *NumberArray[V, C]) ReadJSON(r *json.Reader) bool {
	*a = NumberArray[V, C]{
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

		var codec C
		for {
			if t, _ := r.PeekType(); t == json.ValueTypeNumber {
				if elem, ok := codec.decode(r); ok {
					a.Value = append(a.Value, elem)
				} else {
					skipped = true
				}
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
