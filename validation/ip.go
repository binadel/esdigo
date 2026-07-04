package validation

import (
	"net"

	"github.com/binadel/esdigo/json/types"
	"github.com/binadel/esdigo/utils"
	"github.com/binadel/esdigo/validation/errors"
)

type IP struct {
	String
	version int
}

func (i *IP) Version4() *IP {
	i.version = 4
	return i
}

func (i *IP) Version6() *IP {
	i.version = 6
	return i
}

func (i *IP) Validate(value types.String) Result[net.IP] {
	result := Result[net.IP]{
		Path:    i.Path,
		Errors:  i.validateRaw(value),
		Present: value.Present,
		Defined: value.Defined,
	}

	// Stop on a base error; also skip parsing an allowed null (no string to parse).
	if !result.IsValid() || !value.Valid {
		return result
	}

	str := utils.UnsafeString(value.Value)
	ip := net.ParseIP(str)
	if ip == nil {
		result.Errors = append(result.Errors, errors.InvalidIP)
		return result
	}

	if i.version == 4 {
		ip := ip.To4()
		if ip == nil {
			result.Errors = append(result.Errors, errors.InvalidIPv4)
			return result
		}
	}

	if i.version == 6 {
		if ipv4 := ip.To4(); ipv4 != nil {
			result.Errors = append(result.Errors, errors.InvalidIPv6)
			return result
		}
	}

	result.Value = ip
	return result
}
