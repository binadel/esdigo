package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

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

type Number[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   V
}

func (n *Number[V, C]) IsPresent() bool {
	return n.Present
}

func (n *Number[V, C]) IsDefined() bool {
	return n.Defined
}

func (n *Number[V, C]) IsValid() bool {
	return n.Valid
}

func (n *Number[V, C]) Set(value V) {
	*n = Number[V, C]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *Number[V, C]) SetNull() {
	*n = Number[V, C]{
		Present: true,
	}
}

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

func (n *Number[V, C]) ReadJSON(r *json.Reader) bool {
	*n = Number[V, C]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if !r.NextIsNumber() {
		return r.SkipValue()
	}

	var codec C
	n.Value, n.Valid = codec.decode(r)
	r.SkipWhitespace()

	return r.Error() == nil
}
