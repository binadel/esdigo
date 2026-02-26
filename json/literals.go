package json

func (w *Writer) WriteNull() {
	w.data = append(w.data, "null"...)
}

func (r *Reader) ReadNull() bool {
	if r.err != nil {
		return false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}

	if r.data[r.pos] == 'n' {
		if r.pos+4 <= len(r.data) &&
			r.data[r.pos+1] == 'u' &&
			r.data[r.pos+2] == 'l' &&
			r.data[r.pos+3] == 'l' {
			r.pos += 4
			return true
		}

		for i := 1; i < 4; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false
			}
			if r.data[r.pos+i] != "null"[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'null'")
				return false
			}
		}
	}

	return false
}

func (w *Writer) WriteBoolean(value bool) {
	if value {
		w.data = append(w.data, "true"...)
	} else {
		w.data = append(w.data, "false"...)
	}
}

func (r *Reader) ReadBoolean() (bool, bool) {
	if r.err != nil {
		return false, false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false, false
	}

	c := r.data[r.pos]

	if c == 't' {
		if r.pos+4 <= len(r.data) &&
			r.data[r.pos+1] == 'r' &&
			r.data[r.pos+2] == 'u' &&
			r.data[r.pos+3] == 'e' {
			r.pos += 4
			return true, true
		}

		for i := 1; i < 4; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false, false
			}
			if r.data[r.pos+i] != "true"[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'true'")
				return false, false
			}
		}
	}

	if c == 'f' {
		if r.pos+5 <= len(r.data) &&
			r.data[r.pos+1] == 'a' &&
			r.data[r.pos+2] == 'l' &&
			r.data[r.pos+3] == 's' &&
			r.data[r.pos+4] == 'e' {
			r.pos += 5
			return false, true
		}

		for i := 1; i < 5; i++ {
			if r.pos+i >= len(r.data) {
				r.pos += i
				r.SetEofError()
				return false, false
			}
			if r.data[r.pos+i] != "false"[i] {
				r.pos += i
				r.SetSyntaxError("expected literal 'false'")
				return false, false
			}
		}
	}

	return false, false
}
