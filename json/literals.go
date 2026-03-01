package json

// WriteNull writes the JSON literal null.
func (w *Writer) WriteNull() {
	w.data = append(w.data, 'n', 'u', 'l', 'l')
}

// ReadNull tries to read the JSON literal null.
func (r *Reader) ReadNull() bool {
	if r.err != nil {
		return false
	}

	remain := len(r.data) - r.pos
	if remain < 1 {
		r.SetEofError()
		return false
	}

	// fast path: comparison
	if remain >= 4 && string(r.data[r.pos:r.pos+4]) == "null" {
		r.pos += 4
		return true
	}

	// slow path: error reporting
	if r.data[r.pos] == 'n' {
		const literal = "null"
		for i := 1; i < 4; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false
			}
			if r.data[r.pos+i] != literal[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'null'")
				return false
			}
		}
	}

	return false
}

// WriteBoolean writes the JSON literal "true" or "false" based on boolean argument.
func (w *Writer) WriteBoolean(value bool) {
	if value {
		w.data = append(w.data, 't', 'r', 'u', 'e')
	} else {
		w.data = append(w.data, 'f', 'a', 'l', 's', 'e')
	}
}

// ReadBoolean tries to read the JSON literal "true" or "false".
func (r *Reader) ReadBoolean() (value bool, ok bool) {
	if r.err != nil {
		return false, false
	}

	remain := len(r.data) - r.pos
	if remain < 1 {
		r.SetEofError()
		return false, false
	}

	c := r.data[r.pos]

	if c == 't' {
		// fast path: comparison
		if remain >= 4 && string(r.data[r.pos:r.pos+4]) == "true" {
			r.pos += 4
			return true, true
		}

		// slow path: error reporting
		const literal = "true"
		for i := 1; i < 4; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false, false
			}
			if r.data[r.pos+i] != literal[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'true'")
				return false, false
			}
		}
	} else if c == 'f' {
		// fast path: comparison
		if remain >= 5 && string(r.data[r.pos:r.pos+5]) == "false" {
			r.pos += 5
			return false, true
		}

		// slow path: error reporting
		const literal = "false"
		for i := 1; i < 5; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false, false
			}
			if r.data[r.pos+i] != literal[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'false'")
				return false, false
			}
		}
	}

	return false, false
}
