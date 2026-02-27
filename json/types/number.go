package types

import "github.com/bindadel/esdigo/json"

type Number struct {
	Present bool
	Defined bool
	Valid   bool
	Value   json.NumberValue
}

func (n *Number) Set(value json.NumberValue) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (n *Number) SetInt(value int64) {
	coefficient := uint64(value)
	if value < 0 {
		coefficient = uint64(-value)
	}
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
		Value: json.NumberValue{
			Negative:    value < 0,
			Type:        json.NumberTypeInteger,
			Coefficient: coefficient,
			Exponent:    0,
		},
	}
}

func (n *Number) SetUInt(value uint64) {
	*n = Number{
		Present: true,
		Defined: true,
		Valid:   true,
	}
	n.Value = json.NumberValue{
		Negative:    false,
		Type:        json.NumberTypeInteger,
		Coefficient: value,
		Exponent:    0,
	}
}

func (n *Number) SetNull() {
	*n = Number{
		Present: true,
	}
}

func (n *Number) ShouldWrite() bool {
	return n.Present
}

func (n *Number) WriteJSON(w *json.Writer) bool {
	if n.Defined {
		if n.Valid {
			w.WriteNumber(n.Value.Negative, n.Value.Coefficient, 0, int64(n.Value.Exponent))
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

	if value, ok := r.ReadNumber(); ok {
		r.SkipWhitespace()
		n.Valid = true
		n.Value = value
		return true
	}

	return r.SkipValue()
}
