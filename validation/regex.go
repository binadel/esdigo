package validation

import (
	"regexp"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type Regex struct {
	String
}

func (r *Regex) Validate(value types.String) Result[*regexp.Regexp] {
	result := Result[*regexp.Regexp]{
		Path:    r.Path,
		Errors:  r.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	str := utils.UnsafeString(value.Value)
	regex, err := regexp.Compile(str)
	if err != nil {
		result.Errors = append(result.Errors, errors.InvalidRegex)
		return result
	}

	result.Value = regex
	return result
}
