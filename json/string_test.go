package json

import (
	stdjson "encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestReadString_ValidBasic(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"hello world",
		"123456",
		"!@#$%^&*()",
		"Hello 世界",
		"Привет",
		"مرحبا",
		"😀",
	}

	for _, input := range tests {
		data, _ := stdjson.Marshal(input)

		r := Reader{data: data}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed to read string: %q", input)
		}
		if out != input {
			t.Fatalf("expected %q, got %q", input, out)
		}
	}
}

func TestReadString_Escapes(t *testing.T) {
	tests := map[string]string{
		`"\\""`:    `\`,
		`"\""`:     `"`,
		`"\/"`:     `/`,
		`"\b"`:     "\b",
		`"\f"`:     "\f",
		`"\n"`:     "\n",
		`"\r"`:     "\r",
		`"\t"`:     "\t",
		`"\u0041"`: "A",
		`"\u0061"`: "a",
		`"\u00E9"`: "é",
		`"\u4E16"`: "世",
	}

	for raw, expected := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed reading %s", raw)
		}
		if out != expected {
			t.Fatalf("expected %q got %q", expected, out)
		}
	}
}

func TestReadString_SurrogatePairs(t *testing.T) {
	tests := map[string]string{
		`"\uD83D\uDE00"`: "😀",
		`"\uD834\uDD1E"`: "𝄞", // musical symbol
	}

	for raw, expected := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed reading %s", raw)
		}
		if out != expected {
			t.Fatalf("expected %q got %q", expected, out)
		}
	}
}

func TestReadString_InvalidSurrogates(t *testing.T) {
	tests := []string{
		`"\uD800"`,       // dangling high
		`"\uDC00"`,       // lone low
		`"\uD800\u0041"`, // invalid pair
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("unexpected failure for %s", raw)
		}
		if !strings.ContainsRune(out, utf8.RuneError) {
			t.Fatalf("expected RuneError in %q", out)
		}
	}
}

func TestReadString_InvalidEscapes(t *testing.T) {
	tests := []string{
		`"\x"`,
		`"\a"`,
		`"\u12"`,
		`"\uZZZZ"`,
		`"\u123G"`,
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected failure for %s", raw)
		}
	}
}

func TestReadString_ControlCharacters(t *testing.T) {
	for i := 0; i < 0x20; i++ {
		if i == '\n' || i == '\r' || i == '\t' {
			continue
		}
		raw := "\"" + string(byte(i)) + "\""
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected failure for control char 0x%02x", i)
		}
	}
}

func TestReadString_EOF(t *testing.T) {
	tests := []string{
		`"`,
		`"abc`,
		`"\`,
		`"\u123`,
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected EOF failure for %s", raw)
		}
	}
}

func TestSkipString(t *testing.T) {
	input := `"hello\nworld" 123`
	r := Reader{data: []byte(input)}

	if !r.SkipString() {
		t.Fatal("skip failed")
	}

	if r.pos != len(`"hello\nworld"`) {
		t.Fatalf("unexpected pos %d", r.pos)
	}
}

func TestWriter_RoundTrip(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"Hello 世界",
		"😀",
		"line\nbreak",
		"tab\tchar",
		`quote"slash\`,
	}

	for _, input := range tests {
		var w Writer
		w.WriteString(input)

		r := Reader{data: w.data}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed roundtrip for %q", input)
		}
		if out != input {
			t.Fatalf("roundtrip mismatch: %q != %q", out, input)
		}
	}
}

func TestCompatibilityWithStdlib(t *testing.T) {
	inputs := []string{
		"simple",
		"with\nnewline",
		"😀 emoji",
		"世界",
	}

	for _, input := range inputs {
		stdEncoded, _ := stdjson.Marshal(input)

		var w Writer
		w.WriteString(input)

		if string(w.data) != string(stdEncoded) {
			t.Fatalf("encoding mismatch:\nstd: %s\nour: %s",
				stdEncoded, w.data)
		}

		r := Reader{data: stdEncoded}
		out, ok := r.ReadString()
		if !ok {
			t.Fatal("read failed")
		}
		if out != input {
			t.Fatalf("decode mismatch %q != %q", out, input)
		}
	}
}
