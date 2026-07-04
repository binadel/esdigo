package json

import (
	"encoding/binary"
	"math"
	"math/big"
	"strconv"
)

// uint64MaxCutoff = math.MaxUint64 / 10, used to calculate exact overflow
const uint64MaxCutoff = 1844674407370955161

type NumberType uint8

const (
	NumberTypeInteger = NumberType(iota)
	NumberTypeReal
	NumberTypeBig
	// NumberTypeOverflow marks a number whose decimal magnitude (integer digits +
	// exponent) exceeds maxNumberDigits — too large to materialize as an exact
	// integer/rational. The parser flags it so codecs can reject without
	// re-scanning the token (a big.Rat built from it would exhaust memory).
	//
	// This is a flag, NOT a parse error: a number token is flat, so reading or
	// skipping it is always safe — only the big.Int/big.Rat conversion is
	// dangerous, and the codec simply declines that (Valid=false) while parsing
	// continues. Contrast the nesting-depth limit, which MUST abort as a reader
	// error because a too-deep structure cannot be skipped without recursing into
	// it. Same "resource limit" idea, opposite tiers — for that reason.
	NumberTypeOverflow
)

// maxNumberDigits bounds the decimal magnitude a number may have before the
// parser classifies it NumberTypeOverflow. 65536 digits (~217 kbit) dwarfs any
// real integer while catching amplification such as "1e999999999".
const maxNumberDigits = 1 << 16

type NumberValue struct {
	Negative    bool
	Type        NumberType
	Exponent    int16
	Coefficient uint64
}

func (w *Writer) WriteRawNumber(value []byte) {
	w.data = append(w.data, value...)
}

func (w *Writer) WriteNumber(value NumberValue) {
	if value.Negative {
		w.data = append(w.data, '-')
	}

	w.data = strconv.AppendUint(w.data, value.Coefficient, 10)

	if value.Exponent != 0 {
		w.data = append(w.data, 'e')
		w.data = strconv.AppendInt(w.data, int64(value.Exponent), 10)
	}
}

func (w *Writer) WriteIntNumber(value int64) {
	w.data = strconv.AppendInt(w.data, value, 10)
}

func (w *Writer) WriteUIntNumber(value uint64) {
	w.data = strconv.AppendUint(w.data, value, 10)
}

func (w *Writer) WriteBigIntNumber(value *big.Int) {
	if value == nil {
		w.WriteNull()
	} else {
		w.WriteRawString(value.Text(10))
	}
}

func (w *Writer) WriteFloatNumber(value float64, bitSize int) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		w.WriteNull()
	} else {
		w.data = strconv.AppendFloat(w.data, value, 'g', -1, bitSize)
	}
}

func (w *Writer) WriteBigFloatNumber(value *big.Float) {
	if value == nil || value.IsInf() {
		w.WriteNull()
	} else {
		w.WriteRawString(value.Text('g', -1))
	}
}

func (r *Reader) ReadRawNumber() ([]byte, bool) {
	start := r.pos
	if r.SkipNumber() {
		return r.data[start:r.pos], true
	}
	return nil, false
}

func (r *Reader) ReadNumber() (value NumberValue, raw []byte, ok bool) {
	if r.err != nil {
		return
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return
	}

	start := r.pos

	// minus sign
	value.Negative = r.data[r.pos] == '-'
	if value.Negative {
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}
	}

	var coefficient uint64
	var exponent int
	var explicitExp int
	var trailingZeros int
	var intDigits int
	var big bool

	// integer part
	if digit := uint64(r.data[r.pos] - '0'); digit == 0 {
		intDigits = 1
		r.pos++
		if r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
				r.SetSyntaxError("invalid leading zero in number")
				return
			}
		}
	} else if digit <= 9 {
		coefficient = digit
		intDigits = 1
		r.pos++

		for r.pos < len(r.data) {
			if digit := uint64(r.data[r.pos] - '0'); digit <= 9 {
				intDigits++
				if coefficient < uint64MaxCutoff || (digit < 6 && coefficient == uint64MaxCutoff) {
					coefficient = coefficient*10 + digit
					if digit == 0 {
						trailingZeros++
					} else {
						trailingZeros = 0
					}
				} else if digit == 0 {
					exponent++
				} else {
					big = true
				}
				r.pos++
			} else {
				break
			}
		}
	} else {
		if value.Negative {
			r.SetSyntaxError("expected digit after minus sign in number")
		}
		return
	}

	// fractional part
	if r.pos < len(r.data) && r.data[r.pos] == '.' {
		r.pos++

		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}

		if c := r.data[r.pos]; c < '0' || c > '9' {
			r.SetSyntaxError("expected digit after decimal point in number")
			return
		}

		for r.pos < len(r.data) {
			if digit := uint64(r.data[r.pos] - '0'); digit <= 9 {
				if coefficient < uint64MaxCutoff || (digit < 6 && coefficient == uint64MaxCutoff) {
					coefficient = coefficient*10 + digit
					exponent--
					if digit == 0 {
						if coefficient != 0 {
							trailingZeros++
						}
					} else {
						trailingZeros = 0
					}
				} else if digit != 0 {
					big = true
				}
				r.pos++
			} else {
				break
			}
		}
	}

	// exponent part
	if r.pos < len(r.data) && (r.data[r.pos] == 'e' || r.data[r.pos] == 'E') {
		r.pos++

		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}

		negExp := false
		if c := r.data[r.pos]; c == '-' || c == '+' {
			if c == '-' {
				negExp = true
			}
			r.pos++
			if r.pos >= len(r.data) {
				r.SetEofError()
				return
			}
		}

		if c := r.data[r.pos]; c < '0' || c > '9' {
			r.SetSyntaxError("expected digit after exponent sign in number")
			return
		}

		var exp uint64
		for r.pos < len(r.data) {
			if digit := uint64(r.data[r.pos] - '0'); digit <= 9 {
				if exp < 21474836 {
					exp = exp*10 + digit
				}
				r.pos++
			} else {
				break
			}
		}

		if negExp {
			explicitExp = -int(exp)
			exponent -= int(exp)
		} else {
			explicitExp = int(exp)
			exponent += int(exp)
		}
	}

	// magnitude guard: the materialized value spans roughly intDigits + |explicit
	// exponent| decimal digits; beyond maxNumberDigits it is flagged Overflow so
	// codecs reject it without re-scanning (a big.Rat from it would exhaust memory).
	magnitude := intDigits
	if explicitExp > 0 {
		magnitude += explicitExp
	} else {
		magnitude -= explicitExp
	}
	if magnitude > maxNumberDigits {
		value.Type = NumberTypeOverflow
		return value, r.data[start:r.pos], true
	}

	if exponent > math.MaxInt16 || exponent < math.MinInt16 {
		big = true
	}

	if big {
		value.Type = NumberTypeBig
		return value, r.data[start:r.pos], true
	}

	value.Exponent = int16(exponent)
	value.Coefficient = coefficient

	if coefficient == 0 {
		value.Type = NumberTypeInteger
	} else if trailingZeros >= -exponent {
		value.Type = NumberTypeInteger
	} else {
		value.Type = NumberTypeReal
	}

	value.Exponent = int16(exponent)
	value.Coefficient = coefficient

	return value, r.data[start:r.pos], true
}

// isEightDigits reports whether all 8 bytes of a little-endian word are ASCII
// digits ('0'-'9'). It is exact: a byte b qualifies iff its high nibble is 3 and
// b+6 still has high nibble 3 (i.e. its low nibble is <= 9), so the folded value
// equals 0x33 in every lane exactly when all eight bytes are digits.
func isEightDigits(word uint64) bool {
	return (word&0xf0f0f0f0f0f0f0f0)|(((word+0x0606060606060606)&0xf0f0f0f0f0f0f0f0)>>4) == 0x3333333333333333
}

// skipDigits returns the index of the first non-digit byte in b (or len(b)),
// scanning eight bytes at a time.
func skipDigits(b []byte) int {
	i, n := 0, len(b)
	for i+8 <= n && isEightDigits(binary.LittleEndian.Uint64(b[i:])) {
		i += 8
	}
	for i < n {
		if c := b[i]; c < '0' || c > '9' {
			return i
		}
		i++
	}
	return n
}

func (r *Reader) SkipNumber() (ok bool) {
	if r.err != nil {
		return
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return
	}

	// minus sign
	negative := r.data[r.pos] == '-'
	if negative {
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}
	}

	// integer part
	if c := r.data[r.pos]; c == '0' {
		r.pos++
		if r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
				r.SetSyntaxError("invalid leading zero in number")
				return
			}
		}
	} else if c >= '1' && c <= '9' {
		r.pos++
		r.pos += skipDigits(r.data[r.pos:])
	} else {
		if negative {
			r.SetSyntaxError("expected digit after minus sign in number")
		}
		return
	}

	// fractional part
	if r.pos < len(r.data) && r.data[r.pos] == '.' {
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}

		if c := r.data[r.pos]; c < '0' || c > '9' {
			r.SetSyntaxError("expected digit after decimal point in number")
			return
		}
		r.pos++
		r.pos += skipDigits(r.data[r.pos:])
	}

	// exponent part
	if r.pos < len(r.data) && (r.data[r.pos] == 'e' || r.data[r.pos] == 'E') {
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return
		}

		if r.data[r.pos] == '+' || r.data[r.pos] == '-' {
			r.pos++
			if r.pos >= len(r.data) {
				r.SetEofError()
				return
			}
		}

		if c := r.data[r.pos]; c < '0' || c > '9' {
			r.SetSyntaxError("expected digit after exponent sign in number")
			return
		}
		r.pos++
		r.pos += skipDigits(r.data[r.pos:])
	}

	return true
}
