package types

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/binadel/esdigo/json"
)

// TestSetters exercises Set/SetNull across every wrapper: Set must produce the
// expected JSON, and SetNull must write null.
func TestSetters(t *testing.T) {
	wantWrite := func(name string, v json.ValueWriter, want string) {
		t.Helper()
		if out, ok := writeStr(v); !ok || out != want {
			t.Errorf("%s: wrote (%q, ok=%v), want %q", name, out, ok, want)
		}
	}

	var num Int64
	num.Set(7)
	wantWrite("Int64.Set", &num, "7")
	num.SetNull()
	wantWrite("Int64.SetNull", &num, "null")

	var na Int64Array
	na.Set([]int64{1, 2, 3})
	wantWrite("Int64Array.Set", &na, "[1,2,3]")
	na.SetNull()
	wantWrite("Int64Array.SetNull", &na, "null")

	var ba BooleanArray
	ba.Set([]bool{true, false})
	wantWrite("BooleanArray.Set", &ba, "[true,false]")
	ba.SetNull()
	wantWrite("BooleanArray.SetNull", &ba, "null")

	var sa StringArray
	sa.Set([]string{"a", "b"})
	wantWrite("StringArray.Set", &sa, `["a","b"]`)
	sa.SetNull()
	wantWrite("StringArray.SetNull", &sa, "null")

	// generic Array of elements
	var elem Int64
	elem.Set(5)
	var arr Array[Int64, *Int64]
	arr.Set([]*Int64{&elem})
	wantWrite("Array.Set", &arr, "[5]")
	arr.SetNull()
	wantWrite("Array.SetNull", &arr, "null")

	// generic Object of a nested value
	var p point
	p.X.Set(1)
	p.Y.Set(2)
	var obj Object[point, *point]
	obj.Set(&p)
	wantWrite("Object.Set", &obj, `{"x":1,"y":2}`)
	obj.SetNull()
	wantWrite("Object.SetNull", &obj, "null")

	// Any
	var any Any
	any.Set(json.Value{Type: json.ValueTypeNumber, Payload: []byte("42")})
	wantWrite("Any.Set", &any, "42")
	any.SetNull()
	wantWrite("Any.SetNull", &any, "null")
}

// TestString_SetNilBranches covers the nil/zero → null path of the remaining
// typed string setters.
func TestString_SetNilBranches(t *testing.T) {
	var s String
	s.SetIP(nil)
	assertState(t, "SetIP(nil)", &s, true, false, false)

	s.Set([]byte("x")) // reset to a value first
	s.SetUri((*url.URL)(nil))
	assertState(t, "SetUri(nil)", &s, true, false, false)

	s.Set([]byte("x"))
	s.SetDuration(time.Duration(0))
	assertState(t, "SetDuration(0)", &s, true, false, false)

	// a real IP still writes
	s.SetIP(net.ParseIP("::1"))
	if out, ok := writeStr(&s); !ok || out != `"::1"` {
		t.Errorf("SetIP(::1) wrote (%q, %v)", out, ok)
	}
}

// TestWriteInvalidContainer covers WriteJSON returning false for a defined but
// invalid container (an array that dropped an element).
func TestWriteInvalidContainer(t *testing.T) {
	var a Int64Array
	readCont(&a, `[1,"x"]`) // drops "x" -> Valid=false, but Defined
	if a.Valid {
		t.Fatal("expected the array to be invalid")
	}
	if _, ok := writeStr(&a); ok {
		t.Errorf("WriteJSON of a defined-but-invalid array returned true")
	}
}
