package json

import (
	stdjson "encoding/json"
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"
)

// Inputs ReadString and SkipString must both ACCEPT (well-formed JSON strings,
// including dangling surrogates which decode to RuneError but are still accepted).
var acceptCases = []string{
	`""`, `"hello"`, `"a b c"`, `"!@#$%^&*()"`,
	`"\\"`, `"\""`, `"\/"`, `"\b"`, `"\f"`, `"\n"`, `"\r"`, `"\t"`,
	`"A"`, `"é"`, `"世"`, `"😀"`, `"𝄞"`,
	`"\uD800"`, `"\uDC00"`, `"\uD800A"`, `"\uD800\uD800"`,
	`"世界"`, `"</script>"`, `"<b>&amp;</b>"`,
}

// Inputs both must REJECT.
var rejectCases = []string{
	`"`, `"abc`, `"\`, `"\u`, `"\u1`, `"\u12`, `"\u123`, // truncated / EOF
	`"\x"`, `"\a"`, `"\q"`, `"\ "`, // invalid escape
	`"\uZZZZ"`, `"\u123G"`, `"\uD800\uZZZZ"`, // invalid hex
}

func TestReadString_ValidBasic(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"hello world",
		"123456",
		"!@#$%^&*()",
		"Hello 世界",
		"Привет",
		"مرحبا",
		"😀",
	}

	for _, input := range tests {
		data, _ := stdjson.Marshal(input)

		r := Reader{data: data}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed to read string: %q", input)
		}
		if out != input {
			t.Fatalf("expected %q, got %q", input, out)
		}
	}
}

func TestReadString_Escapes(t *testing.T) {
	tests := map[string]string{
		`"\\""`:    `\`,
		`"\""`:     `"`,
		`"\/"`:     `/`,
		`"\b"`:     "\b",
		`"\f"`:     "\f",
		`"\n"`:     "\n",
		`"\r"`:     "\r",
		`"\t"`:     "\t",
		`"\u0041"`: "A",
		`"\u0061"`: "a",
		`"\u00E9"`: "é",
		`"\u4E16"`: "世",
	}

	for raw, expected := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed reading %s", raw)
		}
		if out != expected {
			t.Fatalf("expected %q got %q", expected, out)
		}
	}
}

func TestReadString_SurrogatePairs(t *testing.T) {
	tests := map[string]string{
		`"\uD83D\uDE00"`: "😀",
		`"\uD834\uDD1E"`: "𝄞", // musical symbol
	}

	for raw, expected := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed reading %s", raw)
		}
		if out != expected {
			t.Fatalf("expected %q got %q", expected, out)
		}
	}
}

func TestReadString_InvalidSurrogates(t *testing.T) {
	tests := []string{
		`"\uD800"`,       // dangling high
		`"\uDC00"`,       // lone low
		`"\uD800\u0041"`, // invalid pair
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("unexpected failure for %s", raw)
		}
		if !strings.ContainsRune(out, utf8.RuneError) {
			t.Fatalf("expected RuneError in %q", out)
		}
	}
}

func TestReadString_InvalidEscapes(t *testing.T) {
	tests := []string{
		`"\x"`,
		`"\a"`,
		`"\u12"`,
		`"\uZZZZ"`,
		`"\u123G"`,
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected failure for %s", raw)
		}
	}
}

func TestReadString_ControlCharacters(t *testing.T) {
	for i := 0; i < 0x20; i++ {
		if i == '\n' || i == '\r' || i == '\t' {
			continue
		}
		raw := "\"" + string(byte(i)) + "\""
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected failure for control char 0x%02x", i)
		}
	}
}

func TestReadString_EOF(t *testing.T) {
	tests := []string{
		`"`,
		`"abc`,
		`"\`,
		`"\u123`,
	}

	for _, raw := range tests {
		r := Reader{data: []byte(raw)}
		_, ok := r.ReadString()
		if ok {
			t.Fatalf("expected EOF failure for %s", raw)
		}
	}
}

func TestSkipString(t *testing.T) {
	input := `"hello\nworld" 123`
	r := Reader{data: []byte(input)}

	if !r.SkipString() {
		t.Fatal("skip failed")
	}

	if r.pos != len(`"hello\nworld"`) {
		t.Fatalf("unexpected pos %d", r.pos)
	}
}

func TestWriter_RoundTrip(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"Hello 世界",
		"😀",
		"line\nbreak",
		"tab\tchar",
		`quote"slash\`,
	}

	for _, input := range tests {
		var w Writer
		w.WriteString(input)

		r := Reader{data: w.data}
		out, ok := r.ReadString()
		if !ok {
			t.Fatalf("failed roundtrip for %q", input)
		}
		if out != input {
			t.Fatalf("roundtrip mismatch: %q != %q", out, input)
		}
	}
}

func TestCompatibilityWithStdlib(t *testing.T) {
	inputs := []string{
		"simple",
		"with\nnewline",
		"😀 emoji",
		"世界",
	}

	for _, input := range inputs {
		stdEncoded, _ := stdjson.Marshal(input)

		var w Writer
		w.WriteString(input)

		if string(w.data) != string(stdEncoded) {
			t.Fatalf("encoding mismatch:\nstd: %s\nour: %s",
				stdEncoded, w.data)
		}

		r := Reader{data: stdEncoded}
		out, ok := r.ReadString()
		if !ok {
			t.Fatal("read failed")
		}
		if out != input {
			t.Fatalf("decode mismatch %q != %q", out, input)
		}
	}
}

func TestSkipString_Accept(t *testing.T) {
	for _, in := range acceptCases {
		r := Reader{data: []byte(in)}
		if !r.SkipString() {
			t.Fatalf("SkipString rejected valid %q (err=%v)", in, r.err)
		}
		if r.pos != len(in) {
			t.Fatalf("SkipString %q: pos=%d want %d", in, r.pos, len(in))
		}
	}
}

func TestSkipString_Reject(t *testing.T) {
	for _, in := range rejectCases {
		r := Reader{data: []byte(in)}
		if r.SkipString() {
			t.Fatalf("SkipString accepted invalid %q", in)
		}
	}
}

// ReadStringBytes and SkipString must agree on accept/reject and end position for
// every input — they duplicate the scanning logic.
func TestReadSkipConsistency(t *testing.T) {
	all := append([]string{}, acceptCases...)
	all = append(all, rejectCases...)
	all = append(all, `123`, `true`, `null`, `{`, `[`) // not-a-string
	// raw (unescaped) control chars inside the string -> both reject
	for i := 0; i < 0x20; i++ {
		all = append(all, "\""+string(rune(i))+"\"")
	}
	for _, in := range all {
		rr := Reader{data: []byte(in)}
		_, rok := rr.ReadStringBytes()

		rs := Reader{data: []byte(in)}
		sok := rs.SkipString()

		if rok != sok {
			t.Fatalf("%q: ReadStringBytes ok=%v but SkipString ok=%v", in, rok, sok)
		}
		if rok && rr.pos != rs.pos {
			t.Fatalf("%q: ReadStringBytes pos=%d but SkipString pos=%d", in, rr.pos, rs.pos)
		}
	}
}

// Unescaped control characters (0x00-0x1F) must be rejected by both.
func TestString_RejectRawControls(t *testing.T) {
	for i := 0; i < 0x20; i++ {
		in := "\"" + string(rune(i)) + "\""
		if _, ok := (&Reader{data: []byte(in)}).ReadStringBytes(); ok {
			t.Fatalf("ReadStringBytes accepted raw control 0x%02x", i)
		}
		if (&Reader{data: []byte(in)}).SkipString() {
			t.Fatalf("SkipString accepted raw control 0x%02x", i)
		}
	}
}

func TestReadStringBytes_NotAString(t *testing.T) {
	for _, in := range []string{`123`, `true`, `null`, `{`, `[`} {
		r := Reader{data: []byte(in)}
		_, ok := r.ReadStringBytes()
		if ok {
			t.Fatalf("%q: expected not-a-string", in)
		}
		if r.err != nil {
			t.Fatalf("%q: not-a-string must not set error, got %v", in, r.err)
		}
		if r.pos != 0 {
			t.Fatalf("%q: pos must not move, got %d", in, r.pos)
		}
	}
}

// Every control char must be writable and decode back to itself (esdigo + stdlib).
func TestWriteEscaped_AllControls(t *testing.T) {
	for i := 0; i < 0x20; i++ {
		s := "a" + string(rune(i)) + "b"
		var w Writer
		w.WriteString(s)

		var std string
		if err := stdjson.Unmarshal(w.data, &std); err != nil || std != s {
			t.Fatalf("control 0x%02x: stdlib decode of %q -> %q err=%v", i, w.data, std, err)
		}
		r := Reader{data: w.data}
		out, ok := r.ReadString()
		if !ok || out != s {
			t.Fatalf("control 0x%02x: esdigo roundtrip of %q -> %q ok=%v", i, w.data, out, ok)
		}
	}
}

// esdigo does NOT HTML-escape (<, >, &) — spec-compliant, differs from stdlib default.
func TestWrite_HTMLCharsRaw(t *testing.T) {
	var w Writer
	w.WriteString("<a> & </a>")
	if got := string(w.data); got != `"<a> & </a>"` {
		t.Fatalf("esdigo should write HTML chars raw, got %q", got)
	}
	var back string
	if err := stdjson.Unmarshal(w.data, &back); err != nil || back != "<a> & </a>" {
		t.Fatalf("stdlib decode: back=%q err=%v", back, err)
	}
}

func TestReadString_SurrogateEdges(t *testing.T) {
	// high + high: each is dangling -> two RuneErrors.
	r := Reader{data: []byte(`"\uD800\uD800"`)}
	out, ok := r.ReadString()
	if !ok {
		t.Fatal("high+high must be accepted")
	}
	if utf8.RuneCountInString(out) != 2 || strings.Count(out, string(utf8.RuneError)) != 2 {
		t.Fatalf("high+high: want 2 RuneErrors, got %q", out)
	}
	// high + non-escape char.
	r2 := Reader{data: []byte(`"\uD800X"`)}
	out2, ok2 := r2.ReadString()
	if !ok2 || out2 != string(utf8.RuneError)+"X" {
		t.Fatalf("high+X: got %q ok=%v", out2, ok2)
	}
	// valid pair followed by more content.
	r3 := Reader{data: []byte(`"😀!"`)}
	out3, ok3 := r3.ReadString()
	if !ok3 || out3 != "😀!" {
		t.Fatalf("pair+more: got %q ok=%v", out3, ok3)
	}
}

func randString(rng *rand.Rand) string {
	var b strings.Builder
	for j, n := 0, rng.Intn(24); j < n; j++ {
		switch rng.Intn(12) {
		case 0:
			b.WriteByte(byte(rng.Intn(0x20))) // control char
		case 1:
			b.WriteByte([]byte{'"', '\\', '/'}[rng.Intn(3)])
		case 2:
			b.WriteByte(byte('<' + rng.Intn(3))) // <, =, > (HTML-sensitive)
		case 3, 4:
			b.WriteByte(byte(0x20 + rng.Intn(0x5f)))
		default:
			for {
				c := rune(rng.Intn(0x110000))
				if (c < 0xD800 || c > 0xDFFF) && utf8.ValidRune(c) {
					b.WriteRune(c)
					break
				}
			}
		}
	}
	return b.String()
}

// Differential vs stdlib: esdigo must decode any stdlib-encoded string back to the
// original (exercises \u escapes, surrogate pairs, HTML/control escapes).
func TestString_DifferentialStdlibRead(t *testing.T) {
	n := 100000
	if testing.Short() {
		n = 3000
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		s := randString(rng)
		enc, _ := stdjson.Marshal(s)
		r := Reader{data: enc}
		out, ok := r.ReadString()
		if !ok || out != s || r.pos != len(enc) {
			t.Fatalf("read stdlib-enc %s (in=%q): out=%q ok=%v pos=%d/%d", enc, s, out, ok, r.pos, len(enc))
		}
	}
}

// Round-trip: esdigo write -> esdigo read == original, and stdlib accepts esdigo output.
func TestString_RoundTripFuzz(t *testing.T) {
	n := 100000
	if testing.Short() {
		n = 3000
	}
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < n; i++ {
		s := randString(rng)
		var w Writer
		w.WriteString(s)

		r := Reader{data: w.data}
		out, ok := r.ReadString()
		if !ok || out != s {
			t.Fatalf("esdigo roundtrip %q -> %s -> %q ok=%v", s, w.data, out, ok)
		}
		var std string
		if err := stdjson.Unmarshal(w.data, &std); err != nil || std != s {
			t.Fatalf("stdlib rejects/mismatches esdigo output %s (in=%q): std=%q err=%v", w.data, s, std, err)
		}
	}
}
