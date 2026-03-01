package json

import (
	"errors"
	"testing"
)

// ========================
// WriteNull
// ========================

func TestWriter_WriteNull(t *testing.T) {
	w := NewWriter()

	w.WriteNull()

	if string(w.data) != "null" {
		t.Fatalf("expected 'null', got %q", string(w.data))
	}
}

func TestWriter_WriteNull_Append(t *testing.T) {
	w := NewWriter()
	w.data = append(w.data, `"name":`...)

	w.WriteNull()

	if string(w.data) != `"name":null` {
		t.Fatalf(`expected '"name":null', got %q`, string(w.data))
	}
}

// ========================
// ReadNull
// ========================

func TestReader_ReadNull_FastPath(t *testing.T) {
	r := NewReader([]byte("null"))

	ok := r.ReadNull()

	if !ok {
		t.Fatal("expected success")
	}
	if r.pos != 4 {
		t.Fatalf("expected pos=4 got %d", r.pos)
	}
}

func TestReader_ReadNull_WrongLiteral(t *testing.T) {
	r := NewReader([]byte("nuxx"))

	ok := r.ReadNull()

	if ok {
		t.Fatal("expected failure")
	}
	if r.err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestReader_ReadNull_EOF(t *testing.T) {
	r := NewReader([]byte("nu"))

	ok := r.ReadNull()

	if ok {
		t.Fatal("expected failure")
	}
	if r.err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestReader_ReadNull_NotStartingWithN(t *testing.T) {
	r := NewReader([]byte("true"))

	ok := r.ReadNull()

	if ok {
		t.Fatal("expected false")
	}
	if r.pos != 0 {
		t.Fatalf("pos should not move")
	}
}

func TestReader_ReadNull_PreExistingError(t *testing.T) {
	r := NewReader([]byte("null"))
	r.err = errors.New("existing")

	ok := r.ReadNull()

	if ok {
		t.Fatal("expected false")
	}
	if r.pos != 0 {
		t.Fatal("reader should not advance")
	}
}

// ========================
// WriteBoolean
// ========================

func TestWriter_WriteBoolean_True(t *testing.T) {
	w := NewWriter()

	w.WriteBoolean(true)

	if string(w.data) != "true" {
		t.Fatalf("expected 'true', got %q", string(w.data))
	}
}

func TestWriter_WriteBoolean_False(t *testing.T) {
	w := NewWriter()

	w.WriteBoolean(false)

	if string(w.data) != "false" {
		t.Fatalf("expected 'false', got %q", string(w.data))
	}
}

// ========================
// ReadBoolean
// ========================

func TestReader_ReadBoolean_True_FastPath(t *testing.T) {
	r := NewReader([]byte("true"))

	v, ok := r.ReadBoolean()

	if !ok || !v {
		t.Fatal("expected true")
	}
	if r.pos != 4 {
		t.Fatalf("expected pos=4 got %d", r.pos)
	}
}

func TestReader_ReadBoolean_False_FastPath(t *testing.T) {
	r := NewReader([]byte("false"))

	v, ok := r.ReadBoolean()

	if !ok || v {
		t.Fatal("expected false")
	}
	if r.pos != 5 {
		t.Fatalf("expected pos=5 got %d", r.pos)
	}
}

func TestReader_ReadBoolean_True_SyntaxError(t *testing.T) {
	r := NewReader([]byte("trux"))

	_, ok := r.ReadBoolean()

	if ok {
		t.Fatal("expected failure")
	}
	if r.err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestReader_ReadBoolean_False_SyntaxError(t *testing.T) {
	r := NewReader([]byte("falsx"))

	_, ok := r.ReadBoolean()

	if ok {
		t.Fatal("expected failure")
	}
	if r.err == nil {
		t.Fatal("expected syntax error")
	}
}

func TestReader_ReadBoolean_EOF(t *testing.T) {
	r := NewReader([]byte("tru"))

	_, ok := r.ReadBoolean()

	if ok {
		t.Fatal("expected failure")
	}
	if r.err == nil {
		t.Fatal("expected EOF error")
	}
}

func TestReader_ReadBoolean_NotBoolean(t *testing.T) {
	r := NewReader([]byte("null"))

	_, ok := r.ReadBoolean()

	if ok {
		t.Fatal("expected false")
	}
	if r.pos != 0 {
		t.Fatal("pos should not move")
	}
}

func TestReader_ReadBoolean_PreExistingError(t *testing.T) {
	r := NewReader([]byte("true"))
	r.err = errors.New("existing")

	_, ok := r.ReadBoolean()

	if ok {
		t.Fatal("expected false")
	}
	if r.pos != 0 {
		t.Fatal("reader should not advance")
	}
}
