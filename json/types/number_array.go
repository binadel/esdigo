package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

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

type NumberArray[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []V
}

func (a *NumberArray[V, C]) IsPresent() bool {
	return a.Present
}

func (a *NumberArray[V, C]) IsDefined() bool {
	return a.Defined
}

func (a *NumberArray[V, C]) IsValid() bool {
	return a.Valid
}

func (a *NumberArray[V, C]) Set(value []V) {
	*a = NumberArray[V, C]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *NumberArray[V, C]) SetNull() {
	*a = NumberArray[V, C]{
		Present: true,
	}
}

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
			if r.NextIsNumber() {
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
