package json

import "fmt"

type Reader struct {
	data []byte
	pos  int
	err  error
}

func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
	}
}

func (r *Reader) Error() error {
	return r.err
}

func (r *Reader) SetEofError() {
	if r.err != nil {
		return
	}

	r.err = ErrUnexpectedEOF
}

func (r *Reader) SetSyntaxError(format string, args ...any) {
	if r.err != nil {
		return
	}

	msg := fmt.Sprintf(format, args...)
	r.err = &SyntaxError{
		Message: msg,
		Offset:  r.pos,
	}
}

func (r *Reader) ReadJSON() (Value, error) {
	if r.err != nil {
		return Value{}, r.err
	}

	r.SkipWhitespace()

	if value, ok := r.ReadValue(); ok {
		if r.pos >= len(r.data) {
			return value, nil
		}
		r.SetSyntaxError("unexpected trailing character '%c'", r.data[r.pos])
		return Value{}, r.err
	}

	return Value{}, r.err
}

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
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return r.SkipNumber()
	case 't':
		b, ok := r.ReadBoolean()
		return ok && b
	case 'f':
		b, ok := r.ReadBoolean()
		return ok && !b
	case 'n':
		return r.ReadNull()
	default:
		r.SetSyntaxError("unexpected character '%c'", c)
		return false
	}
}

func (r *Reader) SkipWhitespace() {
	for r.pos < len(r.data) {
		c := r.data[r.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			r.pos++
		} else {
			break
		}
	}
}

func (r *Reader) consumeByte(expected byte) bool {
	if r.err != nil {
		return false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}

	if r.data[r.pos] == expected {
		r.pos++
		return true
	}
	return false
}
