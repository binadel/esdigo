# esdigo code generation

Generate Go **models** and **validators** from a JSON Schema (2020-12) or OpenAPI
3.1 document, in either **JSON or YAML**. The output has no reflection: each
generated type reads and writes itself through the `json`/`json/types` packages,
and its validator maps a decoded value to a typed, path-aware result.

- **Model** — a struct of `json/types` wrappers with `ReadJSON`/`WriteJSON` and
  `MarshalJSON`/`UnmarshalJSON`.
- **Validator** — `New<T>Validator().Validate(&v)` walks the whole object tree
  (nested objects and array elements included) and returns `*Validated<T>`.
- **Result** — per-field typed values (`r.Email.Value` is a `*mail.Address`), an
  aggregate `IsValid()`, and `Failures()` — a flat list of failing fields, each
  carrying its full path and error codes.

## CLI

The `esdigo-gen` command lives at `gen/cmd/esdigo-gen`:

```sh
go run github.com/binadel/esdigo/gen/cmd/esdigo-gen [flags] <schema.json>
```

| Form | Behavior |
|---|---|
| `esdigo-gen -pkg models schema.json` | one schema → stdout (or `-o out.go`) |
| `esdigo-gen -pkg models < schema.json` | read from stdin |
| `esdigo-gen -pkg models -outdir ./m schema.json` | combined `<pkg>.go` in a directory |
| `esdigo-gen -pkg models -split -outdir ./m schema.json` | one file per type (`asset_response.go`, …) |
| `esdigo-gen -pkg models ./schemas` | a directory → one combined `<pkg>.go` |

Flags: `-pkg` (output package, default `models`), `-name` (root type name; default
derived from the filename), `-o` (combined single-file output, default stdout),
`-outdir` (output directory; writes `<pkg>.go` there, created if missing),
`-split` (write one `<type>.go` per generated type into `-outdir`). `-split` and
`-outdir` work for a single schema, a directory, or stdin.

Input may be **JSON or YAML** — the format is detected automatically, so OpenAPI
specs (usually YAML) work directly (`esdigo-gen -pkg models openapi.yaml`); in a
directory, both `*.json` and `*.yaml`/`*.yml` are read. An **OpenAPI** document is
also detected automatically (its `components.schemas` each become a type). A
**directory** is merged into one namespace: types are deduplicated by name and
`$ref` resolves across files (e.g. `common.json#/$defs/Address`).

## Library

```go
import "github.com/binadel/esdigo/gen"

src, err := gen.Generate(schemaBytes, "models", "User")   // one JSON Schema
src, err := gen.GenerateOpenAPI(docBytes, "models")       // all components.schemas
src, err := gen.GenerateDir(files, "models")              // map[filename]bytes → one file
src, err := gen.GenerateAuto(data, "models", "User")      // detect schema vs OpenAPI

// Split variants return map[filename]bytes — one source file per generated type:
byFile, err := gen.GenerateAutoFiles(data, "models", "User")
byFile, err := gen.GenerateDirFiles(files, "models")
```

## Example

```json
{
  "type": "object",
  "title": "is an API user.",
  "required": ["id", "email"],
  "$defs": {
    "Address": {
      "type": "object",
      "required": ["city"],
      "properties": {
        "city": { "type": "string", "minLength": 1 },
        "zip":  { "type": "string", "pattern": "^[0-9]{5}$" }
      }
    }
  },
  "properties": {
    "id":      { "type": "integer", "minimum": 1 },
    "email":   { "type": "string", "format": "email" },
    "address": { "$ref": "#/$defs/Address" },
    "tags":    { "type": "array", "items": { "type": "string" }, "uniqueItems": true }
  }
}
```

Using the generated code:

```go
var u User
if err := u.UnmarshalJSON(data); err != nil {
    // *json.SyntaxError: the bytes were not well-formed JSON
}

r := NewUserValidator().Validate(&u)
if !r.IsValid() {
    for _, f := range r.Failures() {
        // f serializes to {"path":[...],"errors":[{"code":..,"message":..}]}
        // e.g. {"path":["address","zip"],"errors":[{"code":"PATTERN",...}]}
    }
}

r.Email.Value        // *mail.Address — the parsed, typed value
r.Address.City.Value // "…" — reached through the nested result
```

## Type mapping

| Schema | Model field | Validator | `Result` value |
|---|---|---|---|
| `string` | `types.String` | `*validation.String` | `string` |
| `integer` | `types.Int64` | `*validation.Number[int64]` | `int64` |
| `number` | `types.Float64` | `*validation.Number[float64]` | `float64` |
| `boolean` | `types.Boolean` | `*validation.Boolean` | `bool` |
| `object` / `$ref`→object | `types.Object[T,*T]` | (recurses) | `*ValidatedT` |
| `array` of scalars, no item constraints | `types.StringArray`, `types.Int64Array`, … | `*validation.ScalarArray[V]` | `[]V` |
| `array` (constrained items or objects) | `types.Array[E,*E]` | `*validation.Array[E,*E]` | `[]*E` + per-element |

A `string` with a **`format`** keeps `types.String` but switches its validator and
result:

| format | validator | result |
|---|---|---|
| `email` | `*validation.Email` | `*mail.Address` |
| `ipv4` / `ipv6` | `*validation.IP` | `net.IP` |
| `uri` / `uri-reference` | `*validation.Uri` | `*url.URL` |
| `uuid` | `*validation.Uuid` | `uuid.UUID` |
| `date` / `time` / `date-time` | `*validation.Time` | `time.Time` |
| `duration` | `*validation.Duration` | `time.Duration` |
| `regex` | `*validation.Regex` | `*regexp.Regexp` |
| `hostname` / `json-pointer` | `*validation.Hostname` / `JsonPointer` | `string` |

Unknown formats are ignored (JSON Schema treats `format` as an annotation).

## Constraints

Emitted onto the field validators:

- **string**: `minLength`, `maxLength`, `pattern`, `enum`, `const`
- **number/integer**: `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`,
  `multipleOf`, `enum`, `const`
- **boolean**: `const`
- **array**: `minItems`, `maxItems`, `uniqueItems`
- **object**: `required` (per property)

## Presence and nullability

esdigo models a field as three independent states — **present**, **defined**
(non-null), and **valid**:

- `required` (a property name in the object's `required` list) → `.Required()`; a
  missing field fails.
- Non-nullable (`type: "string"`, i.e. no `"null"`) → `.NotNull()`; a **present**
  `null` fails, but an **absent** field does not (nullability is orthogonal to
  presence).
- Nullable — either `type: ["string", "null"]` (JSON Schema / OpenAPI 3.1) or
  `"nullable": true` (OpenAPI 3.0) — omits `.NotNull()`.

## Notes and limitations

- Array **element** failures carry their index in the path, e.g.
  `["tags", "0"]`; the flat report is built from failing `validation.FieldResult`
  values (the path lives on the result, not the error).
- Directory mode deduplicates types by **name** — two different types with the
  same name silently collapse (last wins).
- Not yet handled: `oneOf`/`anyOf`/`allOf`/`if`-`then`-`else`, `enum`/`const` for
  big-number fields, `minProperties`/`maxProperties`, `dependentRequired`, nested
  arrays (`array` of `array`), and OpenAPI `paths` request/response bodies (only
  `components.schemas` is extracted).
