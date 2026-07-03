package json

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestReadObject_Cases(t *testing.T) {
	valid := []string{
		`{}`,
		`{ }`,
		`{"a":1}`,
		`{"a":1,"b":2}`,
		`{"a":1, "b":2}`,        // whitespace after the value-separator comma
		`{ "a" : 1 , "b" : 2 }`, // whitespace around every token
		"{\n\t\"a\" : 1 ,\n\t\"b\" : 2\n}",
		`{"a":1,"b":2,"c":3}`,
		`{"nested":{"x":1}}`,
		`{"arr":[1,2,3]}`,
		`{"a":null,"b":true,"c":false,"d":"s"}`,
		`{"":0}`, // empty key
	}
	for _, s := range valid {
		r := NewReader([]byte(s))
		obj, ok := r.ReadObject()
		if !ok || r.err != nil {
			t.Errorf("ReadObject(%q): ok=%v err=%v, want success", s, ok, r.err)
			continue
		}
		if r.pos != len(s) {
			t.Errorf("ReadObject(%q): pos=%d, want %d (full consume)", s, r.pos, len(s))
		}
		if obj == nil {
			t.Errorf("ReadObject(%q): returned nil map", s)
		}
	}

	invalid := []string{
		`{`,
		`{"a"`,
		`{"a":}`,
		`{"a":1`,
		`{"a" 1}`,
		`{"a":1,}`,      // trailing comma
		`{,}`,           // leading comma
		`{"a":1 "b":2}`, // missing comma
		`{"a":1,,"b":2}`,
		`{:1}`,
		`{"a":1,"b"}`,
	}
	for _, s := range invalid {
		r := NewReader([]byte(s))
		if _, ok := r.ReadObject(); ok && r.err == nil {
			t.Errorf("ReadObject(%q): accepted, want rejection", s)
		}
	}
}

func TestReadArray_Cases(t *testing.T) {
	valid := []string{
		`[]`,
		`[ ]`,
		`[1]`,
		`[1,2,3]`,
		`[1, 2, 3]`,
		`[ 1 , 2 , 3 ]`,
		"[\n1,\n2\n]",
		`[[1],[2,3]]`,
		`[1,"a",true,false,null,{},[]]`,
		`[{"a":1}, {"b":2}]`,
	}
	for _, s := range valid {
		r := NewReader([]byte(s))
		arr, ok := r.ReadArray()
		if !ok || r.err != nil {
			t.Errorf("ReadArray(%q): ok=%v err=%v, want success", s, ok, r.err)
			continue
		}
		if r.pos != len(s) {
			t.Errorf("ReadArray(%q): pos=%d, want %d (full consume)", s, r.pos, len(s))
		}
		_ = arr
	}

	invalid := []string{
		`[`,
		`[1`,
		`[1,`,
		`[1 2]`,
		`[,]`,
		`[1,]`, // trailing comma
		`[1,,2]`,
		`[1,2`,
	}
	for _, s := range invalid {
		r := NewReader([]byte(s))
		if _, ok := r.ReadArray(); ok && r.err == nil {
			t.Errorf("ReadArray(%q): accepted, want rejection", s)
		}
	}
}

// randJSON builds a random JSON value tree of bounded depth using only tokens on
// which esdigo and encoding/json fully agree (ASCII keys, small ints, no escapes).
func randJSON(rng *rand.Rand, depth int) any {
	kind := rng.Intn(6)
	if depth <= 0 && kind >= 4 {
		kind = rng.Intn(4)
	}
	switch kind {
	case 0:
		return nil
	case 1:
		return rng.Intn(2) == 0
	case 2:
		return rng.Intn(1 << 20)
	case 3:
		return randWord(rng)
	case 4:
		arr := make([]any, rng.Intn(5))
		for i := range arr {
			arr[i] = randJSON(rng, depth-1)
		}
		return arr
	default:
		m := make(map[string]any)
		for i, k := 0, rng.Intn(5); i < k; i++ {
			m[randWord(rng)+strconv.Itoa(i)] = randJSON(rng, depth-1)
		}
		return m
	}
}

func randWord(rng *rand.Rand) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	var sb strings.Builder
	for i, k := 0, rng.Intn(6); i < k; i++ {
		sb.WriteByte(letters[rng.Intn(len(letters))])
	}
	return sb.String()
}

// emitJSON serializes v, sprinkling random JSON whitespace between structural
// tokens — including right after the object/array value-separator comma.
func emitJSON(sb *strings.Builder, v any, rng *rand.Rand) {
	ws := func() {
		for i, k := 0, rng.Intn(3); i < k; i++ {
			sb.WriteByte(" \t\n\r"[rng.Intn(4)])
		}
	}
	switch x := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		if x {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case int:
		sb.WriteString(strconv.Itoa(x))
	case string:
		sb.WriteByte('"')
		sb.WriteString(x)
		sb.WriteByte('"')
	case []any:
		sb.WriteByte('[')
		ws()
		for i, e := range x {
			if i > 0 {
				ws()
				sb.WriteByte(',')
				ws()
			}
			emitJSON(sb, e, rng)
		}
		ws()
		sb.WriteByte(']')
	case map[string]any:
		sb.WriteByte('{')
		ws()
		first := true
		for k, val := range x {
			if !first {
				ws()
				sb.WriteByte(',')
				ws()
			}
			first = false
			sb.WriteByte('"')
			sb.WriteString(k)
			sb.WriteByte('"')
			ws()
			sb.WriteByte(':')
			ws()
			emitJSON(sb, val, rng)
		}
		ws()
		sb.WriteByte('}')
	}
}

// TestStructural_WhitespaceFuzz generates valid JSON with random interior
// whitespace and asserts esdigo's full parse and skip both accept it. It targets
// whitespace handling in the object/array readers (e.g. after a comma).
func TestStructural_WhitespaceFuzz(t *testing.T) {
	n := 50000
	if testing.Short() {
		n = 1000
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		v := randJSON(rng, 4)
		var sb strings.Builder
		emitJSON(&sb, v, rng)
		s := sb.String()

		if !json.Valid([]byte(s)) {
			t.Fatalf("generator emitted invalid JSON: %q", s)
		}

		r := NewReader([]byte(s))
		if _, err := r.ReadJSON(); err != nil {
			t.Fatalf("ReadJSON rejected valid JSON %q: %v", s, err)
		}

		rs := NewReader([]byte(s))
		rs.SkipWhitespace()
		if !rs.SkipValue() {
			t.Fatalf("SkipValue rejected valid JSON %q: %v", s, rs.err)
		}
		rs.SkipWhitespace()
		if rs.pos != len(s) {
			t.Fatalf("SkipValue under-consumed %q: pos=%d want=%d", s, rs.pos, len(s))
		}
	}
}

// TestStructural_CorruptionDifferential deletes one byte from otherwise-valid
// documents and asserts esdigo's accept/reject verdict matches encoding/json's.
// This catches over-acceptance (bad separators, missing tokens) that the
// accept-only whitespace fuzz cannot.
func TestStructural_CorruptionDifferential(t *testing.T) {
	n := 50000
	if testing.Short() {
		n = 1000
	}
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < n; i++ {
		v := randJSON(rng, 4)
		var sb strings.Builder
		emitJSON(&sb, v, rng)
		b := []byte(sb.String())

		if len(b) > 1 {
			k := rng.Intn(len(b))
			switch rng.Intn(3) {
			case 0: // delete a byte
				b = append(b[:k:k], b[k+1:]...)
			case 1: // insert an arbitrary byte
				b = append(b[:k:k], append([]byte{byte(rng.Intn(256))}, b[k:]...)...)
			default: // duplicate a byte
				b = append(b[:k:k], append([]byte{b[k]}, b[k:]...)...)
			}
		}

		want := json.Valid(b)
		r := NewReader(b)
		_, err := r.ReadJSON()
		got := err == nil
		if got != want {
			t.Fatalf("accept/reject disagreement on %q: esdigo=%v stdlib=%v (err=%v)", b, got, want, err)
		}
	}
}
