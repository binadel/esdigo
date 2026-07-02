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
