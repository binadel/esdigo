package types

import "github.com/binadel/esdigo/json"

type Object[V any, PV json.ValueReadWriter[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   PV
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

func (o *Object[V, PV]) Set(value PV) {
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
			if ok := o.Value.WriteJSON(w); !ok {
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

func (o *Object[V, PV]) ReadJSON(r *json.Reader) bool {
	*o = Object[V, PV]{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	o.Defined = true

	o.Value = PV(new(V))
	if o.Valid = o.Value.ReadJSON(r); o.Valid {
		r.SkipWhitespace()
		return true
	}

	return r.SkipValue()
}
