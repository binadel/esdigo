package json

// Comparative benchmarks: esdigo vs encoding/json across several document shapes,
// for reading, writing, validating, and typed struct (de)serialization.
//
// Run:  go test ./json/ -run '^$' -bench 'Benchmark(DOMRead|DOMWrite|Validate|Struct)' -benchmem
//
// Fairness notes:
//   - DOMRead compares esdigo ReadJSON (-> json.Value tree) against Unmarshal into
//     `any` (-> map/[]any tree). Both build a generic DOM. esdigo keeps numbers as
//     raw bytes (lazy parse) where stdlib eagerly parses to float64 — a real design
//     difference, shown as-is.
//   - DOMWrite uses a REUSED esdigo Writer (the intended pooled usage) vs stdlib
//     Marshal (which has no buffer-reuse API). This reflects how each is used.
//   - Struct* compares a hand-written esdigo reader/writer (what codegen emits,
//     using the reader/writer primitives) against reflection-based Unmarshal/Marshal.

import (
	stdjson "encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/binadel/esdigo/utils"
)

// --- payloads ---

var (
	pSmall  = []byte(`{"id":12345,"name":"widget-9000","active":true,"price":19.99,"qty":7,"sku":"WX-900-ABC"}`)
	pNested = []byte(`{"service":"api","version":3,"enabled":true,"limits":{"rps":1000,"burst":50,"window":60.5},` +
		`"regions":["us-east","us-west","eu-central"],"meta":{"owner":"platform","tier":"gold","flags":{"beta":false,"canary":true}}}`)
	pNumbers = buildNumbersPayload()
	pStrings = buildStringsPayload()
	pObjects = buildObjectsPayload()

	pUser = []byte(`{"id":998877,"name":"Jane Doe","email":"jane@example.com","active":true,"age":34,` +
		`"score":91.5,"tags":["admin","early-adopter","beta"],` +
		`"address":{"street":"123 Main St","city":"Springfield","zip":"01234"}}`)
)

func buildNumbersPayload() []byte {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		if i%2 == 0 {
			sb.WriteString(strconv.Itoa(i * 7919))
		} else {
			sb.WriteString(strconv.FormatFloat(float64(i)+0.125, 'g', -1, 64))
		}
	}
	sb.WriteByte(']')
	return []byte(sb.String())
}

func buildStringsPayload() []byte {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 60; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		if i%5 == 0 {
			sb.WriteString(`"line with \"quotes\" and a \n newline"`)
		} else {
			sb.WriteString(`"item-` + strconv.Itoa(i) + `-value"`)
		}
	}
	sb.WriteByte(']')
	return []byte(sb.String())
}

func buildObjectsPayload() []byte {
	var sb strings.Builder
	sb.WriteByte('[')
	for i := 0; i < 50; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{"id":` + strconv.Itoa(i))
		sb.WriteString(`,"name":"user_` + strconv.Itoa(i) + `"`)
		if i%2 == 0 {
			sb.WriteString(`,"active":true`)
		} else {
			sb.WriteString(`,"active":false`)
		}
		sb.WriteString(`,"score":` + strconv.FormatFloat(float64(i)*1.5, 'g', -1, 64))
		sb.WriteString(`,"tags":["a","b"]}`)
	}
	sb.WriteByte(']')
	return []byte(sb.String())
}

var scenarios = []struct {
	name string
	data []byte
}{
	{"Small", pSmall},
	{"Nested", pNested},
	{"Numbers100", pNumbers},
	{"Strings60", pStrings},
	{"Objects50", pObjects},
}

// --- DOM read: esdigo ReadJSON vs stdlib Unmarshal(any) ---

func BenchmarkDOMRead(b *testing.B) {
	for _, s := range scenarios {
		b.Run(s.name+"/esdigo", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				r := NewReader(s.data)
				if _, err := r.ReadJSON(); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(s.name+"/stdlib", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				var v any
				if err := stdjson.Unmarshal(s.data, &v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// --- DOM write: esdigo WriteValue (reused writer) vs stdlib Marshal ---

func BenchmarkDOMWrite(b *testing.B) {
	for _, s := range scenarios {
		ev, err := NewReader(s.data).ReadJSON()
		if err != nil {
			b.Fatal(err)
		}
		var sv any
		if err := stdjson.Unmarshal(s.data, &sv); err != nil {
			b.Fatal(err)
		}

		b.Run(s.name+"/esdigo", func(b *testing.B) {
			w := NewWriter(len(s.data) + 64)
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				w.Reset()
				if !w.WriteValue(ev) {
					b.Fatal("write failed")
				}
			}
		})
		b.Run(s.name+"/stdlib", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				if _, err := stdjson.Marshal(sv); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// --- validate: esdigo skip-scan vs stdlib json.Valid ---

func esdigoValidate(data []byte) bool {
	r := &Reader{data: data}
	r.SkipWhitespace()
	if !r.SkipValue() {
		return false
	}
	r.SkipWhitespace()
	return r.err == nil && r.pos == len(r.data)
}

func BenchmarkValidate(b *testing.B) {
	for _, s := range scenarios {
		b.Run(s.name+"/esdigo", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				if !esdigoValidate(s.data) {
					b.Fatal("invalid")
				}
			}
		})
		b.Run(s.name+"/stdlib", func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(s.data)))
			for i := 0; i < b.N; i++ {
				if !stdjson.Valid(s.data) {
					b.Fatal("invalid")
				}
			}
		})
	}
}

// --- typed struct: hand-written esdigo (de)serializer vs reflection ---

type benchAddr struct {
	Street string `json:"street"`
	City   string `json:"city"`
	Zip    string `json:"zip"`
}

type benchUser struct {
	ID     int64     `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Active bool      `json:"active"`
	Age    int       `json:"age"`
	Score  float64   `json:"score"`
	Tags   []string  `json:"tags"`
	Addr   benchAddr `json:"address"`
}

func readStringArray(r *Reader, out *[]string) bool {
	r.SkipWhitespace()
	if r.ReadNull() {
		return true
	}
	if !r.BeginArray() {
		return false
	}
	r.SkipWhitespace()
	if r.EndArray() {
		return true
	}
	for {
		s, ok := r.ReadString()
		if !ok {
			return false
		}
		*out = append(*out, s)
		r.SkipWhitespace()
		if r.EndArray() {
			return true
		}
		if !r.ValueSeparator() {
			return false
		}
		r.SkipWhitespace()
	}
}

func readBenchAddr(r *Reader, a *benchAddr) bool {
	r.SkipWhitespace()
	if r.ReadNull() {
		return true
	}
	if !r.BeginObject() {
		return false
	}
	r.SkipWhitespace()
	if r.EndObject() {
		return true
	}
	for {
		key, ok := r.ReadString()
		if !ok {
			return false
		}
		r.SkipWhitespace()
		if !r.NameSeparator() {
			return false
		}
		switch key {
		case "street":
			a.Street, ok = r.ReadString()
		case "city":
			a.City, ok = r.ReadString()
		case "zip":
			a.Zip, ok = r.ReadString()
		default:
			ok = r.SkipValue()
		}
		if !ok {
			return false
		}
		r.SkipWhitespace()
		if r.EndObject() {
			return true
		}
		if !r.ValueSeparator() {
			return false
		}
		r.SkipWhitespace()
	}
}

func readBenchUser(r *Reader, u *benchUser) bool {
	r.SkipWhitespace()
	if r.ReadNull() {
		return true
	}
	if !r.BeginObject() {
		return false
	}
	r.SkipWhitespace()
	if r.EndObject() {
		return true
	}
	for {
		key, ok := r.ReadString()
		if !ok {
			return false
		}
		r.SkipWhitespace()
		if !r.NameSeparator() {
			return false
		}
		switch key {
		case "id":
			tok, o := r.ReadRawNumber()
			if o {
				u.ID, _ = strconv.ParseInt(utils.UnsafeString(tok), 10, 64)
			}
			ok = o
		case "name":
			u.Name, ok = r.ReadString()
		case "email":
			u.Email, ok = r.ReadString()
		case "active":
			u.Active, ok = r.ReadBoolean()
		case "age":
			tok, o := r.ReadRawNumber()
			if o {
				n, _ := strconv.ParseInt(utils.UnsafeString(tok), 10, 64)
				u.Age = int(n)
			}
			ok = o
		case "score":
			tok, o := r.ReadRawNumber()
			if o {
				u.Score, _ = strconv.ParseFloat(utils.UnsafeString(tok), 64)
			}
			ok = o
		case "tags":
			ok = readStringArray(r, &u.Tags)
		case "address":
			ok = readBenchAddr(r, &u.Addr)
		default:
			ok = r.SkipValue()
		}
		if !ok {
			return false
		}
		r.SkipWhitespace()
		if r.EndObject() {
			return true
		}
		if !r.ValueSeparator() {
			return false
		}
		r.SkipWhitespace()
	}
}

func writeStringArray(w *Writer, in []string) {
	w.BeginArray()
	for i, s := range in {
		if i > 0 {
			w.ValueSeparator()
		}
		w.WriteString(s)
	}
	w.EndArray()
}

func writeBenchAddr(w *Writer, a *benchAddr) {
	w.BeginObject()
	w.WriteRawString(`"street":`)
	w.WriteString(a.Street)
	w.WriteRawString(`,"city":`)
	w.WriteString(a.City)
	w.WriteRawString(`,"zip":`)
	w.WriteString(a.Zip)
	w.EndObject()
}

// writeBenchUser emits pre-escaped constant keys via WriteRawString (the codegen
// fast path: field names are compile-time-known safe ASCII, so no escape scan).
func writeBenchUser(w *Writer, u *benchUser) {
	w.BeginObject()
	w.WriteRawString(`"id":`)
	w.WriteIntNumber(u.ID)
	w.WriteRawString(`,"name":`)
	w.WriteString(u.Name)
	w.WriteRawString(`,"email":`)
	w.WriteString(u.Email)
	w.WriteRawString(`,"active":`)
	w.WriteBoolean(u.Active)
	w.WriteRawString(`,"age":`)
	w.WriteIntNumber(int64(u.Age))
	w.WriteRawString(`,"score":`)
	w.WriteFloatNumber(u.Score, 64)
	w.WriteRawString(`,"tags":`)
	writeStringArray(w, u.Tags)
	w.WriteRawString(`,"address":`)
	writeBenchAddr(w, &u.Addr)
	w.EndObject()
}

func BenchmarkStructRead(b *testing.B) {
	b.Run("esdigo", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(pUser)))
		for i := 0; i < b.N; i++ {
			var u benchUser
			r := NewReader(pUser)
			if !readBenchUser(r, &u) || r.err != nil {
				b.Fatal("read failed")
			}
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(pUser)))
		for i := 0; i < b.N; i++ {
			var u benchUser
			if err := stdjson.Unmarshal(pUser, &u); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStructWrite(b *testing.B) {
	var u benchUser
	if !readBenchUser(NewReader(pUser), &u) {
		b.Fatal("setup read failed")
	}

	b.Run("esdigo", func(b *testing.B) {
		w := NewWriter(len(pUser) + 64)
		b.ReportAllocs()
		b.SetBytes(int64(len(pUser)))
		for i := 0; i < b.N; i++ {
			w.Reset()
			writeBenchUser(w, &u)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(pUser)))
		for i := 0; i < b.N; i++ {
			if _, err := stdjson.Marshal(&u); err != nil {
				b.Fatal(err)
			}
		}
	})
}
