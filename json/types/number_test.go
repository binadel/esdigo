package types

import (
	"math/big"
	"testing"

	"github.com/binadel/esdigo/json"
)

// readField runs ReadJSON over s and returns the (reader error, ok) pair so
// tests can assert both the tri-state and that no spurious parse error leaked.
func readField(t *testing.T, v interface {
	ReadJSON(*json.Reader) bool
}, s string) bool {
	t.Helper()
	r := json.NewReader([]byte(s))
	ok := v.ReadJSON(r)
	if err := r.Error(); err != nil {
		t.Fatalf("ReadJSON(%q) set unexpected reader error: %v", s, err)
	}
	return ok
}

func writeField(t *testing.T, v interface {
	WriteJSON(*json.Writer) bool
}) string {
	t.Helper()
	w := json.NewWriter(32)
	if !v.WriteJSON(w) {
		t.Fatalf("WriteJSON returned false")
	}
	return string(w.Bytes())
}

func TestInteger_ScalarSigned(t *testing.T) {
	cases := map[string]int64{"0": 0, "42": 42, "-42": -42, "9223372036854775807": 9223372036854775807}
	for in, want := range cases {
		var n Int64
		readField(t, &n, in)
		if !n.Present || !n.Defined || !n.Valid {
			t.Fatalf("%q: tri-state = (%v,%v,%v), want all true", in, n.Present, n.Defined, n.Valid)
		}
		if n.Value != want {
			t.Fatalf("%q: value = %d, want %d", in, n.Value, want)
		}
		if got := writeField(t, &n); got != in {
			t.Fatalf("%q: round-trip = %q", in, got)
		}
	}
}

func TestInteger_ScalarUnsignedFullRange(t *testing.T) {
	const max = "18446744073709551615" // 2^64-1, exceeds int64 max
	var n Uint64
	readField(t, &n, max)
	if !n.Valid || n.Value != 18446744073709551615 {
		t.Fatalf("uint64 max: valid=%v value=%d", n.Valid, n.Value)
	}
	if got := writeField(t, &n); got != max {
		t.Fatalf("uint64 max round-trip = %q", got)
	}
}

func TestInteger_OverflowIsInvalidNotError(t *testing.T) {
	// 2^64 does not fit int64; must be Defined but not Valid, with NO reader error.
	var n Int64
	ok := readField(t, &n, "18446744073709551616")
	if !ok {
		t.Fatalf("ReadJSON returned false (should consume the token)")
	}
	if !n.Defined || n.Valid {
		t.Fatalf("overflow: defined=%v valid=%v, want defined && !valid", n.Defined, n.Valid)
	}
}

func TestInteger_RealIntoIntegerIsInvalid(t *testing.T) {
	var n Int64
	readField(t, &n, "3.14")
	if !n.Defined || n.Valid {
		t.Fatalf("3.14 into Int64: defined=%v valid=%v", n.Defined, n.Valid)
	}
}

func TestNumber_Float(t *testing.T) {
	var n Float64
	readField(t, &n, "3.14")
	if !n.Valid || n.Value != 3.14 {
		t.Fatalf("float: valid=%v value=%v", n.Valid, n.Value)
	}
	if got := writeField(t, &n); got != "3.14" {
		t.Fatalf("float round-trip = %q", got)
	}
}

func TestRaw_Lossless(t *testing.T) {
	// Values no scalar type can hold, preserved exactly through read+write.
	for _, in := range []string{
		"123456789012345678901234567890",
		"-0.000123e+55",
		"1.5e10",
	} {
		var n RawNumber
		readField(t, &n, in)
		if !n.Valid {
			t.Fatalf("%q: not valid", in)
		}
		if got := writeField(t, &n); got != in {
			t.Fatalf("%q: raw round-trip = %q", in, got)
		}
	}
}

func TestBigInt(t *testing.T) {
	const huge = "123456789012345678901234567890"
	var n BigInt
	readField(t, &n, huge)
	if !n.Valid || n.Value == nil || n.Value.String() != huge {
		t.Fatalf("bigint: valid=%v value=%v", n.Valid, n.Value)
	}
	if got := writeField(t, &n); got != huge {
		t.Fatalf("bigint round-trip = %q", got)
	}

	// exponent form is not accepted by the big.Int backing → invalid, no error
	var e BigInt
	readField(t, &e, "1e3")
	if !e.Defined || e.Valid {
		t.Fatalf("1e3 into BigInt: defined=%v valid=%v", e.Defined, e.Valid)
	}
}

func TestBigFloat(t *testing.T) {
	var n BigFloat
	readField(t, &n, "2.5")
	if !n.Valid || n.Value == nil || n.Value.Cmp(big.NewFloat(2.5)) != 0 {
		t.Fatalf("bigfloat: valid=%v value=%v", n.Valid, n.Value)
	}
	if got := writeField(t, &n); got != "2.5" {
		t.Fatalf("bigfloat round-trip = %q", got)
	}
}

func TestNumber_Null(t *testing.T) {
	var n Int64
	readField(t, &n, "null")
	if !n.Present || n.Defined || n.Valid {
		t.Fatalf("null: tri-state = (%v,%v,%v), want present only", n.Present, n.Defined, n.Valid)
	}
	if got := writeField(t, &n); got != "null" {
		t.Fatalf("null round-trip = %q", got)
	}
}

func TestNumberArray_LosslessBig(t *testing.T) {
	// regression: big numbers in an array used to serialise as 0.
	const in = "[1,123456789012345678901234567890,3.14]"
	var a NumberArray
	readField(t, &a, in)
	if !a.Valid || len(a.Value) != 3 {
		t.Fatalf("array: valid=%v len=%d", a.Valid, len(a.Value))
	}
	if got := writeField(t, &a); got != in {
		t.Fatalf("array round-trip = %q", got)
	}
}
