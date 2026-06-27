package types

import (
	"math/big"
	"strconv"
	"unsafe"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

// --- numeric type constraints ---

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type integer interface {
	signed | unsigned
}

type float interface {
	~float32 | ~float64
}

// numberCodec is the behavior parameter of Integer and Number. It is a pure
// token<->value converter: it never touches the Reader (the field envelope owns
// all reader interaction: null, the raw-number read, whitespace and skipping),
// so a token that is consumed but not convertible (e.g. "1e3" into a big.Int)
// simply yields Valid=false instead of a spurious parse error.
//
// Implementations are zero-size struct types, so a codec value carries no state
// and costs nothing to construct. The interface is sealed (unexported methods):
// the set of number backings is closed to this package, and callers select one
// through the exported aliases (Int64, Uint64, BigInt, RawInt, Float64, ...).
type numberCodec[V any] interface {
	// decode converts a raw JSON number token into *dst, reporting whether the
	// token is representable in V. token is never empty.
	decode(token []byte, dst *V) bool
	// write appends the JSON form of v to the writer.
	write(w *json.Writer, v V)
}

// --- scalar integer backing (int/uint of any width) ---

type scalarInt[T integer] struct{}

func (scalarInt[T]) decode(token []byte, dst *T) bool {
	var zero T
	bits := int(unsafe.Sizeof(zero)) * 8
	if zero-1 > zero { // unsigned: 0-1 wraps to the max value, which is > 0
		u, err := strconv.ParseUint(utils.UnsafeString(token), 10, bits)
		if err != nil {
			return false
		}
		*dst = T(u)
		return true
	}
	i, err := strconv.ParseInt(utils.UnsafeString(token), 10, bits)
	if err != nil {
		return false
	}
	*dst = T(i)
	return true
}

func (scalarInt[T]) write(w *json.Writer, v T) {
	var zero T
	if zero-1 > zero {
		w.WriteUIntNumber(uint64(v))
	} else {
		w.WriteIntNumber(int64(v))
	}
}

// --- scalar float backing (float32/float64) ---

type scalarFloat[T float] struct{}

func (scalarFloat[T]) decode(token []byte, dst *T) bool {
	var zero T
	f, err := strconv.ParseFloat(utils.UnsafeString(token), int(unsafe.Sizeof(zero))*8)
	if err != nil {
		return false
	}
	*dst = T(f)
	return true
}

func (scalarFloat[T]) write(w *json.Writer, v T) {
	var zero T
	if unsafe.Sizeof(zero) == 4 {
		w.WriteFloat32(float32(v))
	} else {
		w.WriteFloatNumber(float64(v))
	}
}

// --- big.Int backing (arbitrary-precision integers) ---
//
// SetString(base 10) accepts plain integer syntax with an optional sign. Numbers
// in exponent or fractional form (e.g. "1e3") are rejected here (Valid=false);
// use a Number backing or RawInt to retain those losslessly.

type bigIntCodec struct{}

func (bigIntCodec) decode(token []byte, dst **big.Int) bool {
	z, ok := new(big.Int).SetString(utils.UnsafeString(token), 10)
	if !ok {
		return false
	}
	*dst = z
	return true
}

func (bigIntCodec) write(w *json.Writer, v *big.Int) {
	if v == nil {
		w.WriteNull()
		return
	}
	w.WriteRawString(v.Text(10))
}

// --- big.Float backing (arbitrary-precision reals) ---

type bigFloatCodec struct{}

func (bigFloatCodec) decode(token []byte, dst **big.Float) bool {
	// give the mantissa enough bits to hold every digit of the literal
	// (~3.33 bits per decimal digit; 4 is a safe upper bound).
	prec := uint(len(token)) * 4
	if prec < 64 {
		prec = 64
	}
	f, _, err := big.ParseFloat(utils.UnsafeString(token), 10, prec, big.ToNearestEven)
	if err != nil {
		return false
	}
	*dst = f
	return true
}

func (bigFloatCodec) write(w *json.Writer, v *big.Float) {
	if v == nil || v.IsInf() {
		w.WriteNull()
		return
	}
	w.WriteRawString(v.Text('g', -1))
}

// --- raw backing (lossless, zero-copy on read; shared by Integer and Number) ---
//
// The decoded value aliases the input buffer, like other raw types in this
// package; it is valid only while the source bytes live unmodified.

type rawCodec struct{}

func (rawCodec) decode(token []byte, dst *[]byte) bool {
	*dst = token
	return true
}

func (rawCodec) write(w *json.Writer, v []byte) {
	if len(v) == 0 {
		w.WriteNull()
		return
	}
	w.WriteRawNumber(v)
}
