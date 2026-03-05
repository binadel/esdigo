package json

import (
	"math"
	"strconv"
)

// uint64MaxCutoff = math.MaxUint64 / 10, used to calculate exact overflow
const uint64MaxCutoff = 1844674407370955161

type NumberType uint8

const (
	NumberTypeInteger = NumberType(iota)
	NumberTypeReal
	NumberTypeBig
)

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

func (w *Writer) WriteFloatNumber(value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		w.WriteNull()
	} else {
		w.data = strconv.AppendFloat(w.data, value, 'g', -1, 64)
	}
}

func (r *Reader) ReadRawNumber() ([]byte, bool) {
	start := r.pos
	if r.SkipNumber() {
		return r.data[start:r.pos], true
	}
	return nil, false
}

func (r *Reader) ReadNumber() (value NumberValue, ok bool) {
	if r.err != nil {
		return
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return
	}

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
	var trailingZeros int
	var big bool

	// integer part
	if digit := uint64(r.data[r.pos] - '0'); digit == 0 {
		r.pos++
		if r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
				r.SetSyntaxError("invalid leading zero in number")
				return
			}
		}
	} else if digit <= 9 {
		coefficient = digit
		r.pos++

		for r.pos < len(r.data) {
			if digit := uint64(r.data[r.pos] - '0'); digit <= 9 {
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
			exponent -= int(exp)
		} else {
			exponent += int(exp)
		}
	}

	if exponent > math.MaxInt16 || exponent < math.MinInt16 {
		big = true
	}

	if big {
		value.Type = NumberTypeBig
		return value, true
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

	return value, true
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
		for r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
				r.pos++
			} else {
				break
			}
		}
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
		for r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
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
		for r.pos < len(r.data) {
			if c := r.data[r.pos]; c >= '0' && c <= '9' {
				r.pos++
			} else {
				break
			}
		}
	}

	return true
}
