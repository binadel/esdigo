package errors

import (
	"github.com/binadel/esdigo/json"
	"github.com/binadel/esdigo/validation/rules"
)

const (
	keyCode    = `"code":`
	keyMessage = `,"message":`
)

// BasicError provides error code and message.
// Code identifies the validation rule that failed.
// Message is the human-readable description of the error.
type BasicError struct {
	Code    rules.Code
	Message string
}

// Error returns the human-readable message for the failed validation rule.
func (e *BasicError) Error() string {
	return e.Message
}

// WriteJSON writes the structured JSON representation of the validation error.
func (e *BasicError) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	w.WriteRawString(keyCode)
	w.WriteString(string(e.Code))
	w.WriteRawString(keyMessage)
	w.WriteString(e.Message)
	w.EndObject()
	return true
}
