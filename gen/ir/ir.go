// Package ir is the generator's resolved model: schema nodes mapped to concrete
// Go types, validator chains and result types, decoupled from both the input
// schema and the output templates.
package ir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/binadel/esdigo/gen/schema"
)

// File is a generated Go source file.
type File struct {
	Package  string
	Imports  []string
	Messages []*Message
}

// Message is one generated struct (and its validator).
type Message struct {
	Name   string
	Doc    string
	Fields []*Field
}

// Field is one property of a Message.
type Field struct {
	GoName    string // exported Go field name
	JSONName  string // the JSON property name
	Doc       string
	ModelType string // the types.* wrapper, e.g. "types.String"

	// Validation. Validate is false for a field with no constraints at all.
	Validate      bool
	ByPointer     bool   // number-family validators take &field, others take it by value
	ValidatorType string // e.g. "*validation.String"
	ResultType    string // the Result[T] payload type, e.g. "string"
	NewExpr       string // the fluent constructor, e.g. validation.NewString("x").Required()
}

// HasValidation reports whether any field of the message is validated.
func (m *Message) HasValidation() bool {
	for _, f := range m.Fields {
		if f.Validate {
			return true
		}
	}
	return false
}

// Build resolves a root object schema into a generated file with one message.
func Build(pkg, name string, root *schema.Schema) (*File, error) {
	if root.Type.Primary() != "object" {
		return nil, fmt.Errorf("root schema must be an object, got %q", root.Type.Primary())
	}

	msg := &Message{Name: goName(name), Doc: doc(root)}
	for _, jsonName := range sortedKeys(root.Properties) {
		field, err := buildField(jsonName, root.Properties[jsonName], root.IsRequired(jsonName))
		if err != nil {
			return nil, err
		}
		msg.Fields = append(msg.Fields, field)
	}

	return &File{
		Package:  pkg,
		Imports:  imports(msg.HasValidation()),
		Messages: []*Message{msg},
	}, nil
}

func buildField(jsonName string, s *schema.Schema, required bool) (*Field, error) {
	notNull := !s.Type.Nullable()
	f := &Field{GoName: goName(jsonName), JSONName: jsonName, Doc: doc(s)}

	switch kind := s.Type.Primary(); kind {
	case "string":
		f.ModelType = "types.String"
		f.ValidatorType = "*validation.String"
		f.ResultType = "string"
		f.NewExpr = stringExpr(jsonName, required, notNull, s)
	case "integer":
		f.ModelType = "types.Int64"
		f.ByPointer = true
		f.ValidatorType = "*validation.Number[int64]"
		f.ResultType = "int64"
		f.NewExpr = numberExpr("int64", jsonName, required, notNull, s)
	case "number":
		f.ModelType = "types.Float64"
		f.ByPointer = true
		f.ValidatorType = "*validation.Number[float64]"
		f.ResultType = "float64"
		f.NewExpr = numberExpr("float64", jsonName, required, notNull, s)
	case "boolean":
		f.ModelType = "types.Boolean"
		f.ValidatorType = "*validation.Boolean"
		f.ResultType = "bool"
		f.NewExpr = booleanExpr(jsonName, required, notNull, s)
	default:
		return nil, fmt.Errorf("unsupported type %q for property %q", kind, jsonName)
	}

	f.Validate = required || notNull || hasValueConstraint(s)
	return f, nil
}

func hasValueConstraint(s *schema.Schema) bool {
	return s.MinLength != nil || s.MaxLength != nil || s.Pattern != "" || s.Format != "" ||
		len(s.Enum) > 0 || s.Const != nil ||
		s.Minimum != nil || s.Maximum != nil || s.ExclusiveMinimum != nil ||
		s.ExclusiveMaximum != nil || s.MultipleOf != nil
}

func stringExpr(jsonName string, required, notNull bool, s *schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewString(%q)", jsonName)
	writePresence(&b, required, notNull)
	if s.MinLength != nil {
		fmt.Fprintf(&b, ".MinLength(%d)", *s.MinLength)
	}
	if s.MaxLength != nil {
		fmt.Fprintf(&b, ".MaxLength(%d)", *s.MaxLength)
	}
	if s.Pattern != "" {
		fmt.Fprintf(&b, ".Pattern(%q)", s.Pattern)
	}
	if len(s.Enum) > 0 {
		fmt.Fprintf(&b, ".Enum(%s)", joinRaw(s.Enum))
	}
	if s.Const != nil {
		fmt.Fprintf(&b, ".Const(%s)", raw(s.Const))
	}
	return b.String()
}

func numberExpr(goType, jsonName string, required, notNull bool, s *schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewNumber[%s](%q)", goType, jsonName)
	writePresence(&b, required, notNull)
	if s.Minimum != nil {
		fmt.Fprintf(&b, ".Min(%s)", raw(s.Minimum))
	}
	if s.Maximum != nil {
		fmt.Fprintf(&b, ".Max(%s)", raw(s.Maximum))
	}
	if s.ExclusiveMinimum != nil {
		fmt.Fprintf(&b, ".ExclusiveMin(%s)", raw(s.ExclusiveMinimum))
	}
	if s.ExclusiveMaximum != nil {
		fmt.Fprintf(&b, ".ExclusiveMax(%s)", raw(s.ExclusiveMaximum))
	}
	if s.MultipleOf != nil {
		fmt.Fprintf(&b, ".MultipleOf(%s)", raw(s.MultipleOf))
	}
	if len(s.Enum) > 0 {
		fmt.Fprintf(&b, ".Enum(%s)", joinRaw(s.Enum))
	}
	if s.Const != nil {
		fmt.Fprintf(&b, ".Const(%s)", raw(s.Const))
	}
	return b.String()
}

func booleanExpr(jsonName string, required, notNull bool, s *schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewBoolean(%q)", jsonName)
	writePresence(&b, required, notNull)
	if s.Const != nil {
		fmt.Fprintf(&b, ".Const(%s)", raw(s.Const))
	}
	return b.String()
}

func writePresence(b *strings.Builder, required, notNull bool) {
	if required {
		b.WriteString(".Required()")
	}
	if notNull {
		b.WriteString(".NotNull()")
	}
}

// raw renders a JSON literal as a Go literal — for the value types we emit today
// (numbers, quoted strings, booleans) the JSON and Go spellings coincide.
func raw(m []byte) string {
	return string(bytes.TrimSpace(m))
}

func joinRaw(values []json.RawMessage) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = raw(v)
	}
	return strings.Join(parts, ", ")
}

func doc(s *schema.Schema) string {
	if s.Title != "" {
		return s.Title
	}
	return s.Description
}

func imports(needValidation bool) []string {
	imps := []string{
		"github.com/binadel/esdigo/json",
		"github.com/binadel/esdigo/json/types",
	}
	if needValidation {
		imps = append(imps, "github.com/binadel/esdigo/validation")
	}
	return imps
}

func sortedKeys(m map[string]*schema.Schema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// goName turns a JSON property or schema name into an exported Go identifier:
// separators (_, -, ., space) split words and each word is capitalized.
func goName(s string) string {
	var b strings.Builder
	upNext := true
	for _, r := range s {
		if r == '_' || r == '-' || r == '.' || r == ' ' {
			upNext = true
			continue
		}
		if upNext {
			b.WriteRune(unicode.ToUpper(r))
			upNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
