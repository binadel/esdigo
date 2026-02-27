package types

import "github.com/bindadel/esdigo/json"

type OmittableValue interface {
	ShouldWrite() bool
}

type ValueWriter interface {
	WriteJSON(w *json.Writer) bool
}

type ValueReader interface {
	ReadJSON(r *json.Reader) bool
}

type ValueReadWriter interface {
	ValueWriter
	ValueReader
}
