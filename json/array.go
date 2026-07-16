package json

// BeginArray writes the JSON begin-array token '[' to the writer buffer.
func (w *Writer) BeginArray() {
	w.data = append(w.data, '[')
}

// BeginArray attempts to read the JSON begin-array token '['. On success it also
// records the nesting depth, failing if the maximum depth would be exceeded.
func (r *Reader) BeginArray() bool {
	return r.readByte('[') && r.enterDepth()
}

// EndArray writes the JSON end-array token ']' to the writer buffer.
func (w *Writer) EndArray() {
	w.data = append(w.data, ']')
}

// EndArray attempts to read the JSON end-array token ']', releasing one level of
// nesting depth when it matches.
func (r *Reader) EndArray() bool {
	if r.readByte(']') {
		r.depth--
		return true
	}
	return false
}

// WriteArray serializes a slice of Value as a JSON array.
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

// ReadArray parses a JSON array and returns its elements as a slice of Value.
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

// SkipArray skips a JSON array without allocating or constructing Value
func (r *Reader) SkipArray() bool {
	if r.BeginArray() {
		r.SkipWhitespace()

		if r.EndArray() {
			return true
		}

		for {
			if !r.SkipValue() {
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
