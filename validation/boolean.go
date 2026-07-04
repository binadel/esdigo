package validation

import (
	"fmt"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Boolean struct {
	Path       FieldPath
	required   bool
	notNull    bool
	hasConst   bool
	constValue bool
	constJSON  []byte
}

func NewBoolean(path ...string) *Boolean {
	return &Boolean{
		Path: Field(path),
	}
}

func (b *Boolean) Required() *Boolean {
	b.required = true
	return b
}

func (b *Boolean) NotNull() *Boolean {
	b.notNull = true
	return b
}

// Const requires the value to equal value (JSON-Schema const).
func (b *Boolean) Const(value bool) *Boolean {
	b.hasConst, b.constValue = true, value
	b.constJSON = []byte(fmt.Sprintf("%v", value))
	return b
}

func (b *Boolean) Validate(value types.Boolean) Result[bool] {
	result := Result[bool]{
		Path:    b.Path,
		Present: value.Present,
		Defined: value.Defined,
		Value:   value.Value,
	}

	if b.required && !value.Present {
		result.Errors = append(result.Errors, errors.Required)
		return result
	}
	if b.notNull && !value.Defined {
		result.Errors = append(result.Errors, errors.NotNull)
		return result
	}
	if !value.Valid {
		// A defined value that isn't a boolean is the wrong type; a null (not
		// defined) that reached here is allowed and produces no error.
		if value.Defined {
			result.Errors = append(result.Errors, errors.InvalidBoolean)
		}
		return result
	}

	if b.hasConst && value.Value != b.constValue {
		result.Errors = append(result.Errors, rawParamError(errors.Const, errors.ParamKeyConst, b.constJSON))
	}

	return result
}
