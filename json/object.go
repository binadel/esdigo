package json

func (w *Writer) BeginObject() {
	w.data = append(w.data, '{')
}

func (r *Reader) BeginObject() bool {
	return r.consumeByte('{')
}

func (w *Writer) EndObject() {
	w.data = append(w.data, '}')
}

func (r *Reader) EndObject() bool {
	return r.consumeByte('}')
}

func (w *Writer) NameSeparator() {
	w.data = append(w.data, ':')
}

func (r *Reader) NameSeparator() bool {
	return r.consumeByte(':')
}

func (w *Writer) ValueSeparator() {
	w.data = append(w.data, ',')
}

func (r *Reader) ValueSeparator() bool {
	return r.consumeByte(',')
}

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

		switch v.Type {

		}

		needsComma = true
	}
	w.EndObject()
	return true
}

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
