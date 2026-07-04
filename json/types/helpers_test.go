package types

import (
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/binadel/esdigo/json"
)

// roundtrip reads input into a wrapper, writes it back, re-reads, and asserts the
// two wrapper values are deep-equal and the serialized form is valid JSON. It is
// for VALID inputs (and null); a defined-but-invalid value cannot be written and
// is checked separately. It returns the first parse for further assertions.
func roundtrip[T any, PT json.ValueReadWriter[T]](t *testing.T, input string) T {
	t.Helper()
	var v1, v2 T
	r1 := json.NewReader([]byte(input))
	if !PT(&v1).ReadJSON(r1) || r1.Error() != nil {
		t.Fatalf("%s: read failed: %v", input, r1.Error())
	}
	w := json.NewWriter(64)
	if !PT(&v1).WriteJSON(w) {
		t.Fatalf("%s: WriteJSON returned false", input)
	}
	out := w.Bytes()
	if !stdjson.Valid(out) {
		t.Fatalf("%s: wrote invalid JSON %q", input, out)
	}
	r2 := json.NewReader(out)
	if !PT(&v2).ReadJSON(r2) || r2.Error() != nil {
		t.Fatalf("%s: reread of %q failed: %v", input, out, r2.Error())
	}
	if !reflect.DeepEqual(v1, v2) {
		t.Errorf("%s -> %q: round-trip mismatch\n v1=%#v\n v2=%#v", input, out, v1, v2)
	}
	return v1
}

// readCont reads input into pv and returns whether the reader can continue.
func readCont(pv json.ValueReader, input string) bool {
	return pv.ReadJSON(json.NewReader([]byte(input)))
}

// writeStr writes v and returns the produced JSON plus the WriteJSON result.
func writeStr(v json.ValueWriter) (string, bool) {
	w := json.NewWriter(32)
	ok := v.WriteJSON(w)
	return string(w.Bytes()), ok
}

// assertState checks the (Present, Defined, Valid) tri-state of a wrapper.
func assertState(t *testing.T, name string, v json.OptionalValue, present, defined, valid bool) {
	t.Helper()
	if v.IsPresent() != present || v.IsDefined() != defined || v.IsValid() != valid {
		t.Errorf("%s: state = (present=%v defined=%v valid=%v), want (%v %v %v)",
			name, v.IsPresent(), v.IsDefined(), v.IsValid(), present, defined, valid)
	}
}

// point is a minimal generated-style struct used to exercise the generic
// Array[V,PV] and Object[V,PV] containers.
type point struct {
	X Int64
	Y Int64
}

func (p *point) ReadJSON(r *json.Reader) bool {
	r.SkipWhitespace()
	if r.ReadNull() {
		return true
	}
	if !r.BeginObject() {
		return r.SkipValue()
	}
	r.SkipWhitespace()
	if r.EndObject() {
		return true
	}
	for {
		name, ok := r.ReadString()
		if !ok {
			r.SetSyntaxError("expected a name")
			return false
		}
		r.SkipWhitespace()
		if !r.NameSeparator() {
			r.SetSyntaxError("expected ':'")
			return false
		}
		switch name {
		case "x":
			if !p.X.ReadJSON(r) {
				return false
			}
		case "y":
			if !p.Y.ReadJSON(r) {
				return false
			}
		default:
			if !r.SkipValue() {
				return false
			}
		}
		r.SkipWhitespace()
		if r.EndObject() {
			return true
		}
		if !r.ValueSeparator() {
			r.SetSyntaxError("expected ',' or '}'")
			return false
		}
		r.SkipWhitespace()
	}
}

func (p *point) WriteJSON(w *json.Writer) bool {
	w.BeginObject()
	w.WriteRawString(`"x":`)
	if !p.X.WriteJSON(w) {
		return false
	}
	w.WriteRawString(`,"y":`)
	if !p.Y.WriteJSON(w) {
		return false
	}
	w.EndObject()
	return true
}
