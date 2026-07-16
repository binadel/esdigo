package types

import "testing"

func TestObject(t *testing.T) {
	var o Object[point, *point]
	if !readCont(&o, `{"x":1, "y":2}`) {
		t.Fatal("reader could not continue")
	}
	assertState(t, "Object", &o, true, true, true)
	if o.Value.X.Value != 1 || o.Value.Y.Value != 2 {
		t.Fatalf("value = %+v", o.Value)
	}

	// null
	var n Object[point, *point]
	readCont(&n, "null")
	assertState(t, "Object(null)", &n, true, false, false)

	// extra/unknown keys are skipped, known keys still read
	var e Object[point, *point]
	readCont(&e, `{"x":5,"z":99,"y":6}`)
	assertState(t, "Object(extra keys)", &e, true, true, true)
	if e.Value.X.Value != 5 || e.Value.Y.Value != 6 {
		t.Fatalf("value = %+v", e.Value)
	}
}

func TestObject_RoundTrip(t *testing.T) {
	roundtrip[Object[point, *point], *Object[point, *point]](t, `{"x":1,"y":2}`)
	roundtrip[Object[point, *point], *Object[point, *point]](t, "null")
}
