package json

import "fmt"

type Writer struct {
	data []byte
	err  error
}

func (w *Writer) setError(format string, args ...any) {
	if w.err != nil {
		return
	}

	w.err = fmt.Errorf(format, args...)
}

func (w *Writer) WriteRawByte(value byte) {
	w.data = append(w.data, value)
}

func (w *Writer) WriteRawBytes(value []byte) {
	w.data = append(w.data, value...)
}

func (w *Writer) WriteRawString(value string) {
	w.data = append(w.data, value...)
}
