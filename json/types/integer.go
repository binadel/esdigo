package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

// Integer is a tri-state JSON integer field. V is the in-memory value type and
// C is the codec that reads/writes that representation (see numberCodec). Use
// the aliases below (Int64, Uint64, BigInt, RawInt, ...) rather than naming the
// type parameters directly.
//
// Tri-state semantics:
//   - Present: the field appeared in the input.
//   - Defined: it was present and not null.
//   - Valid:   it was defined and convertible to V.
type Integer[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   V
}

func (n *Integer[V, C]) IsPresent() bool { return n.Present }
func (n *Integer[V, C]) IsDefined() bool { return n.Defined }
func (n *Integer[V, C]) IsValid() bool   { return n.Valid }

func (n *Integer[V, C]) Set(value V) {
	*n = Integer[V, C]{Present: true, Defined: true, Valid: true, Value: value}
}

func (n *Integer[V, C]) SetNull() {
	*n = Integer[V, C]{Present: true}
}

// CreateValue returns a fresh zero value so that *Integer can participate in the
// generic Array/Object containers. The receiver is never read (it may be nil).
func (n *Integer[V, C]) CreateValue() *Integer[V, C] {
	return &Integer[V, C]{}
}

func (n *Integer[V, C]) WriteJSON(w *json.Writer) bool {
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

func (n *Integer[V, C]) ReadJSON(r *json.Reader) bool {
	*n = Integer[V, C]{Present: true}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if token, ok := r.ReadRawNumber(); ok {
		r.SkipWhitespace()
		var codec C
		if codec.decode(token, &n.Value) {
			n.Valid = true
		}
		// A number token was consumed; Valid reflects whether it fit V.
		return true
	}

	// Not a number at all (string/bool/array/object): skip the whole value.
	return r.SkipValue()
}

// Integer backing aliases. Pick by JSON Schema: type:integer + format/bounds.
type (
	Int    = Integer[int, scalarInt[int]]
	Int8   = Integer[int8, scalarInt[int8]]
	Int16  = Integer[int16, scalarInt[int16]]
	Int32  = Integer[int32, scalarInt[int32]]
	Int64  = Integer[int64, scalarInt[int64]]
	Uint   = Integer[uint, scalarInt[uint]]
	Uint8  = Integer[uint8, scalarInt[uint8]]
	Uint16 = Integer[uint16, scalarInt[uint16]]
	Uint32 = Integer[uint32, scalarInt[uint32]]
	Uint64 = Integer[uint64, scalarInt[uint64]]
	BigInt = Integer[*big.Int, bigIntCodec]
	RawInt = Integer[[]byte, rawCodec]
)
