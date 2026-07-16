package json

import "fmt"

// ErrorKind classifies a parse failure so callers can branch on the cause
// without matching Message text — e.g. a depth-limit hit (a possible DoS) wants
// different handling than ordinary malformed input or a truncated stream. The
// set is open for extension: add a constant plus a String case; callers switch
// on it after errors.As.
type ErrorKind uint8

const (
	// KindSyntax when the input violates the JSON grammar (malformed).
	KindSyntax ErrorKind = iota
	// KindTruncated when the input ended in the middle of a value (unexpected EOF).
	KindTruncated
	// KindDepthLimit when a configured resource limit was exceeded (nesting depth).
	KindDepthLimit
)

func (k ErrorKind) String() string {
	switch k {
	case KindSyntax:
		return "syntax"
	case KindTruncated:
		return "truncated"
	case KindDepthLimit:
		return "depth-limit"
	default:
		return "unknown"
	}
}

// SyntaxError is the single error type the reader produces, and it means exactly
// one thing: parsing cannot continue. That happens in only three ways — the bytes
// violate the JSON grammar (KindSyntax), the input ended mid-value
// (KindTruncated), or a limit that cannot be skipped past was hit
// (KindDepthLimit). It is deliberately narrow.
//
// A value that is well-formed JSON but unusable for its target is NOT a
// SyntaxError: a number that overflows, a "1.5" read into an integer, a
// wrong-typed field — the reader reads these fine and keeps going, and the field
// is marked Valid=false on its wrapper instead (see OptionalValue). Business-rule
// failures (min/max, pattern, ...) are a third layer, in the validation package.
// The rule of thumb: reserve a Go error for "cannot continue"; everything else is
// a value status, not an error.
//
// Kind classifies the failure for programmatic handling (via errors.As, then a
// switch on Kind); Message and Offset (the byte position where the failure was
// detected) are for human diagnostics.
type SyntaxError struct {
	Kind    ErrorKind
	Message string
	Offset  int
}

// Error implements the standard Go error interface.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("json %s error at byte offset %d: %s", e.Kind, e.Offset, e.Message)
}

// Position lazily calculates the 1-based line and column number.
// This is intentionally kept out of the main parsing loop to maintain max performance.
// You pass the original JSON []byte to it when you want to log the error.
func (e *SyntaxError) Position(data []byte) (line, column int) {
	line = 1
	column = 1

	limit := e.Offset
	if limit > len(data) {
		limit = len(data)
	}

	for i := 0; i < limit; i++ {
		if data[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
