package types

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestString_Read(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`"hello"`, "hello"},
		{`""`, ""},
		{`"a\"b\\c\n\t"`, "a\"b\\c\n\t"},
		{`"é"`, "é"},
		{`"😀"`, "😀"},
	}
	for _, c := range cases {
		var s String
		if !readCont(&s, c.in) {
			t.Fatalf("String(%s): reader could not continue", c.in)
		}
		assertState(t, "String("+c.in+")", &s, true, true, true)
		if string(s.Value) != c.want {
			t.Errorf("String(%s): value = %q, want %q", c.in, s.Value, c.want)
		}
	}

	// null
	var n String
	readCont(&n, "null")
	assertState(t, "String(null)", &n, true, false, false)

	// wrong type
	for _, in := range []string{"1", "true", "[]", "{}"} {
		var s String
		if !readCont(&s, in) {
			t.Errorf("String(%s): reader could not continue", in)
		}
		assertState(t, "String("+in+")", &s, true, true, false)
	}
}

func TestString_RoundTrip(t *testing.T) {
	roundtrip[String, *String](t, `"hello"`)
	roundtrip[String, *String](t, `""`)
	roundtrip[String, *String](t, `"a\"b\n\t"`)
	roundtrip[String, *String](t, `"é😀<>&"`)
	roundtrip[String, *String](t, "null")
}

func TestString_Setters(t *testing.T) {
	check := func(name string, s *String, want string) {
		t.Helper()
		out, ok := writeStr(s)
		if !ok {
			t.Errorf("%s: WriteJSON returned false", name)
			return
		}
		if out != want {
			t.Errorf("%s: wrote %s, want %s", name, out, want)
		}
	}

	var s String
	s.SetString("hi")
	check("SetString", &s, `"hi"`)

	s.Set([]byte("bytes"))
	check("Set", &s, `"bytes"`)

	s.SetTime(time.Date(2021, 3, 4, 5, 6, 7, 0, time.UTC), time.RFC3339)
	check("SetTime", &s, `"2021-03-04T05:06:07Z"`)

	// zero time -> null
	s.SetTime(time.Time{}, time.RFC3339)
	assertState(t, "SetTime(zero)", &s, true, false, false)

	addr, _ := mail.ParseAddress("Jane <jane@example.com>")
	s.SetEmail(addr)
	check("SetEmail", &s, `"\"Jane\" <jane@example.com>"`)

	s.SetEmail(nil)
	assertState(t, "SetEmail(nil)", &s, true, false, false)

	s.SetIP(net.ParseIP("192.168.0.1"))
	check("SetIP", &s, `"192.168.0.1"`)

	u, _ := url.Parse("https://example.com/x?y=1")
	s.SetUri(u)
	check("SetUri", &s, `"https://example.com/x?y=1"`)

	s.SetUuid(uuid.MustParse("12345678-1234-1234-1234-123456789012"))
	check("SetUuid", &s, `"12345678-1234-1234-1234-123456789012"`)

	s.SetDuration(90 * time.Minute)
	check("SetDuration", &s, `"PT1H30M"`)

	s.SetRegex(regexp.MustCompile(`^a.*z$`))
	check("SetRegex", &s, `"^a.*z$"`)

	s.SetRegex(nil)
	assertState(t, "SetRegex(nil)", &s, true, false, false)
}
