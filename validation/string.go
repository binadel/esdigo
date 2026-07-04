package validation

import (
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

var regexCache = &utils.RegexCache{}

type String struct {
	Path      FieldPath
	required  bool
	notNull   bool
	hasLen    bool
	hasMinLen bool
	hasMaxLen bool
	len       int
	minLen    int
	maxLen    int
	pattern   *regexp.Regexp
}

func NewString(path ...string) *String {
	return &String{
		Path: Field(path),
	}
}

func (s *String) Required() *String {
	s.required = true
	return s
}

func (s *String) NotNull() *String {
	s.notNull = true
	return s
}

func (s *String) Length(length int) *String {
	s.hasLen, s.len = true, length
	return s
}

func (s *String) MinLength(minLength int) *String {
	s.hasMinLen, s.minLen = true, minLength
	return s
}

func (s *String) MaxLength(maxLength int) *String {
	s.hasMaxLen, s.maxLen = true, maxLength
	return s
}

func (s *String) Pattern(pattern string) *String {
	s.pattern = regexCache.MustGet(pattern)
	return s
}

func (s *String) Email() *Email {
	return &Email{*s}
}

func (s *String) IP() *IP {
	return &IP{*s, 0}
}

func (s *String) Regex() *Regex {
	return &Regex{*s}
}

func (s *String) Date(formats ...string) *Time {
	if len(formats) == 0 {
		formats = []string{time.DateOnly}
	}
	return &Time{*s, true, false, formats}
}

func (s *String) Time(formats ...string) *Time {
	if len(formats) == 0 {
		formats = []string{time.TimeOnly}
	}
	return &Time{*s, false, true, formats}
}

func (s *String) DateTime(formats ...string) *Time {
	if len(formats) == 0 {
		formats = []string{time.RFC3339Nano}
	}
	return &Time{*s, false, false, formats}
}

func (s *String) Duration() *Duration {
	return &Duration{*s}
}

func (s *String) Uri() *Uri {
	return &Uri{*s}
}

func (s *String) Uuid() *Uuid {
	return &Uuid{*s, 0}
}

func (s *String) validateRaw(value types.String) []Error {
	var errorList []Error

	if s.required && !value.Present {
		errorList = append(errorList, errors.Required)
		return errorList
	}

	if s.notNull && !value.Defined {
		errorList = append(errorList, errors.NotNull)
		return errorList
	}

	if !value.Valid {
		// A defined value that isn't a usable string is the wrong type; a null (not
		// defined) that reached here is allowed and produces no error.
		if value.Defined {
			errorList = append(errorList, errors.InvalidString)
		}
		return errorList
	}

	str := utils.UnsafeString(value.Value)

	var length int
	if s.hasLen || s.hasMinLen || s.hasMaxLen {
		length = utf8.RuneCountInString(str)
	}

	if s.hasLen {
		if length != s.len {
			errorList = append(errorList, &errors.IntParamError{
				BasicError: errors.Length,
				ParamKey:   errors.ParamKeyLength,
				ParamValue: int64(s.len),
			})
		}
	}

	if s.hasMinLen {
		if length < s.minLen {
			errorList = append(errorList, &errors.IntParamError{
				BasicError: errors.MinLength,
				ParamKey:   errors.ParamKeyMinLength,
				ParamValue: int64(s.minLen),
			})
		}
	}

	if s.hasMaxLen {
		if length > s.maxLen {
			errorList = append(errorList, &errors.IntParamError{
				BasicError: errors.MaxLength,
				ParamKey:   errors.ParamKeyMaxLength,
				ParamValue: int64(s.maxLen),
			})
		}
	}

	if s.pattern != nil {
		if !s.pattern.MatchString(str) {
			errorList = append(errorList, errors.Pattern)
		}
	}

	return errorList
}

func (s *String) Validate(value types.String) Result[string] {
	result := Result[string]{
		Path:    s.Path,
		Errors:  s.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	result.Value = string(value.Value)
	return result
}
