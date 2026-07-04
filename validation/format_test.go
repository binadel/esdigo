package validation

import (
	"testing"

	"github.com/binadel/esdigo/json/types"
)

func TestEmail(t *testing.T) {
	v := readInto[types.String](`"user@example.com"`)
	if r := NewString("e").Email().Validate(v); !r.IsValid() || r.Value == nil {
		t.Errorf("valid email: %s", resultJSON(r))
	}
	v = readInto[types.String](`"not-an-email"`)
	if r := NewString("e").Email().Validate(v); r.IsValid() {
		t.Errorf("invalid email should fail")
	} else {
		mustContain(t, "email", resultJSON(r), "EMAIL")
	}
	// allowed null skips the format parse and stays valid
	if r := NewString("e").Email().Validate(readInto[types.String]("null")); !r.IsValid() {
		t.Errorf("nullable email null should be valid: %s", resultJSON(r))
	}
}

func TestIP(t *testing.T) {
	v := readInto[types.String](`"192.168.0.1"`)
	if r := NewString("ip").IP().Validate(v); !r.IsValid() {
		t.Errorf("valid ipv4: %s", resultJSON(r))
	}
	// v4 required, given v6 -> IPv4 error
	v = readInto[types.String](`"::1"`)
	if r := NewString("ip").IP().Version4().Validate(v); r.IsValid() {
		t.Errorf("ipv6 into v4 should fail")
	} else {
		mustContain(t, "ipv4", resultJSON(r), "IPv4")
	}
	// v6 required, given v4 -> IPv6 error
	v = readInto[types.String](`"192.168.0.1"`)
	if r := NewString("ip").IP().Version6().Validate(v); r.IsValid() {
		t.Errorf("ipv4 into v6 should fail")
	} else {
		mustContain(t, "ipv6", resultJSON(r), "IPv6")
	}
	// garbage -> IP error
	v = readInto[types.String](`"nope"`)
	if r := NewString("ip").IP().Validate(v); r.IsValid() {
		t.Errorf("garbage ip should fail")
	} else {
		mustContain(t, "ip", resultJSON(r), "IP")
	}
}

func TestUri(t *testing.T) {
	v := readInto[types.String](`"https://example.com/x"`)
	if r := NewString("u").Uri().Validate(v); !r.IsValid() || r.Value == nil {
		t.Errorf("valid uri: %s", resultJSON(r))
	}
	v = readInto[types.String](`"://bad"`)
	if r := NewString("u").Uri().Validate(v); r.IsValid() {
		t.Errorf("bad uri should fail")
	} else {
		mustContain(t, "uri", resultJSON(r), "URI")
	}
}

func TestUuid(t *testing.T) {
	v := readInto[types.String](`"123e4567-e89b-12d3-a456-426614174000"`)
	if r := NewString("id").Uuid().Validate(v); !r.IsValid() {
		t.Errorf("valid uuid: %s", resultJSON(r))
	}
	// version mismatch: the above is v1; require v4
	v = readInto[types.String](`"123e4567-e89b-12d3-a456-426614174000"`)
	if r := NewString("id").Uuid().Version(4).Validate(v); r.IsValid() {
		t.Errorf("uuid version mismatch should fail")
	} else {
		mustContain(t, "uuid-version", resultJSON(r), `"version":4`)
	}
	v = readInto[types.String](`"nope"`)
	if r := NewString("id").Uuid().Validate(v); r.IsValid() {
		t.Errorf("garbage uuid should fail")
	} else {
		mustContain(t, "uuid", resultJSON(r), "UUID")
	}
}

func TestTimeFormats(t *testing.T) {
	// date
	v := readInto[types.String](`"2026-07-04"`)
	if r := NewString("d").Date().Validate(v); !r.IsValid() {
		t.Errorf("valid date: %s", resultJSON(r))
	}
	v = readInto[types.String](`"07/04/2026"`)
	if r := NewString("d").Date().Validate(v); r.IsValid() {
		t.Errorf("bad date should fail")
	} else {
		mustContain(t, "date", resultJSON(r), "DATE")
	}
	// time
	v = readInto[types.String](`"13:04:05"`)
	if r := NewString("t").Time().Validate(v); !r.IsValid() {
		t.Errorf("valid time: %s", resultJSON(r))
	}
	v = readInto[types.String](`"nope"`)
	if r := NewString("t").Time().Validate(v); r.IsValid() {
		t.Errorf("bad time should fail")
	} else {
		mustContain(t, "time", resultJSON(r), `"TIME"`)
	}
	// datetime
	v = readInto[types.String](`"2026-07-04T13:04:05Z"`)
	if r := NewString("dt").DateTime().Validate(v); !r.IsValid() {
		t.Errorf("valid datetime: %s", resultJSON(r))
	}
	v = readInto[types.String](`"nope"`)
	if r := NewString("dt").DateTime().Validate(v); r.IsValid() {
		t.Errorf("bad datetime should fail")
	} else {
		mustContain(t, "datetime", resultJSON(r), `"DATE_TIME"`)
	}

	// a custom format is honored
	v = readInto[types.String](`"04/07/2026"`)
	if r := NewString("d").Date("02/01/2006").Validate(v); !r.IsValid() {
		t.Errorf("custom date format: %s", resultJSON(r))
	}
}

func TestDuration(t *testing.T) {
	v := readInto[types.String](`"PT1H30M"`)
	if r := NewString("dur").Duration().Validate(v); !r.IsValid() {
		t.Errorf("valid duration: %s", resultJSON(r))
	}
	v = readInto[types.String](`"1h30m"`)
	if r := NewString("dur").Duration().Validate(v); r.IsValid() {
		t.Errorf("non-ISO8601 duration should fail")
	} else {
		mustContain(t, "duration", resultJSON(r), "DURATION")
	}
}

func TestRegexFormat(t *testing.T) {
	v := readInto[types.String](`"^[a-z]+$"`)
	if r := NewString("re").Regex().Validate(v); !r.IsValid() || r.Value == nil {
		t.Errorf("valid regex: %s", resultJSON(r))
	}
	v = readInto[types.String](`"[unterminated"`)
	if r := NewString("re").Regex().Validate(v); r.IsValid() {
		t.Errorf("bad regex should fail")
	} else {
		mustContain(t, "regex", resultJSON(r), "REGEX")
	}
}

// TestFormatNullGuard confirms every format validator accepts an allowed null.
func TestFormatNullGuard(t *testing.T) {
	null := func() types.String { return readInto[types.String]("null") }
	checks := map[string]bool{
		"email":    isValid(NewString("f").Email().Validate(null())),
		"ip":       isValid(NewString("f").IP().Validate(null())),
		"uri":      isValid(NewString("f").Uri().Validate(null())),
		"uuid":     isValid(NewString("f").Uuid().Validate(null())),
		"date":     isValid(NewString("f").Date().Validate(null())),
		"time":     isValid(NewString("f").Time().Validate(null())),
		"datetime": isValid(NewString("f").DateTime().Validate(null())),
		"duration": isValid(NewString("f").Duration().Validate(null())),
		"regex":    isValid(NewString("f").Regex().Validate(null())),
	}
	for name, valid := range checks {
		if !valid {
			t.Errorf("nullable %s null should be valid", name)
		}
	}
	// but notNull still rejects null
	if isValid(NewString("f").NotNull().Email().Validate(null())) {
		t.Errorf("notNull email null should fail")
	}
}
