package types

import "github.com/bindadel/esdigo/json"

type String struct {
	Present bool
	Defined bool
	Valid   bool
	Value   string
}

func (s *String) Set(value string) {
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (s *String) SetNull() {
	*s = String{
		Present: true,
	}
}

func (s *String) ShouldWrite() bool {
	return s.Present
}

func (s *String) WriteJSON(w *json.Writer) bool {
	if s.Defined {
		if s.Valid {
			w.WriteString(s.Value)
		} else {
			return false
		}
	} else {
		w.WriteNull()
	}
	return true
}

func (s *String) ReadJSON(r *json.Reader) bool {
	*s = String{
		Present: true,
	}

	r.SkipWhitespace()

	if r.ReadNull() {
		return true
	}

	s.Defined = true

	if value, ok := r.ReadString(); ok {
		r.SkipWhitespace()
		s.Valid = true
		s.Value = value
		return true
	}

	return false
}
