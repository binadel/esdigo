package json

import (
	"encoding/binary"
	"math/bits"
	"unicode/utf8"

	"github.com/binadel/esdigo/utils"
)

const hex = "0123456789abcdef"

// hexVal maps a byte to its hexadecimal value, or 0xff if it is not a hex digit.
var hexVal [256]uint8

func init() {
	for i := range hexVal {
		hexVal[i] = 0xff
	}
	for c := byte('0'); c <= '9'; c++ {
		hexVal[c] = c - '0'
	}
	for c := byte('a'); c <= 'f'; c++ {
		hexVal[c] = c - 'a' + 10
	}
	for c := byte('A'); c <= 'F'; c++ {
		hexVal[c] = c - 'A' + 10
	}
}

// stringSpecial sets the high (0x80) bit of each lane of an 8-byte little-endian
// word whose byte must be handled inside a JSON string: a quote ('"'), a backslash
// ('\\'), or a control character (< 0x20). bits.TrailingZeros64(result)>>3 gives
// the offset of the first such byte.
//
// The control test can also flag a 0x20 that immediately follows a control byte
// (a SWAR subtraction borrow), but such a false flag can only ever appear AFTER a
// genuine special byte, so the FIRST flagged lane is always genuine.
func stringSpecial(word uint64) uint64 {
	const ones = uint64(0x0101010101010101)
	const high = uint64(0x8080808080808080)
	q := word ^ 0x2222222222222222 // '"'  -> zero lanes where quote
	s := word ^ 0x5c5c5c5c5c5c5c5c // '\\' -> zero lanes where backslash
	hasQuote := (q - ones) &^ q
	hasSlash := (s - ones) &^ s
	hasCtrl := (word - 0x2020202020202020) &^ word
	return (hasQuote | hasSlash | hasCtrl) & high
}

// indexSpecial returns the index of the first byte in b that must be handled
// inside a JSON string ('"', '\\', or < 0x20), or len(b) if there is none. It
// scans eight bytes at a time (SWAR) with a byte-wise tail.
func indexSpecial(b []byte) int {
	i, n := 0, len(b)
	for i+8 <= n {
		if mask := stringSpecial(binary.LittleEndian.Uint64(b[i:])); mask != 0 {
			return i + bits.TrailingZeros64(mask)>>3
		}
		i += 8
	}
	for i < n {
		if c := b[i]; c == '"' || c == '\\' || c < 0x20 {
			return i
		}
		i++
	}
	return n
}

// WriteString writes value as a quoted, escaped JSON string. Only the characters
// JSON requires are escaped (quote, backslash, and control characters); '<', '>',
// '&' and U+2028/2029 are emitted raw, so the output is not HTML/JS-embed safe.
func (w *Writer) WriteString(value string) {
	w.data = append(w.data, '"')
	w.writeEscapedString(utils.UnsafeBytes(value))
	w.data = append(w.data, '"')
}

// WriteStringBytes is WriteString for a []byte, avoiding a []byte→string
// conversion at the call site.
func (w *Writer) WriteStringBytes(value []byte) {
	w.data = append(w.data, '"')
	w.writeEscapedString(value)
	w.data = append(w.data, '"')
}

// ReadString reads the next JSON string and returns its decoded contents as a
// string. Unlike ReadStringBytes the result is always a copy, so it stays valid
// after the input buffer is reused.
func (r *Reader) ReadString() (string, bool) {
	if bytes, ok := r.ReadStringBytes(); ok {
		return string(bytes), true
	}
	return "", false
}

// ReadStringBytes reads the next JSON string and returns its decoded bytes,
// resolving escapes. It returns false without consuming input if the next value
// is not a string, and sets a reader error only for a malformed string.
//
// ALIASING: when the string contains no escapes the result is a zero-copy
// sub-slice of the input buffer; only an escaped string is a fresh allocation. Do
// not retain the result across a reuse of the input buffer (or copy it, e.g. via
// ReadString) if you might.
func (r *Reader) ReadStringBytes() ([]byte, bool) {
	if r.err != nil {
		return nil, false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return nil, false
	}

	if r.data[r.pos] != '"' {
		return nil, false
	}
	r.pos++

	start := r.pos

	// fast path: scan to the closing quote, a backslash, or a control character
	r.pos += indexSpecial(r.data[r.pos:])

	if r.pos >= len(r.data) {
		r.SetEofError()
		return nil, false
	}

	c := r.data[r.pos]
	if c == '"' {
		// zero allocation success
		value := r.data[start:r.pos]
		r.pos++
		return value, true
	}
	if c == '\\' {
		return r.readEscapedString(start)
	}

	// unescaped control characters are not permitted
	r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
	return nil, false
}

// SkipString validates and advances past the next JSON string without decoding
// it. Like ReadStringBytes it returns false (without consuming) on a non-string,
// and sets a reader error on a malformed one.
func (r *Reader) SkipString() bool {
	if r.err != nil {
		return false
	}

	if r.pos >= len(r.data) {
		r.SetEofError()
		return false
	}

	if r.data[r.pos] != '"' {
		return false
	}
	r.pos++

	for {
		r.pos += indexSpecial(r.data[r.pos:])

		if r.pos >= len(r.data) {
			r.SetEofError()
			return false
		}

		c := r.data[r.pos]
		if c == '"' {
			r.pos++
			return true
		}
		if c < 0x20 {
			r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
			return false
		}

		// backslash: skip the escape
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return false
		}

		switch r.data[r.pos] {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			r.pos++
		case 'u':
			r.pos++
			if !r.skipUnicode() {
				return false
			}
		default:
			r.SetSyntaxError("invalid escape character '\\%c' in string", r.data[r.pos])
			return false
		}
	}
}

func (w *Writer) writeEscapedString(value []byte) {
	for {
		i := indexSpecial(value)
		w.data = append(w.data, value[:i]...)
		if i == len(value) {
			return
		}

		switch c := value[i]; c {
		case '\\', '"':
			w.data = append(w.data, '\\', c)
		case '\b':
			w.data = append(w.data, '\\', 'b')
		case '\f':
			w.data = append(w.data, '\\', 'f')
		case '\n':
			w.data = append(w.data, '\\', 'n')
		case '\r':
			w.data = append(w.data, '\\', 'r')
		case '\t':
			w.data = append(w.data, '\\', 't')
		default:
			// control characters < 0x20
			w.data = append(w.data, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xF])
		}

		value = value[i+1:]
	}
}

func (r *Reader) readEscapedString(start int) ([]byte, bool) {
	// allocate a buffer starting from the clean prefix plus a little headroom
	buffer := make([]byte, 0, r.pos-start+16)
	buffer = append(buffer, r.data[start:r.pos]...)

	for {
		// copy the clean run up to the next special byte in one shot
		runStart := r.pos
		r.pos += indexSpecial(r.data[r.pos:])
		buffer = append(buffer, r.data[runStart:r.pos]...)

		if r.pos >= len(r.data) {
			r.SetEofError()
			return nil, false
		}

		c := r.data[r.pos]
		if c == '"' {
			r.pos++
			return buffer, true
		}
		if c < 0x20 {
			r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
			return nil, false
		}

		// backslash
		r.pos++
		if r.pos >= len(r.data) {
			r.SetEofError()
			return nil, false
		}

		escapeChar := r.data[r.pos]
		switch escapeChar {
		case '"', '\\', '/':
			buffer = append(buffer, escapeChar)
			r.pos++
		case 'b':
			buffer = append(buffer, '\b')
			r.pos++
		case 'f':
			buffer = append(buffer, '\f')
			r.pos++
		case 'n':
			buffer = append(buffer, '\n')
			r.pos++
		case 'r':
			buffer = append(buffer, '\r')
			r.pos++
		case 't':
			buffer = append(buffer, '\t')
			r.pos++
		case 'u':
			r.pos++
			u, ok := r.readUnicode()
			if !ok {
				return nil, false
			}
			// encode the rune back into utf-8 bytes in our buffer
			buffer = utf8.AppendRune(buffer, u)
		default:
			r.SetSyntaxError("invalid escape character '\\%c' in string", escapeChar)
			return nil, false
		}
	}
}

// readUnicode parses a \uXXXX sequence, including handling UTF-16 surrogate pairs.
func (r *Reader) readUnicode() (rune, bool) {
	h, success := r.parseHex4()
	if !success {
		return 0, false
	}

	// check if it's a high surrogate (U+D800 to U+DBFF)
	if h >= 0xD800 && h <= 0xDBFF {
		// expect a matching low surrogate \uXXXX
		if r.pos+6 > len(r.data) || r.data[r.pos] != '\\' || r.data[r.pos+1] != 'u' {
			// dangling high surrogate. Go's standard is to replace with RuneError
			return utf8.RuneError, true
		}

		savedPos := r.pos
		r.pos += 2 // skip '\u'

		r2, success := r.parseHex4()
		if !success {
			return 0, false
		}

		// check if r2 is a valid low surrogate (U+DC00 to U+DFFF)
		if r2 >= 0xDC00 && r2 <= 0xDFFF {
			// combine surrogate pair into a single rune
			combined := (((h - 0xD800) << 10) | (r2 - 0xDC00)) + 0x10000
			return combined, true
		}

		// invalid low surrogate. revert position so the invalid escape can be processed normally,
		// and return RuneError for the high surrogate.
		r.pos = savedPos
		return utf8.RuneError, true
	}

	// unpaired low surrogate (invalid JSON Unicode, but standard fallback is RuneError)
	if h >= 0xDC00 && h <= 0xDFFF {
		return utf8.RuneError, true
	}

	return h, true
}

func (r *Reader) skipUnicode() bool {
	if r.pos+4 > len(r.data) {
		r.pos = len(r.data)
		r.SetEofError()
		return false
	}

	for i := 0; i < 4; i++ {
		if hexVal[r.data[r.pos]] == 0xff {
			r.SetSyntaxError("invalid character '%c' in \\u hexadecimal escape", r.data[r.pos])
			return false
		}
		r.pos++
	}

	return true
}

// parseHex4 reads exactly 4 hexadecimal characters and converts them to a rune.
func (r *Reader) parseHex4() (rune, bool) {
	if r.pos+4 > len(r.data) {
		r.pos = len(r.data)
		r.SetEofError()
		return 0, false
	}

	var val rune
	for i := 0; i < 4; i++ {
		h := hexVal[r.data[r.pos]]
		if h == 0xff {
			r.SetSyntaxError("invalid character '%c' in \\u hexadecimal escape", r.data[r.pos])
			return 0, false
		}
		val = val<<4 | rune(h)
		r.pos++
	}
	return val, true
}
