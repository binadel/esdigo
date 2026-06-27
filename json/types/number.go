package types

import (
	"strconv"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

type Number struct {
	Present  bool
	Defined  bool
	Valid    bool
	negative bool
	dotPos   byte
	expPos   byte
	Value    []byte
}

func (n *Number) IsPresent() bool {
	return n.Present
}

func (n *Number) IsDefined() bool {
	return n.Defined
}

func (n *Number) IsValid() bool {
	return n.Valid
}

func (n *Number) Set(value []byte) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *Number) SetString(value string) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(value),
	}
}

func (n *Number) SetInt(value int64) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   strconv.AppendInt(n.Value, value, 10),
	}
}

func (n *Number) SetUInt(value uint64) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   strconv.AppendUint(n.Value, value, 10),
	}
}

func (n *Number) SetFloat(value float64) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   strconv.AppendFloat(n.Value, value, 'g', -1, 64),
	}
}

func (n *Number) SetNull() {
	*n = Number{
		Present: true,
	}
}

func (n *Number) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			w.WriteRawNumber(n.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (n *Number) ReadJSON(r *json.Reader) bool {
	*n = Number{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if value, ok := r.ReadRawNumber(); ok {
		r.SkipWhitespace()
		n.Valid = true
		n.Value = value
		return true
	}

	return r.SkipValue()
}
