package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Array[V any, PV json.ValueReadWriter[V]] struct {
	Path       FieldPath
	required   bool
	notNull    bool
	exactItems int
	minItems   int
	maxItems   int
}

func NewArray[V any, PV json.ValueReadWriter[V]](path ...string) *Array[V, PV] {
	return &Array[V, PV]{
		Path: Field(path),
	}
}

func (a *Array[V, PV]) Required() *Array[V, PV] {
	a.required = true
	return a
}

func (a *Array[V, PV]) NotNull() *Array[V, PV] {
	a.notNull = true
	return a
}

func (a *Array[V, PV]) ExactItems(exactItems int) *Array[V, PV] {
	a.exactItems = exactItems
	return a
}

func (a *Array[V, PV]) MinItems(minItems int) *Array[V, PV] {
	a.minItems = minItems
	return a
}

func (a *Array[V, PV]) MaxItems(maxItems int) *Array[V, PV] {
	a.maxItems = maxItems
	return a
}

func (a *Array[V, PV]) validateRaw(value types.Array[V, PV]) []Error {
	var errorList []Error

	if a.required && !value.Present {
		errorList = append(errorList, errors.Required)
		return errorList
	}

	if a.notNull && !value.Defined {
		errorList = append(errorList, errors.NotNull)
		return errorList
	}

	if !value.Valid {
		errorList = append(errorList, errors.InvalidString)
		return errorList
	}

	//length := len(value.Value)

	//if a.exactItems > 0 && length == a.exactItems {
	//	errorList = append(errorList, errors.ExactItems)
	//}

	return errorList
}

func (a *Array[V, PV]) Validate(value types.Array[V, PV]) Result[[]V] {
	result := Result[[]V]{
		Path:    a.Path,
		Errors:  a.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	result.Value = value.Value
	return result
}
