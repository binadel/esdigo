package validation

import (
	"net/url"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type Uri struct {
	String
	reference bool
}

// Reference relaxes the check to accept a relative URI reference in addition to an
// absolute URI (JSON-Schema uri-reference vs uri).
func (u *Uri) Reference() *Uri {
	u.reference = true
	return u
}

func (u *Uri) Validate(value types.String) Result[*url.URL] {
	result := Result[*url.URL]{
		Path:    u.Path,
		Errors:  u.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	str := utils.UnsafeString(value.Value)
	uri, err := url.Parse(str)
	if err != nil {
		if u.reference {
			result.Errors = append(result.Errors, errors.InvalidUriReference)
		} else {
			result.Errors = append(result.Errors, errors.InvalidUri)
		}
		return result
	}

	// A plain uri must be absolute; a uri-reference may be relative.
	if !u.reference && !uri.IsAbs() {
		result.Errors = append(result.Errors, errors.InvalidUri)
		return result
	}

	result.Value = uri
	return result
}
