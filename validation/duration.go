package validation

import (
	"time"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
	"github.com/sosodev/duration"
)

type Duration struct {
	String
}

func (d *Duration) Validate(value types.String) Result[time.Duration] {
	result := Result[time.Duration]{
		Path:    d.Path,
		Errors:  d.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	str := utils.UnsafeString(value.Value)
	parsed, err := duration.Parse(str)
	if err != nil {
		result.Errors = append(result.Errors, errors.Duration)
		return result
	}

	// Note: Conversion of IOS 8601 duration to a go duration has a little fuzziness for year and month.
	result.Value = parsed.ToTimeDuration()
	return result
}
