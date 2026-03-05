package json

import "fmt"

// Reader parses JSON from a byte slice. It maintains a position and an error;
// once an error is set, all read methods return failure and no new error is set.
//
// ReadX contract (ReadNull, ReadString, ReadNumber, ReadBoolean, and structural
// reads like BeginObject): each ReadX tries to read X. Exactly one of three
// things happens:
//  1. The value is as expected → return true (pos advanced).
//  2. The value is not of the desired type at all (e.g. expected 'n' for null
//     but got something else) → return false, do not set error, do not increment pos.
//  3. The value starts with the expected character and pos is incremented, but
//     later content is invalid (e.g. "nulx" for null) → set error and return false.
//
// So errors are only set in case 3: after committing to the token we discover
// it is malformed. Callers can use case 2 to try another ReadX or report their
// own error.
type Reader struct {
	data []byte
	pos  int
	err  error
}

// NewReader creates a new Reader over the provided byte slice.
// The input slice is not copied.
func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
	}
}

// Error returns the first error encountered during parsing.
func (r *Reader) Error() error {
	return r.err
}

// SetEofError sets an unexpected EOF error if no prior error exists.
func (r *Reader) SetEofError() {
	if r.err == nil {
		r.err = ErrUnexpectedEOF
	}
}

// SetSyntaxError records a syntax error at the current offset.
// The error is only set if no prior error exists.
func (r *Reader) SetSyntaxError(format string, args ...any) {
	if r.err == nil {
		r.err = &SyntaxError{
			Message: fmt.Sprintf(format, args...),
			Offset:  r.pos,
		}
	}
}

// ReadJSON parses a complete JSON value and ensures no trailing
// non-whitespace characters remain.
func (r *Reader) ReadJSON() (Value, error) {
	if r.err != nil {
		return Value{}, r.err
	}

	r.SkipWhitespace()

	value, ok := r.ReadValue()
	if !ok {
		return Value{}, r.err
	}

	r.SkipWhitespace()

	if r.pos != len(r.data) {
		r.SetSyntaxError("unexpected trailing character '%c'", r.data[r.pos])
		return Value{}, r.err
	}

	return value, nil
}

// SkipWhitespace advances past JSON whitespace characters.
// This is a hot-path function and avoids unnecessary bounds checks.
func (r *Reader) SkipWhitespace() {
	data := r.data
	pos := r.pos

	if pos < len(data) && data[pos] > ' ' {
		return
	}

	for pos < len(data) {
		c := data[pos]
		if c != ' ' && c != '\n' && c != '\r' && c != '\t' {
			break
		}
		pos++
	}

	r.pos = pos
}

// readByte consumes the next byte if it matches expected.
// Returns true if matched and consumed.
func (r *Reader) readByte(expected byte) bool {
	if r.err != nil {
		return false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}

	if r.data[r.pos] != expected {
		return false
	}

	r.pos++
	return true
}
