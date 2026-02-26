package types

import "github.com/bindadel/esdigo/json"

type Boolean struct {
	Present bool
	Defined bool
	Valid   bool
	Value   bool
}

func (b *Boolean) Set(value bool) {
	*b = Boolean{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (b *Boolean) SetNull() {
	*b = Boolean{
		Present: true,
	}
}

func (b *Boolean) ShouldWrite() bool {
	return b.Present
}

func (b *Boolean) WriteJSON(w *json.Writer) bool {
	if b.Defined {
		if b.Valid {
			w.WriteBoolean(b.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (b *Boolean) ReadJSON(r *json.Reader) bool {
	*b = Boolean{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	b.Defined = true

	if value, ok := r.ReadBoolean(); ok {
		r.SkipWhitespace()
		b.Valid = true
		b.Value = value
		return true
	}

	return false
}
