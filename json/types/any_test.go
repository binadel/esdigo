package types

import (
	"testing"

	"github.com/binadel/esdigo/json"
)

func TestAny_Read(t *testing.T) {
	cases := []struct {
		in   string
		want json.ValueType
	}{
		{"42", json.ValueTypeNumber},
		{`"s"`, json.ValueTypeString},
		{"true", json.ValueTypeTrue},
		{"false", json.ValueTypeFalse},
		{"[1,2,3]", json.ValueTypeArray},
		{`{"a":1}`, json.ValueTypeObject},
	}
	for _, c := range cases {
		var a Any
		if !readCont(&a, c.in) {
			t.Fatalf("Any(%s): reader could not continue", c.in)
		}
		assertState(t, "Any("+c.in+")", &a, true, true, true)
		if a.Value.Type != c.want {
			t.Errorf("Any(%s): type = %v, want %v", c.in, a.Value.Type, c.want)
		}
	}

	// null -> present, not defined
	var n Any
	readCont(&n, "null")
	assertState(t, "Any(null)", &n, true, false, false)

	// malformed -> reader cannot continue
	for _, in := range []string{`{"a":`, "[1,", `"abc`, "tru"} {
		var a Any
		if readCont(&a, in) {
			t.Errorf("Any(%s): reader continued on malformed input", in)
		}
	}
}

func TestAny_RoundTrip(t *testing.T) {
	roundtrip[Any, *Any](t, "42")
	roundtrip[Any, *Any](t, `"str"`)
	roundtrip[Any, *Any](t, "true")
	roundtrip[Any, *Any](t, "null")
	roundtrip[Any, *Any](t, "[1,2.5,-3,null,true]")
	roundtrip[Any, *Any](t, `{"a":[1,{"b":"x"}],"c":null}`)
}
