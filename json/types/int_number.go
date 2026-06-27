package types

import (
	"strconv"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

type IntNumber struct {
	Present bool
	Defined bool
	Valid   bool
	Value   int64
}

func (n *IntNumber) IsPresent() bool {
	return n.Present
}

func (n *IntNumber) IsDefined() bool {
	return n.Defined
}

func (n *IntNumber) IsValid() bool {
	return n.Valid
}

func (n *IntNumber) Set(value int64) {
	*n = IntNumber{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *IntNumber) SetNull() {
	*n = IntNumber{
		Present: true,
	}
}

func (n *IntNumber) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			w.WriteIntNumber(n.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (n *IntNumber) ReadJSON(r *json.Reader) bool {
	*n = IntNumber{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if value, ok := r.ReadRawNumber(); ok {
		r.SkipWhitespace()
		integer, err := strconv.ParseInt(utils.UnsafeString(value), 10, 64)
		if err != nil {
			return true
		}

		n.Valid = true
		n.Value = integer
		return true
	}

	return r.SkipValue()
}
