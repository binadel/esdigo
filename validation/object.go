package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Object[T json.ValueReadWriter[T]] struct {
	Path     FieldPath
	required bool
	notNull  bool
}

func NewObject[T json.ValueReadWriter[T]](path ...string) *Object[T] {
	return &Object[T]{
		Path: Field(path),
	}
}

func (o *Object[T]) Required() *Object[T] {
	o.required = true
	return o
}

func (o *Object[T]) NotNull() *Object[T] {
	o.notNull = true
	return o
}

func (o *Object[T]) validateRaw(value types.Object[T]) []Error {
	var errorList []Error

	if o.required && !value.Present {
		errorList = append(errorList, errors.Required)
		return errorList
	}

	if o.notNull && !value.Defined {
		errorList = append(errorList, errors.NotNull)
		return errorList
	}

	if !value.Valid {
		errorList = append(errorList, errors.InvalidString)
		return errorList
	}

	return errorList
}

func (o *Object[T]) Validate(value types.Object[T]) Result[T] {
	result := Result[T]{
		Path:    o.Path,
		Errors:  o.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	result.Value = value.Value
	return result
}
