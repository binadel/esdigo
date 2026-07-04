package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

// JsonPointer validates the JSON-Schema "json-pointer" format (RFC 6901): either
// the empty string (the whole document) or a sequence of "/"-prefixed reference
// tokens in which every "~" is escaped as "~0" or "~1".
type JsonPointer struct {
	String
}

func (p *JsonPointer) Validate(value types.String) Result[string] {
	result := Result[string]{
		Path:    p.Path,
		Errors:  p.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	if !isJsonPointer(utils.UnsafeString(value.Value)) {
		result.Errors = append(result.Errors, errors.InvalidJsonPointer)
		return result
	}

	result.Value = string(value.Value)
	return result
}

// isJsonPointer reports whether s is a valid RFC 6901 JSON pointer.
func isJsonPointer(s string) bool {
	if s == "" { // the empty pointer references the whole document
		return true
	}
	if s[0] != '/' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '~' && (i+1 >= len(s) || (s[i+1] != '0' && s[i+1] != '1')) {
			return false
		}
	}
	return true
}
