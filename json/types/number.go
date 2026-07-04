package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

// The exported aliases pair Number with a codec for each supported backing type.
// Pick the alias that matches the target Go type; the generic Number and the
// codecs are the machinery behind them.
type (
	Int    = Number[int, scalarInt[int]]
	Int8   = Number[int8, scalarInt[int8]]
	Int16  = Number[int16, scalarInt[int16]]
	Int32  = Number[int32, scalarInt[int32]]
	Int64  = Number[int64, scalarInt[int64]]
	UInt   = Number[uint, scalarInt[uint]]
	UInt8  = Number[uint8, scalarInt[uint8]]
	UInt16 = Number[uint16, scalarInt[uint16]]
	UInt32 = Number[uint32, scalarInt[uint32]]
	UInt64 = Number[uint64, scalarInt[uint64]]

	Float32 = Number[float32, scalarFloat[float32]]
	Float64 = Number[float64, scalarFloat[float64]]

	BigInt   = Number[*big.Int, bigIntCodec]
	BigFloat = Number[*big.Float, bigFloatCodec]

	RawNumber = Number[[]byte, rawCodec]
)

// Number is a JSON number field decoded into V by the codec C. V is the Go value
// type (int64, float64, *big.Int, ...) and C is its numberCodec; use the aliases
// above rather than naming the pair directly. It carries the usual tri-state:
// Present, Defined, and Valid. See json.OptionalValue.
type Number[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   V
}

// IsPresent reports whether the field appeared in the input.
func (n *Number[V, C]) IsPresent() bool {
	return n.Present
}

// IsDefined reports whether the field was present and non-null.
func (n *Number[V, C]) IsDefined() bool {
	return n.Defined
}

// IsValid reports whether the number was read and representable as V.
func (n *Number[V, C]) IsValid() bool {
	return n.Valid
}

// Set assigns value and marks the field present, defined, and valid.
func (n *Number[V, C]) Set(value V) {
	*n = Number[V, C]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (n *Number[V, C]) SetNull() {
	*n = Number[V, C]{
		Present: true,
	}
}

// WriteJSON writes the number, or null when the field is not defined. It returns
// false only when the field is defined but invalid.
func (n *Number[V, C]) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			var codec C
			codec.write(w, n.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

// ReadJSON reads a JSON number (or null) into n. It peeks the value type first: a
// non-number is skipped and left Valid=false, and only a real number reaches the
// codec.
//
// The number branch returns r.Error()==nil ("the reader can continue"), NOT
// n.Valid: a number that is read but not representable as V (e.g. "1.5" into an
// integer, or an overflow) is a Valid=false value, not a parse error, so the
// enclosing object or array must keep going.
func (n *Number[V, C]) ReadJSON(r *json.Reader) bool {
	*n = Number[V, C]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if t, _ := r.PeekType(); t == json.ValueTypeNumber {
		var codec C
		n.Value, n.Valid = codec.decode(r)
		r.SkipWhitespace()
		return r.Error() == nil
	}

	return r.SkipValue()
}
