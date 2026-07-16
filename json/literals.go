package json

import "encoding/binary"

// Little-endian word encodings of the JSON literals for the fast-path compare.
const (
	nullU32 = uint32('n') | uint32('u')<<8 | uint32('l')<<16 | uint32('l')<<24
	trueU32 = uint32('t') | uint32('r')<<8 | uint32('u')<<16 | uint32('e')<<24
	falsU32 = uint32('f') | uint32('a')<<8 | uint32('l')<<16 | uint32('s')<<24
)

// WriteNull writes the JSON literal null.
func (w *Writer) WriteNull() {
	w.data = append(w.data, 'n', 'u', 'l', 'l')
}

// ReadNull tries to read the JSON literal null. The fast path is a tiny inlinable
// word compare; error reporting for a malformed 'n...' lives in readNullSlow.
func (r *Reader) ReadNull() bool {
	if r.err == nil && r.pos+4 <= len(r.data) && binary.LittleEndian.Uint32(r.data[r.pos:]) == nullU32 {
		r.pos += 4
		return true
	}
	return r.readNullSlow()
}

func (r *Reader) readNullSlow() bool {
	if r.err != nil {
		return false
	}
	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}
	if r.data[r.pos] != 'n' {
		return false
	}

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
	if r.err == nil && r.pos+4 <= len(r.data) {
		switch binary.LittleEndian.Uint32(r.data[r.pos:]) {
		case trueU32:
			r.pos += 4
			return true, true
		case falsU32:
			if r.pos+5 <= len(r.data) && r.data[r.pos+4] == 'e' {
				r.pos += 5
				return false, true
			}
		}
	}
	return r.readBooleanSlow()
}

func (r *Reader) readBooleanSlow() (value bool, ok bool) {
	if r.err != nil {
		return false, false
	}
	if r.pos >= len(r.data) {
		r.SetEofError()
		return false, false
	}

	c := r.data[r.pos]
	switch c {
	case 't':
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
	case 'f':
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
