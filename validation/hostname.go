package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

// Hostname validates the JSON-Schema "hostname" format (RFC 1123): dot-separated
// labels of letters, digits and hyphens, each 1–63 chars and not hyphen-bounded,
// with a total length of at most 253.
type Hostname struct {
	String
}

func (h *Hostname) Validate(value types.String) Result[string] {
	result := Result[string]{
		Path:    h.Path,
		Errors:  h.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	if !isHostname(utils.UnsafeString(value.Value)) {
		result.Errors = append(result.Errors, errors.InvalidHostname)
		return result
	}

	result.Value = string(value.Value)
	return result
}

// isHostname reports whether s is a valid RFC 1123 hostname.
func isHostname(s string) bool {
	if len(s) < 1 || len(s) > 253 {
		return false
	}

	labelLen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '.':
			if labelLen == 0 || s[i-1] == '-' { // empty label or trailing hyphen
				return false
			}
			labelLen = 0
		case c == '-':
			if labelLen == 0 { // leading hyphen
				return false
			}
			labelLen++
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
			labelLen++
		default:
			return false
		}
		if labelLen > 63 {
			return false
		}
	}

	// the final label must be non-empty and not end with a hyphen
	return labelLen > 0 && s[len(s)-1] != '-'
}
