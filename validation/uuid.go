package validation

import (
	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
	"github.com/google/uuid"
)

type Uuid struct {
	String
	version uuid.Version
}

func (u *Uuid) Version(version byte) *Uuid {
	if version >= 1 && version <= 8 {
		u.version = uuid.Version(version)
	}
	return u
}

func (u *Uuid) Validate(value types.String) Result[uuid.UUID] {
	result := Result[uuid.UUID]{
		Path:    u.Path,
		Errors:  u.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	if !result.IsValid() {
		return result
	}

	str := utils.UnsafeString(value.Value)
	id, err := uuid.Parse(str)
	if err != nil {
		result.Errors = append(result.Errors, errors.InvalidUuid)
		return result
	}

	if u.version != 0 && u.version != id.Version() {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.UuidVersion,
			ParamKey:   errors.ParamKeyVersion,
			ParamValue: int64(u.version),
		})
		return result
	}

	result.Value = id
	return result
}
