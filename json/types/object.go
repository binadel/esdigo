package types

import "github.com/binadel/esdigo/json"

type Object[V any, PV json.ValueReadWriter[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   V
}

func (o *Object[V, PV]) IsPresent() bool {
	return o.Present
}

func (o *Object[V, PV]) IsDefined() bool {
	return o.Defined
}

func (o *Object[V, PV]) IsValid() bool {
	return o.Valid
}

func (o *Object[V, PV]) Set(value V) {
	*o = Object[V, PV]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (o *Object[V, PV]) SetNull() {
	*o = Object[V, PV]{
		Present: true,
	}
}

func (o *Object[V, PV]) WriteJSON(w *json.Writer) bool {
	if o.Defined {
		if o.Valid {
			PV(&o.Value).WriteJSON(w)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (o *Object[V, PV]) ReadJSON(r *json.Reader) bool {
	*o = Object[V, PV]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	o.Defined = true

	if PV(&o.Value).ReadJSON(r) {
		r.SkipWhitespace()
		o.Valid = true
		return true
	}

	return r.SkipValue()
}
