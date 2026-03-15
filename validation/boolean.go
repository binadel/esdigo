package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/validation/errors"
)

type Boolean struct {
	Path     FieldPath
	required bool
	notNull  bool
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
		result.Errors = append(result.Errors, errors.InvalidBoolean)
		return result
	}

	return result
}
