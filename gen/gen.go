// Package gen generates esdigo model + validator Go source from a JSON Schema.
//
// The pipeline is schema.Parse (input) → ir.Build (resolve to Go types and
// validator chains) → emit.File (render gofmt'd source). OpenAPI 3.1 uses the
// same schema dialect, so its component schemas can feed the same pipeline.
package gen

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/binadel/esdigo/gen/emit"
	"github.com/binadel/esdigo/gen/ir"
	"github.com/binadel/esdigo/gen/schema"
)

// Generate turns a JSON Schema document into a Go source file. pkg is the output
// package name and name is the Go type name for the root object.
func Generate(data []byte, pkg, name string) ([]byte, error) {
	file, err := buildSchema(data, pkg, name)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}

func buildSchema(data []byte, pkg, name string) (*ir.File, error) {
	root, err := schema.Parse(data)
	if err != nil {
		return nil, err
	}
	return ir.Build(pkg, name, root)
}

// GenerateOpenAPI turns an OpenAPI 3.1 document into a Go source file, generating
// a type for every schema under components.schemas (their 2020-12 dialect and
// #/components/schemas/ refs flow through the same pipeline).
func GenerateOpenAPI(data []byte, pkg string) ([]byte, error) {
	file, err := buildOpenAPI(data, pkg)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}

func buildOpenAPI(data []byte, pkg string) (*ir.File, error) {
	doc, err := schema.ParseOpenAPI(data)
	if err != nil {
		return nil, err
	}
	if len(doc.Components.Schemas) == 0 {
		return nil, fmt.Errorf("no components.schemas in the OpenAPI document")
	}
	return ir.BuildAll(pkg, doc.Components.Schemas)
}

// GenerateAuto detects the input: an OpenAPI document generates all its component
// schemas (name is ignored); a bare JSON Schema generates the single named root.
func GenerateAuto(data []byte, pkg, name string) ([]byte, error) {
	file, err := buildAuto(data, pkg, name)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}

// GenerateAutoFiles is like GenerateAuto but emits one source file per generated
// type, keyed by a snake_case filename (e.g. "asset_response.go").
func GenerateAutoFiles(data []byte, pkg, name string) (map[string][]byte, error) {
	file, err := buildAuto(data, pkg, name)
	if err != nil {
		return nil, err
	}
	return emit.Files(file)
}

func buildAuto(data []byte, pkg, name string) (*ir.File, error) {
	if schema.IsOpenAPI(data) {
		return buildOpenAPI(data, pkg)
	}
	return buildSchema(data, pkg, name)
}

// GenerateDir generates one Go source file from a set of schema files (filename →
// content). Every file's schemas are merged into a single namespace: a bare
// schema contributes its root (named after the file) and its $defs; an OpenAPI
// document contributes its components.schemas. Types are deduplicated by name and
// $ref resolves across files (e.g. "common.json#/$defs/Address").
func GenerateDir(files map[string][]byte, pkg string) ([]byte, error) {
	file, err := buildDir(files, pkg)
	if err != nil {
		return nil, err
	}
	return emit.File(file)
}

// GenerateDirFiles is like GenerateDir but emits one source file per generated
// type, keyed by a snake_case filename (e.g. "address.go").
func GenerateDirFiles(files map[string][]byte, pkg string) (map[string][]byte, error) {
	file, err := buildDir(files, pkg)
	if err != nil {
		return nil, err
	}
	return emit.Files(file)
}

func buildDir(files map[string][]byte, pkg string) (*ir.File, error) {
	merged := map[string]*schema.Schema{}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic collision resolution across files

	for _, name := range names {
		data := files[name]
		if schema.IsOpenAPI(data) {
			doc, err := schema.ParseOpenAPI(data)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			for key, s := range doc.Components.Schemas {
				merged[key] = s
			}
			continue
		}
		root, err := schema.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		merged[fileBase(name)] = root // the file root, referenceable as "<file>.json"/".yaml"
		for key, s := range root.AllDefs() {
			merged[key] = s
		}
	}

	file, err := ir.BuildAll(pkg, merged)
	if err != nil {
		return nil, err
	}
	if len(file.Messages) == 0 {
		return nil, fmt.Errorf("no object schemas found in the directory")
	}
	return file, nil
}

// fileBase strips a schema filename to its base name, e.g. "person.schema.json",
// "person.yaml", or "dir/person.json" -> "person".
func fileBase(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, ".json")
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	return strings.TrimSuffix(base, ".schema")
}
