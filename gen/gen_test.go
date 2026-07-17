package gen

import (
	"os"
	"sort"
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

// TestGenerateYAMLEquivalent: a YAML schema generates byte-identical output to its
// JSON equivalent — YAML is normalized to JSON at the boundary, so it flows through
// the exact same pipeline.
func TestGenerateYAMLEquivalent(t *testing.T) {
	jsonSchema := []byte(`{"type":"object","required":["id"],"properties":{
		"id":{"type":"string","format":"uuid"},
		"score":{"type":"integer","minimum":0},
		"tags":{"type":"array","items":{"type":"string"}}
	}}`)
	yamlSchema := []byte(`type: object
required: [id]
properties:
  id:
    type: string
    format: uuid
  score:
    type: integer
    minimum: 0
  tags:
    type: array
    items:
      type: string
`)
	fromJSON, err := GenerateAuto(jsonSchema, "m", "T")
	if err != nil {
		t.Fatalf("generate json: %v", err)
	}
	fromYAML, err := GenerateAuto(yamlSchema, "m", "T")
	if err != nil {
		t.Fatalf("generate yaml: %v", err)
	}
	if string(fromJSON) != string(fromYAML) {
		t.Errorf("YAML and JSON should generate identical output.\n--- json ---\n%s\n--- yaml ---\n%s", fromJSON, fromYAML)
	}
}

// TestGenerateOpenAPIYAML: an OpenAPI document authored in YAML is detected and its
// components are generated (the common real-world case — specs are usually YAML).
func TestGenerateOpenAPIYAML(t *testing.T) {
	doc := []byte(`openapi: 3.1.0
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: '#/components/schemas/B'
    B:
      type: object
      properties:
        x:
          type: string
`)
	if !schema.IsOpenAPI(doc) {
		t.Fatalf("should detect YAML OpenAPI")
	}
	out, err := GenerateAuto(doc, "api", "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{"type A struct", "type B struct", "types.Object[B, *B]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

// TestGenerateDirYAML: directory mode picks up .yaml files and cross-file $refs
// resolve across a mix of .json and .yaml.
func TestGenerateDirYAML(t *testing.T) {
	files := map[string][]byte{
		"user.yaml": []byte(`type: object
properties:
  address:
    $ref: 'common.yaml#/$defs/Address'
`),
		"common.yaml": []byte(`$defs:
  Address:
    type: object
    required: [city]
    properties:
      city:
        type: string
`),
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
	if n := strings.Count(got, "type Address struct"); n != 1 {
		t.Errorf("shared Address should be generated once, got %d", n)
	}
}

// TestGenerateAutoFiles: split output puts each type in its own snake_case file,
// carrying only the imports its own message uses (a file with no time field must
// not import "time").
func TestGenerateAutoFiles(t *testing.T) {
	s := `{"type":"object","properties":{
		"when":{"type":"string","format":"date-time"},
		"inner":{"type":"object","properties":{"x":{"type":"string"}}}
	}}`
	files, err := GenerateAutoFiles([]byte(s), "m", "Root")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	root, ok := files["root.go"]
	if !ok {
		t.Fatalf("missing root.go; got files %v", keys(files))
	}
	inner, ok := files["root_inner.go"]
	if !ok {
		t.Fatalf("missing root_inner.go; got files %v", keys(files))
	}

	if !strings.Contains(string(root), "type Root struct") {
		t.Errorf("root.go missing type Root:\n%s", root)
	}
	if !strings.Contains(string(inner), "type RootInner struct") {
		t.Errorf("root_inner.go missing type RootInner:\n%s", inner)
	}
	// import scoping: only root.go has the date-time field
	if !strings.Contains(string(root), `"time"`) {
		t.Errorf("root.go should import time:\n%s", root)
	}
	if strings.Contains(string(inner), `"time"`) {
		t.Errorf("root_inner.go should not import time:\n%s", inner)
	}
	// each file is a standalone source with the header and package clause
	for name, src := range files {
		if !strings.Contains(string(src), "package m") {
			t.Errorf("%s missing package clause", name)
		}
		if !strings.Contains(string(src), "DO NOT EDIT") {
			t.Errorf("%s missing generated header", name)
		}
	}
}

// TestGenerateDirFiles: split output over a directory dedups a shared type into one
// file and resolves cross-file $refs.
func TestGenerateDirFiles(t *testing.T) {
	files := map[string][]byte{
		"a.schema.json": []byte(`{"type":"object","properties":{"addr":{"$ref":"common.json#/$defs/Address"}}}`),
		"b.schema.json": []byte(`{"type":"object","properties":{"addr":{"$ref":"common.json#/$defs/Address"}}}`),
		"common.json":   []byte(`{"$defs":{"Address":{"type":"object","properties":{"city":{"type":"string"}}}}}`),
	}
	out, err := GenerateDirFiles(files, "models")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	for _, name := range []string{"a.go", "b.go", "address.go"} {
		if _, ok := out[name]; !ok {
			t.Errorf("missing %s; got %v", name, keys(out))
		}
	}
	if !strings.Contains(string(out["address.go"]), "type Address struct") {
		t.Errorf("address.go missing type Address:\n%s", out["address.go"])
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestGenerateEnumEscapes: a string enum/const whose JSON spelling uses an escape
// Go's lexer rejects (\/) is decoded and re-quoted, so generation succeeds and the
// emitted Go string literal is valid ("application/json", not "application\/json").
func TestGenerateEnumEscapes(t *testing.T) {
	s := `{"type":"object","properties":{
		"kind":{"type":"string","enum":["application\/json","text\/plain"]},
		"fixed":{"type":"string","const":"a\/b"}
	}}`
	out, err := Generate([]byte(s), "m", "T")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{`.Enum("application/json", "text/plain")`, `.Const("a/b")`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `\/`) {
		t.Errorf("output still contains an invalid Go escape \\/:\n%s", got)
	}
}

// TestGenerateIntegerBounds: numeric bounds on an integer field are emitted as int64
// literals. A value in float syntax that is integral (1e3) is normalized to 1000; a
// large integer is preserved exactly; a fractional or out-of-range bound is a clear
// error rather than code that will not compile.
func TestGenerateIntegerBounds(t *testing.T) {
	ok := `{"type":"object","properties":{"n":{"type":"integer","minimum":1e3,"enum":[1,9007199254740993]}}}`
	out, err := Generate([]byte(ok), "m", "T")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{".Min(1000)", ".Enum(1, 9007199254740993)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}

	frac := `{"type":"object","properties":{"n":{"type":"integer","minimum":1.5}}}`
	if _, err := Generate([]byte(frac), "m", "T"); err == nil {
		t.Errorf("a fractional bound on an integer field should error")
	}
	huge := `{"type":"object","properties":{"n":{"type":"integer","maximum":1e19}}}`
	if _, err := Generate([]byte(huge), "m", "T"); err == nil {
		t.Errorf("an out-of-int64-range bound should error")
	}
}

// TestGenerateFloatBounds: a number (float64) field keeps its JSON bound and enum
// spelling verbatim — a fractional value is valid Go and must not be rejected.
func TestGenerateFloatBounds(t *testing.T) {
	s := `{"type":"object","properties":{"r":{"type":"number","minimum":1.5,"enum":[1.5,2.5,3]}}}`
	out, err := Generate([]byte(s), "m", "T")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{".Min(1.5)", ".Enum(1.5, 2.5, 3)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
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

// TestGenerateNullable30 checks OpenAPI 3.0's "nullable" keyword: a nullable field
// omits .NotNull(), while a non-null field keeps it.
func TestGenerateNullable30(t *testing.T) {
	s := `{"type":"object","required":["a"],"properties":{
		"a":{"type":"string"},
		"b":{"type":"string","nullable":true,"minLength":1}
	}}`
	out, err := Generate([]byte(s), "m", "T")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, `SubPath(base, "a")...).Required().NotNull()`) {
		t.Errorf("non-null required field a should be Required().NotNull():\n%s", got)
	}
	if strings.Contains(got, `SubPath(base, "b")...).NotNull()`) {
		t.Errorf("nullable field b should not have NotNull:\n%s", got)
	}
	if !strings.Contains(got, `SubPath(base, "b")...).MinLength(1)`) {
		t.Errorf("nullable field b should still carry its constraints:\n%s", got)
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
	// unconstrained scalar arrays map to the lean specialized types
	for _, want := range []string{
		"types.StringArray",
		"types.Int64Array",
		"types.Float64Array",
		"types.BooleanArray",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

// TestGenerateLeanArrays: an unconstrained scalar array uses the lean specialized
// type + ScalarArray validator, while a constrained one stays generic.
func TestGenerateLeanArrays(t *testing.T) {
	s := `{"type":"object","properties":{
		"tags":{"type":"array","items":{"type":"string"},"uniqueItems":true},
		"nums":{"type":"array","items":{"type":"integer"}},
		"emails":{"type":"array","items":{"type":"string","format":"email"}}
	}}`
	out, err := Generate([]byte(s), "m", "T")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"types.StringArray", "*validation.ScalarArray[string]",
		"types.Int64Array", "*validation.ScalarArray[int64]",
		"types.Array[types.String, *types.String]", // emails stays generic (format constraint)
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
