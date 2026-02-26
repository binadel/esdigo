package json

import "fmt"

type Reader struct {
	data []byte
	pos  int
	err  error
}

func (r *Reader) setEofError() {
	if r.err != nil {
		return
	}

	r.err = ErrUnexpectedEOF
}

func (r *Reader) setSyntaxError(format string, args ...any) {
	if r.err != nil {
		return
	}

	msg := fmt.Sprintf(format, args...)
	r.err = &SyntaxError{
		Message: msg,
		Offset:  r.pos,
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
