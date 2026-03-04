package json

// Writer is a minimal, high-performance JSON output buffer.
//
// It is append-only and does not perform validation or error tracking.
// The zero value is not usable; use NewWriter to initialize it.
//
// The Writer is intended for low-level or code-generated JSON serialization
// where correctness is guaranteed by construction.
type Writer struct {
	data []byte
}

// NewWriter creates a new Writer with the specified initial capacity.
//
// The capacity defines the initial allocation size of the internal buffer.
// Choosing an appropriate capacity can reduce reallocations.
func NewWriter(capacity int) *Writer {
	return &Writer{
		data: make([]byte, 0, capacity),
	}
}

// Reset clears the buffer while retaining the underlying allocated memory.
//
// This allows the Writer to be reused without additional allocations.
func (w *Writer) Reset() {
	w.data = w.data[:0]
}

// Len returns the number of bytes currently written.
func (w *Writer) Len() int {
	return len(w.data)
}

// Bytes returns the internal buffer.
//
// The returned slice aliases the Writer's internal memory.
// Modifying the returned slice will modify the Writer's contents.
func (w *Writer) Bytes() []byte {
	return w.data
}

// Build returns the accumulated bytes.
//
// The returned slice aliases the internal buffer and is not copied.
// The error return value is always nil and exists for API symmetry
// with higher-level builders.
func (w *Writer) Build() ([]byte, error) {
	return w.data, nil
}

// WriteRawByte appends a single byte to the buffer.
//
// No validation or escaping is performed.
func (w *Writer) WriteRawByte(value byte) {
	w.data = append(w.data, value)
}

// WriteRawBytes appends a byte slice to the buffer.
//
// The bytes are copied into the internal buffer.
func (w *Writer) WriteRawBytes(value []byte) {
	w.data = append(w.data, value...)
}

// WriteRawString appends a string to the buffer.
//
// The string bytes are copied into the internal buffer.
// No validation or escaping is performed.
func (w *Writer) WriteRawString(value string) {
	w.data = append(w.data, value...)
}
