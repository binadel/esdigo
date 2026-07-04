package gen

import (
	"os"
	"testing"
)

// TestGenerateGolden regenerates gen/example/person.go from its schema and
// asserts it still matches the committed, compiled golden file — so any drift in
// the generator (or a change that would break the example) fails here.
func TestGenerateGolden(t *testing.T) {
	schema, err := os.ReadFile("testdata/person.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	got, err := Generate(schema, "example", "Person")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	want, err := os.ReadFile("example/person.go")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated output drifted from gen/example/person.go.\n"+
			"Regenerate the golden file if the change is intended.\n--- got ---\n%s", got)
	}
}

func TestGenerateRejectsNonObjectRoot(t *testing.T) {
	if _, err := Generate([]byte(`{"type":"string"}`), "example", "X"); err == nil {
		t.Errorf("expected error for non-object root")
	}
}
