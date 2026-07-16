package types

import "testing"

func TestBoolean(t *testing.T) {
	// true / false
	for _, tc := range []struct {
		in   string
		want bool
	}{{"true", true}, {"false", false}} {
		var b Boolean
		if !readCont(&b, tc.in) {
			t.Fatalf("Boolean(%s): reader could not continue", tc.in)
		}
		assertState(t, "Boolean("+tc.in+")", &b, true, true, true)
		if b.Value != tc.want {
			t.Errorf("Boolean(%s): value = %v, want %v", tc.in, b.Value, tc.want)
		}
	}

	// null
	var n Boolean
	readCont(&n, "null")
	assertState(t, "Boolean(null)", &n, true, false, false)

	// wrong type: present, defined, invalid — reader continues
	for _, in := range []string{"1", `"true"`, "[]", "{}"} {
		var b Boolean
		if !readCont(&b, in) {
			t.Errorf("Boolean(%s): reader could not continue", in)
		}
		assertState(t, "Boolean("+in+")", &b, true, true, false)
	}
}

func TestBoolean_RoundTrip(t *testing.T) {
	roundtrip[Boolean, *Boolean](t, "true")
	roundtrip[Boolean, *Boolean](t, "false")
	roundtrip[Boolean, *Boolean](t, "null")
}

func TestBoolean_Set(t *testing.T) {
	var b Boolean
	b.Set(true)
	assertState(t, "Set(true)", &b, true, true, true)
	if s, _ := writeStr(&b); s != "true" {
		t.Errorf("Set(true) wrote %q", s)
	}
	b.SetNull()
	assertState(t, "SetNull", &b, true, false, false)
	if s, _ := writeStr(&b); s != "null" {
		t.Errorf("SetNull wrote %q", s)
	}
}
