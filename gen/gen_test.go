package gen

import (
	"os"
	"strings"
	"testing"

	"github.com/binadel/esdigo/gen/schema"
)

// TestGenerateGolden regenerates each committed, compiled golden file from its
// schema and asserts it still matches — so any generator drift (or a change that
// would break the example) fails here.
func TestGenerateGolden(t *testing.T) {
	cases := []struct{ schema, golden, pkg, name string }{
		{"testdata/person.schema.json", "example/person.go", "example", "Person"},
		{"testdata/order.schema.json", "example/order.go", "example", "Order"},
		{"testdata/api.openapi.json", "example/api/api.go", "api", ""},
	}
	for _, c := range cases {
		t.Run(c.golden, func(t *testing.T) {
			schema, err := os.ReadFile(c.schema)
			if err != nil {
				t.Fatalf("read schema: %v", err)
			}
			got, err := GenerateAuto(schema, c.pkg, c.name)
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

// TestGenerateOpenAPI checks component extraction, cross-component $ref, and the
// no-components / detection behavior.
func TestGenerateOpenAPI(t *testing.T) {
	doc := `{"openapi":"3.1.0","components":{"schemas":{
		"A":{"type":"object","properties":{"b":{"$ref":"#/components/schemas/B"}}},
		"B":{"type":"object","properties":{"x":{"type":"string"}}}
	}}}`
	if !schema.IsOpenAPI([]byte(doc)) {
		t.Fatalf("should detect OpenAPI")
	}
	out, err := GenerateOpenAPI([]byte(doc), "api")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{"type A struct", "type B struct", "types.Object[B, *B]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}

	if _, err := GenerateOpenAPI([]byte(`{"openapi":"3.1.0"}`), "api"); err == nil {
		t.Errorf("empty components should error")
	}
	// GenerateAuto routes a bare JSON Schema to the named single-root path.
	single, err := GenerateAuto([]byte(`{"type":"object","properties":{"n":{"type":"string"}}}`), "m", "Thing")
	if err != nil || !strings.Contains(string(single), "type Thing struct") {
		t.Errorf("auto should route a bare schema to Generate: %v", err)
	}
}

// TestGenerateDir checks the directory merge: a cross-file $ref into another
// file's $defs resolves, and the shared type is generated once.
func TestGenerateDir(t *testing.T) {
	files := map[string][]byte{
		"user.schema.json": []byte(`{"type":"object","required":["id"],"properties":{
			"id":{"type":"integer"},
			"address":{"$ref":"common.json#/$defs/Address"}}}`),
		"common.json": []byte(`{"$defs":{"Address":{"type":"object","required":["city"],
			"properties":{"city":{"type":"string"}}}}}`),
	}
	out, err := GenerateDir(files, "models")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{"package models", "type User struct", "type Address struct", "types.Object[Address, *Address]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	if n := strings.Count(got, "type Address struct"); n != 1 {
		t.Errorf("Address should be generated once, got %d", n)
	}
}

// TestGenerateDirDedup: two files referencing the same shared type generate it once.
func TestGenerateDirDedup(t *testing.T) {
	files := map[string][]byte{
		"a.schema.json": []byte(`{"type":"object","properties":{"addr":{"$ref":"common.json#/$defs/Address"}}}`),
		"b.schema.json": []byte(`{"type":"object","properties":{"addr":{"$ref":"common.json#/$defs/Address"}}}`),
		"common.json":   []byte(`{"$defs":{"Address":{"type":"object","properties":{"city":{"type":"string"}}}}}`),
	}
	out, err := GenerateDir(files, "models")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	if n := strings.Count(got, "type Address struct"); n != 1 {
		t.Errorf("shared Address should be generated once, got %d:\n%s", n, got)
	}
	for _, want := range []string{"type A struct", "type B struct"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q", want)
		}
	}
}

// TestGenerateDirFileRootRef resolves a bare "<file>.json" ref to that file's root.
func TestGenerateDirFileRootRef(t *testing.T) {
	files := map[string][]byte{
		"user.schema.json": []byte(`{"type":"object","properties":{"home":{"$ref":"address.json"}}}`),
		"address.json":     []byte(`{"type":"object","required":["city"],"properties":{"city":{"type":"string"}}}`),
	}
	out, err := GenerateDir(files, "models")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{"type User struct", "type Address struct", "types.Object[Address, *Address]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateDirNoObjects(t *testing.T) {
	files := map[string][]byte{"x.json": []byte(`{"type":"string"}`)}
	if _, err := GenerateDir(files, "m"); err == nil {
		t.Errorf("a directory with no object schemas should error")
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

// TestGenerateArrayElements maps each scalar array element type to its wrapper.
func TestGenerateArrayElements(t *testing.T) {
	s := `{"type":"object","properties":{
		"s":{"type":"array","items":{"type":"string"}},
		"i":{"type":"array","items":{"type":"integer"}},
		"n":{"type":"array","items":{"type":"number"}},
		"b":{"type":"array","items":{"type":"boolean"}}
	}}`
	out, err := Generate([]byte(s), "example", "Lists")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"types.Array[types.String, *types.String]",
		"types.Array[types.Int64, *types.Int64]",
		"types.Array[types.Float64, *types.Float64]",
		"types.Array[types.Boolean, *types.Boolean]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

func TestGenerateArrayErrors(t *testing.T) {
	if _, err := Generate([]byte(`{"type":"object","properties":{"x":{"type":"array"}}}`), "example", "X"); err == nil {
		t.Errorf("array without items should error")
	}
	nested := `{"type":"object","properties":{"x":{"type":"array","items":{"type":"array","items":{"type":"string"}}}}}`
	if _, err := Generate([]byte(nested), "example", "X"); err == nil {
		t.Errorf("array of arrays should error (not supported yet)")
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
