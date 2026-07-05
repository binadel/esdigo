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

func TestTypeNameFromFile(t *testing.T) {
	cases := map[string]string{
		"person.schema.json": "person",
		"user-profile.json":  "user-profile",
		"data.json":          "data",
		"/a/b/order.schema.json": "order",
	}
	for path, want := range cases {
		if got := typeNameFromFile(path); got != want {
			t.Errorf("typeNameFromFile(%q) = %q, want %q", path, got, want)
		}
	}
}
