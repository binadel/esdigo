package types

import (
	"math/big"

	"github.com/binadel/esdigo/json"
)

// Number is a tri-state JSON number (real) field. V is the in-memory value type
// and C is the codec that reads/writes it (see numberCodec). It is the mirror of
// Integer for JSON Schema's "number" type. Use the aliases below (Float64,
// Float32, BigFloat, RawNumber, ...) rather than naming the type parameters.
//
// See Integer for the Present/Defined/Valid tri-state model.
type Number[V any, C numberCodec[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   V
}

func (n *Number[V, C]) IsPresent() bool { return n.Present }
func (n *Number[V, C]) IsDefined() bool { return n.Defined }
func (n *Number[V, C]) IsValid() bool   { return n.Valid }

func (n *Number[V, C]) Set(value V) {
	*n = Number[V, C]{Present: true, Defined: true, Valid: true, Value: value}
}

func (n *Number[V, C]) SetNull() {
	*n = Number[V, C]{Present: true}
}

// CreateValue returns a fresh zero value so that *Number can participate in the
// generic Array/Object containers. The receiver is never read (it may be nil).
func (n *Number[V, C]) CreateValue() *Number[V, C] {
	return &Number[V, C]{}
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
	*n = Number[V, C]{Present: true}

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

// Number backing aliases. Pick by JSON Schema: type:number + format.
type (
	Float32   = Number[float32, scalarFloat[float32]]
	Float64   = Number[float64, scalarFloat[float64]]
	BigFloat  = Number[*big.Float, bigFloatCodec]
	RawNumber = Number[[]byte, rawCodec]
)
