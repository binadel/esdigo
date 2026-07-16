package validation

import (
	"testing"

	"github.com/binadel/esdigo/json/types"
)

type intArray = types.Array[types.Int64, *types.Int64]

func readIntArray(s string) intArray { return readInto[intArray](s) }

func newIntArray() *Array[types.Int64, *types.Int64] {
	return NewArray[types.Int64, *types.Int64]("a")
}

func TestArrayPresenceAndNull(t *testing.T) {
	var absent intArray
	if r := newIntArray().Required().Validate(absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	} else {
		mustContain(t, "required", resultJSON(r), "REQUIRED")
	}

	null := readIntArray("null")
	if r := newIntArray().Validate(null); !r.IsValid() || r.Defined {
		t.Errorf("nullable null: valid=%v defined=%v", r.IsValid(), r.Defined)
	}
	if r := newIntArray().NotNull().Validate(null); r.IsValid() {
		t.Errorf("notNull null should fail")
	} else {
		mustContain(t, "notNull", resultJSON(r), "NOT_NULL")
	}

	// wrong type -> ARRAY (not STRING)
	wrong := readIntArray("42")
	if r := newIntArray().Validate(wrong); r.IsValid() {
		t.Errorf("number into array should be invalid")
	} else {
		mustContain(t, "wrong-type", resultJSON(r), "ARRAY")
	}
}

func TestArrayValue(t *testing.T) {
	v := readIntArray("[1,2,3]")
	if r := newIntArray().Validate(v); !r.IsValid() || len(r.Value) != 3 {
		t.Errorf("[1,2,3]: valid=%v len=%d", r.IsValid(), len(r.Value))
	}
}

func TestArrayItemCounts(t *testing.T) {
	cases := []struct {
		name    string
		build   func() *Array[types.Int64, *types.Int64]
		input   string
		code    string
		wantErr bool
	}{
		{"exact-ok", func() *Array[types.Int64, *types.Int64] { return newIntArray().ExactItems(3) }, "[1,2,3]", "", false},
		{"exact-fail", func() *Array[types.Int64, *types.Int64] { return newIntArray().ExactItems(2) }, "[1,2,3]", `"exactItems":2`, true},
		{"exact0-ok", func() *Array[types.Int64, *types.Int64] { return newIntArray().ExactItems(0) }, "[]", "", false},
		{"exact0-fail", func() *Array[types.Int64, *types.Int64] { return newIntArray().ExactItems(0) }, "[1]", `"exactItems":0`, true},
		{"min-ok", func() *Array[types.Int64, *types.Int64] { return newIntArray().MinItems(2) }, "[1,2,3]", "", false},
		{"min-fail", func() *Array[types.Int64, *types.Int64] { return newIntArray().MinItems(4) }, "[1,2,3]", `"minItems":4`, true},
		{"max-ok", func() *Array[types.Int64, *types.Int64] { return newIntArray().MaxItems(3) }, "[1,2,3]", "", false},
		{"max-fail", func() *Array[types.Int64, *types.Int64] { return newIntArray().MaxItems(2) }, "[1,2,3]", `"maxItems":2`, true},
	}
	for _, c := range cases {
		v := readIntArray(c.input)
		r := c.build().Validate(v)
		if r.IsValid() == c.wantErr {
			t.Errorf("%s: IsValid=%v wantErr=%v (%s)", c.name, r.IsValid(), c.wantErr, resultJSON(r))
		}
		if c.code != "" {
			mustContain(t, c.name, resultJSON(r), c.code)
		}
	}
}
