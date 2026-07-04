package validation

import (
	"testing"

	"github.com/binadel/esdigo/json/types"
)

func TestBoolean(t *testing.T) {
	// value
	tru := readInto[types.Boolean]("true")
	if r := NewBoolean("b").Validate(tru); !r.IsValid() || r.Value != true {
		t.Errorf("true: valid=%v value=%v", r.IsValid(), r.Value)
	}
	fls := readInto[types.Boolean]("false")
	if r := NewBoolean("b").Validate(fls); !r.IsValid() || r.Value != false {
		t.Errorf("false: valid=%v value=%v", r.IsValid(), r.Value)
	}

	// absent + required
	var absent types.Boolean
	if r := NewBoolean("b").Required().Validate(absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	} else {
		mustContain(t, "required", resultJSON(r), "REQUIRED")
	}

	// nullable null -> valid; notNull null -> NOT_NULL
	null := readInto[types.Boolean]("null")
	if r := NewBoolean("b").Validate(null); !r.IsValid() || r.Defined {
		t.Errorf("nullable null: valid=%v defined=%v", r.IsValid(), r.Defined)
	}
	if r := NewBoolean("b").NotNull().Validate(null); r.IsValid() {
		t.Errorf("notNull null should fail")
	} else {
		mustContain(t, "notNull", resultJSON(r), "NOT_NULL")
	}

	// wrong type -> BOOLEAN
	wrong := readInto[types.Boolean](`"true"`)
	if r := NewBoolean("b").Validate(wrong); r.IsValid() {
		t.Errorf(`string "true" into bool should be invalid`)
	} else {
		mustContain(t, "wrong-type", resultJSON(r), "BOOLEAN")
	}
}
