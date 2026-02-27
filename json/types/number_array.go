package types

import "github.com/bindadel/esdigo/json"

type NumberArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []json.NumberValue
}

func (a *NumberArray) Set(value []json.NumberValue) {
	*a = NumberArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *NumberArray) SetIntArray(value []int64) {
	*a = NumberArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   make([]json.NumberValue, len(value)),
	}
	for i, v := range value {
		coefficient := uint64(v)
		if v < 0 {
			coefficient = uint64(-v)
		}
		a.Value[i] = json.NumberValue{
			Negative:    v < 0,
			Type:        json.NumberTypeInteger,
			Coefficient: coefficient,
			Exponent:    0,
		}
	}
}

func (a *NumberArray) SetUIntArray(value []uint64) {
	*a = NumberArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   make([]json.NumberValue, len(value)),
	}
	for i, v := range value {
		a.Value[i] = json.NumberValue{
			Negative:    false,
			Type:        json.NumberTypeInteger,
			Coefficient: v,
			Exponent:    0,
		}
	}
}

func (a *NumberArray) SetNull() {
	*a = NumberArray{
		Present: true,
	}
}

func (a *NumberArray) ShouldWrite() bool {
	return a.Present
}

func (a *NumberArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			needsComma := false
			w.BeginArray()
			for _, v := range a.Value {
				if needsComma {
					w.ValueSeparator()
				}
				w.WriteNumber(v.Negative, v.Coefficient, 0, int64(v.Exponent))
				needsComma = true
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
	*a = NumberArray{
		Present: true,
	}

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
			if value, ok := r.ReadNumber(); ok {
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
