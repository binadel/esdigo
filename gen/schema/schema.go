// Package schema is the generator's input model: a subset of JSON Schema
// 2020-12 sufficient to drive esdigo code generation. It is parsed with
// encoding/json — the generator is a build-time tool, so its own input parsing
// is exempt from the library's zero-reflection rule. YAML input is accepted too
// (OpenAPI specs are usually YAML): it is converted to JSON at the boundary so
// the encoding/json path below, with its raw-literal and TypeSet handling, is
// the single source of truth.
package schema

import (
	"bytes"
	"encoding/json"
	"sort"

	"gopkg.in/yaml.v3"
)

// toJSON normalizes an input document to JSON. Bytes that are already valid JSON
// pass through unchanged — a lossless fast path that keeps existing JSON inputs
// byte-identical. Otherwise the document is treated as YAML (a JSON superset) and
// converted through its node tree, which preserves mapping key order (decoding into
// map[string]any then json.Marshal would sort keys and lose field order).
func toJSON(data []byte) ([]byte, error) {
	if json.Valid(data) {
		return data, nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return yamlNodeToJSON(&node)
}

// yamlNodeToJSON renders a yaml.Node as JSON, preserving mapping key order so a
// schema's property order survives the YAML→JSON boundary. Scalars go through the
// node's own decode (yaml.v3 resolves numbers to int/float64, bools, and null), then
// json.Marshal, so bounds and literals render exactly as the map path did before.
func yamlNodeToJSON(n *yaml.Node) ([]byte, error) {
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return []byte("null"), nil
		}
		return yamlNodeToJSON(n.Content[0])
	case yaml.MappingNode:
		var b bytes.Buffer
		b.WriteByte('{')
		for i := 0; i+1 < len(n.Content); i += 2 {
			if i > 0 {
				b.WriteByte(',')
			}
			key, err := json.Marshal(n.Content[i].Value) // object keys are JSON strings
			if err != nil {
				return nil, err
			}
			val, err := yamlNodeToJSON(n.Content[i+1])
			if err != nil {
				return nil, err
			}
			b.Write(key)
			b.WriteByte(':')
			b.Write(val)
		}
		b.WriteByte('}')
		return b.Bytes(), nil
	case yaml.SequenceNode:
		var b bytes.Buffer
		b.WriteByte('[')
		for i, c := range n.Content {
			if i > 0 {
				b.WriteByte(',')
			}
			v, err := yamlNodeToJSON(c)
			if err != nil {
				return nil, err
			}
			b.Write(v)
		}
		b.WriteByte(']')
		return b.Bytes(), nil
	case yaml.AliasNode:
		return yamlNodeToJSON(n.Alias)
	case yaml.ScalarNode:
		var v any
		if err := n.Decode(&v); err != nil {
			return nil, err
		}
		return json.Marshal(v)
	default:
		return []byte("null"), nil
	}
}

// Schema is one JSON Schema node. Numeric bounds and enum/const values are kept
// as raw JSON so their exact literal text can be emitted verbatim, without going
// through a lossy float64.
type Schema struct {
	Type        TypeSet `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`

	// x-esdigo-validate opts an (object) type out of validator generation when
	// explicitly false — for a value the program only produces (e.g. a response) and
	// never validates. Absent means validate (the default); the model (read + write)
	// is always generated regardless.
	Validate *bool `json:"x-esdigo-validate"`

	// nullability: OpenAPI 3.0 uses this keyword; JSON Schema 2020-12 / OpenAPI 3.1
	// use type: [..., "null"] instead. Either marks the value nullable.
	Nullable bool `json:"nullable"`

	// reference
	Ref string `json:"$ref"`

	// object
	Properties    map[string]*Schema `json:"properties"`
	Required      []string           `json:"required"`
	MinProperties *int               `json:"minProperties"`
	MaxProperties *int               `json:"maxProperties"`

	// propertyOrder records the order the "properties" keys appeared in the source
	// (a plain map loses it), captured by UnmarshalJSON so generated struct fields
	// follow the schema's field order. Read it through PropertyNames.
	propertyOrder []string

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

// UnmarshalJSON decodes a Schema and additionally records the source order of its
// "properties" keys — the map alone loses it — so generated struct fields follow the
// schema's field order. The alias type carries the standard field decoding without
// recursing back into this method.
func (s *Schema) UnmarshalJSON(data []byte) error {
	type alias Schema
	if err := json.Unmarshal(data, (*alias)(s)); err != nil {
		return err
	}
	var probe struct {
		Properties json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	order, err := objectKeyOrder(probe.Properties)
	if err != nil {
		return err
	}
	s.propertyOrder = order
	return nil
}

// PropertyNames returns the property names in source order. It falls back to a stable
// alphabetical sort when the order was not captured — e.g. a hand-built Schema value,
// or a malformed object with duplicate keys — so callers always get every property.
func (s *Schema) PropertyNames() []string {
	if len(s.propertyOrder) == len(s.Properties) {
		return s.propertyOrder
	}
	names := make([]string, 0, len(s.Properties))
	for name := range s.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// objectKeyOrder returns the keys of a JSON object literal in source order, reading
// the token stream directly. A null or absent value (or a non-object) yields no keys.
func objectKeyOrder(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if open, ok := tok.(json.Delim); !ok || open != '{' {
		return nil, nil
	}
	var order []string
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return nil, err
		}
		order = append(order, key.(string))
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
	}
	return order, nil
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
