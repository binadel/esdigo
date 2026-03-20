package validation

import (
	"net/url"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type Uri struct {
	String
}

func (u *Uri) Validate(value types.String) Result[*url.URL] {
	result := Result[*url.URL]{
		Path:    u.Path,
		Errors:  u.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	str := utils.UnsafeString(value.Value)
	uri, err := url.Parse(str)
	if err != nil {
		result.Errors = append(result.Errors, errors.InvalidUri)
		return result
	}

	result.Value = uri
	return result
}
