package types

import "github.com/bindadel/esdigo/json"

type StringArray struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []string
}

func (a *StringArray) Set(value []string) {
	*a = StringArray{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (a *StringArray) SetNull() {
	*a = StringArray{
		Present: true,
	}
}

func (a *StringArray) ShouldWrite() bool {
	return a.Present
}

func (a *StringArray) WriteJSON(w *json.Writer) bool {
	if a.Defined {
		if a.Valid {
			needsComma := false
			w.BeginArray()
			for _, v := range a.Value {
				if needsComma {
					w.ValueSeparator()
				}
				w.WriteString(v)
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

func (a *StringArray) ReadJSON(r *json.Reader) bool {
	*a = StringArray{
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
			if value, ok := r.ReadString(); ok {
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
