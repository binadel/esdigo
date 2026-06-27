package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Number struct {
	Path     FieldPath
	required bool
	notNull  bool
}

func NewNumber(path ...string) *Number {
	return &Number{
		Path: Field(path),
	}
}

func (n *Number) Required() *Number {
	n.required = true
	return n
}

func (n *Number) NotNull() *Number {
	n.notNull = true
	return n
}

func (n *Number) validateRaw(value types.Number) []Error {
	var errorList []Error

	if n.required && !value.Present {
		errorList = append(errorList, errors.Required)
		return errorList
	}

	if n.notNull && !value.Defined {
		errorList = append(errorList, errors.NotNull)
		return errorList
	}

	if !value.Valid {
		errorList = append(errorList, errors.InvalidString)
		return errorList
	}

	return errorList
}

func (n *Number) Validate(value types.Number) Result[string] {
	result := Result[string]{
		Path:    n.Path,
		Errors:  n.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	result.Value = string(value.Value)
	return result
}
