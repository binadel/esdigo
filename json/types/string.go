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

// String is a JSON string field with tri-state tracking: Present (the field
// appeared at all), Defined (present and non-null), and Valid (held a real
// string). See json.OptionalValue.
//
// Value holds the decoded bytes. As with json.Reader.ReadStringBytes, an
// unescaped string aliases the reader's input buffer, so do not retain a String
// read from a buffer you intend to reuse.
type String struct {
	Present bool
	Defined bool
	Valid   bool
	Value   []byte
}

// IsPresent reports whether the field appeared in the input.
func (s *String) IsPresent() bool {
	return s.Present
}

// IsDefined reports whether the field was present and non-null.
func (s *String) IsDefined() bool {
	return s.Defined
}

// IsValid reports whether a usable string was read.
func (s *String) IsValid() bool {
	return s.Valid
}

// Set assigns raw bytes and marks the field present, defined, and valid. The
// bytes are not copied.
func (s *String) Set(value []byte) {
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   value,
	}
}

// SetString is Set for a string value (no copy).
func (s *String) SetString(value string) {
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(value),
	}
}

// The Set* methods below hold a typed value in its canonical string form. Each
// treats the type's zero/nil as absent and stores null instead.

// SetRegex stores the pattern text of value, or null if value is nil.
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

// SetTime stores value formatted with layout, or null if value is the zero time.
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

// SetDuration stores value as an ISO-8601 duration, or null if value is zero.
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

// SetEmail stores the RFC 5322 form of value, or null if value is nil.
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

// SetIP stores the textual form of value, or null if value is nil.
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

// SetUri stores the string form of value, or null if value is nil.
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

// SetUuid stores the canonical form of value (the zero UUID is a valid value, so
// it is stored, not treated as null).
func (s *String) SetUuid(value uuid.UUID) {
	str := value.String()
	*s = String{
		Present: true,
		Defined: true,
		Valid:   true,
		Value:   utils.UnsafeBytes(str),
	}
}

// SetNull marks the field present but explicitly null (not defined).
func (s *String) SetNull() {
	*s = String{
		Present: true,
	}
}

// WriteJSON writes the string, or null when the field is not defined. It returns
// false only when the field is defined but invalid.
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

// ReadJSON reads a JSON string (or null) into s. Per the json.ValueReader
// contract the returned bool means "the reader can continue", NOT "the value is
// valid": a non-string is skipped and leaves Valid=false.
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
