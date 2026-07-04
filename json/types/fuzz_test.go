package types

import "testing"

// FuzzWrappers checks that reading arbitrary bytes into every wrapper — and
// writing back any that come out valid — never panics. It exercises the codec
// paths (big.Rat, strconv/big.ParseFloat, the integer conversion) on hostile
// input. The seed corpus runs on every `go test`; use -fuzz to explore further.
func FuzzWrappers(f *testing.F) {
	for _, s := range []string{
		"42", "-9223372036854775808", "18446744073709551615",
		"1.5", "1e3", "1.0", "3.14", "1e999999999", "1e-999999999",
		"0.00000001", "123456789012345678901234567890",
		`"hi"`, `"é😀"`, "true", "false", "null",
		"[1,2,3]", `["a",1,true]`, `{"x":1,"y":2}`, "[[[]]]",
		"", " ", "{", "[", `"`, "1.", "-", "1e", "tru", "\x00",
	} {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		in := string(data)

		var i Int64
		readCont(&i, in)
		var i8 Int8
		readCont(&i8, in)
		var u UInt64
		readCont(&u, in)
		var fl Float64
		readCont(&fl, in)
		var bi BigInt
		readCont(&bi, in)
		var bf BigFloat
		readCont(&bf, in)
		var rn RawNumber
		readCont(&rn, in)
		var s String
		readCont(&s, in)
		var b Boolean
		readCont(&b, in)
		var a Any
		readCont(&a, in)
		var na Int64Array
		readCont(&na, in)
		var sa StringArray
		readCont(&sa, in)
		var arr Array[Int64, *Int64]
		readCont(&arr, in)
		var obj Object[point, *point]
		readCont(&obj, in)

		// writing back a valid result must also not panic
		if i.Valid {
			writeStr(&i)
		}
		if bi.Valid {
			writeStr(&bi)
		}
		if fl.Valid {
			writeStr(&fl)
		}
		if a.Valid {
			writeStr(&a)
		}
		if na.Valid {
			writeStr(&na)
		}
		if obj.Valid {
			writeStr(&obj)
		}
	})
}
