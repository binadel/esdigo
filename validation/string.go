package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type String struct {
	Path     FieldPath
	required bool
	notNull  bool
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

func (s *String) Email() *Email {
	return &Email{*s}
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
