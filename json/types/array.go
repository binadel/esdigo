package types

import "github.com/bindadel/esdigo/json"

type Array[T ValueReadWriter[T]] struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []T
}

func (a *Array[T]) Set(value []T) {
	*a = Array[T]{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *Array[T]) SetNull() {
	*a = Array[T]{
		Present: true,
	}
}

func (a *Array[T]) ShouldWrite() bool {
	return a.Present
}

func (a *Array[T]) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			needsComma := false
			w.BeginArray()
			for _, v := range a.Value {
				if needsComma {
					w.ValueSeparator()
				}
				v.WriteJSON(w)
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

func (a *Array[T]) ReadJSON(r *json.Reader) bool {
	*a = Array[T]{
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

		var factory T
		for {
			item := factory.CreateValue()
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
