package types

import "testing"

func TestArray_Scalars(t *testing.T) {
	var a Array[Int64, *Int64]
	if !readCont(&a, "[1, 2 , 3]") {
		t.Fatal("reader could not continue")
	}
	assertState(t, "Array[1,2,3]", &a, true, true, true)
	if len(a.Value) != 3 || a.Value[0].Value != 1 || a.Value[2].Value != 3 {
		t.Fatalf("value = %v", a.Value)
	}

	// empty array is valid
	var e Array[Int64, *Int64]
	readCont(&e, "[]")
	assertState(t, "Array[]", &e, true, true, true)
	if len(e.Value) != 0 {
		t.Errorf("empty array len = %d", len(e.Value))
	}

	// null
	var n Array[Int64, *Int64]
	readCont(&n, "null")
	assertState(t, "Array(null)", &n, true, false, false)

	// not an array (wrong type at the array level): defined but invalid
	var w Array[Int64, *Int64]
	if !readCont(&w, "42") {
		t.Error("reader could not continue on non-array")
	}
	assertState(t, "Array(42)", &w, true, true, false)
}

// A generic Array keeps a scalar element that decoded but is itself invalid; the
// array stays Valid (unlike the specialized NumberArray, which drops it).
func TestArray_KeepsInvalidElement(t *testing.T) {
	var a Array[Int64, *Int64]
	readCont(&a, `[1,"x",3]`)
	assertState(t, `Array[1,"x",3]`, &a, true, true, true)
	if len(a.Value) != 3 {
		t.Fatalf("len = %d, want 3", len(a.Value))
	}
	if a.Value[1].Valid {
		t.Errorf("element 1 should be invalid")
	}
}

func TestArray_Structs(t *testing.T) {
	var a Array[point, *point]
	if !readCont(&a, `[{"x":1,"y":2}, {"x":-3,"y":4}]`) {
		t.Fatal("reader could not continue")
	}
	assertState(t, "Array[points]", &a, true, true, true)
	if len(a.Value) != 2 || a.Value[0].X.Value != 1 || a.Value[1].Y.Value != 4 {
		t.Fatalf("value = %+v", a.Value)
	}
}

func TestArray_RoundTrip(t *testing.T) {
	roundtrip[Array[Int64, *Int64], *Array[Int64, *Int64]](t, "[1,2,3]")
	roundtrip[Array[Int64, *Int64], *Array[Int64, *Int64]](t, "[]")
	roundtrip[Array[Int64, *Int64], *Array[Int64, *Int64]](t, "null")
	roundtrip[Array[point, *point], *Array[point, *point]](t, `[{"x":1,"y":2},{"x":-3,"y":4}]`)
}
