package json

// BeginObject writes the JSON begin-object token '{' to the writer buffer.
func (w *Writer) BeginObject() {
	w.data = append(w.data, '{')
}

// BeginObject attempts to read the JSON begin-object token '{'.
func (r *Reader) BeginObject() bool {
	return r.readByte('{')
}

// EndObject writes the JSON end-object token '}' to the writer buffer.
func (w *Writer) EndObject() {
	w.data = append(w.data, '}')
}

// EndObject attempts to read the JSON end-object token '}'.
func (r *Reader) EndObject() bool {
	return r.readByte('}')
}

// NameSeparator writes the JSON name/value separator ':'.
func (w *Writer) NameSeparator() {
	w.data = append(w.data, ':')
}

// NameSeparator attempts to read the JSON name/value separator ':'.
func (r *Reader) NameSeparator() bool {
	return r.readByte(':')
}

// ValueSeparator writes the JSON value separator ','.
func (w *Writer) ValueSeparator() {
	w.data = append(w.data, ',')
}

// ValueSeparator attempts to read the JSON value separator ','.
func (r *Reader) ValueSeparator() bool {
	return r.readByte(',')
}

// WriteObject serializes a map[string]Value as a JSON object.
func (w *Writer) WriteObject(value map[string]Value) (ok bool) {
	w.BeginObject()
	needsComma := false
	for k, v := range value {
		if needsComma {
			w.ValueSeparator()
		}

		w.WriteString(k)
		w.NameSeparator()
		if !w.WriteValue(v) {
			return false
		}

		needsComma = true
	}
	w.EndObject()
	return true
}

// ReadObject parses a JSON object and returns its members as a map[string]Value.
func (r *Reader) ReadObject() (map[string]Value, bool) {
	if r.BeginObject() {
		r.SkipWhitespace()

		object := make(map[string]Value)

		if r.EndObject() {
			return object, true
		}

		for {
			if name, ok := r.ReadString(); ok {
				r.SkipWhitespace()

				if r.NameSeparator() {
					if value, ok := r.ReadValue(); ok {
						object[name] = value
					} else {
						r.SetSyntaxError("expected a value after name-separator ':'")
					}
				} else {
					r.SetSyntaxError("expected a name-separator ':' after name")
					return nil, false
				}

				if r.EndObject() {
					return object, true
				}

				if !r.ValueSeparator() {
					r.SetSyntaxError("expected either end-object '}' or value-separator ','")
					return nil, false
				}
			} else {
				r.SetSyntaxError("expected a name after begin-object '{' or value-separator ','")
				return nil, false
			}
		}
	}

	return nil, false
}

// SkipObject skips a JSON object without constructing Value objects.
func (r *Reader) SkipObject() bool {
	if r.BeginObject() {
		r.SkipWhitespace()

		if r.EndObject() {
			return true
		}

		for {
			if r.SkipString() {
				r.SkipWhitespace()

				if r.NameSeparator() {
					if !r.SkipValue() {
						return false
					}
				} else {
					r.SetSyntaxError("expected a name-separator ':' after name")
					return false
				}

				if r.EndObject() {
					return true
				}

				if !r.ValueSeparator() {
					r.SetSyntaxError("expected either end-object '}' or value-separator ','")
					return false
				}

				r.SkipWhitespace()
			} else {
				r.SetSyntaxError("expected a name after begin-object '{' or value-separator ','")
				return false
			}
		}
	}

	return false
}
