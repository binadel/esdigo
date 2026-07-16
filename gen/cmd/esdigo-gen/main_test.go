package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const objectSchema = `{"type":"object","properties":{"name":{"type":"string"}}}`

// nestedSchema has an inline child object, so it generates two types (Root and
// RootInner) — enough to exercise split output.
const nestedSchema = `{"type":"object","properties":{"inner":{"type":"object","properties":{"x":{"type":"string"}}}}}`

func TestRunStdinToStdout(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"-pkg", "demo"}, strings.NewReader(objectSchema), &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	s := out.String()
	if !strings.Contains(s, "package demo") {
		t.Errorf("missing package: %s", s)
	}
	// stdin has no filename, so the root type defaults to Root
	if !strings.Contains(s, "type Root struct") {
		t.Errorf("missing type Root: %s", s)
	}
}

func TestRunFileToFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "person.schema.json")
	if err := os.WriteFile(in, []byte(objectSchema), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "person.go")

	var errb bytes.Buffer
	code := run([]string{"-pkg", "demo", "-o", outPath, in}, nil, io.Discard, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	// the type name is derived from the filename: person.schema.json -> Person
	if !strings.Contains(string(got), "type Person struct") {
		t.Errorf("expected type Person from filename: %s", got)
	}
}

func TestRunNameOverride(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"-pkg", "demo", "-name", "account"}, strings.NewReader(objectSchema), &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "type Account struct") {
		t.Errorf("-name should drive the type (normalized to Account): %s", out.String())
	}
}

func TestRunBadSchemaExitsNonZero(t *testing.T) {
	var errb bytes.Buffer
	if code := run([]string{"-pkg", "demo"}, strings.NewReader(`{"type":"string"}`), io.Discard, &errb); code == 0 {
		t.Errorf("non-object root should exit non-zero")
	}
	if !strings.Contains(errb.String(), "error:") {
		t.Errorf("expected an error message, got %q", errb.String())
	}
}

func TestRunTooManyArgs(t *testing.T) {
	if code := run([]string{"a.json", "b.json"}, nil, io.Discard, io.Discard); code != 2 {
		t.Errorf("two schema files should exit 2, got %d", code)
	}
}

func TestRunDirectory(t *testing.T) {
	in := t.TempDir()
	out := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(in, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("person.schema.json", objectSchema)
	write("account.json", `{"type":"object","properties":{"id":{"type":"integer"}}}`)
	write("notes.txt", "ignored") // non-json is skipped

	var errb bytes.Buffer
	if code := run([]string{"-pkg", "demo", "-outdir", out, in}, nil, io.Discard, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}

	// directory mode writes one combined <pkg>.go with every schema's types
	combined, err := os.ReadFile(filepath.Join(out, "demo.go"))
	if err != nil {
		t.Fatalf("demo.go: %v", err)
	}
	s := string(combined)
	if !strings.Contains(s, "package demo") || !strings.Contains(s, "type Person struct") || !strings.Contains(s, "type Account struct") {
		t.Errorf("combined output wrong: %s", s)
	}
}

func TestRunDirectoryDefaultsOutdirToInput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "thing.schema.json"), []byte(objectSchema), 0o644); err != nil {
		t.Fatal(err)
	}
	var errb bytes.Buffer
	if code := run([]string{"-pkg", "demo", dir}, nil, io.Discard, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "demo.go")); err != nil {
		t.Errorf("demo.go should be written alongside the schemas: %v", err)
	}
}

func TestRunDirectoryEmpty(t *testing.T) {
	if code := run([]string{"-pkg", "demo", t.TempDir()}, nil, io.Discard, io.Discard); code != 1 {
		t.Errorf("empty dir should exit 1, got %d", code)
	}
}

func TestRunDirectoryBadSchema(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"type":`), 0o644); err != nil {
		t.Fatal(err)
	}
	var errb bytes.Buffer
	if code := run([]string{"-pkg", "demo", dir}, nil, io.Discard, &errb); code != 1 {
		t.Errorf("a malformed schema should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "bad.json") {
		t.Errorf("error should name the offending file: %q", errb.String())
	}
}

func TestRunSplitSingle(t *testing.T) {
	out := t.TempDir()
	var errb bytes.Buffer
	code := run([]string{"-pkg", "demo", "-split", "-outdir", out}, strings.NewReader(nestedSchema), io.Discard, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	// each type lands in its own snake_case file
	root, err := os.ReadFile(filepath.Join(out, "root.go"))
	if err != nil {
		t.Fatalf("root.go: %v", err)
	}
	if !strings.Contains(string(root), "type Root struct") || !strings.Contains(string(root), "package demo") {
		t.Errorf("root.go wrong: %s", root)
	}
	inner, err := os.ReadFile(filepath.Join(out, "root_inner.go"))
	if err != nil {
		t.Fatalf("root_inner.go: %v", err)
	}
	if !strings.Contains(string(inner), "type RootInner struct") {
		t.Errorf("root_inner.go missing type RootInner: %s", inner)
	}
}

func TestRunOutdirSingleCreatesDir(t *testing.T) {
	// a not-yet-existing directory is created
	sub := filepath.Join(t.TempDir(), "gen", "out")
	var errb bytes.Buffer
	code := run([]string{"-pkg", "demo", "-outdir", sub}, strings.NewReader(objectSchema), io.Discard, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	// single input with -outdir writes the combined file as <pkg>.go
	got, err := os.ReadFile(filepath.Join(sub, "demo.go"))
	if err != nil {
		t.Fatalf("demo.go: %v", err)
	}
	if !strings.Contains(string(got), "type Root struct") {
		t.Errorf("wrong content: %s", got)
	}
}

func TestRunSplitDirectory(t *testing.T) {
	in := t.TempDir()
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(in, "person.schema.json"), []byte(objectSchema), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(in, "account.json"), []byte(`{"type":"object","properties":{"id":{"type":"integer"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var errb bytes.Buffer
	if code := run([]string{"-pkg", "demo", "-split", "-outdir", out, in}, nil, io.Discard, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	for _, name := range []string{"person.go", "account.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "demo.go")); err == nil {
		t.Errorf("split mode should not also write the combined demo.go")
	}
}

func TestRunSplitRejectsOutputFile(t *testing.T) {
	if code := run([]string{"-split", "-o", "x.go"}, strings.NewReader(objectSchema), io.Discard, io.Discard); code != 2 {
		t.Errorf("-split with -o should exit 2, got %d", code)
	}
}

func TestRunOutputFileAndOutdirConflict(t *testing.T) {
	if code := run([]string{"-o", "x.go", "-outdir", "d"}, strings.NewReader(objectSchema), io.Discard, io.Discard); code != 2 {
		t.Errorf("-o with -outdir should exit 2, got %d", code)
	}
}

func TestTypeNameFromFile(t *testing.T) {
	cases := map[string]string{
		"person.schema.json":     "person",
		"user-profile.json":      "user-profile",
		"data.json":              "data",
		"/a/b/order.schema.json": "order",
	}
	for path, want := range cases {
		if got := typeNameFromFile(path); got != want {
			t.Errorf("typeNameFromFile(%q) = %q, want %q", path, got, want)
		}
	}
}
