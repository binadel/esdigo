package types

import (
	"math"
	"math/big"
	"strconv"
	"unsafe"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

// signed, unsigned, integer and float are the type sets the scalar codecs range
// over, so a single generic implementation covers every width.
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

// numberCodec is the behavior parameter of Number. It reads one JSON number from
// the Reader and converts it to/from the in-memory value V using the json
// package's own scanner (ReadNumber/ReadRawNumber) — types never implement a
// parser themselves.
//
// The field envelope (Number.ReadJSON) handles null and the not-a-number case
// (it peeks with PeekType and skips non-numbers), so decode is invoked only
// when a number token is present and returns whether that number was a valid
// value for V. A number that is read but not convertible (e.g. "1e3" into a
// big.Int, or an out-of-range integer) returns false (Valid=false) without a
// spurious skip, because it has already been consumed.
//
// Implementations are zero-size struct types, so a codec value carries no state
// and costs nothing to construct. The interface is sealed (unexported methods):
// the set of number backings is closed to this package, and callers select one
// through the exported aliases (Int64, UInt64, BigInt, RawNumber, Float64, ...).
type numberCodec[V any] interface {
	decode(r *json.Reader) (V, bool)
	write(w *json.Writer, v V)
}

// scalarInt decodes a JSON number into a fixed-width integer T from the reader's
// structured NumberValue — no re-parsing of the token.
type scalarInt[T integer] struct{}

func (scalarInt[T]) decode(r *json.Reader) (T, bool) {
	num, _, ok := r.ReadNumber()
	if !ok {
		var zero T
		return zero, false
	}
	return intFromNumber[T](num)
}

func (scalarInt[T]) write(w *json.Writer, v T) {
	// zero-1 > zero is true only for unsigned T (it wraps to the max value), so
	// this selects the signed/unsigned writer at compile time per instantiation.
	var zero T
	if zero-1 > zero {
		w.WriteUIntNumber(uint64(v))
	} else {
		w.WriteIntNumber(int64(v))
	}
}

// intFromNumber converts a NumberValue (from the reader's own ReadNumber) into an
// integer of type T, following JSON Schema's "integer" definition: any number
// with a zero fractional part qualifies, so "1e3", "1.0" and "120e-1" convert,
// while "1.5" or out-of-range magnitudes do not.
func intFromNumber[T integer](num json.NumberValue) (value T, ok bool) {
	if num.Type != json.NumberTypeInteger {
		return
	}

	mag := num.Coefficient
	if mag == 0 {
		return 0, true
	}

	exp := int(num.Exponent)
	for exp > 0 { // scale up, guarding uint64 overflow
		if mag > math.MaxUint64/10 {
			return
		}
		mag *= 10
		exp--
	}
	for exp < 0 { // exact: an integer-valued number has the required trailing zeros
		mag /= 10
		exp++
	}

	var zero T
	bits := uint(unsafe.Sizeof(zero)) * 8
	if zero-1 > zero { // unsigned
		if num.Negative {
			return
		}
		if bits < 64 && mag > uint64(1)<<bits-1 {
			return
		}
		return T(mag), true
	}

	if num.Negative {
		cutoff := uint64(1) << (bits - 1) // magnitude of the most-negative value
		if mag > cutoff {
			return
		}
		if mag == cutoff {
			return T(-int64(mag-1) - 1), true // min value, without overflow
		}
		return T(-int64(mag)), true
	}
	if mag > uint64(1)<<(bits-1)-1 {
		return
	}
	return T(mag), true
}

// --- scalar floats ---

// scalarFloat decodes a JSON number into float32/float64 via strconv.ParseFloat
// on the raw token (correctly rounded). A value that overflows the float range
// makes ParseFloat report an error, leaving the field invalid.
type scalarFloat[T float] struct{}

func (scalarFloat[T]) decode(r *json.Reader) (T, bool) {
	var zero T
	token, ok := r.ReadRawNumber()
	if !ok {
		return zero, false
	}
	f, err := strconv.ParseFloat(utils.UnsafeString(token), int(unsafe.Sizeof(zero))*8)
	if err != nil {
		return zero, false
	}
	return T(f), true
}

func (scalarFloat[T]) write(w *json.Writer, v T) {
	var zero T
	if unsafe.Sizeof(zero) == 4 {
		w.WriteFloatNumber(float64(v), 32)
	} else {
		w.WriteFloatNumber(float64(v), 64)
	}
}

type bigIntCodec struct{}

func (bigIntCodec) decode(r *json.Reader) (*big.Int, bool) {
	// ReadNumber classifies the token in its single scan: NumberTypeOverflow means
	// the magnitude is too large to materialize (DoS guard) — reject without the
	// codec re-parsing. big.Rat then builds the exact value from the raw token and
	// its IsInt() enforces JSON Schema's integer rule (rejecting e.g. "1.5").
	num, token, ok := r.ReadNumber()
	if !ok || num.Type == json.NumberTypeOverflow {
		return nil, false
	}
	rat, ok := new(big.Rat).SetString(utils.UnsafeString(token))
	if !ok || !rat.IsInt() {
		return nil, false
	}
	return rat.Num(), true
}

func (bigIntCodec) write(w *json.Writer, v *big.Int) {
	w.WriteBigIntNumber(v)
}

// bigFloatCodec decodes an arbitrary-precision float. It is safe against huge
// exponents by construction — big.Float has a bounded binary exponent, so an
// out-of-range magnitude yields an error rather than a giant allocation.
type bigFloatCodec struct{}

func (bigFloatCodec) decode(r *json.Reader) (*big.Float, bool) {
	token, ok := r.ReadRawNumber()
	if !ok {
		return nil, false
	}
	// give the mantissa enough bits to hold every digit of the literal
	// (~3.33 bits per decimal digit; 4 is a safe upper bound).
	precision := uint(len(token)) * 4
	if precision < 64 {
		precision = 64
	}
	f, _, err := big.ParseFloat(utils.UnsafeString(token), 10, precision, big.ToNearestEven)
	if err != nil {
		return nil, false
	}
	return f, true
}

func (bigFloatCodec) write(w *json.Writer, v *big.Float) {
	w.WriteBigFloatNumber(v)
}

// rawCodec keeps a number as its raw source bytes without interpreting it — the
// backing for RawNumber. It never rejects a well-formed number, so it is the way
// to preserve values that no fixed Go type can hold exactly.
type rawCodec struct{}

func (rawCodec) decode(r *json.Reader) ([]byte, bool) {
	return r.ReadRawNumber()
}

func (rawCodec) write(w *json.Writer, v []byte) {
	if len(v) == 0 { // nothing stored (e.g. a zero-value RawNumber) → null
		w.WriteNull()
		return
	}
	w.WriteRawNumber(v)
}
