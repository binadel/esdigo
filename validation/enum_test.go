package validation

import (
	stdjson "encoding/json"
	"math/big"
	"testing"

	"github.com/binadel/esdigo/json/types"
)

func TestStringEnumConst(t *testing.T) {
	// enum ok / fail
	v := readInto[types.String](`"b"`)
	if r := NewString("s").Enum("a", "b", "c").Validate(v); !r.IsValid() {
		t.Errorf("enum member should pass: %s", resultJSON(r))
	}
	v = readInto[types.String](`"z"`)
	r := NewString("s").Enum("a", "b", "c").Validate(v)
	if r.IsValid() {
		t.Errorf("non-member should fail")
	}
	mustContain(t, "enum-code", resultJSON(r), `"ENUM"`)
	mustContain(t, "enum-param", resultJSON(r), `"enum":["a","b","c"]`)

	// const ok / fail
	v = readInto[types.String](`"x"`)
	if r := NewString("s").Const("x").Validate(v); !r.IsValid() {
		t.Errorf("const match should pass: %s", resultJSON(r))
	}
	v = readInto[types.String](`"y"`)
	r = NewString("s").Const("x").Validate(v)
	if r.IsValid() {
		t.Errorf("const mismatch should fail")
	}
	mustContain(t, "const-code", resultJSON(r), `"CONST"`)
	mustContain(t, "const-param", resultJSON(r), `"const":"x"`)
}

func TestStringEnumEscaping(t *testing.T) {
	// a value containing a quote must stay valid JSON in the echoed param
	v := readInto[types.String](`"nope"`)
	r := NewString("s").Enum(`a"b`, `c\d`).Validate(v)
	out := resultJSON(r)
	mustContain(t, "escaped-enum", out, `"enum":["a\"b","c\\d"]`)
	if !stdjson.Valid([]byte(out)) {
		t.Errorf("escaped enum result is not valid JSON: %s", out)
	}
}

func TestNumberEnumConst(t *testing.T) {
	v := readInto[types.Int64]("2")
	if r := NewNumber[int64]("n").Enum(1, 2, 3).Validate(&v); !r.IsValid() {
		t.Errorf("enum member should pass: %s", resultJSON(r))
	}
	v = readInto[types.Int64]("9")
	r := NewNumber[int64]("n").Enum(1, 2, 3).Validate(&v)
	if r.IsValid() {
		t.Errorf("non-member should fail")
	}
	mustContain(t, "num-enum", resultJSON(r), `"enum":[1,2,3]`)

	v = readInto[types.Int64]("5")
	if r := NewNumber[int64]("n").Const(5).Validate(&v); !r.IsValid() {
		t.Errorf("const match should pass")
	}
	v = readInto[types.Int64]("6")
	r = NewNumber[int64]("n").Const(5).Validate(&v)
	if r.IsValid() {
		t.Errorf("const mismatch should fail")
	}
	mustContain(t, "num-const", resultJSON(r), `"const":5`)

	// float enum
	fv := readInto[types.Float64]("1.5")
	if r := NewNumber[float64]("f").Enum(0.5, 1.5, 2.5).Validate(&fv); !r.IsValid() {
		t.Errorf("float enum member should pass: %s", resultJSON(r))
	}
}

func TestBooleanConst(t *testing.T) {
	v := readInto[types.Boolean]("true")
	if r := NewBoolean("b").Const(true).Validate(v); !r.IsValid() {
		t.Errorf("const true match should pass")
	}
	v = readInto[types.Boolean]("false")
	r := NewBoolean("b").Const(true).Validate(v)
	if r.IsValid() {
		t.Errorf("const mismatch should fail")
	}
	mustContain(t, "bool-const", resultJSON(r), `"const":true`)
}

func TestBigEnumConst(t *testing.T) {
	v := readInto[types.BigInt]("2")
	if r := NewBigInt("n").Enum(big.NewInt(1), big.NewInt(2)).Validate(&v); !r.IsValid() {
		t.Errorf("bigint enum member should pass: %s", resultJSON(r))
	}
	v = readInto[types.BigInt]("9")
	r := NewBigInt("n").Enum(big.NewInt(1), big.NewInt(2)).Validate(&v)
	if r.IsValid() {
		t.Errorf("bigint non-member should fail")
	}
	mustContain(t, "bigint-enum", resultJSON(r), `"enum":[1,2]`)

	v = readInto[types.BigInt]("7")
	if r := NewBigInt("n").Const(big.NewInt(7)).Validate(&v); !r.IsValid() {
		t.Errorf("bigint const match should pass")
	}
	v = readInto[types.BigInt]("8")
	r = NewBigInt("n").Const(big.NewInt(7)).Validate(&v)
	if r.IsValid() {
		t.Errorf("bigint const mismatch should fail")
	}
	mustContain(t, "bigint-const", resultJSON(r), `"const":7`)

	// bigfloat enum
	fe := readInto[types.BigFloat]("2.5")
	if r := NewBigFloat("f").Enum(big.NewFloat(1.5), big.NewFloat(2.5)).Validate(&fe); !r.IsValid() {
		t.Errorf("bigfloat enum member should pass: %s", resultJSON(r))
	}
	fe = readInto[types.BigFloat]("9.5")
	if r := NewBigFloat("f").Enum(big.NewFloat(1.5), big.NewFloat(2.5)).Validate(&fe); r.IsValid() {
		t.Errorf("bigfloat non-member should fail")
	} else {
		mustContain(t, "bigfloat-enum", resultJSON(r), `"enum":[1.5,2.5]`)
	}

	// bigfloat const
	fv := readInto[types.BigFloat]("1.5")
	if r := NewBigFloat("f").Const(big.NewFloat(1.5)).Validate(&fv); !r.IsValid() {
		t.Errorf("bigfloat const match should pass: %s", resultJSON(r))
	}
	fv = readInto[types.BigFloat]("2.5")
	rf := NewBigFloat("f").Const(big.NewFloat(1.5)).Validate(&fv)
	if rf.IsValid() {
		t.Errorf("bigfloat const mismatch should fail")
	}
	mustContain(t, "bigfloat-const", resultJSON(rf), `"const":1.5`)
}
