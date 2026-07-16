# esdigo

A zero-reflection JSON toolkit for Go: a fast reader/writer, typed field
wrappers, fused serialization-and-validation, and a code generator that turns a
JSON Schema or OpenAPI document (JSON or YAML) into models and validators.

## Packages

- **`json`** — a low-level, allocation-conscious JSON `Reader` and `Writer`,
  safe on untrusted input (bounded nesting depth, oversized-number guard).
- **`json/types`** — typed, self-serializing field wrappers (`String`, `Int64`,
  `Object[T,*T]`, `StringArray`, …) carrying a Present/Defined/Valid tri-state.
- **`validation`** — validators that also map: each returns the parsed, typed
  value plus structured, path-aware errors (no reflection; validation is part of
  (de)serialization).
- **`gen`** — generate the above from a schema. See **[gen/README.md](gen/README.md)**.

## Code generation

```sh
go run github.com/binadel/esdigo/gen/cmd/esdigo-gen -pkg models schema.json
```

turns a JSON Schema (2020-12) or OpenAPI 3.1 document — in JSON or YAML — into a
Go model (with `ReadJSON`/`WriteJSON`) and a validator that walks the whole object
tree and reports failures with their full paths. Full guide:
**[gen/README.md](gen/README.md)**.
