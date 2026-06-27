package types

import (
	"strconv"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

type FloatNumber struct {
	Present bool
	Defined bool
	Valid   bool
	Value   float64
}

func (n *FloatNumber) IsPresent() bool {
	return n.Present
}

func (n *FloatNumber) IsDefined() bool {
	return n.Defined
}

func (n *FloatNumber) IsValid() bool {
	return n.Valid
}

func (n *FloatNumber) Set(value float64) {
	*n = FloatNumber{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *FloatNumber) SetNull() {
	*n = FloatNumber{
		Present: true,
	}
}

func (n *FloatNumber) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			w.WriteFloatNumber(n.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (n *FloatNumber) ReadJSON(r *json.Reader) bool {
	*n = FloatNumber{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if value, ok := r.ReadRawNumber(); ok {
		r.SkipWhitespace()
		float, err := strconv.ParseFloat(utils.UnsafeString(value), 64)
		if err != nil {
			return true
		}

		n.Valid = true
		n.Value = float
		return true
	}

	return r.SkipValue()
}
