package json

import (
	"strings"
	"testing"
)

var (
	rawClean   = strings.Repeat("The quick brown fox jumps over the lazy dog. ", 6)
	rawUnicode = strings.Repeat("héllo wörld 世界 😀 ", 8)
	rawEscapes = strings.Repeat("line1\nline2\ttab \"q\" \\s ", 6)

	jsonClean   = `"` + rawClean + `"`
	jsonUnicode = `"` + rawUnicode + `"`
	jsonEscapes = func() string { var w Writer; w.WriteString(rawEscapes); return string(w.data) }()
	jsonShort   = `"hello"`
)

func benchRead(b *testing.B, data string) {
	buf := []byte(data)
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(buf)
		r.ReadStringBytes()
	}
}

func BenchmarkReadString_Clean(b *testing.B)   { benchRead(b, jsonClean) }
func BenchmarkReadString_Unicode(b *testing.B) { benchRead(b, jsonUnicode) }
func BenchmarkReadString_Escapes(b *testing.B) { benchRead(b, jsonEscapes) }
func BenchmarkReadString_Short(b *testing.B)   { benchRead(b, jsonShort) }

func benchWrite(b *testing.B, s string) {
	w := NewWriter(len(s) + 16)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Reset()
		w.WriteString(s)
	}
}

func BenchmarkWriteString_Clean(b *testing.B)   { benchWrite(b, rawClean) }
func BenchmarkWriteString_Unicode(b *testing.B) { benchWrite(b, rawUnicode) }
func BenchmarkWriteString_Escapes(b *testing.B) { benchWrite(b, rawEscapes) }
func BenchmarkWriteString_Short(b *testing.B)   { benchWrite(b, "hello") }
