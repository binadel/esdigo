package json

func (w *Writer) BeginArray() {
	w.data = append(w.data, '[')
}

func (r *Reader) BeginArray() bool {
	return r.consumeByte('[')
}

func (w *Writer) EndArray() {
	w.data = append(w.data, ']')
}

func (r *Reader) EndArray() bool {
	return r.consumeByte(']')
}

func (w *Writer) WriteArray(value []Value) bool {
	w.BeginArray()
	needsComma := false
	for _, item := range value {
		if needsComma {
			w.ValueSeparator()
		}

		if !w.WriteValue(item) {
			return false
		}

		needsComma = true
	}
	w.EndArray()
	return true
}

func (r *Reader) ReadArray() ([]Value, bool) {
	if r.BeginArray() {
		r.SkipWhitespace()

		if r.EndArray() {
			return nil, true
		}

		var array []Value

		for {
			if value, ok := r.ReadValue(); ok {
				array = append(array, value)
			} else {
				r.SetSyntaxError("expected a value after begin-array '[' or value-separator ','")
				return nil, false
			}

			if r.EndArray() {
				return array, true
			}

			if !r.ValueSeparator() {
				r.SetSyntaxError("expected either end-array ']' or value-separator ','")
				return nil, false
			}
		}
	}

	return nil, false
}

func (r *Reader) SkipArray() bool {
	if r.BeginArray() {
		r.SkipWhitespace()

		if r.EndArray() {
			return true
		}

		for {
			if !r.SkipValue() {
				r.SetSyntaxError("expected a value after begin-array '[' or value-separator ','")
				return false
			}

			r.SkipWhitespace()

			if r.EndArray() {
				return true
			}

			if !r.ValueSeparator() {
				r.SetSyntaxError("expected either end-array ']' or value-separator ','")
				return false
			}
		}
	}

	return false
}
