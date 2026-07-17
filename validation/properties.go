package validation

import "github.com/binadel/esdigo/validation/errors"

// Properties validates an object's property count against minProperties /
// maxProperties. The generated Validate counts the object's present fields and
// passes the total here, so the check is reported at the object's own path. esdigo
// models objects as closed structs, so the count is the number of declared
// properties present (unknown properties are not represented).
type Properties struct {
	Path FieldPath

	hasMin bool
	hasMax bool
	min    int
	max    int
}

// NewProperties creates an object property-count validator at the given path.
func NewProperties(path ...string) *Properties {
	return &Properties{Path: Field(path)}
}

func (p *Properties) Min(min int) *Properties { p.hasMin, p.min = true, min; return p }
func (p *Properties) Max(max int) *Properties { p.hasMax, p.max = true, max; return p }

// Validate checks a property count and returns a typed Result carrying the count.
func (p *Properties) Validate(count int) Result[int] {
	result := Result[int]{
		Path:    p.Path,
		Present: true,
		Defined: true,
		Value:   count,
	}
	if p.hasMin && count < p.min {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.MinProperties,
			ParamKey:   errors.ParamKeyMinProperties,
			ParamValue: int64(p.min),
		})
	}
	if p.hasMax && count > p.max {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.MaxProperties,
			ParamKey:   errors.ParamKeyMaxProperties,
			ParamValue: int64(p.max),
		})
	}
	return result
}
