package validation

// RawNumber validates a raw-passthrough number field (types.RawNumber). The value is
// kept verbatim as its JSON bytes, so no numeric bounds apply — RawNumber enforces
// only presence and non-null, and maps the field to Result[[]byte].
type RawNumber struct {
	Path     FieldPath
	required bool
	notNull  bool
}

// NewRawNumber creates a raw-number field validator at the given path.
func NewRawNumber(path ...string) *RawNumber {
	return &RawNumber{Path: Field(path)}
}

func (r *RawNumber) Required() *RawNumber { r.required = true; return r }
func (r *RawNumber) NotNull() *RawNumber  { r.notNull = true; return r }

// Validate checks a raw-number field's presence and null-ness and returns a typed
// Result holding the verbatim JSON bytes.
func (r *RawNumber) Validate(field numberField[[]byte]) Result[[]byte] {
	result, value, done := numberBase[[]byte](r.Path, r.required, r.notNull, field)
	if done {
		return result
	}
	if result.IsValid() {
		result.Value = value
	}
	return result
}
