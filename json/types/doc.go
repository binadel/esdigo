// Package types provides the JSON field wrappers used by generated models.
//
// Each wrapper (Number, String, Boolean, Array, Object, Any, and the array
// variants) reads and writes itself through the json package and carries a
// tri-state — Present, Defined, and Valid (see json.OptionalValue) — so a decoder
// can distinguish absent from null from malformed on a per-field basis, without a
// separate error for each. A value that is well-formed JSON but unusable for its
// field is left Valid=false; that is a status, not a parse error, and it does not
// stop the surrounding parse.
//
// Numbers are converted by zero-size codecs chosen through the exported aliases
// (Int64, Float64, BigInt, RawNumber, ...). Integer conversion follows JSON
// Schema semantics, so "1e3" and "1.0" count as integers.
package types
