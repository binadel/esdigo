package types

import "testing"

func TestBooleanArray(t *testing.T) {
	var a BooleanArray
	if !readCont(&a, "[true, false , true]") {
		t.Fatal("reader could not continue")
	}
	assertState(t, "BooleanArray", &a, true, true, true)
	if len(a.Value) != 3 || a.Value[0] != true || a.Value[1] != false {
		t.Fatalf("value = %v", a.Value)
	}

	// empty and null
	var e BooleanArray
	readCont(&e, "[]")
	assertState(t, "BooleanArray[]", &e, true, true, true)

	var n BooleanArray
	readCont(&n, "null")
	assertState(t, "BooleanArray(null)", &n, true, false, false)

	// a non-boolean element is dropped and marks the array invalid
	var b BooleanArray
	readCont(&b, "[true, 1, false]")
	assertState(t, "BooleanArray[true,1,false]", &b, true, true, false)
	if len(b.Value) != 2 || b.Value[0] != true || b.Value[1] != false {
		t.Errorf("value = %v, want [true false]", b.Value)
	}
}

func TestBooleanArray_RoundTrip(t *testing.T) {
	roundtrip[BooleanArray, *BooleanArray](t, "[true,false,true]")
	roundtrip[BooleanArray, *BooleanArray](t, "[]")
	roundtrip[BooleanArray, *BooleanArray](t, "null")
}
