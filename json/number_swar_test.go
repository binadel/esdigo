package json

import (
	"math/rand"
	"testing"
)

func TestIsEightDigits(t *testing.T) {
	le := func(s string) uint64 {
		var w [8]byte
		copy(w[:], s)
		var v uint64
		for i := 7; i >= 0; i-- {
			v = v<<8 | uint64(w[i])
		}
		return v
	}
	if !isEightDigits(le("01234567")) || !isEightDigits(le("99999999")) || !isEightDigits(le("00000000")) {
		t.Fatal("valid 8-digit words rejected")
	}
	for _, s := range []string{"0123456/", "0123456:", "0123456 ", "a1234567", "01234-67", "0123456."} {
		if isEightDigits(le(s)) {
			t.Fatalf("non-digit word accepted: %q", s)
		}
	}
}

func randNumberish(rng *rand.Rand) []byte {
	var b []byte
	if rng.Intn(2) == 0 {
		b = append(b, '-')
	}
	switch rng.Intn(3) {
	case 0:
		b = append(b, '0')
	default:
		k := 1 + rng.Intn(25)
		b = append(b, byte('1'+rng.Intn(9)))
		for j := 1; j < k; j++ {
			b = append(b, byte('0'+rng.Intn(10)))
		}
	}
	if rng.Intn(2) == 0 {
		b = append(b, '.')
		for j, k := 0, 1+rng.Intn(20); j < k; j++ {
			b = append(b, byte('0'+rng.Intn(10)))
		}
	}
	if rng.Intn(2) == 0 {
		b = append(b, "eE"[rng.Intn(2)])
		if rng.Intn(2) == 0 {
			b = append(b, "+-"[rng.Intn(2)])
		}
		for j, k := 0, 1+rng.Intn(6); j < k; j++ {
			b = append(b, byte('0'+rng.Intn(10)))
		}
	}
	if rng.Intn(3) == 0 { // trailing non-number byte
		b = append(b, ",}] xyz."[rng.Intn(8)])
	}
	if rng.Intn(4) == 0 && len(b) > 1 { // truncate → often invalid
		b = b[:1+rng.Intn(len(b)-1)]
	}
	return b
}

// SkipNumber (SWAR) must agree with the byte-by-byte ReadNumber on validity and
// on where the number ends.
func TestSkipReadNumberConsistency(t *testing.T) {
	n := 300000
	if testing.Short() {
		n = 5000
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		data := randNumberish(rng)

		rs := &Reader{data: data}
		sok := rs.SkipNumber()

		rr := &Reader{data: data}
		_, rok := rr.ReadNumber()

		if sok != rok {
			t.Fatalf("%q: SkipNumber ok=%v ReadNumber ok=%v", data, sok, rok)
		}
		if sok && rs.pos != rr.pos {
			t.Fatalf("%q: SkipNumber pos=%d ReadNumber pos=%d", data, rs.pos, rr.pos)
		}
	}
}
