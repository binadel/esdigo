package json

import (
	"errors"
	"fmt"
)

var ErrUnexpectedEOF = errors.New("unexpected end of json input")

// SyntaxError represents a violation of the JSON specification.
type SyntaxError struct {
	Message string
	Offset  int
}

// Error implements the standard Go error interface.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("json syntax error at byte offset %d: %s", e.Offset, e.Message)
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
