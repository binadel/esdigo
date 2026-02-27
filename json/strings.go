package json

import "unicode/utf8"

const hex = "0123456789abcdef"

func (w *Writer) WriteString(value string) {
	w.data = append(w.data, '"')
	w.writeEscapedString(value)
	w.data = append(w.data, '"')
}

func (r *Reader) ReadRawString() ([]byte, bool) {
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

	// fast path: scan for closing quote, backslash, or invalid control characters
	for {
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
			// escape character found
			return r.readEscapedString(start)
		}

		if c < 0x20 {
			// unescaped control characters are not permitted
			r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
			return nil, false
		}

		r.pos++
	}
}

func (r *Reader) ReadString() (string, bool) {
	if raw, ok := r.ReadRawString(); ok {
		return string(raw), true
	}
	return "", false
}

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

	for r.pos < len(r.data) {
		c := r.data[r.pos]

		if c == '"' {
			r.pos++
			return true
		}

		if c == '\\' {
			// escape character found
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
			continue
		}

		if c < 0x20 {
			// unescaped control characters are not permitted
			r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
			return false
		}

		r.pos++
	}

	// if we exit the loop without returning, we hit EOF before a closing quote
	r.SetEofError()
	return false
}

func (w *Writer) writeEscapedString(value string) {
	start := 0
	for i := 0; i < len(value); i++ {
		c := value[i]

		if c >= 0x20 && c != '\\' && c != '"' {
			continue
		}

		if start < i {
			w.data = append(w.data, value[start:i]...)
		}

		switch c {
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
			w.data = append(
				w.data,
				'\\', 'u', '0', '0',
				hex[c>>4],
				hex[c&0xF],
			)
		}
		start = i + 1
	}
	if start < len(value) {
		w.data = append(w.data, value[start:]...)
	}
}

func (r *Reader) readEscapedString(start int) ([]byte, bool) {
	// allocate a buffer, we start with the capacity of the remaining data
	// to minimize re-allocations, though it may be slightly larger than needed
	buffer := make([]byte, 0, r.pos-start+16)

	// copy the clean, unescaped prefix we already scanned in the fast path
	buffer = append(buffer, r.data[start:r.pos]...)

	for {
		if r.pos >= len(r.data) {
			r.SetEofError()
			return nil, false
		}

		c := r.data[r.pos]

		if c == '"' {
			// skip closing quote and return the buffer
			r.pos++
			return buffer, true
		}

		if c < 0x20 {
			// unescaped control characters are not permitted
			r.SetSyntaxError("invalid unescaped control character 0x%02x in string", c)
			return nil, false
		}

		if c != '\\' {
			// normal character
			buffer = append(buffer, c)
			r.pos++
			continue
		}

		// we hit a backslash
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
			u, success := r.readUnicode()
			if !success {
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
		if c := r.data[r.pos]; (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			r.SetSyntaxError("invalid character '%c' in \\u hexadecimal escape", c)
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
		c := r.data[r.pos]
		val <<= 4
		switch {
		case c >= '0' && c <= '9':
			val |= rune(c - '0')
		case c >= 'a' && c <= 'f':
			val |= rune(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			val |= rune(c - 'A' + 10)
		default:
			r.SetSyntaxError("invalid character '%c' in \\u hexadecimal escape", c)
			return 0, false
		}
		r.pos++
	}
	return val, true
}
