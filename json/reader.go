package json

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

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

// Reset reuses the Reader for new input, clearing position and error.
// This lets a Reader be pooled and reused across parses without allocation.
// The input slice is not copied.
func (r *Reader) Reset(data []byte) {
	r.data = data
	r.pos = 0
	r.err = nil
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

// zeroLanes sets the high (0x80) bit of every 8-byte lane whose byte is exactly
// 0x00, and clears all other bits. Unlike the shorter (x-ones)&^x&0x80 form it is
// EXACT per lane: masking each byte to 7 bits before the add stops a byte's carry
// from leaking into the next lane, so there are no false positives.
func zeroLanes(x uint64) uint64 {
	const lo7 = uint64(0x7f7f7f7f7f7f7f7f)
	const high = uint64(0x8080808080808080)
	// (low7 + 0x7f) sets a lane's high bit iff its low 7 bits are non-zero; OR-ing
	// x folds bit 7 back in, so the high bit ends up set iff the byte is non-zero.
	y := (x & lo7) + lo7
	return ^(y | x | lo7) & high
}

// nonWhitespace sets the high bit of each lane of an 8-byte little-endian word
// whose byte is NOT JSON whitespace (space, tab, LF, CR). Membership must be exact
// (a false positive would skip a real token byte), so it uses zeroLanes rather
// than the cheaper string scanner.
func nonWhitespace(word uint64) uint64 {
	const high = uint64(0x8080808080808080)
	isWS := zeroLanes(word^0x2020202020202020) | // ' '
		zeroLanes(word^0x0909090909090909) | // '\t'
		zeroLanes(word^0x0a0a0a0a0a0a0a0a) | // '\n'
		zeroLanes(word^0x0d0d0d0d0d0d0d0d) // '\r'
	return isWS ^ high
}

// SkipWhitespace advances past JSON whitespace (space, tab, LF, CR). The common
// "already at a token" case is a tiny inlinable check; the whitespace-skipping
// work lives in skipWhitespace so this wrapper stays inlinable in hot callers.
func (r *Reader) SkipWhitespace() {
	if r.pos < len(r.data) && r.data[r.pos] > ' ' {
		return
	}
	r.skipWhitespace()
}

func (r *Reader) skipWhitespace() {
	data := r.data
	pos := r.pos
	n := len(data)

	// SWAR: skip eight whitespace bytes at a time
	for pos+8 <= n {
		if mask := nonWhitespace(binary.LittleEndian.Uint64(data[pos:])); mask != 0 {
			r.pos = pos + bits.TrailingZeros64(mask)>>3
			return
		}
		pos += 8
	}

	// byte-by-byte tail
	for pos < n {
		c := data[pos]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
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
