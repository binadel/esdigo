package types

import (
	"math/big"
	"testing"
)

func TestNumberArray_Int(t *testing.T) {
	var a Int64Array
	if !readCont(&a, "[1, 2 , 3]") {
		t.Fatal("reader could not continue")
	}
	assertState(t, "Int64Array[1,2,3]", &a, true, true, true)
	if len(a.Value) != 3 || a.Value[0] != 1 || a.Value[2] != 3 {
		t.Fatalf("value = %v", a.Value)
	}

	// empty and null
	var e Int64Array
	readCont(&e, "[]")
	assertState(t, "Int64Array[]", &e, true, true, true)

	var n Int64Array
	readCont(&n, "null")
	assertState(t, "Int64Array(null)", &n, true, false, false)
}

// A wrong-typed or non-representable element is dropped and marks the whole array
// invalid, while the good elements are kept.
func TestNumberArray_DropsBadElements(t *testing.T) {
	var a Int64Array
	readCont(&a, `[1,"x",3]`)
	assertState(t, `Int64Array[1,"x",3]`, &a, true, true, false)
	if len(a.Value) != 2 || a.Value[0] != 1 || a.Value[1] != 3 {
		t.Errorf("value = %v, want [1 3]", a.Value)
	}

	var b Int64Array
	readCont(&b, "[1,1.5,3]") // 1.5 is not an integer
	assertState(t, "Int64Array[1,1.5,3]", &b, true, true, false)
	if len(b.Value) != 2 {
		t.Errorf("value = %v, want [1 3]", b.Value)
	}
}

func TestNumberArray_FloatAndBig(t *testing.T) {
	var f Float64Array
	readCont(&f, "[1.5, 2.5, -3]")
	assertState(t, "Float64Array", &f, true, true, true)
	if len(f.Value) != 3 || f.Value[0] != 1.5 {
		t.Fatalf("value = %v", f.Value)
	}

	var b BigIntArray
	readCont(&b, "[1, 1000000000000000000000, 1e3]")
	assertState(t, "BigIntArray", &b, true, true, true)
	if len(b.Value) != 3 {
		t.Fatalf("len = %d", len(b.Value))
	}
	want, _ := new(big.Int).SetString("1000000000000000000000", 10)
	if b.Value[1].Cmp(want) != 0 || b.Value[2].Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("value = %v", b.Value)
	}
}

func TestNumberArray_RoundTrip(t *testing.T) {
	roundtrip[Int64Array, *Int64Array](t, "[1,2,3]")
	roundtrip[Int64Array, *Int64Array](t, "[]")
	roundtrip[Int64Array, *Int64Array](t, "null")
	roundtrip[Float64Array, *Float64Array](t, "[1.5,2.5]")
	roundtrip[BigIntArray, *BigIntArray](t, "[1,2,1000000000000000000000]")
}
