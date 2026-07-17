package errors

import "github.com/binadel/esdigo/json"

const (
	ParamKeyMinimum          = `,"minimum":`
	ParamKeyMaximum          = `,"maximum":`
	ParamKeyExclusiveMinimum = `,"exclusiveMinimum":`
	ParamKeyExclusiveMaximum = `,"exclusiveMaximum":`
	ParamKeyMultipleOf       = `,"multipleOf":`
	ParamKeyFactor           = `,"factor":`
	ParamKeyLength           = `,"length":`
	ParamKeyMinLength        = `,"minLength":`
	ParamKeyMaxLength        = `,"maxLength":`
	ParamKeyExactItems       = `,"exactItems":`
	ParamKeyMinItems         = `,"minItems":`
	ParamKeyMaxItems         = `,"maxItems":`
	ParamKeyMinProperties    = `,"minProperties":`
	ParamKeyMaxProperties    = `,"maxProperties":`
	ParamKeyEnum             = `,"enum":`
	ParamKeyConst            = `,"const":`
	ParamKeyVersion          = `,"version":`
)

// IntParamError is a BasicError carrying one integer parameter (e.g. a length or
// item count), serialized as a JSON number under ParamKey.
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

// NumberParamError is a BasicError carrying one numeric parameter as pre-formatted
// JSON number bytes, so a bound of any numeric type (int, float, big) can be
// reported under ParamKey without knowing its Go type here.
type NumberParamError struct {
	BasicError
	ParamKey   string
	ParamValue []byte
}

func (e *NumberParamError) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	w.WriteRawString(keyCode)
	w.WriteString(string(e.Code))
	w.WriteRawString(keyMessage)
	w.WriteString(e.Message)
	w.WriteRawString(e.ParamKey)
	w.WriteRawNumber(e.ParamValue)
	w.EndObject()
	return true
}

// RawParamError is a BasicError carrying one pre-serialized JSON parameter — an
// enum array or a const value — written verbatim under ParamKey. It lets a
// validator echo the allowed value(s) of any type without knowing them here.
type RawParamError struct {
	BasicError
	ParamKey   string
	ParamValue []byte
}

func (e *RawParamError) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	w.WriteRawString(keyCode)
	w.WriteString(string(e.Code))
	w.WriteRawString(keyMessage)
	w.WriteString(e.Message)
	w.WriteRawString(e.ParamKey)
	w.WriteRawBytes(e.ParamValue)
	w.EndObject()
	return true
}
