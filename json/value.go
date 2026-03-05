package json

type ValueType int

const (
	ValueTypeNull ValueType = iota
	ValueTypeFalse
	ValueTypeTrue
	ValueTypeNumber
	ValueTypeString
	ValueTypeArray
	ValueTypeObject
)

// Value represents a JSON value that holds the type and payload.
type Value struct {
	Type    ValueType
	Payload any
}

// WriteValue writes the next JSON value.
// Writes the value and returns true if a value was successfully written.
func (w *Writer) WriteValue(value Value) (result bool) {
	switch value.Type {
	case ValueTypeNull:
		w.WriteNull()
	case ValueTypeFalse:
		w.WriteBoolean(false)
	case ValueTypeTrue:
		w.WriteBoolean(true)
	case ValueTypeNumber:
		if number, ok := value.Payload.([]byte); ok {
			w.WriteRawNumber(number)
		} else {
			return
		}
	case ValueTypeString:
		if str, ok := value.Payload.(string); ok {
			w.WriteString(str)
		} else {
			return
		}
	case ValueTypeArray:
		if array, ok := value.Payload.([]Value); ok {
			w.WriteArray(array)
		} else {
			return
		}
	case ValueTypeObject:
		if object, ok := value.Payload.(map[string]Value); ok {
			w.WriteObject(object)
		} else {
			return
		}
	default:
		return
	}
	return true
}

// ReadValue reads the next JSON value.
// Returns the value and true if a value was successfully read.
func (r *Reader) ReadValue() (value Value, result bool) {
	if r.err != nil {
		return
	}

	r.SkipWhitespace()

	if r.pos >= len(r.data) {
		r.SetEofError()
		return
	}

	c := r.data[r.pos]
	switch c {
	case '{':
		if payload, ok := r.ReadObject(); ok {
			value.Type = ValueTypeObject
			value.Payload = payload
		} else {
			return
		}
	case '[':
		if payload, ok := r.ReadArray(); ok {
			value.Type = ValueTypeArray
			value.Payload = payload
		} else {
			return
		}
	case '"':
		if payload, ok := r.ReadString(); ok {
			value.Type = ValueTypeString
			value.Payload = payload
		} else {
			return
		}
	case 'n':
		if r.ReadNull() {
			value.Type = ValueTypeNull
		} else {
			return
		}
	case 't', 'f':
		if b, ok := r.ReadBoolean(); ok {
			if b {
				value.Type = ValueTypeTrue
			} else {
				value.Type = ValueTypeFalse
			}
		} else {
			return
		}
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if payload, ok := r.ReadNumber(); ok {
			value.Type = ValueTypeNumber
			value.Payload = payload
		} else {
			return
		}
	default:
		r.SetSyntaxError("unexpected character '%c'", c)
		return
	}

	r.SkipWhitespace()

	return value, true
}

// SkipValue skips over the next JSON value without constructing it.
// Returns true if a value was successfully skipped.
func (r *Reader) SkipValue() bool {
	if r.err != nil {
		return false
	}

	r.SkipWhitespace()

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}

	c := r.data[r.pos]
	switch c {
	case '{':
		return r.SkipObject()
	case '[':
		return r.SkipArray()
	case '"':
		return r.SkipString()
	case 'n':
		return r.ReadNull()
	case 't', 'f':
		_, ok := r.ReadBoolean()
		return ok
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return r.SkipNumber()
	default:
		r.SetSyntaxError("unexpected character '%c'", c)
		return false
	}
}
