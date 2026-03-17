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

func (s *String) Validate(value types.String) Result[string] {
	result := Result[string]{
		Path:    s.Path,
		Present: value.Present,
		Defined: value.Defined,
	}

	if s.required && !value.Present {
		result.Errors = append(result.Errors, errors.Required)
		return result
	}
	if s.notNull && !value.Defined {
		result.Errors = append(result.Errors, errors.NotNull)
		return result
	}
	if !value.Valid {
		result.Errors = append(result.Errors, errors.InvalidString)
		return result
	}

	result.Value = string(value.Value)
	return result
}
