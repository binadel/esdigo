package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Array[T json.ValueReadWriter[T]] struct {
	Path       FieldPath
	required   bool
	notNull    bool
	exactItems int
	minItems   int
	maxItems   int
}

func NewArray[T json.ValueReadWriter[T]](path ...string) *Array[T] {
	return &Array[T]{
		Path: Field(path),
	}
}

func (a *Array[T]) Required() *Array[T] {
	a.required = true
	return a
}

func (a *Array[T]) NotNull() *Array[T] {
	a.notNull = true
	return a
}

func (a *Array[T]) ExactItems(exactItems int) *Array[T] {
	a.exactItems = exactItems
	return a
}

func (a *Array[T]) MinItems(minItems int) *Array[T] {
	a.minItems = minItems
	return a
}

func (a *Array[T]) MaxItems(maxItems int) *Array[T] {
	a.maxItems = maxItems
	return a
}

func (a *Array[T]) validateRaw(value types.Array[T]) []Error {
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

func (a *Array[T]) Validate(value types.Array[T]) Result[[]T] {
	result := Result[[]T]{
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
