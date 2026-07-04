package json

// OptionalValue is the tri-state contract shared by the field wrapper types:
// Present (the field appeared at all), Defined (it was non-null), and Valid (the
// value was usable).
//
// IsValid()==false is the recoverable counterpart to a *SyntaxError: the value
// was well-formed JSON but unusable for THIS field — it overflowed, was the wrong
// type, was out of range, and so on. It never aborts the parse; the enclosing
// object or array keeps reading its other fields. Only a broken byte stream (a
// SyntaxError) stops the reader.
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
