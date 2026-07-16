package validation

import (
	"math/big"

	"github.com/binadel/esdigo/validation/errors"
)

// BigInt validates a decoded arbitrary-precision integer field (types.BigInt),
// mapping it to Result[*big.Int]. It mirrors Number but compares with big.Int.Cmp
// so its bounds are exact at any magnitude. Unset bounds are nil.
type BigInt struct {
	Path     FieldPath
	required bool
	notNull  bool

	min, max     *big.Int
	exMin, exMax *big.Int
	multiple     *big.Int
	constValue   *big.Int
	enum         []*big.Int
	constJSON    []byte
	enumJSON     []byte
}

// NewBigInt creates a big-integer field validator at the given path.
func NewBigInt(path ...string) *BigInt {
	return &BigInt{Path: Field(path)}
}

func (b *BigInt) Required() *BigInt               { b.required = true; return b }
func (b *BigInt) NotNull() *BigInt                { b.notNull = true; return b }
func (b *BigInt) Min(v *big.Int) *BigInt          { b.min = v; return b }
func (b *BigInt) Max(v *big.Int) *BigInt          { b.max = v; return b }
func (b *BigInt) ExclusiveMin(v *big.Int) *BigInt { b.exMin = v; return b }
func (b *BigInt) ExclusiveMax(v *big.Int) *BigInt { b.exMax = v; return b }
func (b *BigInt) MultipleOf(v *big.Int) *BigInt   { b.multiple = v; return b }

// Const requires the value to equal v (JSON-Schema const).
func (b *BigInt) Const(v *big.Int) *BigInt {
	b.constValue = v
	b.constJSON = []byte(v.Text(10))
	return b
}

// Enum requires the value to be one of values (JSON-Schema enum).
func (b *BigInt) Enum(values ...*big.Int) *BigInt {
	b.enum = values
	b.enumJSON = bigIntsJSON(values)
	return b
}

// Validate checks a decoded big-integer field and returns a typed Result.
func (b *BigInt) Validate(field numberField[*big.Int]) Result[*big.Int] {
	result, value, done := numberBase[*big.Int](b.Path, b.required, b.notNull, field)
	if done {
		return result
	}

	if b.min != nil && value.Cmp(b.min) < 0 {
		result.Errors = append(result.Errors, paramError(errors.Minimum, errors.ParamKeyMinimum, []byte(b.min.Text(10))))
	}
	if b.max != nil && value.Cmp(b.max) > 0 {
		result.Errors = append(result.Errors, paramError(errors.Maximum, errors.ParamKeyMaximum, []byte(b.max.Text(10))))
	}
	if b.exMin != nil && value.Cmp(b.exMin) <= 0 {
		result.Errors = append(result.Errors, paramError(errors.ExclusiveMinimum, errors.ParamKeyExclusiveMinimum, []byte(b.exMin.Text(10))))
	}
	if b.exMax != nil && value.Cmp(b.exMax) >= 0 {
		result.Errors = append(result.Errors, paramError(errors.ExclusiveMaximum, errors.ParamKeyExclusiveMaximum, []byte(b.exMax.Text(10))))
	}
	if b.multiple != nil && b.multiple.Sign() != 0 && new(big.Int).Rem(value, b.multiple).Sign() != 0 {
		result.Errors = append(result.Errors, paramError(errors.MultipleOf, errors.ParamKeyMultipleOf, []byte(b.multiple.Text(10))))
	}
	if b.constValue != nil && value.Cmp(b.constValue) != 0 {
		result.Errors = append(result.Errors, rawParamError(errors.Const, errors.ParamKeyConst, b.constJSON))
	}
	if len(b.enum) > 0 && !containsBigInt(b.enum, value) {
		result.Errors = append(result.Errors, rawParamError(errors.Enum, errors.ParamKeyEnum, b.enumJSON))
	}

	if result.IsValid() {
		result.Value = value
	}
	return result
}

func bigIntsJSON(values []*big.Int) []byte {
	b := []byte{'['}
	for i, v := range values {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, v.Text(10)...)
	}
	return append(b, ']')
}

func containsBigInt(list []*big.Int, v *big.Int) bool {
	for _, x := range list {
		if v.Cmp(x) == 0 {
			return true
		}
	}
	return false
}

// BigFloat validates a decoded arbitrary-precision float field (types.BigFloat),
// mapping it to Result[*big.Float]. Bounds compare with big.Float.Cmp; MultipleOf
// is exact for the represented values (computed via big.Rat).
type BigFloat struct {
	Path     FieldPath
	required bool
	notNull  bool

	min, max     *big.Float
	exMin, exMax *big.Float
	multiple     *big.Float
	constValue   *big.Float
	enum         []*big.Float
	constJSON    []byte
	enumJSON     []byte
}

// NewBigFloat creates a big-float field validator at the given path.
func NewBigFloat(path ...string) *BigFloat {
	return &BigFloat{Path: Field(path)}
}

func (b *BigFloat) Required() *BigFloat                 { b.required = true; return b }
func (b *BigFloat) NotNull() *BigFloat                  { b.notNull = true; return b }
func (b *BigFloat) Min(v *big.Float) *BigFloat          { b.min = v; return b }
func (b *BigFloat) Max(v *big.Float) *BigFloat          { b.max = v; return b }
func (b *BigFloat) ExclusiveMin(v *big.Float) *BigFloat { b.exMin = v; return b }
func (b *BigFloat) ExclusiveMax(v *big.Float) *BigFloat { b.exMax = v; return b }
func (b *BigFloat) MultipleOf(v *big.Float) *BigFloat   { b.multiple = v; return b }

// Const requires the value to equal v (JSON-Schema const).
func (b *BigFloat) Const(v *big.Float) *BigFloat {
	b.constValue = v
	b.constJSON = bigFloatBytes(v)
	return b
}

// Enum requires the value to be one of values (JSON-Schema enum).
func (b *BigFloat) Enum(values ...*big.Float) *BigFloat {
	b.enum = values
	b.enumJSON = bigFloatsJSON(values)
	return b
}

// Validate checks a decoded big-float field and returns a typed Result.
func (b *BigFloat) Validate(field numberField[*big.Float]) Result[*big.Float] {
	result, value, done := numberBase[*big.Float](b.Path, b.required, b.notNull, field)
	if done {
		return result
	}

	if b.min != nil && value.Cmp(b.min) < 0 {
		result.Errors = append(result.Errors, paramError(errors.Minimum, errors.ParamKeyMinimum, bigFloatBytes(b.min)))
	}
	if b.max != nil && value.Cmp(b.max) > 0 {
		result.Errors = append(result.Errors, paramError(errors.Maximum, errors.ParamKeyMaximum, bigFloatBytes(b.max)))
	}
	if b.exMin != nil && value.Cmp(b.exMin) <= 0 {
		result.Errors = append(result.Errors, paramError(errors.ExclusiveMinimum, errors.ParamKeyExclusiveMinimum, bigFloatBytes(b.exMin)))
	}
	if b.exMax != nil && value.Cmp(b.exMax) >= 0 {
		result.Errors = append(result.Errors, paramError(errors.ExclusiveMaximum, errors.ParamKeyExclusiveMaximum, bigFloatBytes(b.exMax)))
	}
	if b.multiple != nil && b.multiple.Sign() != 0 && !bigFloatIsMultiple(value, b.multiple) {
		result.Errors = append(result.Errors, paramError(errors.MultipleOf, errors.ParamKeyMultipleOf, bigFloatBytes(b.multiple)))
	}
	if b.constValue != nil && value.Cmp(b.constValue) != 0 {
		result.Errors = append(result.Errors, rawParamError(errors.Const, errors.ParamKeyConst, b.constJSON))
	}
	if len(b.enum) > 0 && !containsBigFloat(b.enum, value) {
		result.Errors = append(result.Errors, rawParamError(errors.Enum, errors.ParamKeyEnum, b.enumJSON))
	}

	if result.IsValid() {
		result.Value = value
	}
	return result
}

func bigFloatBytes(v *big.Float) []byte {
	return []byte(v.Text('g', -1))
}

func bigFloatsJSON(values []*big.Float) []byte {
	b := []byte{'['}
	for i, v := range values {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, bigFloatBytes(v)...)
	}
	return append(b, ']')
}

func containsBigFloat(list []*big.Float, v *big.Float) bool {
	for _, x := range list {
		if v.Cmp(x) == 0 {
			return true
		}
	}
	return false
}

// bigFloatIsMultiple reports whether value/factor is an integer, computed exactly
// on the values' exact rationals.
func bigFloatIsMultiple(value, factor *big.Float) bool {
	vr, _ := value.Rat(nil)
	fr, _ := factor.Rat(nil)
	if vr == nil || fr == nil { // a non-finite value cannot reach here
		return true
	}
	return new(big.Rat).Quo(vr, fr).IsInt()
}
