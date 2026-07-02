package types

import "github.com/binadel/esdigo/json"

type Array[V any, PV json.ValueReadWriter[V]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []PV
}

func (a *Array[V, PV]) IsPresent() bool {
	return a.Present
}

func (a *Array[V, PV]) IsDefined() bool {
	return a.Defined
}

func (a *Array[V, PV]) IsValid() bool {
	return a.Valid
}

func (a *Array[V, PV]) Set(value []PV) {
	*a = Array[V, PV]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *Array[V, PV]) SetNull() {
	*a = Array[V, PV]{
		Present: true,
	}
}

func (a *Array[V, PV]) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			w.BeginArray()
			for i, v := range a.Value {
				if i > 0 {
					w.ValueSeparator()
				}
				if ok := v.WriteJSON(w); !ok {
					return false
				}
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

func (a *Array[V, PV]) ReadJSON(r *json.Reader) bool {
	*a = Array[V, PV]{
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
			item := PV(new(V))
			if item.ReadJSON(r) {
				a.Value = append(a.Value, item)
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
