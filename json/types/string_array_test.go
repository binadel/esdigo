package types

import "testing"

func TestStringArray(t *testing.T) {
	var a StringArray
	if !readCont(&a, `["a", "b" , "c"]`) {
		t.Fatal("reader could not continue")
	}
	assertState(t, "StringArray", &a, true, true, true)
	if len(a.Value) != 3 || a.Value[1] != "b" {
		t.Fatalf("value = %v", a.Value)
	}

	// empty and null
	var e StringArray
	readCont(&e, "[]")
	assertState(t, "StringArray[]", &e, true, true, true)

	var n StringArray
	readCont(&n, "null")
	assertState(t, "StringArray(null)", &n, true, false, false)

	// a non-string element is dropped and marks the array invalid
	var b StringArray
	readCont(&b, `["a", 1, "c"]`)
	assertState(t, `StringArray["a",1,"c"]`, &b, true, true, false)
	if len(b.Value) != 2 || b.Value[0] != "a" || b.Value[1] != "c" {
		t.Errorf("value = %v, want [a c]", b.Value)
	}
}

func TestStringArray_RoundTrip(t *testing.T) {
	roundtrip[StringArray, *StringArray](t, `["a","b","c"]`)
	roundtrip[StringArray, *StringArray](t, `["a\"b","é😀"]`)
	roundtrip[StringArray, *StringArray](t, "[]")
	roundtrip[StringArray, *StringArray](t, "null")
}
