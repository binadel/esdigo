package types

import "github.com/bindadel/esdigo/json"

type ValueFactory[T any] interface {
	CreateValue() T
}

type OmittableValue interface {
	ShouldWrite() bool
}

type ValueWriter interface {
	WriteJSON(w *json.Writer) bool
}

type ValueReader interface {
	ReadJSON(r *json.Reader) bool
}

type ValueReadWriter[T any] interface {
	ValueFactory[T]
	ValueWriter
	ValueReader
}
