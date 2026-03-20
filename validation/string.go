package validation

import (
	"regexp"
	"unicode/utf8"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

var regexCache = &utils.RegexCache{}

type String struct {
	Path     FieldPath
	required bool
	notNull  bool
	len      int
	minLen   int
	maxLen   int
	pattern  *regexp.Regexp
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
	s.len = length
	return s
}

func (s *String) MinLength(minLength int) *String {
	s.minLen = minLength
	return s
}

func (s *String) MaxLength(maxLength int) *String {
	s.maxLen = maxLength
	return s
}

func (s *String) Pattern(pattern string) *String {
	s.pattern = regexCache.MustGet(pattern)
	return s
}

func (s *String) Email() *Email {
	return &Email{*s}
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
		errorList = append(errorList, errors.InvalidString)
		return errorList
	}

	str := utils.UnsafeString(value.Value)

	var length int
	if s.len > 0 || s.minLen > 0 || s.maxLen > 0 {
		length = utf8.RuneCountInString(str)
	}

	if s.len > 0 {
		if length != s.len {
			errorList = append(errorList, &errors.IntParamError{
				BasicError: errors.Length,
				ParamKey:   errors.ParamKeyLength,
				ParamValue: int64(s.len),
			})
		}
	}

	if s.minLen > 0 {
		if length < s.minLen {
			errorList = append(errorList, &errors.IntParamError{
				BasicError: errors.MinLength,
				ParamKey:   errors.ParamKeyMinLength,
				ParamValue: int64(s.minLen),
			})
		}
	}

	if s.maxLen > 0 {
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
