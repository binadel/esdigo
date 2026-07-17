// Package ir is the generator's resolved model: schema nodes mapped to concrete
// Go types, validator chains and result types, decoupled from both the input
// schema and the output templates.
package ir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
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
	Name      string
	Doc       string
	Fields    []*Field
	Validated bool     // whether to emit validators (x-esdigo-validate; false = produce-only)
	Imports   []string // the import set when this message is emitted in its own file

	// object-level property-count constraints (JSON Schema minProperties/maxProperties),
	// checked against the number of present fields.
	MinProperties *int
	MaxProperties *int

	// Union, when non-nil, makes this message a discriminated union (a tagged
	// type selected by a property value) instead of a struct: it has Variants
	// rather than Fields. A parent still references it as an ordinary object field.
	Union *Union
}

// Union describes a discriminated oneOf/anyOf: the property whose value selects a
// variant, and the variants (each a generated object message) keyed by tag value.
type Union struct {
	Discriminator string    // the JSON property carrying the variant tag
	Variants      []Variant // sorted by tag for deterministic output
}

// Variant is one arm of a Union: a discriminator value and the object message it
// decodes to.
type Variant struct {
	Tag    string // the discriminator value that selects this variant
	GoType string // the variant's generated message name
}

// validatedOf reports whether a type should get validators, from its
// x-esdigo-validate flag. Absent (nil) means yes (the default); only an explicit
// false — a produce-only value the program never validates — turns them off. The
// model (marshal/write and unmarshal/read) is always generated in either case, so a
// type stays usable as a nested field regardless of this flag.
func validatedOf(s *schema.Schema) bool {
	return s.Validate == nil || *s.Validate
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

	// Per-element validation for array fields: the generated loop builds a fresh
	// element validator at the indexed path for each element (so failures carry
	// their element index), and results collect into a <Field>Items slice.
	ElemValidate  bool
	ElemIsObject  bool   // object element (recurse) vs scalar element
	ElemNewExpr   string // per-element validator constructor (uses the loop's path var `p`)
	ElemItemsType string // the <Field>Items slice element type
	ElemByValue   bool   // scalar element: Validate takes the dereferenced *elem
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
}

// Build resolves a root object schema (plus its $defs and inline nested objects)
// into a generated file. Each object becomes its own message.
func Build(pkg, name string, root *schema.Schema) (*File, error) {
	defs, err := normalizeDefs(root.AllDefs())
	if err != nil {
		return nil, err
	}
	b := &builder{defs: defs, built: map[string]bool{}}

	validated := validatedOf(root)
	switch {
	case isUnionSchema(root):
		if err := b.buildUnion(goName(name), root, validated); err != nil {
			return nil, err
		}
	case isObjectSchema(root):
		if err := b.buildMessage(goName(name), root, validated); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("root schema must be an object or a discriminated oneOf, got %q", root.Type.Primary())
	}
	// Emit every definition too, referenced or not.
	for _, key := range sortedKeys(b.defs) {
		if err := b.buildNamed(key, b.defs[key]); err != nil {
			return nil, err
		}
	}

	return b.file(pkg), nil
}

// BuildAll resolves a set of named schemas (e.g. OpenAPI components) into a
// generated file: a message for every object schema, with $ref resolved against
// the whole set. Non-object schemas become messages only where a ref inlines them.
func BuildAll(pkg string, schemas map[string]*schema.Schema) (*File, error) {
	defs, err := normalizeDefs(schemas)
	if err != nil {
		return nil, err
	}
	b := &builder{defs: defs, built: map[string]bool{}}

	for _, key := range sortedKeys(b.defs) {
		if err := b.buildNamed(key, b.defs[key]); err != nil {
			return nil, err
		}
	}

	return b.file(pkg), nil
}

// buildNamed builds the message for one named schema (a $def or component): a
// discriminated union, an object, or nothing (e.g. a scalar def, which becomes a
// message only where a $ref inlines it).
func (b *builder) buildNamed(key string, s *schema.Schema) error {
	if isUnionSchema(s) {
		return b.buildUnion(key, s, validatedOf(s))
	}
	if err := checkComposition(key, s); err != nil {
		return err
	}
	if !isObjectSchema(s) {
		return nil
	}
	return b.buildMessage(key, s, validatedOf(s))
}

// file assembles the generated File: messages sorted by name (Go allows forward
// references), the aggregate import set for the combined output, and each
// message's own import set for per-file (split) output.
func (b *builder) file(pkg string) *File {
	sort.Slice(b.messages, func(i, j int) bool { return b.messages[i].Name < b.messages[j].Name })
	fileImports := map[string]bool{}
	for _, m := range b.messages {
		m.Imports = messageImports(m)
		for _, imp := range m.Imports {
			fileImports[imp] = true
		}
	}
	return &File{
		Package:  pkg,
		Imports:  sortedImports(fileImports),
		Messages: b.messages,
	}
}

// isObjectSchema reports whether s should produce (or reference) a generated struct:
// an explicit type:object, or an allOf composition with no conflicting scalar type.
func isObjectSchema(s *schema.Schema) bool {
	if s.Type.Primary() == "object" {
		return true
	}
	return s.Type.Primary() == "" && len(s.AllOf) > 0
}

// checkComposition rejects the composition keywords the generator does not model,
// so an unsupported oneOf/anyOf/if/not is a clear generation error instead of being
// silently dropped (which would emit a struct that ignores the constraint). allOf is
// handled by mergedObject, and a discriminated union or the [X, null] nullable idiom
// are handled by the callers before this point; both are intentionally allowed here.
func checkComposition(name string, s *schema.Schema) error {
	switch {
	case s.Not != nil:
		return fmt.Errorf("schema %q uses not, which is not supported", name)
	case s.If != nil:
		return fmt.Errorf("schema %q uses if/then/else, which is not yet supported", name)
	case len(s.OneOf) > 0 || len(s.AnyOf) > 0:
		if isUnionSchema(s) || nullableIdiomOf(s) != nil {
			return nil
		}
		return fmt.Errorf("schema %q uses oneOf/anyOf that is neither a discriminated union nor the [X, null] nullable form, which is not supported", name)
	}
	return nil
}

// unionMembers returns a schema's oneOf list, or its anyOf list, or nil. oneOf and
// anyOf are treated the same for codegen: with a discriminator each is a tagged
// union, and the [X, null] form of either collapses to a nullable X.
func unionMembers(s *schema.Schema) []*schema.Schema {
	if len(s.OneOf) > 0 {
		return s.OneOf
	}
	return s.AnyOf
}

// isNullSchema reports whether m is the bare null schema, {"type":"null"}.
func isNullSchema(m *schema.Schema) bool {
	return m.Ref == "" && m.Type.Primary() == "" && m.Type.Nullable() &&
		len(m.AllOf) == 0 && len(m.OneOf) == 0 && len(m.AnyOf) == 0
}

// nullableIdiomOf detects the "X or null" idiom — a oneOf/anyOf of exactly two
// members, one of them the bare null schema — and returns the non-null member (so
// the field is generated as a nullable X). It returns nil when the shape does not
// match.
func nullableIdiomOf(s *schema.Schema) *schema.Schema {
	members := unionMembers(s)
	if len(members) != 2 {
		return nil
	}
	switch {
	case isNullSchema(members[0]) && !isNullSchema(members[1]):
		return members[1]
	case isNullSchema(members[1]) && !isNullSchema(members[0]):
		return members[0]
	default:
		return nil
	}
}

// isUnionSchema reports whether s is a discriminated union the generator can turn
// into a tagged type: a oneOf/anyOf with a discriminator, that is not the [X, null]
// nullable idiom.
func isUnionSchema(s *schema.Schema) bool {
	return len(unionMembers(s)) > 0 && s.Discriminator != nil && nullableIdiomOf(s) == nil
}

// buildUnion produces the union message named goName for a discriminated oneOf/anyOf
// schema s (once). Each variant references an object schema; the discriminator's
// mapping (or, absent a mapping, each variant's schema name) supplies the tag values.
func (b *builder) buildUnion(name string, s *schema.Schema, validated bool) error {
	if b.built[name] {
		return nil
	}
	b.built[name] = true

	prop := s.Discriminator.PropertyName
	if prop == "" {
		return fmt.Errorf("union %q: discriminator has no propertyName", name)
	}

	variants, err := b.unionVariants(name, s)
	if err != nil {
		return err
	}

	b.messages = append(b.messages, &Message{
		Name:      name,
		Doc:       doc(s),
		Validated: validated,
		Union:     &Union{Discriminator: prop, Variants: variants},
	})
	return nil
}

// unionVariants resolves a union's variants to generated object messages, building
// each. With an explicit discriminator mapping the tags come from it; otherwise every
// member must be a $ref and its schema name is the tag.
func (b *builder) unionVariants(name string, s *schema.Schema) ([]Variant, error) {
	refs := map[string]string{} // tag -> reference (a $ref or a bare schema name)
	if len(s.Discriminator.Mapping) > 0 {
		for tag, ref := range s.Discriminator.Mapping {
			refs[tag] = ref
		}
	} else {
		for _, m := range unionMembers(s) {
			if m.Ref == "" {
				return nil, fmt.Errorf("union %q: every variant must be a $ref when the discriminator has no mapping", name)
			}
			refs[refName(m.Ref)] = m.Ref
		}
	}

	tags := make([]string, 0, len(refs))
	for tag := range refs {
		tags = append(tags, tag)
	}
	sort.Strings(tags) // deterministic output

	variants := make([]Variant, 0, len(tags))
	for _, tag := range tags {
		goType, err := b.unionVariantType(name, tag, refs[tag])
		if err != nil {
			return nil, err
		}
		variants = append(variants, Variant{Tag: tag, GoType: goType})
	}
	return variants, nil
}

// unionVariantType resolves one variant reference (a $ref or bare schema name) to a
// generated object message, building it, and returns its Go type name.
func (b *builder) unionVariantType(union, tag, ref string) (string, error) {
	target := ref
	if strings.Contains(ref, "/") {
		target = refName(ref)
	}
	key := goName(target)
	sub, ok := b.defs[key]
	if !ok {
		return "", fmt.Errorf("union %q variant %q: unresolved reference %q", union, tag, ref)
	}
	if !isObjectSchema(sub) {
		return "", fmt.Errorf("union %q variant %q: %q must reference an object schema", union, tag, ref)
	}
	// A variant carries its own x-esdigo-validate flag, like any shared $ref target.
	if err := b.buildMessage(key, sub, validatedOf(sub)); err != nil {
		return "", err
	}
	return key, nil
}

// mergedObject flattens an object schema's effective properties and required set,
// expanding allOf: each subschema's properties/required are merged in (a $ref
// subschema is resolved), then the schema's own properties/required override. So a
// derived type inlines its base's fields — matching esdigo's one-struct-per-type
// model. A plain object (no allOf) yields its own properties unchanged.
func (b *builder) mergedObject(s *schema.Schema) (order []string, props map[string]*schema.Schema, required map[string]bool, err error) {
	props = map[string]*schema.Schema{}
	required = map[string]bool{}
	if err := b.mergeInto(s, &order, props, required, map[*schema.Schema]bool{}); err != nil {
		return nil, nil, nil, err
	}
	return order, props, required, nil
}

// mergeInto accumulates s's allOf subschemas (base) then s's own properties/required
// into props/required, tracking first-seen field order in *order (base fields first,
// then this schema's own; an override keeps the field's original position). seen
// breaks cycles from a self-referential allOf.
func (b *builder) mergeInto(s *schema.Schema, order *[]string, props map[string]*schema.Schema, required map[string]bool, seen map[*schema.Schema]bool) error {
	if seen[s] {
		return nil
	}
	seen[s] = true

	for _, sub := range s.AllOf {
		resolved := sub
		if sub.Ref != "" {
			target, ok := b.defs[goName(refName(sub.Ref))]
			if !ok {
				return fmt.Errorf("unresolved $ref %q in allOf", sub.Ref)
			}
			resolved = target
		}
		if err := b.mergeInto(resolved, order, props, required, seen); err != nil {
			return err
		}
	}
	for _, name := range s.PropertyNames() {
		if _, exists := props[name]; !exists {
			*order = append(*order, name)
		}
		props[name] = s.Properties[name]
	}
	for _, name := range s.Required {
		required[name] = true
	}
	return nil
}

// buildMessage produces the message named goName for object schema s (once).
// validated is whether this message emits validators (inline children share their
// parent's flag; a shared $ref target uses its own).
func (b *builder) buildMessage(name string, s *schema.Schema, validated bool) error {
	if b.built[name] {
		return nil
	}
	b.built[name] = true

	if err := checkComposition(name, s); err != nil {
		return err
	}

	order, props, required, err := b.mergedObject(s)
	if err != nil {
		return err
	}
	if len(props) == 0 {
		return fmt.Errorf("object type %q has no properties; a generated struct would silently drop all data (additionalProperties/maps are not supported)", name)
	}

	msg := &Message{Name: name, Doc: doc(s), Validated: validated, MinProperties: s.MinProperties, MaxProperties: s.MaxProperties}
	for _, jsonName := range order {
		field, err := b.buildField(name, jsonName, props[jsonName], required[jsonName], validated)
		if err != nil {
			return err
		}
		msg.Fields = append(msg.Fields, field)
	}
	b.messages = append(b.messages, msg)
	return nil
}

// buildField resolves one property. Objects and $refs become types.Object fields
// (registering / referencing the child message); everything else is a scalar.
func (b *builder) buildField(msgName, jsonName string, s *schema.Schema, required bool, validated bool) (*Field, error) {
	if s.Ref != "" {
		return b.buildRefField(msgName, jsonName, s, required, validated)
	}
	// "X or null" collapses to a nullable X, decoded exactly like a bare X.
	if inner := nullableIdiomOf(s); inner != nil {
		nn := *inner
		nn.Nullable = true
		return b.buildField(msgName, jsonName, &nn, required, validated)
	}
	// A discriminated oneOf/anyOf becomes a tagged union type, referenced like an
	// inline object.
	if isUnionSchema(s) {
		child := msgName + goName(jsonName)
		if err := b.buildUnion(child, s, validated); err != nil {
			return nil, err
		}
		return objectField(jsonName, child, required, !s.IsNullable()), nil
	}
	if err := checkComposition(jsonName, s); err != nil {
		return nil, err
	}
	if isObjectSchema(s) {
		child := msgName + goName(jsonName)
		if err := b.buildMessage(child, s, validated); err != nil {
			return nil, err
		}
		return objectField(jsonName, child, required, !s.IsNullable()), nil
	}

	switch kind := s.Type.Primary(); kind {
	case "array":
		return b.buildArrayField(msgName, jsonName, s, required, validated)
	case "string", "integer", "number", "boolean":
		return buildScalarField(jsonName, s, required)
	default:
		return nil, fmt.Errorf("unsupported type %q for property %q", kind, jsonName)
	}
}

// buildArrayField resolves a type:array property to a generic
// types.Array[Elem,*Elem] field with an array-level validator (item counts,
// uniqueness). The element type comes from items; an object element registers a
// child message. Per-element validation is not composed yet.
func (b *builder) buildArrayField(msgName, jsonName string, s *schema.Schema, required bool, validated bool) (*Field, error) {
	if s.Items == nil {
		return nil, fmt.Errorf("array property %q has no items schema", jsonName)
	}
	elem, elemIsObject, elemSchema, err := b.arrayElem(msgName, jsonName, s.Items, validated)
	if err != nil {
		return nil, err
	}

	notNull := !s.IsNullable()
	f := &Field{GoName: goName(jsonName), JSONName: jsonName, Doc: doc(s)}

	// Lean path: a scalar element with no per-element constraints uses the unboxed
	// specialized array + ScalarArray validator (no boxing, direct-comparison
	// uniqueness). Everything else uses the generic array.
	if leanArray, leanElem, ok := leanArrayFor(elemSchema); ok && !elemIsObject && !hasValueConstraint(elemSchema) {
		f.ModelType = leanArray
		f.Validate = required || notNull || hasArrayConstraint(s)
		if f.Validate {
			f.ValidatorType = fmt.Sprintf("*validation.ScalarArray[%s]", leanElem)
			f.ResultType = "[]" + leanElem
			f.NewExpr = scalarArrayExpr(leanElem, jsonName, required, notNull, s)
			f.ByPointer = true // ScalarArray.Validate takes the wrapper as an interface
		}
		return f, nil
	}

	f.ModelType = fmt.Sprintf("types.Array[%s, *%s]", elem, elem)
	// An object/union element, or a scalar element with its own constraints, means the
	// array validates per element — so the field must validate even when the array
	// itself is optional/nullable and otherwise unconstrained. Otherwise emit would
	// skip the field (its per-element loop and results) while still importing strconv
	// for it, producing code that drops element validation and fails to compile.
	elemValidates := elemIsObject || hasValueConstraint(elemSchema)
	f.Validate = required || notNull || hasArrayConstraint(s) || elemValidates
	if f.Validate {
		f.ValidatorType = fmt.Sprintf("*validation.Array[%s, *%s]", elem, elem)
		f.ResultType = fmt.Sprintf("[]*%s", elem)
		f.NewExpr = arrayExpr(elem, jsonName, required, notNull, s)
	}

	// Per-element validation: recurse into an object element, or validate a scalar
	// element that carries value constraints. The element validator is (re)built per
	// element at the indexed path `p` the generated loop computes.
	if elemIsObject {
		f.ElemValidate = true
		f.ElemIsObject = true
		f.ElemNewExpr = fmt.Sprintf("New%sValidator(p...)", elem)
		f.ElemItemsType = "*Validated" + elem
		f.Imports = append(f.Imports, "strconv")
	} else if hasValueConstraint(elemSchema) {
		ef, err := buildScalarField(jsonName, elemSchema, false)
		if err != nil {
			return nil, err
		}
		f.ElemValidate = true
		// Reuse the scalar field's constructor, retargeting its path to the loop's
		// indexed path variable `p`.
		f.ElemNewExpr = strings.Replace(ef.NewExpr, subPath(jsonName), "p...", 1)
		f.ElemItemsType = "validation.Result[" + ef.ResultType + "]"
		f.ElemByValue = true
		f.Imports = append(f.Imports, ef.Imports...)
		f.Imports = append(f.Imports, "strconv")
	}
	return f, nil
}

// arrayElem returns the Go element type for an array's items schema (a scalar
// wrapper or a generated child message), whether it is an object, and the
// resolved items schema (used to build the per-element validator).
func (b *builder) arrayElem(msgName, jsonName string, items *schema.Schema, validated bool) (elem string, isObject bool, resolved *schema.Schema, err error) {
	if items.Ref != "" {
		key := goName(refName(items.Ref))
		target, ok := b.defs[key]
		if !ok {
			return "", false, nil, fmt.Errorf("unresolved $ref %q in items of %q", items.Ref, jsonName)
		}
		if isObjectSchema(target) {
			child := key
			targetValidated := validatedOf(target)
			if err := b.checkRefValidator(validated, targetValidated, child, jsonName); err != nil {
				return "", false, nil, err
			}
			if err := b.buildMessage(child, target, targetValidated); err != nil {
				return "", false, nil, err
			}
			return child, true, target, nil
		}
		if isUnionSchema(target) {
			// A union element behaves like an object element: it validates through
			// New<Union>Validator, exactly the object-element path below.
			targetValidated := validatedOf(target)
			if err := b.checkRefValidator(validated, targetValidated, key, jsonName); err != nil {
				return "", false, nil, err
			}
			if err := b.buildUnion(key, target, targetValidated); err != nil {
				return "", false, nil, err
			}
			return key, true, target, nil
		}
		return b.arrayElem(msgName, jsonName, target, validated)
	}

	if isObjectSchema(items) {
		child := msgName + goName(jsonName) + "Item"
		if err := b.buildMessage(child, items, validated); err != nil {
			return "", false, nil, err
		}
		return child, true, items, nil
	}
	if isUnionSchema(items) {
		child := msgName + goName(jsonName) + "Item"
		if err := b.buildUnion(child, items, validated); err != nil {
			return "", false, nil, err
		}
		return child, true, items, nil
	}

	switch kind := items.Type.Primary(); kind {
	case "string":
		return "types.String", false, items, nil
	case "integer", "number":
		if _, ok := bigNumberKind(kind, items.Format); ok {
			return "", false, nil, fmt.Errorf("big-number array elements (format %q) are not yet supported in %q", items.Format, jsonName)
		}
		if rawNumberFormats[items.Format] {
			return "", false, nil, fmt.Errorf("raw number array elements (format %q) are not yet supported in %q", items.Format, jsonName)
		}
		return "types." + numericAlias[numericGoType(kind, items.Format)], false, items, nil
	case "boolean":
		return "types.Boolean", false, items, nil
	default:
		return "", false, nil, fmt.Errorf("unsupported array element type %q in %q", kind, jsonName)
	}
}

func hasArrayConstraint(s *schema.Schema) bool {
	return s.MinItems != nil || s.MaxItems != nil || s.UniqueItems
}

// leanArrayFor maps a scalar element schema to the unboxed specialized array model
// and its Go element type, for the lean scalar-array path. Integer/number honor the
// element's numeric format (e.g. []int32 -> types.Int32Array).
func leanArrayFor(s *schema.Schema) (arrayModel, elem string, ok bool) {
	switch kind := s.Type.Primary(); kind {
	case "string":
		return "types.StringArray", "string", true
	case "boolean":
		return "types.BooleanArray", "bool", true
	case "integer", "number":
		goType := numericGoType(kind, s.Format)
		return "types." + numericAlias[goType] + "Array", goType, true
	}
	return "", "", false
}

func arrayExpr(elem, jsonName string, required, notNull bool, s *schema.Schema) string {
	return arrayExprCtor(fmt.Sprintf("validation.NewArray[%s, *%s]", elem, elem), jsonName, required, notNull, s)
}

func scalarArrayExpr(elem, jsonName string, required, notNull bool, s *schema.Schema) string {
	return arrayExprCtor(fmt.Sprintf("validation.NewScalarArray[%s]", elem), jsonName, required, notNull, s)
}

// arrayExprCtor renders an array validator constructor plus the shared array-level
// constraint chain (min/max items, uniqueness).
func arrayExprCtor(ctor, jsonName string, required, notNull bool, s *schema.Schema) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s(%s)", ctor, subPath(jsonName))
	writePresence(&b, required, notNull)
	if s.MinItems != nil {
		fmt.Fprintf(&b, ".MinItems(%d)", *s.MinItems)
	}
	if s.MaxItems != nil {
		fmt.Fprintf(&b, ".MaxItems(%d)", *s.MaxItems)
	}
	if s.UniqueItems {
		b.WriteString(".UniqueItems()")
	}
	return b.String()
}

// buildRefField resolves a $ref: an object target becomes a shared reference to
// that definition's message; a scalar target is inlined as the field's schema.
func (b *builder) buildRefField(msgName, jsonName string, s *schema.Schema, required bool, validated bool) (*Field, error) {
	key := goName(refName(s.Ref))
	target, ok := b.defs[key]
	if !ok {
		return nil, fmt.Errorf("unresolved $ref %q for property %q", s.Ref, jsonName)
	}
	if isUnionSchema(target) {
		child := key
		targetValidated := validatedOf(target)
		if err := b.checkRefValidator(validated, targetValidated, child, jsonName); err != nil {
			return nil, err
		}
		if err := b.buildUnion(child, target, targetValidated); err != nil {
			return nil, err
		}
		return objectField(jsonName, child, required, !s.IsNullable()), nil
	}
	if isObjectSchema(target) {
		child := key
		// A shared named target carries its OWN x-esdigo-validate flag, not the
		// referrer's — so it stays usable from every parent that references it.
		targetValidated := validatedOf(target)
		if err := b.checkRefValidator(validated, targetValidated, child, jsonName); err != nil {
			return nil, err
		}
		if err := b.buildMessage(child, target, targetValidated); err != nil {
			return nil, err
		}
		// A bare $ref has no type (notNull); a sibling "type" may relax it to
		// nullable, e.g. {"type":["object","null"],"$ref":...}.
		return objectField(jsonName, child, required, !s.IsNullable()), nil
	}
	return b.buildField(msgName, jsonName, target, required, validated)
}

// checkRefValidator rejects a validated parent that nests a non-validated ($ref)
// child. A parent that emits validators recurses into New<Child>Validator, which a
// child marked x-esdigo-validate:false does not generate, so the code would not
// compile — a contradiction: a produce-only type cannot sit inside a value you
// validate on the way in.
func (b *builder) checkRefValidator(parentValidated, childValidated bool, child, jsonName string) error {
	if parentValidated && !childValidated {
		return fmt.Errorf("property %q references %q, which is x-esdigo-validate:false and has no validator, but its parent is validated; set %q to validate, or mark the parent x-esdigo-validate:false too", jsonName, child, child)
	}
	return nil
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

func buildScalarField(jsonName string, s *schema.Schema, required bool) (*Field, error) {
	notNull := !s.IsNullable()
	f := &Field{GoName: goName(jsonName), JSONName: jsonName, Doc: doc(s)}

	var expr string
	var err error
	switch s.Type.Primary() {
	case "string":
		f.ModelType = "types.String"
		f.ValidatorType = "*validation.String"
		f.ResultType = "string"
		expr, err = stringExpr(jsonName, required, notNull, s)
		// A recognized format switches to its specific validator; the base String
		// constraints already applied above still run before the format check.
		if fi, ok := formats[s.Format]; ok {
			f.ValidatorType = fi.validatorType
			f.ResultType = fi.resultType
			expr += fi.method
			f.Imports = fi.imports
		}
	case "integer", "number":
		kind := s.Type.Primary()
		if rawNumberFormats[s.Format] {
			if hasValueConstraint(s) {
				return nil, fmt.Errorf("property %q: a raw number field (format %q) keeps the value verbatim and cannot enforce numeric constraints; remove them or use a concrete numeric format", jsonName, s.Format)
			}
			// RawNumber has only a presence/null validator (no value bounds); when the
			// field needs neither (nullable + optional) it stays a model-only passthrough.
			f.ModelType = "types.RawNumber"
			f.ByPointer = true
			f.ValidatorType = "*validation.RawNumber"
			f.ResultType = "[]byte"
			expr = rawNumberExpr(jsonName, required, notNull)
		} else if bigKind, ok := bigNumberKind(kind, s.Format); ok {
			f.ModelType = "types." + bigKind
			f.ByPointer = true
			f.ValidatorType = "*validation." + bigKind
			f.ResultType = bigResultType(bigKind)
			f.Imports = []string{"math/big"} // the Result[*big.Int]/[*big.Float] payload
			expr, err = bigNumberExpr(bigKind, jsonName, required, notNull, s)
		} else {
			goType := numericGoType(kind, s.Format)
			f.ModelType = "types." + numericAlias[goType]
			f.ByPointer = true
			f.ValidatorType = "*validation.Number[" + goType + "]"
			f.ResultType = goType
			expr, err = numberExpr(goType, jsonName, required, notNull, s)
		}
	case "boolean":
		f.ModelType = "types.Boolean"
		f.ValidatorType = "*validation.Boolean"
		f.ResultType = "bool"
		expr = booleanExpr(jsonName, required, notNull, s)
	}
	if err != nil {
		return nil, fmt.Errorf("property %q: %w", jsonName, err)
	}
	f.NewExpr = expr

	f.Validate = required || notNull || hasValueConstraint(s)
	return f, nil
}

// refName is the target name of a $ref. It takes a JSON-pointer fragment's last
// token ("#/$defs/Address" or "common.json#/$defs/Address" -> "Address") or, for a
// bare file reference ("address.json", "address.yaml", "./dir/address.json"), the
// filename without its .schema/.json/.yaml/.yml extension. The caller normalizes
// the result with goName.
func refName(ref string) string {
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		if frag := ref[i+1:]; strings.ContainsRune(frag, '/') {
			return frag[strings.LastIndexByte(frag, '/')+1:]
		}
		ref = ref[:i]
	}
	if k := strings.LastIndexAny(ref, `/\`); k >= 0 {
		ref = ref[k+1:]
	}
	ref = strings.TrimSuffix(ref, ".json")
	ref = strings.TrimSuffix(ref, ".yaml")
	ref = strings.TrimSuffix(ref, ".yml")
	return strings.TrimSuffix(ref, ".schema")
}

// normalizeDefs re-keys a schema set by Go type name, so $ref resolution and
// generation agree regardless of how names were spelled in the schema (and so
// same-named types across merged files deduplicate). Deterministic on collisions.
func normalizeDefs(defs map[string]*schema.Schema) (map[string]*schema.Schema, error) {
	if len(defs) == 0 {
		return nil, nil
	}
	out := make(map[string]*schema.Schema, len(defs))
	origin := make(map[string]string, len(defs)) // Go name -> the source key that claimed it
	for _, k := range sortedKeys(defs) {
		name := goName(k)
		if prev, taken := origin[name]; taken {
			// Same Go name from two source keys: fine if identical (a shared type),
			// a conflict if they differ (one would silently overwrite the other).
			if !reflect.DeepEqual(out[name], defs[k]) {
				return nil, fmt.Errorf("schemas %q and %q both map to the Go type name %q but differ", prev, k, name)
			}
			continue
		}
		origin[name] = k
		out[name] = defs[k]
	}
	return out, nil
}

func hasValueConstraint(s *schema.Schema) bool {
	return s.MinLength != nil || s.MaxLength != nil || s.Pattern != "" || formatKnown(s.Format) ||
		len(s.Enum) > 0 || s.Const != nil ||
		s.Minimum != nil || s.Maximum != nil || s.ExclusiveMinimum != nil ||
		s.ExclusiveMaximum != nil || s.MultipleOf != nil
}

func stringExpr(jsonName string, required, notNull bool, s *schema.Schema) (string, error) {
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
		list, err := joinStrings(s.Enum)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Enum(%s)", list)
	}
	if s.Const != nil {
		lit, err := goString(s.Const)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Const(%s)", lit)
	}
	return b.String(), nil
}

func numberExpr(goType, jsonName string, required, notNull bool, s *schema.Schema) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewNumber[%s](%s)", goType, subPath(jsonName))
	writePresence(&b, required, notNull)

	bound := func(method string, m json.RawMessage) error {
		if m == nil {
			return nil
		}
		lit, err := goNumber(goType, m)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s(%s)", method, lit)
		return nil
	}
	for _, e := range []struct {
		method string
		val    json.RawMessage
	}{
		{".Min", s.Minimum},
		{".Max", s.Maximum},
		{".ExclusiveMin", s.ExclusiveMinimum},
		{".ExclusiveMax", s.ExclusiveMaximum},
		{".MultipleOf", s.MultipleOf},
	} {
		if err := bound(e.method, e.val); err != nil {
			return "", err
		}
	}
	if len(s.Enum) > 0 {
		list, err := joinNumbers(goType, s.Enum)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Enum(%s)", list)
	}
	if s.Const != nil {
		lit, err := goNumber(goType, s.Const)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Const(%s)", lit)
	}
	return b.String(), nil
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

// rawNumberExpr builds a RawNumber validator chain — presence and null only; a raw
// number keeps its value verbatim, so no numeric constraints apply.
func rawNumberExpr(jsonName string, required, notNull bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.NewRawNumber(%s)", subPath(jsonName))
	writePresence(&b, required, notNull)
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

// raw renders a JSON literal verbatim as a Go literal. It fits values whose JSON
// and Go spellings always coincide: booleans, and float64 numbers (JSON number
// syntax is a subset of Go's float literal syntax). Strings and integer bounds need
// the escape- and range-aware helpers below.
func raw(m []byte) string {
	return string(bytes.TrimSpace(m))
}

// goString renders a JSON string value as an equivalent Go string literal. JSON
// permits escapes Go's lexer rejects (notably \/), so the value is decoded and
// re-quoted rather than emitted verbatim.
func goString(m json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(m, &s); err != nil {
		return "", err
	}
	return strconv.Quote(s), nil
}

// goNumber renders a JSON number as a Go literal of the target numeric type. A
// float literal is emitted verbatim (JSON number syntax is a subset of Go's float
// literal syntax); an integer literal is normalized and range-checked (see goInt).
func goNumber(goType string, m json.RawMessage) (string, error) {
	if isFloatType(goType) {
		return raw(m), nil
	}
	return goInt(goType, m)
}

func isFloatType(goType string) bool {
	return goType == "float32" || goType == "float64"
}

// goInt renders a JSON number as a Go literal of the integer type goType. An integer
// value is emitted exactly (preserving magnitudes beyond float64 precision); a value
// written in float syntax is accepted only when it is integral (e.g. 1e3 -> 1000). A
// fractional value, or one outside goType's range (including a negative on an
// unsigned type), cannot be represented by the field and is an error rather than code
// that will not compile.
func goInt(goType string, m json.RawMessage) (string, error) {
	lit := raw(m)
	r, ok := new(big.Rat).SetString(lit)
	if !ok {
		return "", fmt.Errorf("invalid number %q", lit)
	}
	if !r.IsInt() {
		return "", fmt.Errorf("non-integer bound %q on an integer field", lit)
	}
	n := r.Num()
	lo, hi := intBounds(goType)
	if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
		return "", fmt.Errorf("bound %q out of range for %s", lit, goType)
	}
	return n.String(), nil
}

// intBounds returns the inclusive [min, max] range of the integer type goType. The
// platform-width int/uint are treated as 64-bit.
func intBounds(goType string) (min, max *big.Int) {
	bits := intBits(goType)
	one := big.NewInt(1)
	if strings.HasPrefix(goType, "uint") {
		max = new(big.Int).Sub(new(big.Int).Lsh(one, uint(bits)), one)
		return big.NewInt(0), max
	}
	max = new(big.Int).Sub(new(big.Int).Lsh(one, uint(bits-1)), one)
	min = new(big.Int).Neg(new(big.Int).Lsh(one, uint(bits-1)))
	return min, max
}

func intBits(goType string) int {
	switch goType {
	case "int8", "uint8":
		return 8
	case "int16", "uint16":
		return 16
	case "int32", "uint32":
		return 32
	default: // int, int64, uint, uint64
		return 64
	}
}

// intFormats/floatFormats map an OpenAPI/JSON Schema numeric "format" to a Go type.
// Both OpenAPI names (int32/int64/float/double) and Go names (uint32/float32/...) are
// accepted; an unrecognized format falls back to the default width (JSON Schema
// treats format as an annotation, so it must not fail).
var intFormats = map[string]string{
	"int": "int", "int8": "int8", "int16": "int16", "int32": "int32", "int64": "int64",
	"uint": "uint", "uint8": "uint8", "uint16": "uint16", "uint32": "uint32", "uint64": "uint64",
}

var floatFormats = map[string]string{
	"float": "float32", "double": "float64", "float32": "float32", "float64": "float64",
}

// numericAlias maps a Go numeric type to its esdigo model/validator wrapper alias
// (types.<alias>, types.<alias>Array), e.g. "int32" -> "Int32", "uint" -> "UInt".
var numericAlias = map[string]string{
	"int": "Int", "int8": "Int8", "int16": "Int16", "int32": "Int32", "int64": "Int64",
	"uint": "UInt", "uint8": "UInt8", "uint16": "UInt16", "uint32": "UInt32", "uint64": "UInt64",
	"float32": "Float32", "float64": "Float64",
}

// numericGoType resolves an integer/number schema to its Go value type, honoring a
// recognized format and defaulting to int64/float64.
func numericGoType(kind, format string) string {
	switch kind {
	case "integer":
		if t, ok := intFormats[format]; ok {
			return t
		}
		return "int64"
	case "number":
		if t, ok := floatFormats[format]; ok {
			return t
		}
		return "float64"
	}
	return ""
}

// bigIntFormats/bigFloatFormats select the arbitrary-precision backings: an integer
// with a big-int format decodes into *big.Int (types.BigInt), a number with a
// big-float format into *big.Float (types.BigFloat) — validated exactly at any
// magnitude by the dedicated validators (not the generic Number[V]).
var bigIntFormats = map[string]bool{"bigint": true, "biginteger": true}

var bigFloatFormats = map[string]bool{"bigfloat": true, "bignumber": true, "decimal": true}

// rawNumberFormats select types.RawNumber — the JSON number is preserved as raw bytes.
// It has a presence/null validator only (validation.RawNumber), so a raw field may be
// required or non-null but a numeric constraint on it is a generation error.
var rawNumberFormats = map[string]bool{"raw": true, "rawnumber": true}

// bigNumberKind returns the big-number wrapper ("BigInt"/"BigFloat") a format selects,
// or ok=false for an ordinary fixed-width type.
func bigNumberKind(kind, format string) (string, bool) {
	switch kind {
	case "integer":
		if bigIntFormats[format] {
			return "BigInt", true
		}
	case "number":
		if bigFloatFormats[format] {
			return "BigFloat", true
		}
	}
	return "", false
}

func bigResultType(bigKind string) string {
	if bigKind == "BigInt" {
		return "*big.Int"
	}
	return "*big.Float"
}

// bigNumberExpr builds a BigInt/BigFloat validator chain. Bounds and enum/const are
// wrapped in validation.Big{Int,Float}FromString so the exact literal survives — a
// big value cannot be an untyped Go constant.
func bigNumberExpr(bigKind, jsonName string, required, notNull bool, s *schema.Schema) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "validation.New%s(%s)", bigKind, subPath(jsonName))
	writePresence(&b, required, notNull)

	bound := func(method string, m json.RawMessage) error {
		if m == nil {
			return nil
		}
		lit, err := bigLit(bigKind, m)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s(%s)", method, lit)
		return nil
	}
	for _, e := range []struct {
		method string
		val    json.RawMessage
	}{
		{".Min", s.Minimum},
		{".Max", s.Maximum},
		{".ExclusiveMin", s.ExclusiveMinimum},
		{".ExclusiveMax", s.ExclusiveMaximum},
		{".MultipleOf", s.MultipleOf},
	} {
		if err := bound(e.method, e.val); err != nil {
			return "", err
		}
	}
	if len(s.Enum) > 0 {
		list, err := joinBig(bigKind, s.Enum)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Enum(%s)", list)
	}
	if s.Const != nil {
		lit, err := bigLit(bigKind, s.Const)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, ".Const(%s)", lit)
	}
	return b.String(), nil
}

// bigLit renders a JSON number as a validation.Big{Int,Float}FromString call. A
// BigInt value is normalized to its exact decimal integer form (a fractional value is
// an error); a BigFloat value is validated and passed through verbatim.
func bigLit(bigKind string, m json.RawMessage) (string, error) {
	lit := raw(m)
	if bigKind == "BigInt" {
		r, ok := new(big.Rat).SetString(lit)
		if !ok {
			return "", fmt.Errorf("invalid number %q", lit)
		}
		if !r.IsInt() {
			return "", fmt.Errorf("non-integer bound %q on a bigint field", lit)
		}
		return fmt.Sprintf("validation.BigIntFromString(%q)", r.Num().String()), nil
	}
	if _, ok := new(big.Float).SetString(lit); !ok {
		return "", fmt.Errorf("invalid number %q", lit)
	}
	return fmt.Sprintf("validation.BigFloatFromString(%q)", lit), nil
}

func joinBig(bigKind string, values []json.RawMessage) (string, error) {
	parts := make([]string, len(values))
	for i, v := range values {
		lit, err := bigLit(bigKind, v)
		if err != nil {
			return "", err
		}
		parts[i] = lit
	}
	return strings.Join(parts, ", "), nil
}

// joinStrings renders JSON string values as a comma-separated list of Go string
// literals (a string field's enum).
func joinStrings(values []json.RawMessage) (string, error) {
	parts := make([]string, len(values))
	for i, v := range values {
		lit, err := goString(v)
		if err != nil {
			return "", err
		}
		parts[i] = lit
	}
	return strings.Join(parts, ", "), nil
}

// joinNumbers renders JSON number values as a comma-separated list of Go literals
// of the target numeric type (a number/integer field's enum).
func joinNumbers(goType string, values []json.RawMessage) (string, error) {
	parts := make([]string, len(values))
	for i, v := range values {
		lit, err := goNumber(goType, v)
		if err != nil {
			return "", err
		}
		parts[i] = lit
	}
	return strings.Join(parts, ", "), nil
}

func doc(s *schema.Schema) string {
	if s.Title != "" {
		return strings.TrimSpace(s.Title)
	}
	return strings.TrimSpace(s.Description)
}

// messageImports is the import set for a single message. The combined file's imports
// are the union of these across messages. json (marshal/read) is always used, and
// json/types whenever the struct has fields (its wrappers, incl. types.Object for
// nested refs). The validator and its field result-type imports are used only when
// the message emits validators — a non-validated type needs neither. References to
// sibling types live in the same package, so they need no import.
func messageImports(m *Message) []string {
	set := map[string]bool{"github.com/binadel/esdigo/json": true}
	if len(m.Fields) > 0 {
		set["github.com/binadel/esdigo/json/types"] = true
	}
	if m.Validated {
		set["github.com/binadel/esdigo/validation"] = true
		for _, f := range m.Fields {
			for _, imp := range f.Imports {
				set[imp] = true
			}
		}
	}
	return sortedImports(set)
}

func sortedImports(set map[string]bool) []string {
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
