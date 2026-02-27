package types

import "github.com/bindadel/esdigo/json"

type Object[T ValueReadWriter] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   T
}

func (o *Object[T]) Set(value T) {
	*o = Object[T]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (o *Object[T]) SetNull() {
	*o = Object[T]{
		Present: true,
	}
}

func (o *Object[T]) ShouldWrite() bool {
	return o.Present
}

func (o *Object[T]) WriteJSON(w *json.Writer) bool {
	if o.Defined {
		if o.Valid {
			o.Value.WriteJSON(w)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (o *Object[T]) ReadJSON(r *json.Reader) bool {
	*o = Object[T]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	o.Defined = true

	if o.Value.ReadJSON(r) {
		r.SkipWhitespace()
		o.Valid = true
		return true
	}

	return r.SkipValue()
}
