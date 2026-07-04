package gen

import (
	"os"
	"strings"
	"testing"
)

// TestGenerateGolden regenerates each committed, compiled golden file from its
// schema and asserts it still matches — so any generator drift (or a change that
// would break the example) fails here.
func TestGenerateGolden(t *testing.T) {
	cases := []struct{ schema, golden, name string }{
		{"testdata/person.schema.json", "example/person.go", "Person"},
		{"testdata/order.schema.json", "example/order.go", "Order"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			schema, err := os.ReadFile(c.schema)
			if err != nil {
				t.Fatalf("read schema: %v", err)
			}
			got, err := Generate(schema, "example", c.name)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			want, err := os.ReadFile(c.golden)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("generated output drifted from %s.\n"+
					"Regenerate the golden file if the change is intended.\n--- got ---\n%s", c.golden, got)
			}
		})
	}
}

func TestGenerateRejectsNonObjectRoot(t *testing.T) {
	if _, err := Generate([]byte(`{"type":"string"}`), "example", "X"); err == nil {
		t.Errorf("expected error for non-object root")
	}
}

// TestGenerateFormats exercises the full string-format table: each format maps to
// the right validator type, builder call, Result payload and import.
func TestGenerateFormats(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"ip":   {"type": "string", "format": "ipv4"},
			"when": {"type": "string", "format": "date-time"},
			"day":  {"type": "string", "format": "date"},
			"dur":  {"type": "string", "format": "duration"},
			"re":   {"type": "string", "format": "regex"},
			"host": {"type": "string", "format": "hostname"},
			"ptr":  {"type": "string", "format": "json-pointer"},
			"ref":  {"type": "string", "format": "uri-reference"}
		}
	}`
	out, err := Generate([]byte(schema), "example", "Formats")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)

	for _, want := range []string{
		"*validation.IP", ".IP().Version4()", "validation.Result[net.IP]", `"net"`,
		"*validation.Time", ".DateTime()", ".Date()", "validation.Result[time.Time]", `"time"`,
		"*validation.Duration", ".Duration()", "validation.Result[time.Duration]",
		"*validation.Regex", ".Regex()", "validation.Result[*regexp.Regexp]", `"regexp"`,
		"*validation.Hostname", ".Hostname()",
		"*validation.JsonPointer", ".JsonPointer()",
		".UriReference()", `"net/url"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated output missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateUnresolvedRef(t *testing.T) {
	s := `{"type":"object","properties":{"x":{"$ref":"#/$defs/Missing"}}}`
	if _, err := Generate([]byte(s), "example", "X"); err == nil {
		t.Errorf("expected error for unresolved $ref")
	}
}

// TestGenerateDraft07Definitions confirms draft-07 "definitions" + "#/definitions/"
// refs resolve the same as 2020-12 "$defs".
func TestGenerateDraft07Definitions(t *testing.T) {
	s := `{"type":"object",
		"definitions":{"Inner":{"type":"object","properties":{"a":{"type":"string"}}}},
		"properties":{"inner":{"$ref":"#/definitions/Inner"}}}`
	out, err := Generate([]byte(s), "example", "Root")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{"type Inner struct", "types.Object[Inner, *Inner]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

// TestGenerateUnknownFormatIgnored confirms an unrecognized format falls back to a
// plain string validator (JSON Schema treats format as an annotation).
func TestGenerateUnknownFormatIgnored(t *testing.T) {
	out, err := Generate([]byte(`{"type":"object","properties":{"x":{"type":"string","format":"weird"}}}`), "example", "X")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if strings.Contains(string(out), "*validation.String") == false {
		t.Errorf("unknown format should map to plain *validation.String:\n%s", out)
	}
}
