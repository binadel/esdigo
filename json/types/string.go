package types

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
)

type String struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []byte
}

func (s *String) IsPresent() bool {
	return s.Present
}

func (s *String) IsDefined() bool {
	return s.Defined
}

func (s *String) IsValid() bool {
	return s.Valid
}

func (s *String) SetNull() {
	*s = String{
		Present: true,
	}
}

func (s *String) Set(value []byte) {
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

func (s *String) SetString(value string) {
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(value),
	}
}

func (s *String) WriteJSON(w *json.Writer) bool {
	if s.Defined {
		if s.Valid {
			w.WriteStringBytes(s.Value)
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

	if value, ok := r.ReadStringBytes(); ok {
		r.SkipWhitespace()
		s.Valid = true
		s.Value = value
		return true
	}

	return r.SkipValue()
}
