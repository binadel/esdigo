package json

// OptionalValue is the tri-state contract shared by the field wrapper types.
type OptionalValue interface {
	IsPresent() bool
	IsDefined() bool
	IsValid() bool
}

// ValueReader reads a JSON value into the receiver.
type ValueReader interface {
	ReadJSON(r *Reader) bool
}

// ValueWriter writes the receiver as a JSON value.
type ValueWriter interface {
	WriteJSON(w *Writer) bool
}

// ValueReadWriter constrains a pointer type *V whose base type V can read and
// write itself as JSON. Array[V, PV] and Object[V, PV] use PV to allocate and
// populate a V in place — no reflection and no CreateValue factory method:
// a fresh element is just `var v V`, addressed through `PV(&v)`.
type ValueReadWriter[V any] interface {
	*V
	ValueReader
	ValueWriter
}
