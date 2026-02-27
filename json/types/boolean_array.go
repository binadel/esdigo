package types

import "github.com/bindadel/esdigo/json"

type BooleanArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []bool
}

func (a *BooleanArray) Set(value []bool) {
	*a = BooleanArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *BooleanArray) SetNull() {
	*a = BooleanArray{
		Present: true,
	}
}

func (a *BooleanArray) ShouldWrite() bool {
	return a.Present
}

func (a *BooleanArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			needsComma := false
			w.BeginArray()
			for _, v := range a.Value {
				if needsComma {
					w.ValueSeparator()
				}
				w.WriteBoolean(v)
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

func (a *BooleanArray) ReadJSON(r *json.Reader) bool {
	*a = BooleanArray{
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
			if value, ok := r.ReadBoolean(); ok {
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
