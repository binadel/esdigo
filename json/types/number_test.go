package types

import (
	"math"
	"math/big"
	"testing"
)

func TestNumber_Int64_TriState(t *testing.T) {
	// value: present, defined, valid
	var v Int64
	if !readCont(&v, "42") {
		t.Fatal("read of 42 could not continue")
	}
	assertState(t, "Int64(42)", &v, true, true, true)
	if v.Value != 42 {
		t.Errorf("Value = %d, want 42", v.Value)
	}

	// null: present, not defined, not valid
	var n Int64
	if !readCont(&n, "null") {
		t.Fatal("read of null could not continue")
	}
	assertState(t, "Int64(null)", &n, true, false, false)

	// wrong type: present, defined, not valid — reader still continues
	for _, in := range []string{`"x"`, "true", "[1]", "{}"} {
		var w Int64
		if !readCont(&w, in) {
			t.Errorf("Int64(%s): reader could not continue", in)
		}
		assertState(t, "Int64("+in+")", &w, true, true, false)
	}
}

func TestNumber_Int64_Values(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
		want  int64
	}{
		{"0", true, 0},
		{"7", true, 7},
		{"-42", true, -42},
		{"9223372036854775807", true, math.MaxInt64},
		{"-9223372036854775808", true, math.MinInt64},
		// JSON Schema integers: zero fractional part
		{"1e3", true, 1000},
		{"1.0", true, 1},
		{"120e-1", true, 12},
		{"1E2", true, 100},
		// not an integer
		{"1.5", false, 0},
		{"0.1", false, 0},
		// out of range / overflow
		{"9223372036854775808", false, 0}, // MaxInt64 + 1
		{"99999999999999999999999", false, 0},
		{"1e999999999", false, 0},
	}
	for _, c := range cases {
		var v Int64
		readCont(&v, c.in)
		if v.Valid != c.valid {
			t.Errorf("Int64(%s): valid = %v, want %v", c.in, v.Valid, c.valid)
			continue
		}
		if c.valid && v.Value != c.want {
			t.Errorf("Int64(%s): value = %d, want %d", c.in, v.Value, c.want)
		}
	}
}

func TestNumber_IntegerRanges(t *testing.T) {
	// int8
	int8Cases := map[string]bool{"127": true, "-128": true, "128": false, "-129": false, "0": true}
	for in, valid := range int8Cases {
		var v Int8
		readCont(&v, in)
		if v.Valid != valid {
			t.Errorf("Int8(%s): valid = %v, want %v", in, v.Valid, valid)
		}
	}
	// uint8: rejects negatives and >255
	uint8Cases := map[string]bool{"0": true, "255": true, "256": false, "-1": false}
	for in, valid := range uint8Cases {
		var v UInt8
		readCont(&v, in)
		if v.Valid != valid {
			t.Errorf("UInt8(%s): valid = %v, want %v", in, v.Valid, valid)
		}
	}
	// uint64 max
	var u UInt64
	readCont(&u, "18446744073709551615")
	if !u.Valid || u.Value != math.MaxUint64 {
		t.Errorf("UInt64 max: valid=%v value=%d", u.Valid, u.Value)
	}
}

func TestNumber_Float(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
		want  float64
	}{
		{"3.14", true, 3.14},
		{"-0.5", true, -0.5},
		{"1e10", true, 1e10},
		{"42", true, 42}, // an integer is a valid float
		{"1e-3", true, 0.001},
		{`"x"`, false, 0},
		{"true", false, 0},
	}
	for _, c := range cases {
		var v Float64
		readCont(&v, c.in)
		if v.Valid != c.valid {
			t.Errorf("Float64(%s): valid = %v, want %v", c.in, v.Valid, c.valid)
			continue
		}
		if c.valid && v.Value != c.want {
			t.Errorf("Float64(%s): value = %v, want %v", c.in, v.Value, c.want)
		}
	}
	// float32 overflow -> invalid
	var f Float32
	readCont(&f, "1e40")
	if f.Valid {
		t.Errorf("Float32(1e40): accepted, want invalid (overflow)")
	}
}

func TestNumber_BigInt(t *testing.T) {
	ok := map[string]string{
		"1000":                           "1000",
		"1e3":                            "1000",
		"1.0":                            "1",
		"120e-1":                         "12",
		"1.5E2":                          "150",
		"-5":                             "-5",
		"0":                              "0",
		"123456789012345678901234567890": "123456789012345678901234567890",
	}
	for in, want := range ok {
		var v BigInt
		readCont(&v, in)
		if !v.Valid {
			t.Errorf("BigInt(%s): invalid, want %s", in, want)
			continue
		}
		wantN, _ := new(big.Int).SetString(want, 10)
		if v.Value.Cmp(wantN) != 0 {
			t.Errorf("BigInt(%s): value = %v, want %s", in, v.Value, want)
		}
	}
	for _, in := range []string{"1.5", "2e-1", "1e999999999", "1e70000", `"x"`} {
		var v BigInt
		readCont(&v, in)
		if v.Valid {
			t.Errorf("BigInt(%s): valid, want rejected", in)
		}
	}
}

func TestNumber_BigFloat_Raw(t *testing.T) {
	var bf BigFloat
	readCont(&bf, "3.14159265358979")
	if !bf.Valid {
		t.Errorf("BigFloat(3.14159...): invalid")
	}
	// raw number stores the exact token bytes
	var rn RawNumber
	readCont(&rn, "3.14e5")
	if !rn.Valid || string(rn.Value) != "3.14e5" {
		t.Errorf("RawNumber(3.14e5): valid=%v value=%q", rn.Valid, rn.Value)
	}
}

func TestNumber_RoundTrip(t *testing.T) {
	roundtrip[Int64, *Int64](t, "-42")
	roundtrip[Int64, *Int64](t, "0")
	roundtrip[Int64, *Int64](t, "9223372036854775807")
	roundtrip[UInt64, *UInt64](t, "18446744073709551615")
	roundtrip[Float64, *Float64](t, "3.14")
	roundtrip[Float64, *Float64](t, "-0.5")
	roundtrip[BigInt, *BigInt](t, "123456789012345678901234567890")
	roundtrip[BigInt, *BigInt](t, "1e30")
	roundtrip[BigFloat, *BigFloat](t, "2.5")
	roundtrip[RawNumber, *RawNumber](t, "3.14e5")
	// null round-trips (defined=false)
	roundtrip[Int64, *Int64](t, "null")
	roundtrip[BigInt, *BigInt](t, "null")
}

func TestNumber_Write(t *testing.T) {
	// NaN / Inf serialize as null (JSON has no such values)
	var f Float64
	f.Set(math.NaN())
	if s, _ := writeStr(&f); s != "null" {
		t.Errorf("NaN Float64 wrote %q, want null", s)
	}
	f.Set(math.Inf(1))
	if s, _ := writeStr(&f); s != "null" {
		t.Errorf("Inf Float64 wrote %q, want null", s)
	}

	// zero-value (not defined) writes null
	var z Int64
	if s, _ := writeStr(&z); s != "null" {
		t.Errorf("zero Int64 wrote %q, want null", s)
	}

	// a defined-but-invalid value cannot be serialized
	var bad Int64
	readCont(&bad, "1.5")
	if bad.Valid {
		t.Fatal("expected 1.5 into Int64 to be invalid")
	}
	if _, ok := writeStr(&bad); ok {
		t.Errorf("WriteJSON of invalid Int64 returned true")
	}
}
