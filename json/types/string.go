package types

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"time"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/utils"
	"github.com/google/uuid"
	"github.com/sosodev/duration"
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

func (s *String) SetRegex(value *regexp.Regexp) {
	if value == nil {
		s.SetNull()
		return
	}

	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetTime(value time.Time, layout string) {
	if value.IsZero() {
		s.SetNull()
		return
	}

	str := value.Format(layout)
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetDuration(value time.Duration) {
	if value == 0 {
		s.SetNull()
		return
	}

	str := duration.Format(value)
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetEmail(value *mail.Address) {
	if value == nil {
		s.SetNull()
		return
	}

	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetIP(value net.IP) {
	if value == nil {
		s.SetNull()
		return
	}

	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetUri(value *url.URL) {
	if value == nil {
		s.SetNull()
		return
	}

	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetUuid(value uuid.UUID) {
	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

func (s *String) SetNull() {
	*s = String{
		Present: true,
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

	if s.Value, s.Valid = r.ReadStringBytes(); s.Valid {
		r.SkipWhitespace()
		return true
	}

	return r.SkipValue()
}
