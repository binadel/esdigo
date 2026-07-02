package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Object[V any, PV json.ValueReadWriter[V]] struct {
	Path     FieldPath
	required bool
	notNull  bool
}

func NewObject[V any, PV json.ValueReadWriter[V]](path ...string) *Object[V, PV] {
	return &Object[V, PV]{
		Path: Field(path),
	}
}

func (o *Object[V, PV]) Required() *Object[V, PV] {
	o.required = true
	return o
}

func (o *Object[V, PV]) NotNull() *Object[V, PV] {
	o.notNull = true
	return o
}

func (o *Object[V, PV]) validateRaw(value types.Object[V, PV]) []Error {
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

func (o *Object[V, PV]) Validate(value types.Object[V, PV]) Result[V] {
	result := Result[V]{
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
