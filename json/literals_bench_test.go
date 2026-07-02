package json

import "testing"

func BenchmarkReadNull(b *testing.B) {
	data := []byte("null")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.ReadNull()
	}
}

func BenchmarkReadBool_True(b *testing.B) {
	data := []byte("true")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.ReadBoolean()
	}
}

func BenchmarkReadBool_False(b *testing.B) {
	data := []byte("false")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.ReadBoolean()
	}
}

// not-a-literal (the common "peek fails" path, e.g. ReadNull on a non-null value)
func BenchmarkReadNull_Miss(b *testing.B) {
	data := []byte("12345")
	r := &Reader{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		r.ReadNull()
	}
}
