package validation

import (
	"math/big"
	"testing"

	"github.com/binadel/esdigo/json/types"
)

func TestNumberPresenceAndNull(t *testing.T) {
	// absent (never read) + required -> REQUIRED
	var absent types.Int64
	if r := NewNumber[int64]("age").Required().Validate(&absent); r.IsValid() {
		t.Fatalf("required+absent should fail: %s", resultJSON(r))
	} else {
		mustContain(t, "required", resultJSON(r), "REQUIRED")
	}

	// nullable null -> valid, not defined, no value error
	null := readInto[types.Int64]("null")
	if r := NewNumber[int64]("age").Validate(&null); !r.IsValid() || r.Defined {
		t.Errorf("nullable null: valid=%v defined=%v", r.IsValid(), r.Defined)
	}

	// notNull + null -> NOT_NULL
	if r := NewNumber[int64]("age").NotNull().Validate(&null); r.IsValid() {
		t.Errorf("notNull null should fail")
	} else {
		mustContain(t, "notNull", resultJSON(r), "NOT_NULL")
	}

	// notNull without required: an ABSENT field is fine (only present-null fails).
	var absent2 types.Int64
	if r := NewNumber[int64]("age").NotNull().Validate(&absent2); !r.IsValid() {
		t.Errorf("notNull absent (not required) should be valid: %v", r.Errors)
	}
}

func TestNumberReasons(t *testing.T) {
	// wrong type (string into int) -> NUMBER
	wrong := readInto[types.Int64](`"x"`)
	if r := NewNumber[int64]("n").Validate(&wrong); r.IsValid() {
		t.Errorf("string into int should be invalid")
	} else {
		mustContain(t, "wrong-type", resultJSON(r), "NUMBER")
	}

	// real into int -> INTEGER
	real := readInto[types.Int64]("1.5")
	if r := NewNumber[int64]("n").Validate(&real); r.IsValid() {
		t.Errorf("1.5 into int should be invalid")
	} else {
		mustContain(t, "not-integer", resultJSON(r), "INTEGER")
	}

	// overflow -> OUT_OF_RANGE
	over := readInto[types.Int64]("99999999999999999999")
	if r := NewNumber[int64]("n").Validate(&over); r.IsValid() {
		t.Errorf("overflow into int should be invalid")
	} else {
		mustContain(t, "overflow", resultJSON(r), "OUT_OF_RANGE")
	}
}

func TestNumberBounds(t *testing.T) {
	valid := readInto[types.Int64]("5")
	if r := NewNumber[int64]("n").Min(1).Max(10).Validate(&valid); !r.IsValid() || r.Value != 5 {
		t.Errorf("5 in [1,10]: valid=%v value=%v", r.IsValid(), r.Value)
	}

	cases := []struct {
		name    string
		build   func() *Number[int64]
		input   string
		code    string
		wantErr bool
	}{
		{"min-ok", func() *Number[int64] { return NewNumber[int64]("n").Min(5) }, "5", "", false},
		{"min-fail", func() *Number[int64] { return NewNumber[int64]("n").Min(5) }, "4", `"minimum":5`, true},
		{"max-fail", func() *Number[int64] { return NewNumber[int64]("n").Max(5) }, "6", `"maximum":5`, true},
		{"exmin-fail", func() *Number[int64] { return NewNumber[int64]("n").ExclusiveMin(5) }, "5", `"exclusiveMinimum":5`, true},
		{"exmax-fail", func() *Number[int64] { return NewNumber[int64]("n").ExclusiveMax(5) }, "5", `"exclusiveMaximum":5`, true},
		{"multiple-ok", func() *Number[int64] { return NewNumber[int64]("n").MultipleOf(3) }, "9", "", false},
		{"multiple-fail", func() *Number[int64] { return NewNumber[int64]("n").MultipleOf(3) }, "10", `"multipleOf":3`, true},
	}
	for _, c := range cases {
		v := readInto[types.Int64](c.input)
		r := c.build().Validate(&v)
		if r.IsValid() == c.wantErr {
			t.Errorf("%s: IsValid=%v wantErr=%v (%s)", c.name, r.IsValid(), c.wantErr, resultJSON(r))
		}
		if c.code != "" {
			mustContain(t, c.name, resultJSON(r), c.code)
		}
	}
}

func TestNumberFloat(t *testing.T) {
	v := readInto[types.Float64]("2.5")
	if r := NewNumber[float64]("f").MultipleOf(0.5).Max(3).Validate(&v); !r.IsValid() || r.Value != 2.5 {
		t.Errorf("2.5 mult 0.5 max 3: %s", resultJSON(r))
	}
	v = readInto[types.Float64]("2.5")
	if r := NewNumber[float64]("f").MultipleOf(2).Validate(&v); r.IsValid() {
		t.Errorf("2.5 is not a multiple of 2")
	}

	// MultipleOf(0) is a no-op (avoids a divide-by-zero), never an error
	iv := readInto[types.Int64]("7")
	if r := NewNumber[int64]("n").MultipleOf(0).Validate(&iv); !r.IsValid() {
		t.Errorf("MultipleOf(0) should be a no-op: %s", resultJSON(r))
	}
}

func TestBigIntExtraBounds(t *testing.T) {
	var absent types.BigInt
	if r := NewBigInt("n").Required().Validate(&absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	}
	null := readInto[types.BigInt]("null")
	if r := NewBigInt("n").NotNull().Validate(&null); r.IsValid() {
		t.Errorf("notNull null should fail")
	}
	if r := NewBigInt("n").Validate(&null); !r.IsValid() {
		t.Errorf("nullable null should pass: %s", resultJSON(r))
	}

	v := readInto[types.BigInt]("5")
	if r := NewBigInt("n").Min(big.NewInt(10)).Validate(&v); r.IsValid() {
		t.Errorf("5 < min 10 should fail")
	} else {
		mustContain(t, "bigint-min", resultJSON(r), `"minimum":10`)
	}
	v = readInto[types.BigInt]("5")
	if r := NewBigInt("n").ExclusiveMin(big.NewInt(5)).Validate(&v); r.IsValid() {
		t.Errorf("5 is not > exclusiveMin 5")
	} else {
		mustContain(t, "bigint-exmin", resultJSON(r), `"exclusiveMinimum":5`)
	}
	v = readInto[types.BigInt]("5")
	if r := NewBigInt("n").ExclusiveMax(big.NewInt(5)).Validate(&v); r.IsValid() {
		t.Errorf("5 is not < exclusiveMax 5")
	} else {
		mustContain(t, "bigint-exmax", resultJSON(r), `"exclusiveMaximum":5`)
	}
}

func TestBigFloatExtraBounds(t *testing.T) {
	var absent types.BigFloat
	if r := NewBigFloat("f").Required().Validate(&absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	}
	null := readInto[types.BigFloat]("null")
	if r := NewBigFloat("f").NotNull().Validate(&null); r.IsValid() {
		t.Errorf("notNull null should fail")
	}

	v := readInto[types.BigFloat]("2.0")
	if r := NewBigFloat("f").Min(big.NewFloat(5)).Validate(&v); r.IsValid() {
		t.Errorf("2 < min 5 should fail")
	} else {
		mustContain(t, "bigfloat-min", resultJSON(r), `"minimum":5`)
	}
	v = readInto[types.BigFloat]("5.0")
	if r := NewBigFloat("f").ExclusiveMin(big.NewFloat(5)).Validate(&v); r.IsValid() {
		t.Errorf("5 is not > exclusiveMin 5")
	} else {
		mustContain(t, "bigfloat-exmin", resultJSON(r), `"exclusiveMinimum":5`)
	}
	v = readInto[types.BigFloat]("5.0")
	if r := NewBigFloat("f").ExclusiveMax(big.NewFloat(5)).Validate(&v); r.IsValid() {
		t.Errorf("5 is not < exclusiveMax 5")
	} else {
		mustContain(t, "bigfloat-exmax", resultJSON(r), `"exclusiveMaximum":5`)
	}
}

func TestBigIntValidator(t *testing.T) {
	v := readInto[types.BigInt]("42")
	if r := NewBigInt("n").Min(big.NewInt(0)).Max(big.NewInt(100)).Validate(&v); !r.IsValid() || r.Value.Int64() != 42 {
		t.Errorf("bigint 42 in [0,100]: %s", resultJSON(r))
	}

	v = readInto[types.BigInt]("150")
	if r := NewBigInt("n").Max(big.NewInt(100)).Validate(&v); r.IsValid() {
		t.Errorf("bigint 150 > 100 should fail")
	} else {
		mustContain(t, "bigint-max", resultJSON(r), `"maximum":100`)
	}

	v = readInto[types.BigInt]("10")
	if r := NewBigInt("n").MultipleOf(big.NewInt(3)).Validate(&v); r.IsValid() {
		t.Errorf("bigint 10 not multiple of 3")
	} else {
		mustContain(t, "bigint-mult", resultJSON(r), "MULTIPLE_OF")
	}

	// precise reasons still apply through numberBase
	v = readInto[types.BigInt]("1.5")
	if r := NewBigInt("n").Validate(&v); r.IsValid() {
		t.Errorf("bigint 1.5 should be INTEGER-invalid")
	} else {
		mustContain(t, "bigint-real", resultJSON(r), "INTEGER")
	}
}

func TestBigFloatValidator(t *testing.T) {
	v := readInto[types.BigFloat]("3.0")
	if r := NewBigFloat("f").MultipleOf(big.NewFloat(1.5)).Validate(&v); !r.IsValid() { // 3.0/1.5 = 2
		t.Errorf("bigfloat 3.0 mult 1.5 should pass: %s", resultJSON(r))
	}
	v = readInto[types.BigFloat]("3.5")
	if r := NewBigFloat("f").MultipleOf(big.NewFloat(1.5)).Validate(&v); r.IsValid() {
		t.Errorf("bigfloat 3.5 not multiple of 1.5")
	}
	v = readInto[types.BigFloat]("3.14")
	if r := NewBigFloat("f").Min(big.NewFloat(0)).Max(big.NewFloat(10)).Validate(&v); !r.IsValid() {
		t.Errorf("bigfloat 3.14 in [0,10]: %s", resultJSON(r))
	}
}
