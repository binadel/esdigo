package types

import (
	"strconv"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

type UIntNumber struct {
	Present bool
	Defined bool
	Valid   bool
	Value   uint64
}

func (n *UIntNumber) IsPresent() bool {
	return n.Present
}

func (n *UIntNumber) IsDefined() bool {
	return n.Defined
}

func (n *UIntNumber) IsValid() bool {
	return n.Valid
}

func (n *UIntNumber) Set(value uint64) {
	*n = UIntNumber{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *UIntNumber) SetNull() {
	*n = UIntNumber{
		Present: true,
	}
}

func (n *UIntNumber) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			w.WriteUIntNumber(n.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (n *UIntNumber) ReadJSON(r *json.Reader) bool {
	*n = UIntNumber{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	n.Defined = true

	if value, ok := r.ReadRawNumber(); ok {
		r.SkipWhitespace()
		integer, err := strconv.ParseUint(utils.UnsafeString(value), 10, 64)
		if err != nil {
			return true
		}

		n.Valid = true
		n.Value = integer
		return true
	}

	return r.SkipValue()
}
