package errors

import "github.com/binadel/esdigo/json"

const (
	ParamKeyMinimum   = `,"minimum":`
	ParamKeyMaximum   = `,"maximum":`
	ParamKeyFactor    = `,"factor":`
	ParamKeyLength    = `,"length":`
	ParamKeyMinLength = `,"minLength":`
	ParamKeyMaxLength = `,"maxLength":`
	ParamKeyVersion   = `,"version":`
)

type IntParamError struct {
	BasicError
	ParamKey   string
	ParamValue int64
}

func (e *IntParamError) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	w.WriteRawString(keyCode)
	w.WriteString(string(e.Code))
	w.WriteRawString(keyMessage)
	w.WriteString(e.Message)
	w.WriteRawString(e.ParamKey)
	w.WriteIntNumber(e.ParamValue)
	w.EndObject()
	return true
}
