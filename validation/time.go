package validation

import (
	"time"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type Time struct {
	String
	onlyDate bool
	onlyTime bool
	formats  []string
}

func (t *Time) Validate(value types.String) Result[time.Time] {
	result := Result[time.Time]{
		Path:    t.Path,
		Errors:  t.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	str := utils.UnsafeString(value.Value)
	for _, format := range t.formats {
		if parsed, err := time.Parse(format, str); err == nil {
			result.Value = parsed
			return result
		}
	}

	if t.onlyDate {
		result.Errors = append(result.Errors, errors.Date)
	} else if t.onlyTime {
		result.Errors = append(result.Errors, errors.Time)
	} else {
		result.Errors = append(result.Errors, errors.DateTime)
	}

	return result
}
