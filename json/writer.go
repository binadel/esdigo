package json

import "fmt"

type Writer struct {
	data []byte
	err  error
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Error() error {
	return w.err
}

func (w *Writer) SetError(format string, args ...any) {
	if w.err != nil {
		return
	}

	w.err = fmt.Errorf(format, args...)
}

func (w *Writer) Build() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.data, nil
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
