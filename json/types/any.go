package types

import "github.com/binadel/esdigo/json"

type Any struct {
	Present bool
	Defined bool
	Valid   bool
	Value   json.Value
}

func (a *Any) IsPresent() bool {
	return a.Present
}

func (a *Any) IsDefined() bool {
	return a.Defined
}

func (a *Any) IsValid() bool {
	return a.Valid
}

func (a *Any) Set(value json.Value) {
	*a = Any{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *Any) SetNull() {
	*a = Any{
		Present: true,
	}
}

func (a *Any) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			if ok := w.WriteValue(a.Value); !ok {
				return false
			}
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (a *Any) ReadJSON(r *json.Reader) bool {
	*a = Any{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	a.Defined = true

	if a.Value, a.Valid = r.ReadValue(); a.Valid {
		r.SkipWhitespace()
		return true
	}

	return false
}
