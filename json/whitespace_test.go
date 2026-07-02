package json

import (
	"math/rand"
	"strings"
	"testing"
)

// refSkipWS is the obvious byte-by-byte reference the SWAR version must match.
func refSkipWS(data []byte, pos int) int {
	for pos < len(data) {
		c := data[pos]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		pos++
	}
	return pos
}

func TestSkipWhitespace_Cases(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"x", 0},
		{"   x", 3},
		{"\t\n\r x", 4},
		{"        ", 8},                      // 8 spaces, all whitespace
		{strings.Repeat(" ", 40) + "y", 40},  // long run > 8
		{"\n" + strings.Repeat(" ", 20), 21}, // all whitespace, > 8
		{" !", 1},                            // TRAP: '!' (0x21) right after space
		{"\t\x08", 1},                        // TRAP: 0x08 right after tab
		{"\n\x0b", 1},                        // TRAP: 0x0b (VT) after LF (NOT json ws)
		{"\r\x0c", 1},                        // TRAP: 0x0c (FF) after CR (NOT json ws)
		{" \x00", 1},                         // control char after space
		{"   {", 3},
	}
	for _, c := range cases {
		r := &Reader{data: []byte(c.in)}
		r.SkipWhitespace()
		if r.pos != c.want {
			t.Fatalf("%q: pos=%d want %d", c.in, r.pos, c.want)
		}
	}
}

// Fuzz against the reference, heavy on whitespace and the SWAR borrow-trap bytes
// (0x21, 0x08, 0x0b, 0x0c — each is a json-ws byte XOR 1).
func TestSkipWhitespace_Fuzz(t *testing.T) {
	n := 300000
	if testing.Short() {
		n = 5000
	}
	rng := rand.New(rand.NewSource(1))
	traps := []byte{0x21, 0x08, 0x0b, 0x0c}
	for i := 0; i < n; i++ {
		m := rng.Intn(40)
		data := make([]byte, m)
		for j := range data {
			switch rng.Intn(8) {
			case 0:
				data[j] = ' '
			case 1:
				data[j] = '\t'
			case 2:
				data[j] = '\n'
			case 3:
				data[j] = '\r'
			case 4:
				data[j] = traps[rng.Intn(len(traps))]
			default:
				data[j] = byte(rng.Intn(256))
			}
		}
		start := rng.Intn(m + 1)
		r := &Reader{data: data, pos: start}
		r.SkipWhitespace()
		if want := refSkipWS(data, start); r.pos != want {
			t.Fatalf("data=%v start=%d: got %d want %d", data, start, r.pos, want)
		}
	}
}

func BenchmarkSkipWhitespace_Heavy(b *testing.B) {
	data := []byte("\n" + strings.Repeat(" ", 40) + "x")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.SkipWhitespace()
	}
}

func BenchmarkSkipWhitespace_None(b *testing.B) {
	data := []byte("xyz")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.SkipWhitespace()
	}
}
