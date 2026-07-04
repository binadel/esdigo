package validation

import (
	"testing"

	"github.com/binadel/esdigo/json/types"
)

func TestStringPresenceAndNull(t *testing.T) {
	var absent types.String
	if r := NewString("s").Required().Validate(absent); r.IsValid() {
		t.Errorf("required+absent should fail")
	} else {
		mustContain(t, "required", resultJSON(r), "REQUIRED")
	}

	null := readInto[types.String]("null")
	if r := NewString("s").Validate(null); !r.IsValid() || r.Defined {
		t.Errorf("nullable null: valid=%v defined=%v", r.IsValid(), r.Defined)
	}
	if r := NewString("s").NotNull().Validate(null); r.IsValid() {
		t.Errorf("notNull null should fail")
	} else {
		mustContain(t, "notNull", resultJSON(r), "NOT_NULL")
	}

	// wrong type (number into string) -> STRING
	wrong := readInto[types.String]("42")
	if r := NewString("s").Validate(wrong); r.IsValid() {
		t.Errorf("number into string should be invalid")
	} else {
		mustContain(t, "wrong-type", resultJSON(r), "STRING")
	}
}

func TestStringValue(t *testing.T) {
	v := readInto[types.String](`"hello"`)
	if r := NewString("s").Validate(v); !r.IsValid() || r.Value != "hello" {
		t.Errorf("hello: valid=%v value=%q", r.IsValid(), r.Value)
	}
}

func TestStringLength(t *testing.T) {
	cases := []struct {
		name    string
		build   func() *String
		input   string
		code    string
		wantErr bool
	}{
		{"len-ok", func() *String { return NewString("s").Length(3) }, `"abc"`, "", false},
		{"len-fail", func() *String { return NewString("s").Length(3) }, `"ab"`, `"length":3`, true},
		{"len0-ok", func() *String { return NewString("s").Length(0) }, `""`, "", false},
		{"len0-fail", func() *String { return NewString("s").Length(0) }, `"x"`, `"length":0`, true},
		{"min-ok", func() *String { return NewString("s").MinLength(2) }, `"ab"`, "", false},
		{"min-fail", func() *String { return NewString("s").MinLength(2) }, `"a"`, `"minLength":2`, true},
		{"max-ok", func() *String { return NewString("s").MaxLength(2) }, `"ab"`, "", false},
		{"max-fail", func() *String { return NewString("s").MaxLength(2) }, `"abc"`, `"maxLength":2`, true},
		// multi-byte runes count as runes, not bytes: "aé" is 3 bytes but 2 runes
		{"rune-ok", func() *String { return NewString("s").Length(2) }, `"aé"`, "", false},
	}
	for _, c := range cases {
		v := readInto[types.String](c.input)
		r := c.build().Validate(v)
		if r.IsValid() == c.wantErr {
			t.Errorf("%s: IsValid=%v wantErr=%v (%s)", c.name, r.IsValid(), c.wantErr, resultJSON(r))
		}
		if c.code != "" {
			mustContain(t, c.name, resultJSON(r), c.code)
		}
	}
}

func TestStringPattern(t *testing.T) {
	v := readInto[types.String](`"abc123"`)
	if r := NewString("s").Pattern(`^[a-z]+[0-9]+$`).Validate(v); !r.IsValid() {
		t.Errorf("pattern match should pass: %s", resultJSON(r))
	}
	v = readInto[types.String](`"ABC"`)
	if r := NewString("s").Pattern(`^[a-z]+$`).Validate(v); r.IsValid() {
		t.Errorf("pattern mismatch should fail")
	} else {
		mustContain(t, "pattern", resultJSON(r), "PATTERN")
	}
}
