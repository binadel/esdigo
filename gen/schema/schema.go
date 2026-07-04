// Package schema is the generator's input model: a subset of JSON Schema
// 2020-12 sufficient to drive esdigo code generation. It is parsed with
// encoding/json — the generator is a build-time tool, so its own input parsing
// is exempt from the library's zero-reflection rule.
package schema

import "encoding/json"

// Schema is one JSON Schema node. Numeric bounds and enum/const values are kept
// as raw JSON so their exact literal text can be emitted verbatim, without going
// through a lossy float64.
type Schema struct {
	Type        TypeSet `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`

	// object
	Properties map[string]*Schema `json:"properties"`
	Required   []string           `json:"required"`

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

// Parse unmarshals a JSON Schema document.
func Parse(data []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
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
