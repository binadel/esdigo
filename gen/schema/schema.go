// Package schema is the generator's input model: a subset of JSON Schema
// 2020-12 sufficient to drive esdigo code generation. It is parsed with
// encoding/json — the generator is a build-time tool, so its own input parsing
// is exempt from the library's zero-reflection rule. YAML input is accepted too
// (OpenAPI specs are usually YAML): it is converted to JSON at the boundary so
// the encoding/json path below, with its raw-literal and TypeSet handling, is
// the single source of truth.
package schema

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// toJSON normalizes an input document to JSON. Bytes that are already valid JSON
// pass through unchanged — a lossless fast path that keeps existing JSON inputs
// byte-identical. Otherwise the document is treated as YAML (a JSON superset) and
// converted: yaml.v3 decodes mappings to map[string]any and numbers to int/float64,
// which json.Marshal renders back to faithful literals for the raw-JSON bounds.
func toJSON(data []byte) ([]byte, error) {
	if json.Valid(data) {
		return data, nil
	}
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}

// Schema is one JSON Schema node. Numeric bounds and enum/const values are kept
// as raw JSON so their exact literal text can be emitted verbatim, without going
// through a lossy float64.
type Schema struct {
	Type        TypeSet `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`

	// x-esdigo-io flags which generated code this (object) type needs: "in"
	// (reader + validators), "out" (marshal + writer, no validators), or "" / "both"
	// (everything). See gen/ir.Direction.
	IO string `json:"x-esdigo-io"`

	// nullability: OpenAPI 3.0 uses this keyword; JSON Schema 2020-12 / OpenAPI 3.1
	// use type: [..., "null"] instead. Either marks the value nullable.
	Nullable bool `json:"nullable"`

	// reference
	Ref string `json:"$ref"`

	// object
	Properties map[string]*Schema `json:"properties"`
	Required   []string           `json:"required"`

	// composition: allOf merges every subschema's properties and required list into
	// this one (JSON Schema intersection; OpenAPI uses it for object inheritance).
	AllOf []*Schema `json:"allOf"`

	// other composition keywords. oneOf/anyOf are unions, if/then/else is
	// conditional, and not is negation. They are parsed so the generator can reject
	// the ones it does not yet model with a clear error, rather than silently
	// dropping the constraint and emitting a struct that ignores it.
	OneOf []*Schema `json:"oneOf"`
	AnyOf []*Schema `json:"anyOf"`
	Not   *Schema   `json:"not"`
	If    *Schema   `json:"if"`
	Then  *Schema   `json:"then"`
	Else  *Schema   `json:"else"`

	// discriminator (OpenAPI): names the property whose value selects a oneOf/anyOf
	// variant, with an optional value->schema mapping. Its presence turns a
	// oneOf/anyOf into a generated tagged-union type.
	Discriminator *Discriminator `json:"discriminator"`

	// named subschemas: "$defs" (2020-12 / OpenAPI 3.1) or "definitions" (draft-07)
	Defs        map[string]*Schema `json:"$defs"`
	Definitions map[string]*Schema `json:"definitions"`

	// string
	MinLength *int   `json:"minLength"`
	MaxLength *int   `json:"maxLength"`
	Pattern   string `json:"pattern"`
	Format    string `json:"format"`

	// number / integer
	Minimum          json.RawMessage `json:"minimum"`
	Maximum          json.RawMessage `json:"maximum"`
	ExclusiveMinimum json.RawMessage `json:"exclusiveMinimum"`
	ExclusiveMaximum json.RawMessage `json:"exclusiveMaximum"`
	MultipleOf       json.RawMessage `json:"multipleOf"`

	// array
	MinItems    *int    `json:"minItems"`
	MaxItems    *int    `json:"maxItems"`
	UniqueItems bool    `json:"uniqueItems"`
	Items       *Schema `json:"items"`

	// any
	Enum  []json.RawMessage `json:"enum"`
	Const json.RawMessage   `json:"const"`
}

// Discriminator is the OpenAPI discriminator object: the property that carries the
// variant tag, and an optional map from tag value to the variant schema ($ref or a
// bare schema name).
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping"`
}

// Parse unmarshals a JSON (or YAML) Schema document.
func Parse(data []byte) (*Schema, error) {
	data, err := toJSON(data)
	if err != nil {
		return nil, err
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// OpenAPIDoc is the slice of an OpenAPI 3.1 document the generator reads: its
// component schemas (which use the JSON Schema 2020-12 dialect).
type OpenAPIDoc struct {
	OpenAPI    string `json:"openapi"`
	Components struct {
		Schemas map[string]*Schema `json:"schemas"`
	} `json:"components"`
}

// ParseOpenAPI unmarshals an OpenAPI document (JSON or YAML).
func ParseOpenAPI(data []byte) (*OpenAPIDoc, error) {
	data, err := toJSON(data)
	if err != nil {
		return nil, err
	}
	var doc OpenAPIDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// IsOpenAPI reports whether data looks like an OpenAPI document rather than a bare
// JSON Schema — it has an "openapi" version or a components.schemas map.
func IsOpenAPI(data []byte) bool {
	doc, err := ParseOpenAPI(data)
	if err != nil {
		return false
	}
	return doc.OpenAPI != "" || len(doc.Components.Schemas) > 0
}

// AllDefs merges the named subschemas from "$defs" and "definitions", with "$defs"
// taking precedence on a key collision.
func (s *Schema) AllDefs() map[string]*Schema {
	if len(s.Defs) == 0 && len(s.Definitions) == 0 {
		return nil
	}
	all := make(map[string]*Schema, len(s.Defs)+len(s.Definitions))
	for k, v := range s.Definitions {
		all[k] = v
	}
	for k, v := range s.Defs {
		all[k] = v
	}
	return all
}

// IsNullable reports whether null is an allowed value — via the OpenAPI 3.0
// "nullable" keyword or a "null" entry in the type set (JSON Schema / OpenAPI 3.1).
func (s *Schema) IsNullable() bool {
	return s.Nullable || s.Type.Nullable()
}

// IsRequired reports whether name is in this (object) schema's required list.
func (s *Schema) IsRequired(name string) bool {
	for _, r := range s.Required {
		if r == name {
			return true
		}
	}
	return false
}

// TypeSet is a JSON Schema "type" — either a single type ("string") or a set
// (["string","null"]). It normalizes both forms to a slice.
type TypeSet []string

func (t *TypeSet) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*t = arr
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = TypeSet{s}
	return nil
}

// Primary returns the first non-null type, e.g. "string" for ["string","null"].
func (t TypeSet) Primary() string {
	for _, name := range t {
		if name != "null" {
			return name
		}
	}
	return ""
}

// Nullable reports whether "null" is one of the allowed types.
func (t TypeSet) Nullable() bool {
	for _, name := range t {
		if name == "null" {
			return true
		}
	}
	return false
}
