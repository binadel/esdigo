package schema

import "testing"

// TestToJSONPassThrough: valid JSON is returned byte-identical (no lossy YAML
// round-trip), so existing JSON inputs and goldens are unaffected.
func TestToJSONPassThrough(t *testing.T) {
	in := []byte(`{"type":"object","minimum":1e3,"enum":["a","b"]}`)
	out, err := toJSON(in)
	if err != nil {
		t.Fatalf("toJSON: %v", err)
	}
	if string(out) != string(in) {
		t.Errorf("valid JSON should pass through unchanged:\n got %s\nwant %s", out, in)
	}
}

// TestToJSONFromYAML converts a YAML document to equivalent JSON.
func TestToJSONFromYAML(t *testing.T) {
	yaml := []byte("type: object\nrequired:\n  - id\nproperties:\n  id:\n    type: string\n")
	out, err := toJSON(yaml)
	if err != nil {
		t.Fatalf("toJSON: %v", err)
	}
	// The converted JSON must parse into the same Schema an equivalent JSON doc would.
	got, err := Parse(out)
	if err != nil {
		t.Fatalf("parse converted: %v", err)
	}
	if got.Type.Primary() != "object" {
		t.Errorf("type = %q, want object", got.Type.Primary())
	}
	if len(got.Required) != 1 || got.Required[0] != "id" {
		t.Errorf("required = %v, want [id]", got.Required)
	}
	if _, ok := got.Properties["id"]; !ok {
		t.Errorf("missing property id: %+v", got.Properties)
	}
}

// TestParseYAML: Parse accepts YAML directly and preserves a numeric bound's
// literal (yaml.v3 decodes an integer as int, not a lossy float).
func TestParseYAML(t *testing.T) {
	yaml := []byte("type: object\nproperties:\n  n:\n    type: integer\n    minimum: 5\n")
	s, err := Parse(yaml)
	if err != nil {
		t.Fatalf("parse yaml: %v", err)
	}
	if got := string(s.Properties["n"].Minimum); got != "5" {
		t.Errorf("minimum = %q, want \"5\"", got)
	}
}

// TestIsOpenAPIYAML detects an OpenAPI document authored in YAML.
func TestIsOpenAPIYAML(t *testing.T) {
	yaml := []byte("openapi: 3.1.0\ncomponents:\n  schemas:\n    A:\n      type: object\n")
	if !IsOpenAPI(yaml) {
		t.Errorf("should detect YAML OpenAPI document")
	}
}
