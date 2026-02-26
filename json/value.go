package json

type ValueType int

const (
	ValueTypeInvalid ValueType = iota
	ValueTypeNull
	ValueTypeFalse
	ValueTypeTrue
	ValueTypeNumber
	ValueTypeString
	ValueTypeArray
	ValueTypeObject
)

type Value struct {
	Type    ValueType
	Payload any
}

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
			w.SetError("failed to cast number payload")
			return
		}
	case ValueTypeString:
		if str, ok := value.Payload.(string); ok {
			w.WriteString(str)
		} else {
			w.SetError("failed to cast string payload")
			return
		}
	case ValueTypeArray:
		if array, ok := value.Payload.([]Value); ok {
			w.WriteArray(array)
		} else {
			w.SetError("failed to cast array payload")
			return
		}
	case ValueTypeObject:
		if object, ok := value.Payload.(map[string]Value); ok {
			w.WriteObject(object)
		} else {
			w.SetError("failed to cast object payload")
			return
		}
	default:
		w.SetError("invalid value type: %v", value.Type)
		return
	}
	return true
}

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
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if payload, ok := r.ReadNumber(); ok {
			value.Type = ValueTypeNumber
			value.Payload = payload
		} else {
			return
		}
	case 't':
		if _, ok := r.ReadBoolean(); ok {
			value.Type = ValueTypeTrue
		} else {
			return
		}
	case 'f':
		if _, ok := r.ReadBoolean(); ok {
			value.Type = ValueTypeFalse
		} else {
			return
		}
	case 'n':
		if r.ReadNull() {
			value.Type = ValueTypeNull
		} else {
			return
		}
	default:
		return
	}

	r.SkipWhitespace()

	return value, true
}
