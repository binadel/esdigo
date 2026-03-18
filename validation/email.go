package validation

import (
	"net/mail"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type Email struct {
	String
}

func (e *Email) Validate(value types.String) Result[*mail.Address] {
	result := Result[*mail.Address]{
		Path:    e.Path,
		Errors:  e.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	str := utils.UnsafeString(value.Value)
	email, err := mail.ParseAddress(str)
	if err != nil {
		result.Errors = append(result.Errors, errors.InvalidEmail)
		return result
	}

	result.Value = email
	return result
}
