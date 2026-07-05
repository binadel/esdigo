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

	// Validation. Validate is false for a scalar field with no constraints at all;
	// object fields always validate (they recurse into the child validator).
	Validate      bool
	ByPointer     bool     // number-family validators take &field, others take it by value
	ValidatorType string   // scalar: e.g. "*validation.String"
	ResultType    string   // scalar: the Result[T] payload type, e.g. "string"
	NewExpr       string   // scalar: the fluent constructor
	Imports       []string // extra imports this field's result type needs

	// Object fields (types.Object[Child,*Child]) recurse into the child validator.
	IsObject      bool
	ChildType     string // the child message name, e.g. "OrderCustomer"
	ChildNewExpr  string // the child validator constructor, e.g. NewOrderCustomerValidator(...)
	ObjectNewExpr string // the object-level (presence/null/type) validator constructor
}

// formatInfo maps a JSON Schema string "format" to the format-specific validator:
// the trailing builder call, the resulting validator and Result types, and the
// extra imports the Result type pulls in. The model type stays types.String.
type formatInfo struct {
	method        string
	validatorType string
	resultType    string
	imports       []string
}

var formats = map[string]formatInfo{
	"email":         {".Email()", "*validation.Email", "*mail.Address", []string{"net/mail"}},
	"ipv4":          {".IP().Version4()", "*validation.IP", "net.IP", []string{"net"}},
	"ipv6":          {".IP().Version6()", "*validation.IP", "net.IP", []string{"net"}},
	"uri":           {".Uri()", "*validation.Uri", "*url.URL", []string{"net/url"}},
	"uri-reference": {".UriReference()", "*validation.Uri", "*url.URL", []string{"net/url"}},
	"uuid":          {".Uuid()", "*validation.Uuid", "uuid.UUID", []string{"github.com/google/uuid"}},
	"date":          {".Date()", "*validation.Time", "time.Time", []string{"time"}},
	"time":          {".Time()", "*validation.Time", "time.Time", []string{"time"}},
	"date-time":     {".DateTime()", "*validation.Time", "time.Time", []string{"time"}},
	"duration":      {".Duration()", "*validation.Duration", "time.Duration", []string{"time"}},
	"regex":         {".Regex()", "*validation.Regex", "*regexp.Regexp", []string{"regexp"}},
	"hostname":      {".Hostname()", "*validation.Hostname", "string", nil},
	"json-pointer":  {".JsonPointer()", "*validation.JsonPointer", "string", nil},
}

// formatKnown reports whether name is a format the generator maps to a validator.
// Unknown formats are ignored (JSON Schema treats format as an annotation).
func formatKnown(name string) bool {
	_, ok := formats[name]
	return ok
}

// builder accumulates the messages (root, inline nested objects and $defs) of one
// generated file, resolving $ref against the merged definitions.
type builder struct {
	defs     map[string]*schema.Schema
	messages []*Message
	built    map[string]bool // message names already produced (dedup)
	imports  map[string]bool // extra result-type imports
}

// Build resolves a root object schema (plus its $defs and inline nested objects)
// into a generated file. Each object becomes its own message.
func Build(pkg, name string, root *schema.Schema) (*File, error) {
	if root.Type.Primary() != "object" {
		return nil, fmt.Errorf("root schema must be an object, got %q", root.Type.Primary())
	}

	b := &builder{
		defs:    root.AllDefs(),
		built:   map[string]bool{},
		imports: map[string]bool{},
	}

	if err := b.buildMessage(goName(name), root); err != nil {
		return nil, err
	}
	// Emit every object definition too, referenced or not.
	for _, key := range sortedKeys(b.defs) {
		if b.defs[key].Type.Primary() == "object" {
			if err := b.buildMessage(goName(key), b.defs[key]); err != nil {
				return nil, err
			}
		}
	}

	sort.Slice(b.messages, func(i, j int) bool { return b.messages[i].Name < b.messages[j].Name })

	return &File{
		Package:  pkg,
		Imports:  buildImports(b.imports),
		Messages: b.messages,
	}, nil
}

// buildMessage produces the message named goName for object schema s (once).
func (b *builder) buildMessage(name string, s *schema.Schema) error {
	if b.built[name] {
		return nil
	}
	b.built[name] = true

	msg := &Message{Name: name, Doc: doc(s)}
	for _, jsonName := range sortedKeys(s.Properties) {
		field, err := b.buildField(name, jsonName, s.Properties[jsonName], s.IsRequired(jsonName))
		if err != nil {
			return err
		}
		for _, imp := range field.Imports {
			b.imports[imp] = true
		}
		msg.Fields = append(msg.Fields, field)
	}
	b.messages = append(b.messages, msg)
	return nil
}

// buildField resolves one property. Objects and $refs become types.Object fields
// (registering / referencing the child message); everything else is a scalar.
func (b *builder) buildField(msgName, jsonName string, s *schema.Schema, required bool) (*Field, error) {
	if s.Ref != "" {
		return b.buildRefField(msgName, jsonName, s, required)
	}

	switch kind := s.Type.Primary(); kind {
	case "object":
		child := msgName + goName(jsonName)
		if err := b.buildMessage(child, s); err != nil {
			return nil, err
		}
		return objectField(jsonName, child, required, !s.Type.Nullable()), nil
	case "string", "integer", "number", "boolean":
		return buildScalarField(jsonName, s, required), nil
	default:
		return nil, fmt.Errorf("unsupported type %q for property %q", kind, jsonName)
	}
}

// buildRefField resolves a $ref: an object target becomes a shared reference to
// that definition's message; a scalar target is inlined as the field's schema.
func (b *builder) buildRefField(msgName, jsonName string, s *schema.Schema, required bool) (*Field, error) {
	key := refName(s.Ref)
	target, ok := b.defs[key]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q for property %q", s.Ref, jsonName)
	}
	if target.Type.Primary() == "object" {
		child := goName(key)
		if err := b.buildMessage(child, target); err != nil {
			return nil, err
		}
		// A bare $ref has no type (notNull); a sibling "type" may relax it to
		// nullable, e.g. {"type":["object","null"],"$ref":...}.
		return objectField(jsonName, child, required, !s.Type.Nullable()), nil
	}
	return b.buildField(msgName, jsonName, target, required)
}

// objectField builds a types.Object[Child,*Child] field. It always validates: the
// object-level validator checks presence/null/type, and the generated Validate
// recurses into the child validator to check the child's own fields.
func objectField(jsonName, child string, required, notNull bool) *Field {
	f := &Field{
		GoName:    goName(jsonName),
		JSONName:  jsonName,
		ModelType: fmt.Sprintf("types.Object[%s, *%s]", child, child),
		IsObject:  true,
		ChildType: child,
		Validate:  true,
	}

	var obj strings.Builder
	fmt.Fprintf(&obj, "validation.NewObject[%s, *%s](%s)", child, child, subPath(jsonName))
	writePresence(&obj, required, notNull)
	f.ObjectNewExpr = obj.String()

	f.ChildNewExpr = fmt.Sprintf("New%sValidator(%s)", child, subPath(jsonName))
	return f
}

// subPath renders the path argument for a field validator: the base slice threaded
// from the parent, with this field's JSON name appended.
func subPath(jsonName string) string {
	return fmt.Sprintf("validation.SubPath(base, %q)...", jsonName)
}

func buildScalarField(jsonName string, s *schema.Schema, required bool) *Field {
	notNull := !s.Type.Nullable()
	f := &Field{GoName: goName(jsonName), JSONName: jsonName, Doc: doc(s)}

	switch s.Type.Primary() {
	case "string":
		f.ModelType = "types.String"
		f.ValidatorType = "*validation.String"
		f.ResultType = "string"
		f.NewExpr = stringExpr(jsonName, required, notNull, s)
		// A recognized format switches to its specific validator; the base String
		// constraints already applied above still run before the format check.
		if fi, ok := formats[s.Format]; ok {
			f.ValidatorType = fi.validatorType
			f.ResultType = fi.resultType
			f.NewExpr += fi.method
			f.Imports = fi.imports
		}
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
	}

	f.Validate = required || notNull || hasValueConstraint(s)
	return f
}

// refName returns the last path segment of a JSON pointer $ref, e.g. "Address"
// for "#/$defs/Address".
func refName(ref string) string {
	if i := strings.LastIndexByte(ref, '/'); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

func hasValueConstraint(s *schema.Schema) bool {
	return s.MinLength != nil || s.MaxLength != nil || s.Pattern != "" || formatKnown(s.Format) ||
		len(s.Enum) > 0 || s.Const != nil ||
		s.Minimum != nil || s.Maximum != nil || s.ExclusiveMinimum != nil ||
		s.ExclusiveMaximum != nil || s.MultipleOf != nil
}

func stringExpr(jsonName string, required, notNull bool, s *schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewString(%s)", subPath(jsonName))
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
	fmt.Fprintf(&b, "validation.NewNumber[%s](%s)", goType, subPath(jsonName))
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
	fmt.Fprintf(&b, "validation.NewBoolean(%s)", subPath(jsonName))
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

func buildImports(extra map[string]bool) []string {
	// Every generated message carries a validator (with an object-level Result), so
	// validation is always imported alongside json and types.
	set := map[string]bool{
		"github.com/binadel/esdigo/json":       true,
		"github.com/binadel/esdigo/json/types": true,
		"github.com/binadel/esdigo/validation": true,
	}
	for imp := range extra {
		set[imp] = true
	}
	list := make([]string, 0, len(set))
	for imp := range set {
		list = append(list, imp)
	}
	sort.Strings(list)
	return list
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
