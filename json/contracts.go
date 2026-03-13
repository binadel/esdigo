package json

type OptionalValue interface {
	IsPresent() bool
	IsDefined() bool
	IsValid() bool
}

type ValueFactory[T any] interface {
	CreateValue() T
}

type ValueWriter interface {
	WriteJSON(w *Writer) bool
}

type ValueReader interface {
	ReadJSON(r *Reader) bool
}

type ValueReadWriter[T any] interface {
	ValueFactory[T]
	ValueWriter
	ValueReader
}
