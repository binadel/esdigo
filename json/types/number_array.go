package types

import (
	"strconv"

	"github.com/binadel/esdigo/json"
)

// NumberArray is a flat array of JSON numbers held as their raw token bytes,
// which is lossless for values of any magnitude or precision. Each element
// aliases the input buffer on read (zero-copy), like the other raw types here,
// so it is valid only while the source bytes live unmodified.
type NumberArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   [][]byte
}

func (a *NumberArray) IsPresent() bool { return a.Present }
func (a *NumberArray) IsDefined() bool { return a.Defined }
func (a *NumberArray) IsValid() bool   { return a.Valid }

func (a *NumberArray) Set(value [][]byte) {
	*a = NumberArray{Present: true, Defined: true, Valid: true, Value: value}
}

func (a *NumberArray) SetIntArray(value []int64) {
	out := make([][]byte, len(value))
	for i, v := range value {
		out[i] = strconv.AppendInt(nil, v, 10)
	}
	*a = NumberArray{Present: true, Defined: true, Valid: true, Value: out}
}

func (a *NumberArray) SetUIntArray(value []uint64) {
	out := make([][]byte, len(value))
	for i, v := range value {
		out[i] = strconv.AppendUint(nil, v, 10)
	}
	*a = NumberArray{Present: true, Defined: true, Valid: true, Value: out}
}

func (a *NumberArray) SetNull() {
	*a = NumberArray{Present: true}
}

func (a *NumberArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				w.WriteRawNumber(v)
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

func (a *NumberArray) ReadJSON(r *json.Reader) bool {
	*a = NumberArray{Present: true}

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

		for {
			if value, ok := r.ReadRawNumber(); ok {
				a.Value = append(a.Value, value)
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
