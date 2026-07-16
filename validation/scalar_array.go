package validation

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation/errors"
)

// scalarArrayField is what ScalarArray reads from a lean, unboxed scalar array
// wrapper (types.StringArray, types.Int64Array, types.BooleanArray, ...): the
// tri-state plus the element slice.
type scalarArrayField[V any] interface {
	json.OptionalValue
	Elements() []V
}

// ScalarArray validates a lean scalar array (types.StringArray, types.Int64Array,
// ...) at the array level: presence, null, item counts and uniqueness. Because its
// elements are unboxed and comparable, uniqueness is a direct set membership rather
// than the generic Array's per-element serialization. It maps to Result[[]V]. For
// arrays whose elements need their own validation, use the generic Array.
type ScalarArray[V comparable] struct {
	Path     FieldPath
	required bool
	notNull  bool

	hasExactItems bool
	hasMinItems   bool
	hasMaxItems   bool
	uniqueItems   bool
	exactItems    int
	minItems      int
	maxItems      int
}

// NewScalarArray creates a lean scalar-array validator at the given path.
func NewScalarArray[V comparable](path ...string) *ScalarArray[V] {
	return &ScalarArray[V]{Path: Field(path)}
}

func (a *ScalarArray[V]) Required() *ScalarArray[V] { a.required = true; return a }
func (a *ScalarArray[V]) NotNull() *ScalarArray[V]  { a.notNull = true; return a }

func (a *ScalarArray[V]) ExactItems(exactItems int) *ScalarArray[V] {
	a.hasExactItems, a.exactItems = true, exactItems
	return a
}

func (a *ScalarArray[V]) MinItems(minItems int) *ScalarArray[V] {
	a.hasMinItems, a.minItems = true, minItems
	return a
}

func (a *ScalarArray[V]) MaxItems(maxItems int) *ScalarArray[V] {
	a.hasMaxItems, a.maxItems = true, maxItems
	return a
}

// UniqueItems requires every element to be distinct (JSON-Schema uniqueItems).
func (a *ScalarArray[V]) UniqueItems() *ScalarArray[V] {
	a.uniqueItems = true
	return a
}

// Validate checks a decoded lean scalar array and returns a typed Result.
func (a *ScalarArray[V]) Validate(field scalarArrayField[V]) Result[[]V] {
	result := Result[[]V]{
		Path:    a.Path,
		Present: field.IsPresent(),
		Defined: field.IsDefined(),
	}

	if a.required && !field.IsPresent() {
		result.Errors = append(result.Errors, errors.Required)
		return result
	}
	if a.notNull && field.IsPresent() && !field.IsDefined() {
		result.Errors = append(result.Errors, errors.NotNull)
		return result
	}
	if !field.IsValid() {
		// A defined value that isn't a valid array is the wrong type; a null (not
		// defined) that reached here is allowed and produces no error.
		if field.IsDefined() {
			result.Errors = append(result.Errors, errors.InvalidArray)
		}
		return result
	}

	values := field.Elements()
	length := len(values)

	if a.hasExactItems && length != a.exactItems {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.ExactItems,
			ParamKey:   errors.ParamKeyExactItems,
			ParamValue: int64(a.exactItems),
		})
	}
	if a.hasMinItems && length < a.minItems {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.MinItems,
			ParamKey:   errors.ParamKeyMinItems,
			ParamValue: int64(a.minItems),
		})
	}
	if a.hasMaxItems && length > a.maxItems {
		result.Errors = append(result.Errors, &errors.IntParamError{
			BasicError: errors.MaxItems,
			ParamKey:   errors.ParamKeyMaxItems,
			ParamValue: int64(a.maxItems),
		})
	}
	if a.uniqueItems && hasDuplicateScalars(values) {
		result.Errors = append(result.Errors, errors.UniqueItems)
	}

	if result.IsValid() {
		result.Value = values
	}
	return result
}

// hasDuplicateScalars reports whether values contains a repeated element.
func hasDuplicateScalars[V comparable](values []V) bool {
	seen := make(map[V]struct{}, len(values))
	for _, v := range values {
		if _, dup := seen[v]; dup {
			return true
		}
		seen[v] = struct{}{}
	}
	return false
}
