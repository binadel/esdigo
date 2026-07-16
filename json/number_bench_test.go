package json

import (
	"strings"
	"testing"
)

func benchReadNum(b *testing.B, s string) {
	buf := []byte(s)
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(buf)
		r.ReadNumber()
	}
}
func benchSkipNum(b *testing.B, s string) {
	buf := []byte(s)
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(buf)
		r.SkipNumber()
	}
}

func BenchmarkReadNumber_Int(b *testing.B)     { benchReadNum(b, "12345") }
func BenchmarkReadNumber_BigInt(b *testing.B)  { benchReadNum(b, "12345678901234567") }
func BenchmarkReadNumber_Float(b *testing.B)   { benchReadNum(b, "3.14159265358979") }
func BenchmarkReadNumber_Exp(b *testing.B)     { benchReadNum(b, "-6.022e23") }
func BenchmarkReadNumber_Small(b *testing.B)   { benchReadNum(b, "7") }
func BenchmarkReadNumber_LongInt(b *testing.B) { benchReadNum(b, strings.Repeat("9", 40)) }
func BenchmarkSkipNumber_Int(b *testing.B)     { benchSkipNum(b, "12345") }
func BenchmarkSkipNumber_LongInt(b *testing.B) { benchSkipNum(b, strings.Repeat("9", 40)) }
func BenchmarkSkipNumber_Float(b *testing.B)   { benchSkipNum(b, "3.14159265358979") }
