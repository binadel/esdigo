package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Array[V any, PV json.ValueReadWriter[V]] struct {
	Path          FieldPath
	required      bool
	notNull       bool
	hasExactItems bool
	hasMinItems   bool
	hasMaxItems   bool
	exactItems    int
	minItems      int
	maxItems      int
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
	a.hasExactItems, a.exactItems = true, exactItems
	return a
}

func (a *Array[V, PV]) MinItems(minItems int) *Array[V, PV] {
	a.hasMinItems, a.minItems = true, minItems
	return a
}

func (a *Array[V, PV]) MaxItems(maxItems int) *Array[V, PV] {
	a.hasMaxItems, a.maxItems = true, maxItems
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
		// A defined value that isn't an array is the wrong type; a null (not
		// defined) that reached here is allowed and produces no error.
		if value.Defined {
			errorList = append(errorList, errors.InvalidArray)
		}
		return errorList
	}

	length := len(value.Value)

	if a.hasExactItems && length != a.exactItems {
		errorList = append(errorList, &errors.IntParamError{
			BasicError: errors.ExactItems,
			ParamKey:   errors.ParamKeyExactItems,
			ParamValue: int64(a.exactItems),
		})
	}

	if a.hasMinItems && length < a.minItems {
		errorList = append(errorList, &errors.IntParamError{
			BasicError: errors.MinItems,
			ParamKey:   errors.ParamKeyMinItems,
			ParamValue: int64(a.minItems),
		})
	}

	if a.hasMaxItems && length > a.maxItems {
		errorList = append(errorList, &errors.IntParamError{
			BasicError: errors.MaxItems,
			ParamKey:   errors.ParamKeyMaxItems,
			ParamValue: int64(a.maxItems),
		})
	}

	return errorList
}

func (a *Array[V, PV]) Validate(value types.Array[V, PV]) Result[[]PV] {
	result := Result[[]PV]{
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
