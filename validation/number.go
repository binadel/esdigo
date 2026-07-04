package validation

import (
	"fmt"
	"math"

	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation/errors"
)

// orderedNumber is the set of scalar number types Number validates. Big numbers
// (not comparable with <) have their own validators (BigInt, BigFloat).
type orderedNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// numberField is what the numeric validators read from a decoded wrapper
// (types.Int64, types.BigInt, ...): the tri-state plus the value and the
// classification of why it may be invalid.
type numberField[V any] interface {
	json.OptionalValue
	Unwrap() (V, json.NumberType)
}

// Aliases for the scalar backings, mirroring the types wrappers.
type (
	Int    = Number[int]
	Int8   = Number[int8]
	Int16  = Number[int16]
	Int32  = Number[int32]
	Int64  = Number[int64]
	UInt   = Number[uint]
	UInt8  = Number[uint8]
	UInt16 = Number[uint16]
	UInt32 = Number[uint32]
	UInt64 = Number[uint64]

	Float32 = Number[float32]
	Float64 = Number[float64]
)

// Number validates a decoded numeric field of type V and maps it back to a typed
// Result[V]. Besides presence/null it enforces the JSON-Schema numeric bounds and,
// when the value is unusable, reports a precise reason: not a number, not an
// integer, or out of range (derived from the field's json.NumberType).
type Number[V orderedNumber] struct {
	Path     FieldPath
	required bool
	notNull  bool

	hasMin, hasMax     bool
	hasExMin, hasExMax bool
	hasMultiple        bool
	hasConst           bool
	min, max           V
	exMin, exMax       V
	multiple           V
	constValue         V
	enum               []V
	constJSON          []byte
	enumJSON           []byte
}

// NewNumber creates a numeric field validator for V at the given path, e.g.
// NewNumber[int64]("user", "age").
func NewNumber[V orderedNumber](path ...string) *Number[V] {
	return &Number[V]{
		Path: Field(path),
	}
}

func (n *Number[V]) Required() *Number[V] {
	n.required = true
	return n
}

func (n *Number[V]) NotNull() *Number[V] {
	n.notNull = true
	return n
}

// Min requires the value to be >= v (JSON-Schema minimum).
func (n *Number[V]) Min(v V) *Number[V] {
	n.hasMin, n.min = true, v
	return n
}

// Max requires the value to be <= v (JSON-Schema maximum).
func (n *Number[V]) Max(v V) *Number[V] {
	n.hasMax, n.max = true, v
	return n
}

// ExclusiveMin requires the value to be strictly > v.
func (n *Number[V]) ExclusiveMin(v V) *Number[V] {
	n.hasExMin, n.exMin = true, v
	return n
}

// ExclusiveMax requires the value to be strictly < v.
func (n *Number[V]) ExclusiveMax(v V) *Number[V] {
	n.hasExMax, n.exMax = true, v
	return n
}

// MultipleOf requires the value to be an integer multiple of v.
func (n *Number[V]) MultipleOf(v V) *Number[V] {
	n.hasMultiple, n.multiple = true, v
	return n
}

// Const requires the value to equal v (JSON-Schema const).
func (n *Number[V]) Const(v V) *Number[V] {
	n.hasConst, n.constValue = true, v
	n.constJSON = []byte(fmt.Sprintf("%v", v))
	return n
}

// Enum requires the value to be one of values (JSON-Schema enum).
func (n *Number[V]) Enum(values ...V) *Number[V] {
	n.enum = values
	n.enumJSON = numbersJSON(values)
	return n
}

// Validate checks a decoded numeric field and returns a typed Result. On success
// Result.Value holds the number; otherwise Result.Errors describes the failure.
func (n *Number[V]) Validate(field numberField[V]) Result[V] {
	result, value, done := numberBase[V](n.Path, n.required, n.notNull, field)
	if done {
		return result
	}
	n.checkBounds(value, &result)
	if result.IsValid() {
		result.Value = value
	}
	return result
}

func (n *Number[V]) checkBounds(value V, result *Result[V]) {
	if n.hasMin && value < n.min {
		result.Errors = append(result.Errors, numberBoundError(errors.Minimum, errors.ParamKeyMinimum, n.min))
	}
	if n.hasMax && value > n.max {
		result.Errors = append(result.Errors, numberBoundError(errors.Maximum, errors.ParamKeyMaximum, n.max))
	}
	if n.hasExMin && value <= n.exMin {
		result.Errors = append(result.Errors, numberBoundError(errors.ExclusiveMinimum, errors.ParamKeyExclusiveMinimum, n.exMin))
	}
	if n.hasExMax && value >= n.exMax {
		result.Errors = append(result.Errors, numberBoundError(errors.ExclusiveMaximum, errors.ParamKeyExclusiveMaximum, n.exMax))
	}
	if n.hasMultiple && !isMultipleOf(value, n.multiple) {
		result.Errors = append(result.Errors, numberBoundError(errors.MultipleOf, errors.ParamKeyMultipleOf, n.multiple))
	}
	if n.hasConst && value != n.constValue {
		result.Errors = append(result.Errors, rawParamError(errors.Const, errors.ParamKeyConst, n.constJSON))
	}
	if len(n.enum) > 0 && !containsOrdered(n.enum, value) {
		result.Errors = append(result.Errors, rawParamError(errors.Enum, errors.ParamKeyEnum, n.enumJSON))
	}
}

// numbersJSON renders values as a JSON array of numbers, reusing the %v form the
// bound errors use so any scalar V serializes without knowing its type here.
func numbersJSON[V orderedNumber](values []V) []byte {
	b := []byte{'['}
	for i, v := range values {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, fmt.Sprintf("%v", v)...)
	}
	return append(b, ']')
}

func containsOrdered[V orderedNumber](list []V, v V) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// isMultipleOf reports whether value is an integer multiple of factor. It compares
// through float64, so it is exact for values within float64's integer range; for
// exact checks on large integers use BigInt.
func isMultipleOf[V orderedNumber](value, factor V) bool {
	if factor == 0 {
		return true
	}
	return math.Mod(float64(value), float64(factor)) == 0
}

// numberBase runs the checks common to every numeric validator — presence, null,
// and the precise reason for an unusable value. done is true when the caller
// should return result as-is; otherwise value holds the number to bounds-check.
func numberBase[V any](path FieldPath, required, notNull bool, field numberField[V]) (result Result[V], value V, done bool) {
	result.Path = path
	result.Present = field.IsPresent()
	result.Defined = field.IsDefined()

	if required && !field.IsPresent() {
		result.Errors = append(result.Errors, errors.Required)
		return result, value, true
	}
	if notNull && !field.IsDefined() {
		result.Errors = append(result.Errors, errors.NotNull)
		return result, value, true
	}
	if !field.IsValid() {
		// A defined value that isn't a usable number gets a precise reason; a null
		// (not defined) that reached here is allowed and produces no error.
		if field.IsDefined() {
			_, typ := field.Unwrap()
			result.Errors = append(result.Errors, reasonError(typ))
		}
		return result, value, true
	}

	value, _ = field.Unwrap()
	return result, value, false
}

// reasonError turns a number's classification into the matching precise error.
func reasonError(typ json.NumberType) Error {
	switch typ {
	case json.NumberTypeReal:
		return errors.NotInteger
	case json.NumberTypeInvalid:
		return errors.InvalidNumber
	default: // Big, Overflow, or an integer that didn't fit the target width
		return errors.OutOfRange
	}
}

// paramError builds a bound error carrying the offending bound as pre-formatted
// JSON number bytes.
func paramError(base errors.BasicError, key string, value []byte) Error {
	return &errors.NumberParamError{
		BasicError: base,
		ParamKey:   key,
		ParamValue: value,
	}
}

// rawParamError builds an error carrying a pre-serialized JSON parameter (an enum
// array or a const value) written verbatim.
func rawParamError(base errors.BasicError, key string, value []byte) Error {
	return &errors.RawParamError{
		BasicError: base,
		ParamKey:   key,
		ParamValue: value,
	}
}

func numberBoundError[V orderedNumber](base errors.BasicError, key string, bound V) Error {
	return paramError(base, key, []byte(fmt.Sprintf("%v", bound)))
}
