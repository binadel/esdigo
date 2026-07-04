// Package json is a zero-reflection JSON reader and writer built for validation
// and code generation.
//
// Reader decodes JSON from a []byte and Writer appends JSON to a buffer; both are
// low-level, allocation-conscious, and reusable across inputs via Reset. The
// reader is safe on untrusted input — it bounds object/array nesting depth and
// rejects oversized numbers — and reports any failure it cannot continue past as
// a *SyntaxError.
//
// For schema-less JSON, ReadValue and ReadJSON build an untyped DOM of Value
// nodes. Typed, per-field decoding lives in the sibling types package, whose
// wrappers read and write themselves through this package's primitives.
package json
