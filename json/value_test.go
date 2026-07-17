package json

import (
	stdjson "encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// writeThenRead serializes v with WriteValue and parses the output back.
func writeThenRead(t *testing.T, v Value) (Value, []byte) {
	t.Helper()
	w := NewWriter(64)
	if !w.WriteValue(v) {
		t.Fatalf("WriteValue returned false for %#v", v)
	}
	out := w.Bytes()
	if !stdjson.Valid(out) {
		t.Fatalf("WriteValue produced invalid JSON %q", out)
	}
	r := NewReader(out)
	v2, err := r.ReadJSON()
	if err != nil {
		t.Fatalf("re-read of %q failed: %v", out, err)
	}
	return v2, out
}

// TestReadRawValue checks that ReadRawValue returns the exact bytes a value spans
// (leading whitespace trimmed), leaves the reader positioned just after it, and can
// be re-parsed. This backs discriminated-union decoding, which peeks a tag and then
// re-reads the captured bytes into the chosen variant.
func TestReadRawValue(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`{"a":1}`, `{"a":1}`},
		{`  {"a": [1, 2], "b": {"c": 3}}  `, `{"a": [1, 2], "b": {"c": 3}}`},
		{`"x" , 5`, `"x"`},
		{`123,`, `123`},
		{`[1,{"k":"}"}] rest`, `[1,{"k":"}"}]`},
	}
	for _, tc := range cases {
		r := NewReader([]byte(tc.in))
		raw, ok := r.ReadRawValue()
		if !ok {
			t.Errorf("ReadRawValue(%q) returned false", tc.in)
			continue
		}
		if string(raw) != tc.want {
			t.Errorf("ReadRawValue(%q) = %q, want %q", tc.in, raw, tc.want)
		}
		// The captured bytes must be valid, re-readable JSON.
		if _, err := NewReader(raw).ReadJSON(); err != nil {
			t.Errorf("captured %q is not valid JSON: %v", raw, err)
		}
	}

	// A malformed value fails.
	if _, ok := NewReader([]byte(`{`)).ReadRawValue(); ok {
		t.Errorf("ReadRawValue on a truncated object should fail")
	}
}

// FuzzReadRawValue checks the core invariant of ReadRawValue on arbitrary input:
// whatever bytes it captures form exactly one JSON value — a fresh reader reads a
// value from them and consumes all of it (ReadJSON rejects trailing bytes). It must
// never panic. This backs discriminated-union decoding, which re-parses the captured
// bytes into the chosen variant.
func FuzzReadRawValue(f *testing.F) {
	for _, s := range []string{
		"{}", `{"a":1}`, "[1,2,3]", `"x"`, "123", "-0.5e2", "true", "false", "null",
		"  {}  ", `{"k":"}"}`, `["\"]"]`, "[[[]]]", `{"a":{"b":[1,{"c":2}]}}`,
		"", " ", "{", "]", `"unterminated`, `{"a":1} trailing`, "1e999999", "\x00",
	} {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		raw, ok := NewReader(data).ReadRawValue()
		if !ok {
			return
		}
		if _, err := NewReader(raw).ReadJSON(); err != nil {
			t.Fatalf("ReadRawValue captured %q which does not re-parse as one value: %v", raw, err)
		}
	})
}

func TestWriteValue_RoundTrip_Cases(t *testing.T) {
	cases := []string{
		`null`, `true`, `false`,
		`0`, `-0`, `42`, `-42`, `3.14`, `-0.99`,
		`1e3`, `1E3`, `1e+30`, `1e-30`, `1.5e10`,
		`123456789012345678901234567890`, `0.0001`,
		`""`, `"hello"`, `"a\"b\\c\n\t\r\f\b"`, `"/"`,
		`"é"`, `"😀"`, `"<>&"`,
		`[]`, `[1,2,3]`, `["a",true,null,{}]`,
		`{}`, `{"a":1}`, `{"k":[1,{"n":-2.5}],"s":"x"}`,
		`[[[[]]]]`, `{"a":{"b":{"c":null}}}`,
	}
	for _, s := range cases {
		r := NewReader([]byte(s))
		v1, err := r.ReadJSON()
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		v2, out := writeThenRead(t, v1)
		if !reflect.DeepEqual(v1, v2) {
			t.Errorf("round-trip mismatch for %q -> %q:\n v1=%#v\n v2=%#v", s, out, v1, v2)
		}
	}
}

// TestWriteValue_RoundTrip_Fuzz reads random JSON into the DOM, writes it, and
// re-reads: the two DOM values must be deep-equal and the output valid JSON.
func TestWriteValue_RoundTrip_Fuzz(t *testing.T) {
	n := 50000
	if testing.Short() {
		n = 1000
	}
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < n; i++ {
		var sb strings.Builder
		emitJSON(&sb, randJSON(rng, 4), rng)
		s := sb.String()

		r := NewReader([]byte(s))
		v1, err := r.ReadJSON()
		if err != nil {
			t.Fatalf("parse generated %q: %v", s, err)
		}

		v2, out := writeThenRead(t, v1)
		if !reflect.DeepEqual(v1, v2) {
			t.Fatalf("round-trip mismatch:\n in=%q\n out=%q\n v1=%#v\n v2=%#v", s, out, v1, v2)
		}
	}
}
