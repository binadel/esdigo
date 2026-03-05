package json

import "testing"

func TestReadNumber(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      NumberValue
		wantOk    bool
		wantError bool
	}{
		// --- Basic Integers ---
		{"Zero", "0", NumberValue{Negative: false, Coefficient: 0, Exponent: 0, Type: NumberTypeInteger}, true, false},
		{"Single Digit", "7", NumberValue{Negative: false, Coefficient: 7, Exponent: 0, Type: NumberTypeInteger}, true, false},
		{"Multiple Digits", "42", NumberValue{Negative: false, Coefficient: 42, Exponent: 0, Type: NumberTypeInteger}, true, false},
		{"Negative Integer", "-42", NumberValue{Negative: true, Coefficient: 42, Exponent: 0, Type: NumberTypeInteger}, true, false},
		{"Large Number uint64", "-18446744073709551615", NumberValue{Negative: true, Coefficient: 18446744073709551615, Exponent: 0, Type: NumberTypeInteger}, true, false},

		// --- Basic Reals ---
		{"Fractional", "3.14", NumberValue{Negative: false, Coefficient: 314, Exponent: -2, Type: NumberTypeReal}, true, false},
		{"Negative Fractional", "-0.99", NumberValue{Negative: true, Coefficient: 99, Exponent: -2, Type: NumberTypeReal}, true, false},

		// --- Exponent Notation ---
		{"Positive Exponent", "1e3", NumberValue{Negative: false, Coefficient: 1, Exponent: 3, Type: NumberTypeInteger}, true, false},
		{"Capital E Exponent", "1E3", NumberValue{Negative: false, Coefficient: 1, Exponent: 3, Type: NumberTypeInteger}, true, false},
		{"Explicit Positive Exponent", "1e+3", NumberValue{Negative: false, Coefficient: 1, Exponent: 3, Type: NumberTypeInteger}, true, false},
		{"Negative Exponent", "5e-1", NumberValue{Negative: false, Coefficient: 5, Exponent: -1, Type: NumberTypeReal}, true, false},

		// --- Trailing Zeros & Number Type Classification (The Optimization Fix) ---
		{"Exact Integer Division", "120e-1", NumberValue{Negative: false, Coefficient: 120, Exponent: -1, Type: NumberTypeInteger}, true, false},
		{"Multiple Trailing Zeros Integer", "1200e-2", NumberValue{Negative: false, Coefficient: 1200, Exponent: -2, Type: NumberTypeInteger}, true, false},
		{"Insufficient Trailing Zeros", "1200e-3", NumberValue{Negative: false, Coefficient: 1200, Exponent: -3, Type: NumberTypeReal}, true, false},
		{"Zero with Fractional", "0.0", NumberValue{Negative: false, Coefficient: 0, Exponent: -1, Type: NumberTypeInteger}, true, false},
		{"Zero with Many Fractions", "0.0000", NumberValue{Negative: false, Coefficient: 0, Exponent: -4, Type: NumberTypeInteger}, true, false},
		{"Zero with Exponent", "0e-10", NumberValue{Negative: false, Coefficient: 0, Exponent: -10, Type: NumberTypeInteger}, true, false},
		{"Fractional zero but real value", "1.000", NumberValue{Negative: false, Coefficient: 1000, Exponent: -3, Type: NumberTypeInteger}, true, false},
		{"Fractional trailing zero real", "1.0010", NumberValue{Negative: false, Coefficient: 10010, Exponent: -4, Type: NumberTypeReal}, true, false}, // 10010 has 1 trailing zero, exp -4. 1 < 4, so Real.

		// --- Big Numbers and Overflow Boundaries ---
		// MaxInt16 is 32767. MinInt16 is -32768.
		{"Max Int16 Exponent", "1e32767", NumberValue{Negative: false, Coefficient: 1, Exponent: 32767, Type: NumberTypeInteger}, true, false},
		{"Min Int16 Exponent Real", "1e-32768", NumberValue{Negative: false, Coefficient: 1, Exponent: -32768, Type: NumberTypeReal}, true, false},
		{"Exceed Max Int16 Exponent", "1e32768", NumberValue{Type: NumberTypeBig}, true, false},
		{"Exceed Min Int16 Exponent", "1e-32769", NumberValue{Type: NumberTypeBig}, true, false},
		{"Large Number Beyond uint64", "18446744073709551616", NumberValue{Type: NumberTypeBig}, true, false},

		// --- Internal Offset Resolution (Preventing Silent Corruption Fix) ---
		// Fraction creates a massive negative exponent, "e" offsets it back to 0.
		{"Huge Internal Offset", "1.000000000000000000000e19", NumberValue{Negative: false, Coefficient: 10000000000000000000, Exponent: 0, Type: NumberTypeInteger}, true, false},
		{"Huge Internal Offset", "1.000000000000000000000e25", NumberValue{Negative: false, Coefficient: 10000000000000000000, Exponent: 6, Type: NumberTypeInteger}, true, false},

		// --- Syntax Errors & Invalid Formats ---
		{"Leading Zero Invalid", "01", NumberValue{}, false, true},
		{"Leading Zero Invalid Negative", "-01", NumberValue{}, false, true},
		{"Minus Sign Only", "-", NumberValue{}, false, true},
		{"Minus Sign No Digits", "-a", NumberValue{}, false, true},
		{"Dot Only", ".", NumberValue{}, false, false},
		{"Trailing Dot", "1.", NumberValue{}, false, true},
		{"Dot No Digits", "1.a", NumberValue{}, false, true},
		{"E Only", "e", NumberValue{}, false, false},
		{"Trailing E", "1e", NumberValue{}, false, true},
		{"E No Digits", "1ea", NumberValue{}, false, true},
		{"Double Minus Exponent", "1e--1", NumberValue{}, false, true},
		{"Empty String", "", NumberValue{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{data: []byte(tt.input), pos: 0}
			val, ok := r.ReadNumber()

			if tt.wantError {
				if r.err == nil {
					t.Errorf("ReadNumber() expected error but got none")
				}
				return // Expected error, no need to check further
			}

			if r.err != nil {
				t.Fatalf("ReadNumber() unexpected error: %v", r.err)
			}

			if ok != tt.wantOk {
				t.Errorf("ReadNumber() ok = %v, wantOk %v", ok, tt.wantOk)
			}

			if val != tt.want {
				t.Errorf("ReadNumber() returned %+v, want %+v", val, tt.want)
			}
		})
	}
}

func TestSkipNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantPos int
	}{
		{"Single Digit Comma", "7,", 1},
		{"Multi Digit Space", "42 ", 2},
		{"Negative Bracket", "-42]", 3},
		{"Fractional Brace", "3.14}", 4},
		{"Exponent Array", "1.5e-10]", 7},
		{"Zero End Of String", "0", 1},
		{"Stop at Invalid char", "123abc", 3},
		{"Stop at String boundary", "123\"x", 3},
		{"Complex Number Skip", "-0.000123e+55,", 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reader{data: []byte(tt.input), pos: 0}
			ok := r.SkipNumber()

			if !ok {
				t.Errorf("SkipNumber() expected success but got failed")
			}

			if r.pos != tt.wantPos {
				t.Errorf("SkipNumber() ended at pos %d, want pos %d", r.pos, tt.wantPos)
			}
		})
	}
}
