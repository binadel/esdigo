package types

import (
	"testing"

	"github.com/binadel/esdigo/json"
)

// These benchmarks exist to prove the codec design is allocation-free on the
// scalar/raw hot path. Run: go test ./json/types/ -bench . -benchmem -run ^$
//
// The codec (numberCodec) is a generic *constraint*, not a runtime interface
// value: `var codec C` is a zero-size concrete struct and codec.decode is a
// static dictionary call, so the read/write paths box nothing.

func BenchmarkRead_Int64(b *testing.B) {
	tok := []byte("1234567890")
	r := json.NewReader(nil)
	var n Int64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(tok)
		n.ReadJSON(r)
	}
}

func BenchmarkRead_Uint64(b *testing.B) {
	tok := []byte("18446744073709551615")
	r := json.NewReader(nil)
	var n Uint64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(tok)
		n.ReadJSON(r)
	}
}

func BenchmarkRead_Float64(b *testing.B) {
	tok := []byte("3.141592653589793")
	r := json.NewReader(nil)
	var n Float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(tok)
		n.ReadJSON(r)
	}
}

func BenchmarkRead_RawNumber(b *testing.B) {
	tok := []byte("123456789012345678901234567890")
	r := json.NewReader(nil)
	var n RawNumber
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(tok)
		n.ReadJSON(r)
	}
}

// Contrast: big.Int is inherently allocating (heap mantissa) — expected, and
// not the hot path. Shown so the difference from scalar/raw is explicit.
func BenchmarkRead_BigInt(b *testing.B) {
	tok := []byte("123456789012345678901234567890")
	r := json.NewReader(nil)
	var n BigInt
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(tok)
		n.ReadJSON(r)
	}
}

func BenchmarkWrite_Int64(b *testing.B) {
	var n Int64
	n.Set(1234567890)
	w := json.NewWriter(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Reset()
		n.WriteJSON(w)
	}
}

func BenchmarkWrite_Float64(b *testing.B) {
	var n Float64
	n.Set(3.141592653589793)
	w := json.NewWriter(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Reset()
		n.WriteJSON(w)
	}
}

func BenchmarkWrite_RawNumber(b *testing.B) {
	var n RawNumber
	n.Set([]byte("123456789012345678901234567890"))
	w := json.NewWriter(64)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Reset()
		n.WriteJSON(w)
	}
}
